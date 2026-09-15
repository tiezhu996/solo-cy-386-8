package service

import (
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/marketpal/marketpal/internal/constants"
	"github.com/marketpal/marketpal/internal/dto"
	"github.com/marketpal/marketpal/internal/model"
	"github.com/marketpal/marketpal/internal/repository"
	"github.com/marketpal/marketpal/internal/util"
	"gorm.io/gorm"
)

// 商品审核闭环端到端回归测试。
//
// 约束（按需求执行）：
//   - 每类数据独立重建：每个用例在 t.TempDir() 中新建独立 SQLite 文件库，不复用、不共享内存。
//   - 不使用内存替身：所有断言都走真实 repository + 真实建表（AutoMigrate）+ 真实事务/FOR UPDATE。
//   - 不使用单连接串行化：连接池允许多连接，打开 WAL 与 busy_timeout，并发用例以多协程真实竞争。
//   - 开启外键约束，复现生产 PostgreSQL 的真实约束（待审记录 reviewer_id 必须为 NULL）。
//   - 持久化/并发结果一律以操作完成后的数据库回读为准。
//   - 每个失败断言都带“阶段(stage)”前缀，便于定位是发布/可见性/审核/重提/并发中的哪一环。

// regEnv 一组完整真实装配。
type regEnv struct {
	db      *gorm.DB
	product *ProductService
	review  *ProductReviewService
	cart    *CartService
	order   *OrderService
}

// regModels AutoMigrate 的全量模型（与生产 database.Connect 保持一致）。
var regModels = []interface{}{
	&model.User{}, &model.Product{}, &model.Favorite{}, &model.Address{},
	&model.CartItem{}, &model.Order{}, &model.Message{}, &model.Review{},
	&model.AuditLog{}, &model.ProductReview{},
}

// newRegDB 新建独立文件库：WAL + 外键 + busy_timeout，连接池放开并发。
func newRegDB(t *testing.T) *gorm.DB {
	t.Helper()
	// glebarez(modernc) 通过 _pragma 让连接池中的每个新连接都继承外键/WAL/busy_timeout，
	// 不能只在单连接上执行 PRAGMA（连接池复用时会丢失）。
	// _txlock=immediate 使写事务在 BEGIN 即申请 RESERVED 锁：多协程竞争时由 SQLite 在多连接上
	// 真实排队（不是单连接串行化），配合 busy_timeout 与条件更新，确定性地模拟 PostgreSQL 行锁裁决。
	dsn := fmt.Sprintf(
		"file:%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(20000)&_txlock=immediate",
		filepath.ToSlash(filepath.Join(t.TempDir(), "reg.db")),
	)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("阶段=建库 open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("阶段=建库 get sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(8)
	sqlDB.SetMaxIdleConns(8)
	if err := db.AutoMigrate(regModels...); err != nil {
		t.Fatalf("阶段=建库 auto migrate: %v", err)
	}
	return db
}

// newRegEnv 装配真实仓储/服务（与 cmd/server/main.go 装配方式一致）。
func newRegEnv(t *testing.T) *regEnv {
	t.Helper()
	db := newRegDB(t)
	logger := slog.New(slog.NewTextHandler(discardWriter{}, nil))
	productRepo := repository.NewProductRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)
	addressRepo := repository.NewAddressRepository(db)
	cartRepo := repository.NewCartRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	reviewRepo := repository.NewProductReviewRepository(db)
	return &regEnv{
		db:      db,
		product: NewProductService(db, productRepo, favoriteRepo, reviewRepo, logger),
		review:  NewProductReviewService(db, reviewRepo, productRepo, logger),
		cart:    NewCartService(cartRepo, productRepo, logger),
		order:   NewOrderService(db, orderRepo, productRepo, addressRepo, cartRepo, logger),
	}
}

// regCreateUser 建真实用户。
func regCreateUser(t *testing.T, db *gorm.DB, username, role string) *model.User {
	t.Helper()
	u := &model.User{Username: username, PasswordHash: "hash", Nickname: username, Role: role, CreditScore: 100, Status: "active"}
	if err := db.Create(u).Error; err != nil {
		t.Fatalf("阶段=建用户 %s: %v", username, err)
	}
	return u
}

// regPublish 通过 service 发布商品（真实走事务双写）。
func regPublish(t *testing.T, env *regEnv, sellerID uint, suffix string) *model.Product {
	t.Helper()
	p, err := env.product.Create(sellerID, dto.ProductCreateRequest{
		Title:         "回归商品-" + suffix,
		Description:   "回归测试专用商品描述不少于五个字",
		OriginalPrice: 1000,
		Price:         500,
		Condition:     constants.ProductConditionAlmostNew,
		Category:      constants.ProductCategoryDigital,
	})
	if err != nil {
		t.Fatalf("阶段=发布 seller=%d: %v", sellerID, err)
	}
	return p
}

