package service

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/marketpal/marketpal/internal/constants"
	"github.com/marketpal/marketpal/internal/dto"
	"github.com/marketpal/marketpal/internal/model"
	"github.com/marketpal/marketpal/internal/repository"
	"gorm.io/gorm"
)

// newReviewIntegrationDB 打开开启外键约束的内存 SQLite，并按生产方式 AutoMigrate 全部模型。
func newReviewIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_fk=1"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		t.Fatalf("enable fk: %v", err)
	}
	models := []interface{}{
		&model.User{}, &model.Product{}, &model.Favorite{}, &model.Address{},
		&model.CartItem{}, &model.Order{}, &model.Message{}, &model.Review{},
		&model.AuditLog{}, &model.ProductReview{},
	}
	if err := db.AutoMigrate(models...); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	// 每个测试用独立内存库：先清空。
	db.Exec("DELETE FROM product_reviews")
	db.Exec("DELETE FROM products")
	db.Exec("DELETE FROM users")
	return db
}

// TestCreateProductPersistsProductAndPendingReview 发布链路：商品与首轮待审记录必须在同一事务内一起落库，
// 待审记录 reviewer_id 为 NULL（PostgreSQL/SQLite 外键约束下不得写 0 伪造账号）。
func TestCreateProductPersistsProductAndPendingReview(t *testing.T) {
	db := newReviewIntegrationDB(t)
	logger := slog.New(slog.NewTextHandler(discardWriter{}, nil))
	seller := &model.User{Username: "seller_it", PasswordHash: "h", Nickname: "卖家", Role: "user"}
	if err := db.Create(seller).Error; err != nil {
		t.Fatalf("create seller: %v", err)
	}
	pRepo := repository.NewProductRepository(db)
	fRepo := repository.NewFavoriteRepository(db)
	rRepo := repository.NewProductReviewRepository(db)
	svc := NewProductService(db, pRepo, fRepo, rRepo, logger)

	product, err := svc.Create(seller.ID, dto.ProductCreateRequest{
		Title: "iPad Air", Description: "二手平板，几乎全新", OriginalPrice: 4799, Price: 2899,
		Condition: "almost_new", Category: "digital",
	})
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if product.ReviewStatus != constants.ProductReviewPending || product.Status != constants.ProductStatusOffShelf {
		t.Fatalf("new product must be pending/off_shelf, got %s/%s", product.ReviewStatus, product.Status)
	}

	// 商品落库。
	var pc int64
	db.Model(&model.Product{}).Where("id = ?", product.ID).Count(&pc)
	if pc != 1 {
		t.Fatal("product not persisted")
	}
	// 首轮待审记录落库，且 reviewer_id 为 NULL。
	var rv model.ProductReview
	if err := db.Where("product_id = ? AND round = 1", product.ID).First(&rv).Error; err != nil {
		t.Fatalf("pending review not persisted: %v", err)
	}
	if rv.Status != constants.ProductReviewPending {
		t.Fatalf("expected pending review, got %s", rv.Status)
	}
	if rv.ReviewerID != nil {
		t.Fatalf("pending review reviewer_id must be NULL, got %d", *rv.ReviewerID)
	}
}

// failingReviewRepo 第二步入库必失败，用来验证事务整体回滚、不留半条商品。
type failingReviewRepo struct {
	inner repository.ProductReviewRepository
}

func (f *failingReviewRepo) CreateTx(tx *gorm.DB, review *model.ProductReview) error {
	return errInjectedReviewFailure
}
func (f *failingReviewRepo) GetByID(id uint) (*model.ProductReview, error) {
	return f.inner.GetByID(id)
}
func (f *failingReviewRepo) GetByIDForUpdate(tx *gorm.DB, id uint) (*model.ProductReview, error) {
	return f.inner.GetByIDForUpdate(tx, id)
}
func (f *failingReviewRepo) GetLatestByProduct(productID uint) (*model.ProductReview, error) {
	return f.inner.GetLatestByProduct(productID)
}
func (f *failingReviewRepo) GetPendingByProductForUpdate(tx *gorm.DB, productID uint) (*model.ProductReview, error) {
	return f.inner.GetPendingByProductForUpdate(tx, productID)
}
func (f *failingReviewRepo) MaxRound(tx *gorm.DB, productID uint) (int, error) {
	return f.inner.MaxRound(tx, productID)
}
func (f *failingReviewRepo) DecideForUpdate(tx *gorm.DB, id uint, reviewerID *uint, status, reason string, reviewedAt interface{}) error {
	return f.inner.DecideForUpdate(tx, id, reviewerID, status, reason, reviewedAt)
}
func (f *failingReviewRepo) List(query map[string]interface{}, page, pageSize int) ([]model.ProductReview, int64, error) {
	return f.inner.List(query, page, pageSize)
}
func (f *failingReviewRepo) ListByProduct(productID uint) ([]model.ProductReview, error) {
	return f.inner.ListByProduct(productID)
}

var errInjectedReviewFailure = errors.New("injected review insert failure")

