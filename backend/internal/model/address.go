package model

import "time"

// Address 收货地址实体。
type Address struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"not null;index" json:"user_id"`
	ReceiverName string    `gorm:"size:32;not null" json:"receiver_name"`
	Phone        string    `gorm:"size:20;not null" json:"phone"`
	Province     string    `gorm:"size:32;not null" json:"province"`
	City         string    `gorm:"size:32;not null" json:"city"`
	District     string    `gorm:"size:32" json:"district"`
	Detail       string    `gorm:"size:255;not null" json:"detail"`
	IsDefault    bool      `gorm:"not null;default:false" json:"is_default"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