// regReadProduct 从数据库重新回读商品。
func regReadProduct(t *testing.T, db *gorm.DB, id uint) model.Product {
	t.Helper()
	var p model.Product
	if err := db.First(&p, id).Error; err != nil {
		t.Fatalf("阶段=回读商品 id=%d: %v", id, err)
	}
	return p
}

// regReadReviews 回读某商品全部审核记录（按轮次升序）。
func regReadReviews(t *testing.T, db *gorm.DB, productID uint) []model.ProductReview {
	t.Helper()
	var list []model.ProductReview
	if err := db.Where("product_id = ?", productID).Order("round ASC").Find(&list).Error; err != nil {
		t.Fatalf("阶段=回读审核记录 product=%d: %v", productID, err)
	}
	return list
}

func regAppErrCode(t *testing.T, stage string, err error) int {
	t.Helper()
	if err == nil {
		t.Fatalf("阶段=%s：期望返回错误，实际成功", stage)
	}
	var ae *util.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("阶段=%s：期望 AppError，实际 %T %v", stage, err, err)
	}
	return ae.Code
}

// regFailReviewRepo 在真实仓储外包一层：仅让“创建审核记录”这一步必失败，用于断言事务整体回滚。
// 它不是内存替身——其余所有方法都委托给真实 repository（连接真实文件库与真实约束）。
type regFailReviewRepo struct {
	inner repository.ProductReviewRepository
}

func (f *regFailReviewRepo) CreateTx(tx *gorm.DB, review *model.ProductReview) error {
	return errRegReviewInsert
}
func (f *regFailReviewRepo) GetByID(id uint) (*model.ProductReview, error) {
	return f.inner.GetByID(id)
}
func (f *regFailReviewRepo) GetByIDForUpdate(tx *gorm.DB, id uint) (*model.ProductReview, error) {
	return f.inner.GetByIDForUpdate(tx, id)
}
func (f *regFailReviewRepo) GetLatestByProduct(productID uint) (*model.ProductReview, error) {
	return f.inner.GetLatestByProduct(productID)
}
func (f *regFailReviewRepo) GetPendingByProductForUpdate(tx *gorm.DB, productID uint) (*model.ProductReview, error) {
	return f.inner.GetPendingByProductForUpdate(tx, productID)
}
func (f *regFailReviewRepo) MaxRound(tx *gorm.DB, productID uint) (int, error) {
	return f.inner.MaxRound(tx, productID)
}
func (f *regFailReviewRepo) DecideForUpdate(tx *gorm.DB, id uint, reviewerID *uint, status, reason string, reviewedAt interface{}) error {
	return f.inner.DecideForUpdate(tx, id, reviewerID, status, reason, reviewedAt)
}
func (f *regFailReviewRepo) List(query map[string]interface{}, page, pageSize int) ([]model.ProductReview, int64, error) {
	return f.inner.List(query, page, pageSize)
}
func (f *regFailReviewRepo) ListByProduct(productID uint) ([]model.ProductReview, error) {
	return f.inner.ListByProduct(productID)
}

var errRegReviewInsert = errors.New("injected review insert failure")

// TestReg_RollbackWhenReviewInsertFails 故障注入：审核记录写库失败时，同事务内刚插入的商品必须整体回滚，
// 不留下半条数据（无孤儿商品、无孤儿审核记录）。
func TestReg_RollbackWhenReviewInsertFails(t *testing.T) {
	db := newRegDB(t)
	logger := slog.New(slog.NewTextHandler(discardWriter{}, nil))
	seller := regCreateUser(t, db, "reg_seller_rollback", constants.UserRoleUser)
	productRepo := repository.NewProductRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)
	failRepo := &regFailReviewRepo{inner: repository.NewProductReviewRepository(db)}
	svc := NewProductService(db, productRepo, favoriteRepo, failRepo, logger)

	if _, err := svc.Create(seller.ID, dto.ProductCreateRequest{
		Title: "回滚验证", Description: "审核记录写入失败时商品必须一起回滚的描述",
		OriginalPrice: 100, Price: 50, Condition: constants.ProductConditionBrandNew, Category: constants.ProductCategoryOther,
	}); err == nil {
		t.Fatalf("阶段=注入失败发布：应返回错误")
	}

	var pc, rc int64
	db.Model(&model.Product{}).Where("seller_id = ?", seller.ID).Count(&pc)
	db.Model(&model.ProductReview{}).Where("seller_id = ?", seller.ID).Count(&rc)
	if pc != 0 || rc != 0 {
		t.Fatalf("阶段=回滚后回读：不应留下任何商品/审核记录，实际商品=%d 记录=%d", pc, rc)
	}
}