// TestCreateProductRollsBackWhenReviewFails 审核记录写入失败时，商品必须一并回滚。
func TestCreateProductRollsBackWhenReviewFails(t *testing.T) {
	db := newReviewIntegrationDB(t)
	logger := slog.New(slog.NewTextHandler(discardWriter{}, nil))
	seller := &model.User{Username: "seller_rb", PasswordHash: "h", Nickname: "卖家", Role: "user"}
	if err := db.Create(seller).Error; err != nil {
		t.Fatalf("create seller: %v", err)
	}
	pRepo := repository.NewProductRepository(db)
	fRepo := repository.NewFavoriteRepository(db)
	rRepo := &failingReviewRepo{inner: repository.NewProductReviewRepository(db)}
	svc := NewProductService(db, pRepo, fRepo, rRepo, logger)

	if _, err := svc.Create(seller.ID, dto.ProductCreateRequest{
		Title: "Rollback", Description: "回滚验证专用商品", OriginalPrice: 100, Price: 50,
		Condition: "brand_new", Category: "other",
	}); err == nil {
		t.Fatal("expected error when review insert fails")
	}

	var pc int64
	db.Model(&model.Product{}).Where("seller_id = ?", seller.ID).Count(&pc)
	if pc != 0 {
		t.Fatalf("product must be rolled back, found %d orphan product(s)", pc)
	}
}

// TestApprovedOnSaleProductEditResubmits 已上架且审核通过的商品，卖家修改后必须立即下架并生成新一轮待审记录。
func TestApprovedOnSaleProductEditResubmits(t *testing.T) {
	db := newReviewIntegrationDB(t)
	logger := slog.New(slog.NewTextHandler(discardWriter{}, nil))
	seller := &model.User{Username: "seller_edit", PasswordHash: "h", Nickname: "卖家", Role: "user"}
	if err := db.Create(seller).Error; err != nil {
		t.Fatalf("create seller: %v", err)
	}
	pRepo := repository.NewProductRepository(db)
	fRepo := repository.NewFavoriteRepository(db)
	rRepo := repository.NewProductReviewRepository(db)
	svc := NewProductService(db, pRepo, fRepo, rRepo, logger)
	reviewSvc := NewProductReviewService(db, rRepo, pRepo, logger)

	product, err := svc.Create(seller.ID, dto.ProductCreateRequest{
		Title: "Kindle", Description: "电子书阅读器", OriginalPrice: 599, Price: 299,
		Condition: "lightly_used", Category: "books",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	rv, err := rRepo.GetPendingByProductForUpdate(db, product.ID)
	if err != nil {
		t.Fatalf("pending review: %v", err)
	}
	// 管理员审核通过 → 在售。
	admin := &model.User{Username: "admin_edit", PasswordHash: "h", Nickname: "管理员", Role: "admin"}
	if err := db.Create(admin).Error; err != nil {
		t.Fatalf("create admin: %v", err)
	}
	if _, err := reviewSvc.Decide(admin.ID, rv.ID, dto.ProductReviewDecisionRequest{Action: "approve"}); err != nil {
		t.Fatalf("approve: %v", err)
	}

	// 卖家修改已上架商品 → 立即下架 + 新一轮待审。
	newPrice := 279.0
	updated, err := svc.Update(seller.ID, product.ID, dto.ProductUpdateRequest{Price: &newPrice})
	if err != nil {
		t.Fatalf("update approved product: %v", err)
	}
	if updated.Status != constants.ProductStatusOffShelf || updated.ReviewStatus != constants.ProductReviewPending || updated.ReviewRound != 2 {
		t.Fatalf("expected off_shelf/pending/round2, got %s/%s/round%d", updated.Status, updated.ReviewStatus, updated.ReviewRound)
	}
	var rc int64
	db.Model(&model.ProductReview{}).Where("product_id = ?", product.ID).Count(&rc)
	if rc != 2 {
		t.Fatalf("expected 2 review records, got %d", rc)
	}
	var pending model.ProductReview
	if err := db.Where("product_id = ? AND round = 2 AND status = ?", product.ID, constants.ProductReviewPending).First(&pending).Error; err != nil {
		t.Fatalf("round 2 pending review missing: %v", err)
	}
	if pending.ReviewerID != nil {
		t.Fatalf("new pending round reviewer_id must be NULL")
	}

	// 待审中重复修改必须被拒绝且不新增记录。
	again := 269.0
	if _, err := svc.Update(seller.ID, product.ID, dto.ProductUpdateRequest{Price: &again}); err == nil {
		t.Fatal("expected duplicate update while pending to be rejected")
	}
	db.Model(&model.ProductReview{}).Where("product_id = ?", product.ID).Count(&rc)
	if rc != 2 {
		t.Fatalf("duplicate resubmit must not add a review record, got %d", rc)
	}

	// 非卖家不能修改。
	other := &model.User{Username: "other_edit", PasswordHash: "h", Nickname: "路人", Role: "user"}
	if err := db.Create(other).Error; err != nil {
		t.Fatalf("create other: %v", err)
	}
	if _, err := svc.Update(other.ID, product.ID, dto.ProductUpdateRequest{Price: &again}); err == nil {
		t.Fatal("expected non-seller update to be rejected")
	}
}
