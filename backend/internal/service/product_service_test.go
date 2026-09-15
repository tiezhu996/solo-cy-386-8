package service

import (
	"log/slog"
	"strconv"
	"testing"
	"time"

	"github.com/marketpal/marketpal/internal/constants"
	"github.com/marketpal/marketpal/internal/dto"
	"github.com/marketpal/marketpal/internal/model"
	"github.com/marketpal/marketpal/internal/repository"
	"gorm.io/gorm"
)

// fakeProductRepo 内存版商品仓储。
type fakeProductRepo struct {
	products map[uint]*model.Product
	seq      uint
}

func newFakeProductRepo() *fakeProductRepo {
	return &fakeProductRepo{products: map[uint]*model.Product{}}
}

func (f *fakeProductRepo) CreateTx(tx *gorm.DB, p *model.Product) error { return f.Create(p) }
func (f *fakeProductRepo) Create(p *model.Product) error {
	f.seq++
	p.ID = f.seq
	f.products[p.ID] = p
	return nil
}
func (f *fakeProductRepo) GetByID(id uint) (*model.Product, error) {
	if p, ok := f.products[id]; ok {
		return p, nil
	}
	return nil, repository.ErrNotFound
}
func (f *fakeProductRepo) GetByIDForUpdate(tx *gorm.DB, id uint) (*model.Product, error) {
	return f.GetByID(id)
}
func (f *fakeProductRepo) List(query map[string]interface{}, sortBy string, page, pageSize int) ([]model.Product, int64, error) {
	return nil, 0, nil
}
func (f *fakeProductRepo) ListBySeller(sellerID uint, page, pageSize int) ([]model.Product, int64, error) {
	return nil, 0, nil
}
func (f *fakeProductRepo) ListByIDs(ids []uint) ([]model.Product, error) { return nil, nil }
func (f *fakeProductRepo) Update(p *model.Product) error {
	if _, ok := f.products[p.ID]; !ok {
		return repository.ErrNotFound
	}
	f.products[p.ID] = p
	return nil
}
func (f *fakeProductRepo) UpdateTx(tx *gorm.DB, p *model.Product) error { return f.Update(p) }
func (f *fakeProductRepo) IncrViewCount(id uint) error                  { return nil }
func (f *fakeProductRepo) IncrFavoriteCount(tx *gorm.DB, id uint, delta int) error {
	return nil
}
func (f *fakeProductRepo) UpdateStatusForUpdate(tx *gorm.DB, id uint, status string) error {
	if p, ok := f.products[id]; ok {
		p.Status = status
		return nil
	}
	return repository.ErrNotFound
}
func (f *fakeProductRepo) UpdateReviewForUpdate(tx *gorm.DB, id uint, updates map[string]interface{}) error {
	p, ok := f.products[id]
	if !ok {
		return repository.ErrNotFound
	}
	if v, ok := updates["status"].(string); ok {
		p.Status = v
	}
	if v, ok := updates["review_status"].(string); ok {
		p.ReviewStatus = v
	}
	if v, ok := updates["reject_reason"].(string); ok {
		p.RejectReason = v
	}
	return nil
}
func (f *fakeProductRepo) RestoreOnSaleIfApprovedForUpdate(tx *gorm.DB, id uint) (bool, error) {
	p, ok := f.products[id]
	if !ok {
		return false, repository.ErrNotFound
	}
	if p.ReviewStatus == constants.ProductReviewApproved {
		p.Status = constants.ProductStatusOnSale
		return true, nil
	}
	return false, nil
}

// fakeProductReviewRepo 内存版审核仓储（记录条件更新次数以模拟并发唯一结果）。
type fakeProductReviewRepo struct {
	reviews map[uint]*model.ProductReview
	seq     uint
}

func newFakeProductReviewRepo() *fakeProductReviewRepo {
	return &fakeProductReviewRepo{reviews: map[uint]*model.ProductReview{}}
}

