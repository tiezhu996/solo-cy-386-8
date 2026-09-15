package model

import (
	"time"

	"gorm.io/gorm"
)

// Product 商品实体：成色/分类/状态枚举见 internal/constants/enums.go。
type Product struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	SellerID      uint           `gorm:"not null;index" json:"seller_id"`
	Title         string         `gorm:"size:128;not null" json:"title"`
	Description   string         `gorm:"type:text;not null" json:"description"`
	OriginalPrice float64        `gorm:"type:numeric(12,2);not null" json:"original_price"`
	Price         float64        `gorm:"type:numeric(12,2);not null" json:"price"`
	Condition     string         `gorm:"size:32;not null;default:almost_new" json:"condition"`
	Category      string         `gorm:"size:32;not null;default:other" json:"category"`
	Images        string         `gorm:"type:text" json:"images"` // 逗号分隔的图片 URL 列表
	Status        string         `gorm:"size:32;not null;default:on_sale" json:"status"`
	ViewCount     int            `gorm:"not null;default:0" json:"view_count"`
	FavoriteCount int            `gorm:"not null;default:0" json:"favorite_count"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	Seller *User `gorm:"foreignKey:SellerID" json:"seller,omitempty"`
}
