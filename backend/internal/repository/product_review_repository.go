package repository

import (
	"errors"
	"fmt"

	"github.com/marketpal/marketpal/internal/model"
	"gorm.io/gorm"
)

// ErrReviewAlreadyDecided 审核记录已被其他管理员处理（并发审核幂等哨兵错误）。
var ErrReviewAlreadyDecided = errors.New("product review already decided")

// ProductReviewRepository 商品审核记录仓储接口。
type ProductReviewRepository interface {
	CreateTx(tx *gorm.DB, review *model.ProductReview) error
	GetByID(id uint) (*model.ProductReview, error)
	GetByIDForUpdate(tx *gorm.DB, id uint) (*model.ProductReview, error)
	GetLatestByProduct(productID uint) (*model.ProductReview, error)
	GetPendingByProductForUpdate(tx *gorm.DB, productID uint) (*model.ProductReview, error)
	MaxRound(tx *gorm.DB, productID uint) (int, error)
	// DecideForUpdate 条件更新：仅当记录仍为 pending_review 时写入结果。
	// 并发审核只有一个事务 RowsAffected=1，从 SQL 层保证“同一商品并发审核只产生一个结果”。
	DecideForUpdate(tx *gorm.DB, id uint, reviewerID *uint, status, reason string, reviewedAt interface{}) error
	List(query map[string]interface{}, page, pageSize int) ([]model.ProductReview, int64, error)
	ListByProduct(productID uint) ([]model.ProductReview, error)
}

type productReviewRepo struct {
	db *gorm.DB
}

// NewProductReviewRepository 构造商品审核仓储。
func NewProductReviewRepository(db *gorm.DB) ProductReviewRepository {
	return &productReviewRepo{db: db}
}

func (r *productReviewRepo) CreateTx(tx *gorm.DB, review *model.ProductReview) error {
	if tx == nil {
		tx = r.db
	}
	if err := tx.Create(review).Error; err != nil {
		return fmt.Errorf("create product review product=%d round=%d: %w", review.ProductID, review.Round, err)
	}
	return nil
}

func (r *productReviewRepo) GetByID(id uint) (*model.ProductReview, error) {
	var rv model.ProductReview
	err := r.db.Preload("Product").Preload("Reviewer").First(&rv, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get product review %d: %w", id, err)
	}
	return &rv, nil
}

func (r *productReviewRepo) GetByIDForUpdate(tx *gorm.DB, id uint) (*model.ProductReview, error) {
	var rv model.ProductReview
	err := tx.Clauses(clauseLocking()).First(&rv, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get product review %d for update: %w", id, err)
	}
	return &rv, nil
}

func (r *productReviewRepo) GetLatestByProduct(productID uint) (*model.ProductReview, error) {
	var rv model.ProductReview
	err := r.db.Where("product_id = ?", productID).Order("round DESC").First(&rv).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get latest review of product %d: %w", productID, err)
	}
	return &rv, nil
}

func (r *productReviewRepo) GetPendingByProductForUpdate(tx *gorm.DB, productID uint) (*model.ProductReview, error) {
	var rv model.ProductReview
	err := tx.Clauses(clauseLocking()).
		Where("product_id = ? AND status = ?", productID, "pending_review").
		Order("round DESC").First(&rv).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get pending review of product %d for update: %w", productID, err)
	}
	return &rv, nil
}

func (r *productReviewRepo) MaxRound(tx *gorm.DB, productID uint) (int, error) {
	if tx == nil {
		tx = r.db
	}
	var maxRound *int
	if err := tx.Model(&model.ProductReview{}).Where("product_id = ?", productID).
		Select("MAX(round)").Scan(&maxRound).Error; err != nil {
		return 0, fmt.Errorf("max review round of product %d: %w", productID, err)
	}
	if maxRound == nil {
		return 0, nil
	}
	return *maxRound, nil
}

func (r *productReviewRepo) DecideForUpdate(tx *gorm.DB, id uint, reviewerID *uint, status, reason string, reviewedAt interface{}) error {
	if tx == nil {
		tx = r.db
	}
	res := tx.Model(&model.ProductReview{}).
		Where("id = ? AND status = ?", id, "pending_review").
		Updates(map[string]interface{}{
			"status":      status,
			"reason":      reason,
			"reviewer_id": reviewerID, // *uint：裁决写真实管理员 ID，不裁决的路径根本不会调用本方法
			"reviewed_at": reviewedAt,
		})
	if res.Error != nil {
		return fmt.Errorf("decide product review %d -> %s: %w", id, status, res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrReviewAlreadyDecided
	}
	return nil
}

func (r *productReviewRepo) List(query map[string]interface{}, page, pageSize int) ([]model.ProductReview, int64, error) {
	var list []model.ProductReview
	var total int64
	q := r.db.Model(&model.ProductReview{})
	if v, ok := query["status"]; ok && v != "" {
		q = q.Where("product_reviews.status = ?", v)
	}
	if v, ok := query["product_id"]; ok && numVal(v) > 0 {
		q = q.Where("product_reviews.product_id = ?", v)
	}
	if v, ok := query["seller_id"]; ok && numVal(v) > 0 {
		q = q.Where("product_reviews.seller_id = ?", v)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count product reviews: %w", err)
	}
	if err := q.Preload("Product").Preload("Product.Seller").Preload("Reviewer").
		Order("product_reviews.id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("list product reviews: %w", err)
	}
	return list, total, nil
}

func (r *productReviewRepo) ListByProduct(productID uint) ([]model.ProductReview, error) {
	var list []model.ProductReview
	if err := r.db.Preload("Reviewer").Preload("Product").
		Where("product_id = ?", productID).Order("round DESC").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("list reviews of product %d: %w", productID, err)
	}
	return list, nil
}