func (f *fakeProductReviewRepo) CreateTx(tx *gorm.DB, rv *model.ProductReview) error {
	f.seq++
	rv.ID = f.seq
	f.reviews[rv.ID] = rv
	return nil
}
func (f *fakeProductReviewRepo) GetByID(id uint) (*model.ProductReview, error) {
	if rv, ok := f.reviews[id]; ok {
		return rv, nil
	}
	return nil, repository.ErrNotFound
}
func (f *fakeProductReviewRepo) GetByIDForUpdate(tx *gorm.DB, id uint) (*model.ProductReview, error) {
	return f.GetByID(id)
}
func (f *fakeProductReviewRepo) GetLatestByProduct(productID uint) (*model.ProductReview, error) {
	var latest *model.ProductReview
	for _, rv := range f.reviews {
		if rv.ProductID == productID && (latest == nil || rv.Round > latest.Round) {
			latest = rv
		}
	}
	if latest == nil {
		return nil, repository.ErrNotFound
	}
	return latest, nil
}
func (f *fakeProductReviewRepo) GetPendingByProductForUpdate(tx *gorm.DB, productID uint) (*model.ProductReview, error) {
	rv, err := f.GetLatestByProduct(productID)
	if err != nil {
		return nil, err
	}
	if rv.Status != constants.ProductReviewPending {
		return nil, repository.ErrNotFound
	}
	return rv, nil
}
func (f *fakeProductReviewRepo) MaxRound(tx *gorm.DB, productID uint) (int, error) {
	rv, err := f.GetLatestByProduct(productID)
	if err != nil {
		return 0, nil
	}
	return rv.Round, nil
}
func (f *fakeProductReviewRepo) DecideForUpdate(tx *gorm.DB, id uint, reviewerID *uint, status, reason string, reviewedAt interface{}) error {
	rv, ok := f.reviews[id]
	if !ok {
		return repository.ErrNotFound
	}
	if rv.Status != constants.ProductReviewPending {
		return repository.ErrReviewAlreadyDecided
	}
	rv.Status = status
	rv.Reason = reason
	rv.ReviewerID = reviewerID
	if t, ok := reviewedAt.(*time.Time); ok {
		rv.ReviewedAt = t
	}
	return nil
}
func (f *fakeProductReviewRepo) List(query map[string]interface{}, page, pageSize int) ([]model.ProductReview, int64, error) {
	return nil, 0, nil
}
func (f *fakeProductReviewRepo) ListByProduct(productID uint) ([]model.ProductReview, error) {
	return nil, nil
}

// fakeFavoriteRepo 内存版收藏仓储。
type fakeFavoriteRepo struct {
	favs map[string]bool
}

func newFakeFavoriteRepo() *fakeFavoriteRepo {
	return &fakeFavoriteRepo{favs: map[string]bool{}}
}
func (f *fakeFavoriteRepo) Create(fav *model.Favorite) error {
	f.favs[key(fav.UserID, fav.ProductID)] = true
	return nil
}
func (f *fakeFavoriteRepo) Delete(userID, productID uint) error {
	if !f.favs[key(userID, productID)] {
		return repository.ErrNotFound
	}
	delete(f.favs, key(userID, productID))
	return nil
}
func (f *fakeFavoriteRepo) Exists(userID, productID uint) (bool, error) {
	return f.favs[key(userID, productID)], nil
}
func (f *fakeFavoriteRepo) ListByUser(userID uint, page, pageSize int) ([]model.Favorite, int64, error) {
	return nil, 0, nil
}

func key(a, b uint) string {
	return strconv.FormatUint(uint64(a), 10) + "-" + strconv.FormatUint(uint64(b), 10)
}

// uintPtr 测试辅助：构造审核人 ID 指针。
func uintPtr(v uint) *uint { return &v }

func newTestProductService() (*ProductService, *fakeProductRepo, *fakeFavoriteRepo, *fakeProductReviewRepo) {
	pr := newFakeProductRepo()
	fr := newFakeFavoriteRepo()
	rr := newFakeProductReviewRepo()
	logger := slog.New(slog.NewTextHandler(discardWriter{}, nil))
	return NewProductService(nil, pr, fr, rr, logger), pr, fr, rr
}

func TestProductServiceCreate(t *testing.T) {
	svc, pr, _, _ := newTestProductService()
	req := dto.ProductCreateRequest{
		Title: "iPhone 13", Description: "九成新 iPhone 13 128G", OriginalPrice: 5999,
		Price: 3999, Condition: "almost_new", Category: "digital",
	}
	product, err := svc.Create(1, req)
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	// 新发布商品必须先进入待审核并下架。
	if product.Status != constants.ProductStatusOffShelf {
		t.Fatalf("expected off_shelf before review, got %s", product.Status)
	}
	if product.ReviewStatus != constants.ProductReviewPending {
		t.Fatalf("expected pending_review, got %s", product.ReviewStatus)
	}
	if _, ok := pr.products[product.ID]; !ok {
		t.Fatal("product not persisted")
	}
}

func TestProductServiceInvalidCategory(t *testing.T) {
	svc, _, _, _ := newTestProductService()
	req := dto.ProductCreateRequest{
		Title: "x", Description: "yy", OriginalPrice: 1, Price: 1,
		Condition: "almost_new", Category: "unknown",
	}
	if _, err := svc.Create(1, req); err == nil {
		t.Fatal("expected error for invalid category")
	}
}

