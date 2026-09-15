package repository

import (
	"errors"
	"fmt"

	"github.com/marketpal/marketpal/internal/model"
	"gorm.io/gorm"
)

// OrderRepository 订单仓储接口。
type OrderRepository interface {
	Create(order *model.Order) error
	CreateWithTx(tx *gorm.DB, order *model.Order) error
	GetByID(id uint) (*model.Order, error)
	GetByIDForUpdate(tx *gorm.DB, id uint) (*model.Order, error)
	ListByBuyer(buyerID uint, status string, page, pageSize int) ([]model.Order, int64, error)
	ListBySeller(sellerID uint, status string, page, pageSize int) ([]model.Order, int64, error)
	UpdateStatusForUpdate(tx *gorm.DB, id uint, status string, updates map[string]interface{}) error
	Update(order *model.Order) error
}

type orderRepo struct {
	db *gorm.DB
}

// NewOrderRepository 构造订单仓储。
func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepo{db: db}
}

func (r *orderRepo) Create(order *model.Order) error {
	if err := r.db.Create(order).Error; err != nil {
		return fmt.Errorf("create order: %w", err)
	}
	return nil
}

// CreateWithTx 在指定事务中创建订单（订单创建必须与商品锁定同事务，避免跨连接死锁）。
func (r *orderRepo) CreateWithTx(tx *gorm.DB, order *model.Order) error {
	if err := tx.Create(order).Error; err != nil {
		return fmt.Errorf("create order with tx: %w", err)
	}
	return nil
}

func (r *orderRepo) GetByID(id uint) (*model.Order, error) {
	var o model.Order
	err := r.db.Preload("Product").Preload("Product.Seller").Preload("Buyer").Preload("Seller").Preload("Address").First(&o, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get order by id %d: %w", id, err)
	}
	return &o, nil
}

func (r *orderRepo) GetByIDForUpdate(tx *gorm.DB, id uint) (*model.Order, error) {
	var o model.Order
	err := tx.Clauses(clauseLocking()).First(&o, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get order by id %d for update: %w", id, err)
	}
	return &o, nil
}

func (r *orderRepo) ListByBuyer(buyerID uint, status string, page, pageSize int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64
	q := r.db.Model(&model.Order{}).Where("buyer_id = ?", buyerID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count buyer orders: %w", err)
	}
	if err := q.Preload("Product").Preload("Product.Seller").Preload("Address").Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&orders).Error; err != nil {
		return nil, 0, fmt.Errorf("list buyer orders: %w", err)
	}
	return orders, total, nil
}

func (r *orderRepo) ListBySeller(sellerID uint, status string, page, pageSize int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64
	q := r.db.Model(&model.Order{}).Where("seller_id = ?", sellerID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count seller orders: %w", err)
	}
	if err := q.Preload("Product").Preload("Buyer").Preload("Address").Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&orders).Error; err != nil {
		return nil, 0, fmt.Errorf("list seller orders: %w", err)
	}
	return orders, total, nil
}

func (r *orderRepo) UpdateStatusForUpdate(tx *gorm.DB, id uint, status string, updates map[string]interface{}) error {
	updates["status"] = status
	res := tx.Model(&model.Order{}).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		return fmt.Errorf("update order %d status to %s: %w", id, status, res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *orderRepo) Update(order *model.Order) error {
	if err := r.db.Save(order).Error; err != nil {
		return fmt.Errorf("update order %d: %w", order.ID, err)
	}
	return nil
}
