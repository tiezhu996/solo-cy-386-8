package model

import "time"

// ProductReview 商品审核记录实体：每次提交审核生成一条，管理员只能对待审核记录给出一次结果。
// 审核枚举见 internal/constants/enums.go（ProductReview*）。
type ProductReview struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	ProductID  uint       `gorm:"not null;index;uniqueIndex:uk_product_round" json:"product_id"`
	SellerID   uint       `gorm:"not null;index" json:"seller_id"`
	Round      int        `gorm:"not null;default:1;uniqueIndex:uk_product_round" json:"round"` // 第几轮送审
	Status     string     `gorm:"size:32;not null;default:pending_review;index" json:"status"`  // pending_review/approved/rejected
	Reason     string     `gorm:"type:text" json:"reason"`                                      // 驳回原因（通过时为空）
	ReviewerID uint       `gorm:"index" json:"reviewer_id"`                                     // 审核管理员 ID（待审核时为 0）
	Reviewer   *User      `gorm:"foreignKey:ReviewerID" json:"reviewer,omitempty"`
	Seller     *User      `gorm:"foreignKey:SellerID" json:"seller,omitempty"`
	Product    *Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	ReviewedAt *time.Time `json:"reviewed_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