func TestProductServiceReviewFlow(t *testing.T) {
	svc, pr, _, rr := newTestProductService()
	req := dto.ProductCreateRequest{
		Title: "Kindle", Description: "二手电子书阅读器", OriginalPrice: 599, Price: 299,
		Condition: "lightly_used", Category: "books",
	}
	product, _ := svc.Create(1, req)

	// 待审核中重复修改不能重复改状态。
	title := "Kindle Paperwhite"
	if _, err := svc.Update(1, product.ID, dto.ProductUpdateRequest{Title: &title}); err == nil {
		t.Fatal("expected CodeReviewDuplicate when updating pending product")
	}

	// 管理员驳回需要原因（通过 review service 直接驱动仓储模拟）。
	rv, _ := rr.GetPendingByProductForUpdate(nil, product.ID)
	if rv == nil || rv.Round != 1 {
		t.Fatalf("expected round 1 pending review, got %+v", rv)
	}
	now := time.Now()
	if err := rr.DecideForUpdate(nil, rv.ID, uintPtr(99), constants.ProductReviewRejected, "图片不清晰", &now); err != nil {
		t.Fatalf("first decide failed: %v", err)
	}
	// 模拟审核 service 同步商品审核结果（真实链路在 ProductReviewService.Decide 中完成）。
	if err := pr.UpdateReviewForUpdate(nil, product.ID, map[string]interface{}{
		"status": constants.ProductStatusOffShelf, "review_status": constants.ProductReviewRejected, "reject_reason": "图片不清晰",
	}); err != nil {
		t.Fatalf("apply reject to product failed: %v", err)
	}
	// 并发重复审核：同一轮记录不能再改状态。
	if err := rr.DecideForUpdate(nil, rv.ID, uintPtr(100), constants.ProductReviewApproved, "", &now); err == nil {
		t.Fatal("expected ErrReviewAlreadyDecided on concurrent duplicate review")
	}

	// 驳回后卖家修改重提：进入新一轮待审核，复审期间继续下架。
	desc := "已补图，功能完好"
	updated, err := svc.Update(1, product.ID, dto.ProductUpdateRequest{Description: &desc})
	if err != nil {
		t.Fatalf("resubmit after reject failed: %v", err)
	}
	if updated.ReviewStatus != constants.ProductReviewPending || updated.Status != constants.ProductStatusOffShelf {
		t.Fatalf("expected pending + off_shelf during re-review, got %s/%s", updated.ReviewStatus, updated.Status)
	}
	if updated.ReviewRound != 2 {
		t.Fatalf("expected round 2, got %d", updated.ReviewRound)
	}
	rv2, _ := rr.GetPendingByProductForUpdate(nil, product.ID)
	if rv2.Round != 2 {
		t.Fatalf("expected round 2 pending review, got %d", rv2.Round)
	}

	// 通过后商品恢复在售（模拟审核 service 同步）。
	if err := rr.DecideForUpdate(nil, rv2.ID, uintPtr(99), constants.ProductReviewApproved, "", &now); err != nil {
		t.Fatalf("approve failed: %v", err)
	}
	if err := pr.UpdateReviewForUpdate(nil, product.ID, map[string]interface{}{
		"status": constants.ProductStatusOnSale, "review_status": constants.ProductReviewApproved, "reject_reason": "",
	}); err != nil {
		t.Fatalf("apply approve failed: %v", err)
	}
	if got := pr.products[product.ID]; got.Status != constants.ProductStatusOnSale || got.ReviewStatus != constants.ProductReviewApproved {
		t.Fatalf("expected on_sale/approved, got %s/%s", got.Status, got.ReviewStatus)
	}

	// 同一轮记录再次给出结果必须失败（并发审核只有一个结果）。
	if err := rr.DecideForUpdate(nil, rv2.ID, uintPtr(100), constants.ProductReviewRejected, "again", &now); err == nil {
		t.Fatal("expected duplicate decide blocked after approval")
	}
}

func TestProductServiceFavoriteUnfavorite(t *testing.T) {
	svc, _, fr, _ := newTestProductService()
	req := dto.ProductCreateRequest{
		Title: "Book", Description: "二手图书", OriginalPrice: 50, Price: 20,
		Condition: "lightly_used", Category: "books",
	}
	product, _ := svc.Create(1, req)
	// 待审核商品不可收藏。
	if err := svc.Favorite(2, product.ID); err == nil {
		t.Fatal("expected favorite blocked before approval")
	}
	product.ReviewStatus = constants.ProductReviewApproved
	if err := svc.Favorite(2, product.ID); err != nil {
		t.Fatalf("Favorite() error: %v", err)
	}
	if ok, _ := fr.Exists(2, product.ID); !ok {
		t.Fatal("favorite not recorded")
	}
	if err := svc.Unfavorite(2, product.ID); err != nil {
		t.Fatalf("Unfavorite() error: %v", err)
	}
	if ok, _ := fr.Exists(2, product.ID); ok {
		t.Fatal("favorite should be removed")
	}
}
