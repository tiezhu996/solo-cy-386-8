package model

import "time"

// Favorite 收藏实体：user_id + product_id 唯一，防止重复收藏。
type Favorite struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;uniqueIndex:uk_fav_user_product" json:"user_id"`
	ProductID uint      `gorm:"not null;uniqueIndex:uk_fav_user_product" json:"product_id"`
	CreatedAt time.Time `json:"created_at"`

	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}
