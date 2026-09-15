package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户实体：角色字段用于 JWT + RBAC 鉴权，信用积分由评价系统维护。
type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Username     string         `gorm:"size:32;uniqueIndex;not null" json:"username"`
	PasswordHash string         `gorm:"size:255;not null" json:"-"`
	Nickname     string         `gorm:"size:32;not null" json:"nickname"`
	Email        string         `gorm:"size:128" json:"email"`
	Phone        string         `gorm:"size:20" json:"phone"`
	Avatar       string         `gorm:"size:512" json:"avatar"`
	Role         string         `gorm:"size:16;not null;default:user" json:"role"`
	CreditScore  int            `gorm:"not null;default:100" json:"credit_score"`
	Status       string         `gorm:"size:16;not null;default:active" json:"status"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
