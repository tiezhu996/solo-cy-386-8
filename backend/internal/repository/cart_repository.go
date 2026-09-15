package repository

import (
	"errors"
	"fmt"

	"github.com/marketpal/marketpal/internal/model"
	"gorm.io/gorm"
)

// CartRepository 购物车仓储接口。
type CartRepository interface {
	Create(item *model.CartItem) error
	GetByID(id uint) (*model.CartItem, error)
	GetByUserAndProduct(userID, productID uint) (*model.CartItem, error)
	ListByUser(userID uint) ([]model.CartItem, error)
	Update(item *model.CartItem) error
	Delete(id, userID uint) error
	Clear(userID uint) error
}

type cartRepo struct {
	db *gorm.DB
}

// NewCartRepository 构造购物车仓储。
func NewCartRepository(db *gorm.DB) CartRepository {
	return &cartRepo{db: db}
}

func (r *cartRepo) Create(item *model.CartItem) error {
	if err := r.db.Create(item).Error; err != nil {
		return fmt.Errorf("create cart item: %w", err)
	}
	return nil
}

func (r *cartRepo) GetByID(id uint) (*model.CartItem, error) {
	var it model.CartItem
	err := r.db.First(&it, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get cart item by id %d: %w", id, err)
	}
	return &it, nil
}

func (r *cartRepo) GetByUserAndProduct(userID, productID uint) (*model.CartItem, error) {
	var it model.CartItem
	err := r.db.Where("user_id = ? AND product_id = ?", userID, productID).First(&it).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get cart user=%d product=%d: %w", userID, productID, err)
	}
	return &it, nil
}

func (r *cartRepo) ListByUser(userID uint) ([]model.CartItem, error) {
	var items []model.CartItem
	if err := r.db.Preload("Product").Preload("Product.Seller").Where("user_id = ?", userID).Order("id DESC").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list cart user=%d: %w", userID, err)
	}
	return items, nil
}

func (r *cartRepo) Update(item *model.CartItem) error {
	if err := r.db.Save(item).Error; err != nil {
		return fmt.Errorf("update cart item %d: %w", item.ID, err)
	}
	return nil
}

func (r *cartRepo) Delete(id, userID uint) error {
	res := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.CartItem{})
	if res.Error != nil {
		return fmt.Errorf("delete cart item id=%d user=%d: %w", id, userID, res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *cartRepo) Clear(userID uint) error {
	res := r.db.Where("user_id = ?", userID).Delete(&model.CartItem{})
	if res.Error != nil {
		return fmt.Errorf("clear cart user=%d: %w", userID, res.Error)
	}
	return nil
}