// TestReg_CreatePersistsProductAndPendingReview 新发布：商品与首轮待审记录在同一事务一起保存，
// 待审记录 reviewer_id 为 NULL（不伪造审核人），商品立即下架。
func TestReg_CreatePersistsProductAndPendingReview(t *testing.T) {
	env := newRegEnv(t)
	seller := regCreateUser(t, env.db, "reg_seller_publish", constants.UserRoleUser)
	p := regPublish(t, env, seller.ID, "publish")

	got := regReadProduct(t, env.db, p.ID)
	if got.ReviewStatus != constants.ProductReviewPending {
		t.Fatalf("阶段=回读商品：review_status 期望 %s，实际 %s", constants.ProductReviewPending, got.ReviewStatus)
	}
	if got.Status != constants.ProductStatusOffShelf {
		t.Fatalf("阶段=回读商品：审核前 status 期望 %s，实际 %s", constants.ProductStatusOffShelf, got.Status)
	}
	if got.ReviewRound != 1 {
		t.Fatalf("阶段=回读商品：review_round 期望 1，实际 %d", got.ReviewRound)
	}
	reviews := regReadReviews(t, env.db, p.ID)
	if len(reviews) != 1 {
		t.Fatalf("阶段=回读审核记录：期望恰好 1 条首轮记录，实际 %d 条", len(reviews))
	}
	r := reviews[0]
	if r.Status != constants.ProductReviewPending || r.Round != 1 {
		t.Fatalf("阶段=回读审核记录：期望 pending/round1，实际 %s/round%d", r.Status, r.Round)
	}
	if r.ReviewerID != nil {
		t.Fatalf("阶段=回读审核记录：待审记录 reviewer_id 必须为 NULL，实际 %d", *r.ReviewerID)
	}
	if r.ReviewedAt != nil {
		t.Fatalf("阶段=回读审核记录：待审记录 reviewed_at 必须为空")
	}
}

// TestReg_PendingInvisibleInHallSearchDetail 待审商品对大厅/搜索不可见；
// 详情对其他用户与匿名用户不可见，卖家本人可见。
func TestReg_PendingInvisibleInHallSearchDetail(t *testing.T) {
	env := newRegEnv(t)
	seller := regCreateUser(t, env.db, "reg_seller_vis", constants.UserRoleUser)
	other := regCreateUser(t, env.db, "reg_other_vis", constants.UserRoleUser)
	p := regPublish(t, env, seller.ID, "visibility")

	// 大厅（无条件）
	list, err := env.product.List(dto.ProductQuery{Page: 1, PageSize: 50}, other.ID)
	if err != nil {
		t.Fatalf("阶段=大厅列表：%v", err)
	}
	for _, v := range list.List {
		if v.ID == p.ID {
			t.Fatalf("阶段=大厅列表：待审商品 id=%d 不应出现在大厅", p.ID)
		}
	}
	// 搜索（分类 + 排序走同一 List 查询；关键词 ILIKE 为 PostgreSQL 方言，仅生产 PG 覆盖）
	list, err = env.product.List(dto.ProductQuery{Category: constants.ProductCategoryDigital, SortBy: "time_desc", Page: 1, PageSize: 50}, other.ID)
	if err != nil {
		t.Fatalf("阶段=搜索：%v", err)
	}
	if list.Total != 0 {
		t.Fatalf("阶段=搜索：待审商品不应被搜索到，total=%d", list.Total)
	}

	// 详情：其他买家不可见
	if _, err := env.product.GetDetail(p.ID, other.ID, constants.UserRoleUser); err == nil {
		t.Fatalf("阶段=详情-其他用户：待审商品应对其他用户不可见")
	}
	// 详情：匿名（viewerID=0）不可见
	if _, err := env.product.GetDetail(p.ID, 0, ""); err == nil {
		t.Fatalf("阶段=详情-匿名：待审商品应对匿名用户不可见")
	}
	// 详情：卖家本人可见
	if _, err := env.product.GetDetail(p.ID, seller.ID, constants.UserRoleUser); err != nil {
		t.Fatalf("阶段=详情-卖家：卖家应能查看自己的待审商品，err=%v", err)
	}
	// 详情：管理员可见
	admin := regCreateUser(t, env.db, "reg_admin_vis", constants.UserRoleAdmin)
	if _, err := env.product.GetDetail(p.ID, admin.ID, constants.UserRoleAdmin); err != nil {
		t.Fatalf("阶段=详情-管理员：管理员应能查看待审商品，err=%v", err)
	}
}

