package database

import (
	"fmt"
	"time"

	"github.com/marketpal/marketpal/internal/constants"
	"github.com/marketpal/marketpal/internal/model"
	"gorm.io/gorm"
)

// backfillLegacyProducts 审核功能上线时一次性回填存量商品：
// 存量商品（含此前已上架/已售出/已下架）一律视为审核通过，保持原有售卖状态不受影响；
// 并补写一轮 approved 审核记录，保证审核历史可回读。幂等：已有审核记录的商品跳过。
func backfillLegacyProducts(db *gorm.DB) error {
	var legacy []model.Product
	if err := db.Where("review_status = ?", constants.ProductReviewPending).Find(&legacy).Error; err != nil {
		return fmt.Errorf("query legacy products: %w", err)
	}
	if len(legacy) == 0 {
		return nil
	}
	now := time.Now()
	return db.Transaction(func(tx *gorm.DB) error {
		for i := range legacy {
			p := legacy[i]
			var count int64
			if err := tx.Model(&model.ProductReview{}).Where("product_id = ?", p.ID).Count(&count).Error; err != nil {
				return fmt.Errorf("count reviews of product %d: %w", p.ID, err)
			}
			if count > 0 {
				continue
			}
			if err := tx.Model(&model.Product{}).Where("id = ?", p.ID).
				Updates(map[string]interface{}{
					"review_status": constants.ProductReviewApproved,
					"review_round":  1,
					"reject_reason": "",
				}).Error; err != nil {
				return fmt.Errorf("approve legacy product %d: %w", p.ID, err)
			}
			rec := model.ProductReview{
				ProductID:  p.ID,
				SellerID:   p.SellerID,
				Round:      1,
				Status:     constants.ProductReviewApproved,
				Reason:     "存量商品，审核功能上线前发布，默认通过",
				ReviewedAt: &now,
				CreatedAt:  now,
				UpdatedAt:  now,
			}
			if err := tx.Create(&rec).Error; err != nil {
				return fmt.Errorf("seed review record of product %d: %w", p.ID, err)
			}
		}
		return nil
	})
}
