package model

import "time"

// Message 站内私信实体：买卖双方就商品细节沟通，通过 WebSocket 实时推送。
type Message struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	SenderID   uint      `gorm:"not null;index" json:"sender_id"`
	ReceiverID uint      `gorm:"not null;index" json:"receiver_id"`
	ProductID  uint      `gorm:"not null;index" json:"product_id"`
	Content    string    `gorm:"size:1000;not null" json:"content"`
	IsRead     bool      `gorm:"not null;default:false" json:"is_read"`
	CreatedAt  time.Time `json:"created_at"`

	Sender   *User    `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	Receiver *User    `gorm:"foreignKey:ReceiverID" json:"receiver,omitempty"`
	Product  *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}