// TestReg_PendingBlockedFromCartAndOrder 待审商品不能加购、不能下单。
func TestReg_PendingBlockedFromCartAndOrder(t *testing.T) {
	env := newRegEnv(t)
	seller := regCreateUser(t, env.db, "reg_seller_block", constants.UserRoleUser)
	buyer := regCreateUser(t, env.db, "reg_buyer_block", constants.UserRoleUser)
	p := regPublish(t, env, seller.ID, "block")

	if _, err := env.cart.Add(buyer.ID, dto.CartAddRequest{ProductID: p.ID, Quantity: 1}); err == nil {
		t.Fatalf("阶段=加购：待审商品不应允许加入购物车")
	}
	var cartCount int64
	env.db.Model(&model.CartItem{}).Where("user_id = ?", buyer.ID).Count(&cartCount)
	if cartCount != 0 {
		t.Fatalf("阶段=加购回读：失败不应留下购物车记录，实际 %d 条", cartCount)
	}

	addr := &model.Address{UserID: buyer.ID, ReceiverName: "买家", Phone: "13800000000", Province: "省", City: "市", Detail: "详细地址", IsDefault: true}
	if err := env.db.Create(addr).Error; err != nil {
		t.Fatalf("阶段=建地址：%v", err)
	}
	if _, err := env.order.Create(buyer.ID, dto.OrderCreateRequest{ProductID: p.ID, AddressID: addr.ID, Quantity: 1}); err == nil {
		t.Fatalf("阶段=下单：待审商品不应允许下单")
	}
	var orderCount int64
	env.db.Model(&model.Order{}).Where("buyer_id = ?", buyer.ID).Count(&orderCount)
	if orderCount != 0 {
		t.Fatalf("阶段=下单回读：失败不应留下订单，实际 %d 条", orderCount)
	}
}

// TestReg_AdminApproveMakesProductSalable 管理员通过后：记录写入真实审核人，商品上架，
// 大厅/详情/加购/下单全链路放开；下单后商品售出。
func TestReg_AdminApproveMakesProductSalable(t *testing.T) {
	env := newRegEnv(t)
	seller := regCreateUser(t, env.db, "reg_seller_app", constants.UserRoleUser)
	buyer := regCreateUser(t, env.db, "reg_buyer_app", constants.UserRoleUser)
	admin := regCreateUser(t, env.db, "reg_admin_app", constants.UserRoleAdmin)
	p := regPublish(t, env, seller.ID, "approve")
	rid := regReadReviews(t, env.db, p.ID)[0].ID

	if _, err := env.review.Decide(admin.ID, rid, dto.ProductReviewDecisionRequest{Action: "approve"}); err != nil {
		t.Fatalf("阶段=管理员通过：%v", err)
	}
	rv := regReadReviews(t, env.db, p.ID)[0]
	if rv.Status != constants.ProductReviewApproved || rv.ReviewerID == nil || *rv.ReviewerID != admin.ID || rv.ReviewedAt == nil {
		t.Fatalf("阶段=回读审核记录：approved/reviewer=%d/reviewed_at 非空 期望，实际 %s/reviewer=%v", admin.ID, rv.Status, rv.ReviewerID)
	}
	got := regReadProduct(t, env.db, p.ID)
	if got.ReviewStatus != constants.ProductReviewApproved || got.Status != constants.ProductStatusOnSale {
		t.Fatalf("阶段=回读商品：通过后期望 approved/on_sale，实际 %s/%s", got.ReviewStatus, got.Status)
	}
	// 大厅可见（关键词 ILIKE 为 PostgreSQL 方言，SQLite 回归用分类过滤走同一 List/review_status 条件）
	list, err := env.product.List(dto.ProductQuery{Category: constants.ProductCategoryDigital, Page: 1, PageSize: 50}, buyer.ID)
	if err != nil {
		t.Fatalf("阶段=大厅列表：%v", err)
	}
	if list.Total == 0 {
		t.Fatalf("阶段=大厅列表：通过后应可见，total=%d", list.Total)
	}
	// 加购
	if _, err := env.cart.Add(buyer.ID, dto.CartAddRequest{ProductID: p.ID, Quantity: 1}); err != nil {
		t.Fatalf("阶段=加购：通过后应可加购，err=%v", err)
	}
	// 下单
	addr := &model.Address{UserID: buyer.ID, ReceiverName: "买家", Phone: "13800000000", Province: "省", City: "市", Detail: "地址", IsDefault: true}
	if err := env.db.Create(addr).Error; err != nil {
		t.Fatalf("阶段=建地址：%v", err)
	}
	order, err := env.order.Create(buyer.ID, dto.OrderCreateRequest{ProductID: p.ID, AddressID: addr.ID, Quantity: 1})
	if err != nil {
		t.Fatalf("阶段=下单：通过后应可下单，err=%v", err)
	}
	sold := regReadProduct(t, env.db, p.ID)
	if sold.Status != constants.ProductStatusSold {
		t.Fatalf("阶段=回读商品：下单后期望 sold，实际 %s", sold.Status)
	}
	var oc int64
	env.db.Model(&model.Order{}).Where("id = ? AND status = ?", order.ID, constants.OrderStatusPendingPayment).Count(&oc)
	if oc != 1 {
		t.Fatalf("阶段=回读订单：期望 1 条待付款订单")
	}
}

