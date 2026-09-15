package model

import "time"

// Review 评价实体：交易完成后双方互评，评价影响用户信用积分。
// (order_id, reviewer_id) 联合唯一，保证同一订单同一用户只能评价一次。
type Review struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	OrderID    uint      `gorm:"not null;uniqueIndex:uk_review_order_reviewer" json:"order_id"`
	ProductID  uint      `gorm:"not null;index" json:"product_id"`
	ReviewerID uint      `gorm:"not null;uniqueIndex:uk_review_order_reviewer" json:"reviewer_id"`
	RevieweeID uint      `gorm:"not null;index" json:"reviewee_id"`
	Rating     string    `gorm:"size:16;not null" json:"rating"` // good/neutral/bad
	Content    string    `gorm:"size:500" json:"content"`
	CreatedAt  time.Time `json:"created_at"`

	Order    *Order `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Reviewer *User  `gorm:"foreignKey:ReviewerID" json:"reviewer,omitempty"`
	Reviewee *User  `gorm:"foreignKey:RevieweeID" json:"reviewee,omitempty"`
}
