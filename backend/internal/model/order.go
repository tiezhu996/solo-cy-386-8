package model

import (
	"time"

	"gorm.io/gorm"
)

// Order 订单实体：状态机见 internal/constants/enums.go 与 internal/service/order_service.go。
type Order struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	OrderNo      string         `gorm:"size:40;uniqueIndex;not null" json:"order_no"`
	BuyerID      uint           `gorm:"not null;index" json:"buyer_id"`
	SellerID     uint           `gorm:"not null;index" json:"seller_id"`
	ProductID    uint           `gorm:"not null;index" json:"product_id"`
	AddressID    uint           `gorm:"not null" json:"address_id"`
	Quantity     int            `gorm:"not null;default:1" json:"quantity"`
	TotalPrice   float64        `gorm:"type:numeric(12,2);not null" json:"total_price"`
	Status       string         `gorm:"size:32;not null;default:pending_payment" json:"status"`
	Remark       string         `gorm:"size:255" json:"remark"`
	PaidAt       *time.Time     `json:"paid_at,omitempty"`
	ShippedAt    *time.Time     `json:"shipped_at,omitempty"`
	ReceivedAt   *time.Time     `json:"received_at,omitempty"`
	CompletedAt  *time.Time     `json:"completed_at,omitempty"`
	CancelledAt  *time.Time     `json:"cancelled_at,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	Buyer   *User    `gorm:"foreignKey:BuyerID" json:"buyer,omitempty"`
	Seller  *User    `gorm:"foreignKey:SellerID" json:"seller,omitempty"`
	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Address *Address `gorm:"foreignKey:AddressID" json:"address,omitempty"`
}