// TestReg_AdminRejectRequiresReasonAndStaysOffShelf 管理员驳回：必须填写原因；
// 驳回后商品保持下架、大厅不可见、不能加购下单；驳回原因回写到商品与记录。
func TestReg_AdminRejectRequiresReasonAndStaysOffShelf(t *testing.T) {
	env := newRegEnv(t)
	seller := regCreateUser(t, env.db, "reg_seller_rej", constants.UserRoleUser)
	buyer := regCreateUser(t, env.db, "reg_buyer_rej", constants.UserRoleUser)
	admin := regCreateUser(t, env.db, "reg_admin_rej", constants.UserRoleAdmin)
	p := regPublish(t, env, seller.ID, "reject")
	rid := regReadReviews(t, env.db, p.ID)[0].ID

	_, err := env.review.Decide(admin.ID, rid, dto.ProductReviewDecisionRequest{Action: "reject"})
	if code := regAppErrCode(t, "驳回-不填原因", err); code != constants.CodeReviewReasonEmpty {
		t.Fatalf("阶段=驳回-不填原因：期望错误码 %d，实际 %d", constants.CodeReviewReasonEmpty, code)
	}
	reason := "主图模糊且价格与描述不符"
	if _, err := env.review.Decide(admin.ID, rid, dto.ProductReviewDecisionRequest{Action: "reject", Reason: reason}); err != nil {
		t.Fatalf("阶段=驳回：%v", err)
	}
	got := regReadProduct(t, env.db, p.ID)
	if got.ReviewStatus != constants.ProductReviewRejected || got.Status != constants.ProductStatusOffShelf || got.RejectReason != reason {
		t.Fatalf("阶段=回读商品：驳回后期望 rejected/off_shelf/原因已写入，实际 %s/%s/%q", got.ReviewStatus, got.Status, got.RejectReason)
	}
	list, err := env.product.List(dto.ProductQuery{Category: constants.ProductCategoryDigital, Page: 1, PageSize: 50}, buyer.ID)
	if err != nil {
		t.Fatalf("阶段=大厅列表：%v", err)
	}
	if list.Total != 0 {
		t.Fatalf("阶段=大厅列表：驳回商品不应可见，total=%d", list.Total)
	}
	if _, err := env.cart.Add(buyer.ID, dto.CartAddRequest{ProductID: p.ID, Quantity: 1}); err == nil {
		t.Fatalf("阶段=加购：驳回商品不应允许加购")
	}
}

