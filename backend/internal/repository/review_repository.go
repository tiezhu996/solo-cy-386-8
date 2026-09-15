package repository

import (
	"errors"
	"fmt"

	"github.com/marketpal/marketpal/internal/model"
	"gorm.io/gorm"
)

// ReviewRepository 评价仓储接口。
type ReviewRepository interface {
	Create(review *model.Review) error
	CreateWithTx(tx *gorm.DB, review *model.Review) error
	GetByOrderID(orderID uint) (*model.Review, error)
	ListByUser(userID uint, page, pageSize int) ([]model.Review, int64, error)
	ListByProduct(productID uint, page, pageSize int) ([]model.Review, int64, error)
}

type reviewRepo struct {
	db *gorm.DB
}

// NewReviewRepository 构造评价仓储。
func NewReviewRepository(db *gorm.DB) ReviewRepository {
	return &reviewRepo{db: db}
}

func (r *reviewRepo) Create(review *model.Review) error {
	if err := r.db.Create(review).Error; err != nil {
		return fmt.Errorf("create review: %w", err)
	}
	return nil
}

// CreateWithTx 在指定事务中创建评价（与信用积分更新同事务）。
func (r *reviewRepo) CreateWithTx(tx *gorm.DB, review *model.Review) error {
	if err := tx.Create(review).Error; err != nil {
		return fmt.Errorf("create review with tx: %w", err)
	}
	return nil
}

func (r *reviewRepo) GetByOrderID(orderID uint) (*model.Review, error) {
	var rev model.Review
	err := r.db.Where("order_id = ?", orderID).First(&rev).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get review by order %d: %w", orderID, err)
	}
	return &rev, nil
}

func (r *reviewRepo) ListByUser(userID uint, page, pageSize int) ([]model.Review, int64, error) {
	var list []model.Review
	var total int64
	q := r.db.Model(&model.Review{}).Where("reviewer_id = ? OR reviewee_id = ?", userID, userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count reviews: %w", err)
	}
	if err := q.Preload("Reviewer").Preload("Order").Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("list reviews: %w", err)
	}
	return list, total, nil
}

func (r *reviewRepo) ListByProduct(productID uint, page, pageSize int) ([]model.Review, int64, error) {
	var list []model.Review
	var total int64
	q := r.db.Model(&model.Review{}).Where("product_id = ?", productID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count product reviews: %w", err)
	}
	if err := q.Preload("Reviewer").Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("list product reviews: %w", err)
	}
	return list, total, nil
}
