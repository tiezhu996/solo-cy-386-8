package model

import (
	"time"

	"gorm.io/gorm"
)

// Product 商品实体：成色/分类/状态枚举见 internal/constants/enums.go。
type Product struct {
	ID            uint    `gorm:"primaryKey" json:"id"`
	SellerID      uint    `gorm:"not null;index" json:"seller_id"`
	Title         string  `gorm:"size:128;not null" json:"title"`
	Description   string  `gorm:"type:text;not null" json:"description"`
	OriginalPrice float64 `gorm:"type:numeric(12,2);not null" json:"original_price"`
	Price         float64 `gorm:"type:numeric(12,2);not null" json:"price"`
	Condition     string  `gorm:"size:32;not null;default:almost_new" json:"condition"`
	Category      string  `gorm:"size:32;not null;default:other" json:"category"`
	Images        string  `gorm:"type:text" json:"images"` // 逗号分隔的图片 URL 列表
	Status        string  `gorm:"size:32;not null;default:on_sale" json:"status"`
	// ReviewStatus 平台审核状态：pending_review/approved/rejected（枚举见 constants/enums.go）。
	// 新发布与修改重提期间恒为 pending_review 且 Status=off_shelf；只有 approved 才可能在售。
	ReviewStatus  string         `gorm:"size:32;not null;default:pending_review;index" json:"review_status"`
	ReviewRound   int            `gorm:"not null;default:1" json:"review_round"` // 审核轮次：首次发布=1，每次重提 +1
	RejectReason  string         `gorm:"type:text" json:"reject_reason"`         // 最近一次驳回原因（通过后清空）
	ViewCount     int            `gorm:"not null;default:0" json:"view_count"`
	FavoriteCount int            `gorm:"not null;default:0" json:"favorite_count"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	Seller *User `gorm:"foreignKey:SellerID" json:"seller,omitempty"`
}