// TestReg_ResubmitAfterRejectThenApprove 驳回后卖家修改重提：新一轮待审、继续下架、生成新记录；
// 复审通过后重新上架；待审中重复修改不产生新状态。
func TestReg_ResubmitAfterRejectThenApprove(t *testing.T) {
	env := newRegEnv(t)
	seller := regCreateUser(t, env.db, "reg_seller_resub", constants.UserRoleUser)
	admin := regCreateUser(t, env.db, "reg_admin_resub", constants.UserRoleAdmin)
	p := regPublish(t, env, seller.ID, "resubmit")
	rid1 := regReadReviews(t, env.db, p.ID)[0].ID
	if _, err := env.review.Decide(admin.ID, rid1, dto.ProductReviewDecisionRequest{Action: "reject", Reason: "描述过短"}); err != nil {
		t.Fatalf("阶段=一审驳回：%v", err)
	}

	// 驳回后修改重提
	newDesc := "这是补充后的商品描述，成色好，配件齐全。"
	updated, err := env.product.Update(seller.ID, p.ID, dto.ProductUpdateRequest{Description: &newDesc})
	if err != nil {
		t.Fatalf("阶段=修改重提：%v", err)
	}
	if updated.ReviewStatus != constants.ProductReviewPending || updated.Status != constants.ProductStatusOffShelf || updated.ReviewRound != 2 {
		t.Fatalf("阶段=回读商品：重提后期望 pending/off_shelf/round2，实际 %s/%s/round%d", updated.ReviewStatus, updated.Status, updated.ReviewRound)
	}
	reviews := regReadReviews(t, env.db, p.ID)
	if len(reviews) != 2 {
		t.Fatalf("阶段=回读审核记录：期望 2 轮记录，实际 %d 条", len(reviews))
	}
	r2 := reviews[1]
	if r2.Round != 2 || r2.Status != constants.ProductReviewPending || r2.ReviewerID != nil {
		t.Fatalf("阶段=回读审核记录：第二轮应为 pending 且无审核人，实际 round=%d %s reviewer=%v", r2.Round, r2.Status, r2.ReviewerID)
	}

	// 待审中重复修改：拒绝，记录数不增加
	dupDesc := "待审中再次修改的描述内容"
	if _, err := env.product.Update(seller.ID, p.ID, dto.ProductUpdateRequest{Description: &dupDesc}); err == nil {
		t.Fatalf("阶段=待审中重复修改：应被拒绝")
	}
	if n := len(regReadReviews(t, env.db, p.ID)); n != 2 {
		t.Fatalf("阶段=待审中重复修改-回读：记录数应仍为 2，实际 %d", n)
	}

	// 复审通过
	if _, err := env.review.Decide(admin.ID, r2.ID, dto.ProductReviewDecisionRequest{Action: "approve"}); err != nil {
		t.Fatalf("阶段=复审通过：%v", err)
	}
	got := regReadProduct(t, env.db, p.ID)
	if got.ReviewStatus != constants.ProductReviewApproved || got.Status != constants.ProductStatusOnSale || got.Description != newDesc || got.RejectReason != "" {
		t.Fatalf("阶段=复审通过-回读：expected approved/on_sale/新描述/原因清空，实际 %s/%s/%q/%q", got.ReviewStatus, got.Status, got.Description, got.RejectReason)
	}
}

// TestReg_SellerListEnterEditApprovedOnSaleProduct 已上架且审核通过商品，
// 从卖家“我的发布”列表能拿到该商品，进入修改保存后立即下架并生成新一轮待审记录。
func TestReg_SellerListEnterEditApprovedOnSaleProduct(t *testing.T) {
	env := newRegEnv(t)
	seller := regCreateUser(t, env.db, "reg_seller_edit", constants.UserRoleUser)
	admin := regCreateUser(t, env.db, "reg_admin_edit", constants.UserRoleAdmin)
	p := regPublish(t, env, seller.ID, "edit")
	rid := regReadReviews(t, env.db, p.ID)[0].ID
	if _, err := env.review.Decide(admin.ID, rid, dto.ProductReviewDecisionRequest{Action: "approve"}); err != nil {
		t.Fatalf("阶段=一审通过：%v", err)
	}

	// 模拟卖家从“我的发布”列表进入：ListBySeller 必须能看到该在售商品
	mine, err := env.product.ListBySeller(seller.ID, 1, 50)
	if err != nil {
		t.Fatalf("阶段=我的发布列表：%v", err)
	}
	var target *dto.ProductVO
	for i := range mine.List {
		if mine.List[i].ID == p.ID {
			v := mine.List[i]
			target = &v
		}
	}
	if target == nil {
		t.Fatalf("阶段=我的发布列表：已上架商品必须出现在卖家列表")
	}
	if target.Status != constants.ProductStatusOnSale || target.ReviewStatus != constants.ProductReviewApproved {
		t.Fatalf("阶段=我的发布列表：期望 on_sale/approved，实际 %s/%s", target.Status, target.ReviewStatus)
	}

	// 保存修改 → 立即下架 + 新一轮待审
	newTitle := "回归商品-已修改标题"
	updated, err := env.product.Update(seller.ID, p.ID, dto.ProductUpdateRequest{Title: &newTitle})
	if err != nil {
		t.Fatalf("阶段=保存修改：%v", err)
	}
	if updated.Status != constants.ProductStatusOffShelf || updated.ReviewStatus != constants.ProductReviewPending || updated.ReviewRound != 2 || updated.Title != newTitle {
		t.Fatalf("阶段=回读商品：修改后期望 off_shelf/pending/round2/新标题，实际 %s/%s/round%d/%s", updated.Status, updated.ReviewStatus, updated.ReviewRound, updated.Title)
	}
	rvs := regReadReviews(t, env.db, p.ID)
	if len(rvs) != 2 || rvs[1].Status != constants.ProductReviewPending {
		t.Fatalf("阶段=回读审核记录：应新增第二轮待审，实际 %d 条", len(rvs))
	}
}

