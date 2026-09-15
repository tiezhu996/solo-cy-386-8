package service

import (
	"log/slog"
	"strconv"
	"testing"

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

func (f *fakeProductRepo) CreateWithTx(tx *gorm.DB, p *model.Product) error { return f.Create(p) }
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
func (f *fakeProductRepo) IncrViewCount(id uint) error { return nil }
func (f *fakeProductRepo) IncrFavoriteCount(tx *gorm.DB, id uint, delta int) error {
	return nil
}
func (f *fakeProductRepo) UpdateStatusForUpdate(tx *gorm.DB, id uint, status string) error {
	return nil
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

func key(a, b uint) string { return strconv.FormatUint(uint64(a), 10) + "-" + strconv.FormatUint(uint64(b), 10) }

func newTestProductService() (*ProductService, *fakeProductRepo, *fakeFavoriteRepo) {
	pr := newFakeProductRepo()
	fr := newFakeFavoriteRepo()
	logger := slog.New(slog.NewTextHandler(discardWriter{}, nil))
	return NewProductService(pr, fr, logger), pr, fr
}

func TestProductServiceCreate(t *testing.T) {
	svc, pr, _ := newTestProductService()
	req := dto.ProductCreateRequest{
		Title: "iPhone 13", Description: "九成新 iPhone 13 128G", OriginalPrice: 5999,
		Price: 3999, Condition: "almost_new", Category: "digital",
	}
	product, err := svc.Create(1, req)
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if product.Status != "on_sale" {
		t.Fatalf("expected on_sale, got %s", product.Status)
	}
	if _, ok := pr.products[product.ID]; !ok {
		t.Fatal("product not persisted")
	}
}

func TestProductServiceInvalidCategory(t *testing.T) {
	svc, _, _ := newTestProductService()
	req := dto.ProductCreateRequest{
		Title: "x", Description: "yy", OriginalPrice: 1, Price: 1,
		Condition: "almost_new", Category: "unknown",
	}
	if _, err := svc.Create(1, req); err == nil {
		t.Fatal("expected error for invalid category")
	}
}

func TestProductServiceFavoriteUnfavorite(t *testing.T) {
	svc, _, fr := newTestProductService()
	req := dto.ProductCreateRequest{
		Title: "Book", Description: "二手图书", OriginalPrice: 50, Price: 20,
		Condition: "lightly_used", Category: "books",
	}
	product, _ := svc.Create(1, req)
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
