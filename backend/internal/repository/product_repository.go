package repository

import (
	"errors"
	"fmt"

	"github.com/marketpal/marketpal/internal/model"
	"gorm.io/gorm"
)

// ProductRepository 商品仓储接口。
type ProductRepository interface {
	Create(product *model.Product) error
	GetByID(id uint) (*model.Product, error)
	GetByIDForUpdate(tx *gorm.DB, id uint) (*model.Product, error)
	List(query map[string]interface{}, sortBy string, page, pageSize int) ([]model.Product, int64, error)
	ListBySeller(sellerID uint, page, pageSize int) ([]model.Product, int64, error)
	ListByIDs(ids []uint) ([]model.Product, error)
	Update(product *model.Product) error
	IncrViewCount(id uint) error
	IncrFavoriteCount(tx *gorm.DB, id uint, delta int) error
	UpdateStatusForUpdate(tx *gorm.DB, id uint, status string) error
}

type productRepo struct {
	db *gorm.DB
}

// NewProductRepository 构造商品仓储。
func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepo{db: db}
}

func (r *productRepo) Create(product *model.Product) error {
	if err := r.db.Create(product).Error; err != nil {
		return fmt.Errorf("create product: %w", err)
	}
	return nil
}

func (r *productRepo) GetByID(id uint) (*model.Product, error) {
	var p model.Product
	err := r.db.Preload("Seller").First(&p, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get product by id %d: %w", id, err)
	}
	return &p, nil
}

func (r *productRepo) GetByIDForUpdate(tx *gorm.DB, id uint) (*model.Product, error) {
	var p model.Product
	err := tx.Clauses(clauseLocking()).First(&p, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get product by id %d for update: %w", id, err)
	}
	return &p, nil
}

// numVal 将查询参数中的数值提取为 float64（0 视为未设置）。
func numVal(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case uint:
		return float64(n)
	default:
		return 0
	}
}

func (r *productRepo) List(query map[string]interface{}, sortBy string, page, pageSize int) ([]model.Product, int64, error) {
	var products []model.Product
	var total int64
	q := r.db.Model(&model.Product{})
	if v, ok := query["keyword"]; ok && v != "" {
		kw := "%" + fmt.Sprintf("%v", v) + "%"
		q = q.Where("title ILIKE ? OR description ILIKE ?", kw, kw)
	}
	if v, ok := query["category"]; ok && v != "" {
		q = q.Where("category = ?", v)
	}
	if v, ok := query["condition"]; ok && v != "" {
		q = q.Where("condition = ?", v)
	}
	if v, ok := query["min_price"]; ok && numVal(v) > 0 {
		q = q.Where("price >= ?", v)
	}
	if v, ok := query["max_price"]; ok && numVal(v) > 0 {
		q = q.Where("price <= ?", v)
	}
	if v, ok := query["status"]; ok && v != "" {
		q = q.Where("status = ?", v)
	}
	if v, ok := query["seller_id"]; ok && numVal(v) > 0 {
		q = q.Where("seller_id = ?", v)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count products: %w", err)
	}
	order := "id DESC"
	switch sortBy {
	case "price":
		order = "price ASC"
	case "price_desc":
		order = "price DESC"
	case "time":
		order = "created_at ASC"
	case "time_desc":
		order = "created_at DESC"
	}
	if err := q.Preload("Seller").Order(order).Offset((page - 1) * pageSize).Limit(pageSize).Find(&products).Error; err != nil {
		return nil, 0, fmt.Errorf("list products: %w", err)
	}
	return products, total, nil
}

func (r *productRepo) ListBySeller(sellerID uint, page, pageSize int) ([]model.Product, int64, error) {
	var products []model.Product
	var total int64
	q := r.db.Model(&model.Product{}).Where("seller_id = ?", sellerID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count seller products: %w", err)
	}
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&products).Error; err != nil {
		return nil, 0, fmt.Errorf("list seller products: %w", err)
	}
	return products, total, nil
}

func (r *productRepo) ListByIDs(ids []uint) ([]model.Product, error) {
	if len(ids) == 0 {
		return []model.Product{}, nil
	}
	var products []model.Product
	if err := r.db.Where("id IN ?", ids).Find(&products).Error; err != nil {
		return nil, fmt.Errorf("list products by ids: %w", err)
	}
	return products, nil
}

func (r *productRepo) Update(product *model.Product) error {
	if err := r.db.Save(product).Error; err != nil {
		return fmt.Errorf("update product %d: %w", product.ID, err)
	}
	return nil
}

func (r *productRepo) IncrViewCount(id uint) error {
	res := r.db.Model(&model.Product{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + 1"))
	if res.Error != nil {
		return fmt.Errorf("incr view count of product %d: %w", id, res.Error)
	}
	return nil
}

func (r *productRepo) IncrFavoriteCount(tx *gorm.DB, id uint, delta int) error {
	if tx == nil {
		tx = r.db
	}
	res := tx.Model(&model.Product{}).Where("id = ?", id).UpdateColumn("favorite_count", gorm.Expr("favorite_count + ?", delta))
	if res.Error != nil {
		return fmt.Errorf("update favorite count of product %d: %w", id, res.Error)
	}
	return nil
}

func (r *productRepo) UpdateStatusForUpdate(tx *gorm.DB, id uint, status string) error {
	res := tx.Model(&model.Product{}).Where("id = ?", id).Update("status", status)
	if res.Error != nil {
		return fmt.Errorf("update product %d status to %s: %w", id, status, res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