// TestReg_NonSellerCannotModify 非卖家不能修改商品（管理员也不能走卖家修改链路）。
func TestReg_NonSellerCannotModify(t *testing.T) {
	env := newRegEnv(t)
	seller := regCreateUser(t, env.db, "reg_seller_ns", constants.UserRoleUser)
	other := regCreateUser(t, env.db, "reg_other_ns", constants.UserRoleUser)
	admin := regCreateUser(t, env.db, "reg_admin_ns", constants.UserRoleAdmin)
	p := regPublish(t, env, seller.ID, "nonseller")

	title := "路人修改"
	for _, tc := range []struct {
		stage string
		uid   uint
	}{
		{"修改-其他用户", other.ID},
		{"修改-管理员", admin.ID},
	} {
		_, err := env.product.Update(tc.uid, p.ID, dto.ProductUpdateRequest{Title: &title})
		if code := regAppErrCode(t, tc.stage, err); code != constants.CodeForbidden {
			t.Fatalf("阶段=%s：期望无权限 %d，实际 %d", tc.stage, constants.CodeForbidden, code)
		}
	}
	got := regReadProduct(t, env.db, p.ID)
	if got.Title == title || got.ReviewRound != 1 {
		t.Fatalf("阶段=回读商品：被拒绝的修改不得落库，title=%q round=%d", got.Title, got.ReviewRound)
	}
}

// TestReg_DuplicateModifyWhilePendingRejected 待审中重复修改被拒且不新增审核记录（状态幂等）。
func TestReg_DuplicateModifyWhilePendingRejected(t *testing.T) {
	env := newRegEnv(t)
	seller := regCreateUser(t, env.db, "reg_seller_dup", constants.UserRoleUser)
	p := regPublish(t, env, seller.ID, "dup")

	for i := 0; i < 3; i++ {
		title := fmt.Sprintf("重复修改-%d", i)
		if _, err := env.product.Update(seller.ID, p.ID, dto.ProductUpdateRequest{Title: &title}); err == nil {
			t.Fatalf("阶段=待审重复修改 第%d次：应被拒绝", i+1)
		}
	}
	if n := len(regReadReviews(t, env.db, p.ID)); n != 1 {
		t.Fatalf("阶段=回读审核记录：重复修改不得新增记录，实际 %d 条", n)
	}
	got := regReadProduct(t, env.db, p.ID)
	if got.ReviewStatus != constants.ProductReviewPending || got.ReviewRound != 1 {
		t.Fatalf("阶段=回读商品：状态应保持 pending/round1，实际 %s/round%d", got.ReviewStatus, got.ReviewRound)
	}
}

