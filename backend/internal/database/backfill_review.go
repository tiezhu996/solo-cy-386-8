package database

import (
	"fmt"
	"time"

	"github.com/marketpal/marketpal/internal/constants"
	"github.com/marketpal/marketpal/internal/model"
	"gorm.io/gorm"
)

// ensureProductReviewSchema 审核表上线后的结构修正（幂等）：
// 待审核记录没有审核人，reviewer_id 必须可空。旧版本 AutoMigrate 可能已把列建成 NOT NULL
// （GORM 不会在后续迁移中收放非空约束），这里显式放开；并显式建立到 users 的外键（已存在时跳过）。
func ensureProductReviewSchema(db *gorm.DB) error {
	if !db.Migrator().HasTable(&model.ProductReview{}) {
		return nil
	}
	if err := db.Exec("ALTER TABLE product_reviews ALTER COLUMN reviewer_id DROP NOT NULL").Error; err != nil {
		return fmt.Errorf("drop not null on product_reviews.reviewer_id: %w", err)
	}
	// ON DELETE SET NULL：管理员账号注销不影响审核留痕。
	if err := db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'fk_product_reviews_reviewer'
			) THEN
				ALTER TABLE product_reviews
					ADD CONSTRAINT fk_product_reviews_reviewer
					FOREIGN KEY (reviewer_id) REFERENCES users(id) ON DELETE SET NULL;
			END IF;
		END$$;`).Error; err != nil {
		return fmt.Errorf("add fk product_reviews.reviewer_id: %w", err)
	}
	return nil
}

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