// TestReg_ConcurrentDecisionOnlyOneResult 并发审核唯一结果：多管理员/多操作同时对同一待审记录裁决，
// 恰好一个成功、其余失败；最终只写入一个审核人、一个结果，商品状态与记录一致。
// 多轮重复（3 轮）以确保不是调度巧合。
func TestReg_ConcurrentDecisionOnlyOneResult(t *testing.T) {
	const competitors = 12
	for iter := 0; iter < 3; iter++ {
		env := newRegEnv(t)
		seller := regCreateUser(t, env.db, fmt.Sprintf("reg_seller_cc_%d", iter), constants.UserRoleUser)
		p := regPublish(t, env, seller.ID, fmt.Sprintf("cc-%d", iter))
		rid := regReadReviews(t, env.db, p.ID)[0].ID
		admins := make([]*model.User, 0, competitors)
		for i := 0; i < competitors; i++ {
			admins = append(admins, regCreateUser(t, env.db, fmt.Sprintf("reg_admin_cc_%d_%d", iter, i), constants.UserRoleAdmin))
		}

		start := make(chan struct{})
		var wg sync.WaitGroup
		var mu sync.Mutex
		var successCodes, failCodes []int
		wg.Add(competitors)
		for i := 0; i < competitors; i++ {
			i := i
			go func() {
				defer wg.Done()
				<-start
				req := dto.ProductReviewDecisionRequest{Action: "approve"}
				if i%2 == 1 {
					req = dto.ProductReviewDecisionRequest{Action: "reject", Reason: fmt.Sprintf("并发驳回-%d", i)}
				}
				_, err := env.review.Decide(admins[i].ID, rid, req)
				mu.Lock()
				defer mu.Unlock()
				if err == nil {
					successCodes = append(successCodes, i)
				} else {
					var ae *util.AppError
					if errors.As(err, &ae) {
						failCodes = append(failCodes, ae.Code)
					}
				}
			}()
		}
		close(start)
		wg.Wait()

		stage := fmt.Sprintf("并发裁决 第%d轮", iter+1)
		if len(successCodes) != 1 {
			t.Fatalf("阶段=%s：恰好一个裁决应成功，实际成功 %d 个（失败 %d 个）", stage, len(successCodes), len(failCodes))
		}
		for _, c := range failCodes {
			if c != constants.CodeReviewStateInvalid {
				t.Fatalf("阶段=%s：失败方应为重复审核错误码 %d，实际出现 %d", stage, constants.CodeReviewStateInvalid, c)
			}
		}
		// 回读：记录唯一结果
		reviews := regReadReviews(t, env.db, p.ID)
		if len(reviews) != 1 {
			t.Fatalf("阶段=%s-回读：记录仍应只有 1 条，实际 %d 条", stage, len(reviews))
		}
		rv := reviews[0]
		winner := admins[successCodes[0]]
		if rv.ReviewerID == nil || *rv.ReviewerID != winner.ID || rv.ReviewedAt == nil {
			t.Fatalf("阶段=%s-回读：唯一审核人应为 admin[%d]=%d，实际 reviewer=%v", stage, successCodes[0], winner.ID, rv.ReviewerID)
		}
		// 商品状态与记录一致
		got := regReadProduct(t, env.db, p.ID)
		wantStatus := constants.ProductStatusOnSale
		if successCodes[0]%2 == 1 {
			wantStatus = constants.ProductStatusOffShelf
		}
		if got.ReviewStatus != rv.Status || got.Status != wantStatus {
			t.Fatalf("阶段=%s-回读：商品状态与审核结果不一致，product=%s/%s review=%s wantStatus=%s", stage, got.ReviewStatus, got.Status, rv.Status, wantStatus)
		}
	}
}

// TestReg_RejectedStaysOffShelfDuringReReview 复审期间继续下架：
// 已售商品在卖家修改进入复审后，即使订单取消也不会被恢复上架。
func TestReg_RejectedStaysOffShelfDuringReReview(t *testing.T) {
	env := newRegEnv(t)
	seller := regCreateUser(t, env.db, "reg_seller_rev", constants.UserRoleUser)
	buyer := regCreateUser(t, env.db, "reg_buyer_rev", constants.UserRoleUser)
	admin := regCreateUser(t, env.db, "reg_admin_rev", constants.UserRoleAdmin)
	p := regPublish(t, env, seller.ID, "rereview")
	rid := regReadReviews(t, env.db, p.ID)[0].ID
	if _, err := env.review.Decide(admin.ID, rid, dto.ProductReviewDecisionRequest{Action: "approve"}); err != nil {
		t.Fatalf("阶段=一审通过：%v", err)
	}
	addr := &model.Address{UserID: buyer.ID, ReceiverName: "买家", Phone: "13800000000", Province: "省", City: "市", Detail: "地址"}
	if err := env.db.Create(addr).Error; err != nil {
		t.Fatalf("阶段=建地址：%v", err)
	}
	order, err := env.order.Create(buyer.ID, dto.OrderCreateRequest{ProductID: p.ID, AddressID: addr.ID, Quantity: 1})
	if err != nil {
		t.Fatalf("阶段=下单：%v", err)
	}
	// 卖家修改已售/已通过商品 → 复审中、下架
	newTitle := "订单存在时修改标题"
	if _, err := env.product.Update(seller.ID, p.ID, dto.ProductUpdateRequest{Title: &newTitle}); err != nil {
		t.Fatalf("阶段=卖家修改触发复审：%v", err)
	}
	if got := regReadProduct(t, env.db, p.ID); got.ReviewStatus != constants.ProductReviewPending || got.Status != constants.ProductStatusOffShelf {
		t.Fatalf("阶段=复审中回读：期望 pending/off_shelf，实际 %s/%s", got.ReviewStatus, got.Status)
	}
	// 取消订单：复审中的商品必须继续下架
	if _, err := env.order.Cancel(buyer.ID, order.ID, "buyer"); err != nil {
		t.Fatalf("阶段=取消订单：%v", err)
	}
	if got := regReadProduct(t, env.db, p.ID); got.ReviewStatus != constants.ProductReviewPending || got.Status != constants.ProductStatusOffShelf {
		t.Fatalf("阶段=取消后回读：复审期间必须继续下架，实际 %s/%s", got.ReviewStatus, got.Status)
	}
}
