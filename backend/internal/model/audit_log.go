package model

import "time"

// AuditLog 操作审计日志实体：由 audit 中间件与 service 埋点写入。
type AuditLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index" json:"user_id"`
	Username   string    `gorm:"size:32" json:"username"`
	Module     string    `gorm:"size:32;index" json:"module"`
	Action     string    `gorm:"size:64;index" json:"action"`
	ResourceID string    `gorm:"size:64" json:"resource_id"`
	Detail     string    `gorm:"type:text" json:"detail"`
	IP         string    `gorm:"size:64" json:"ip"`
	RequestID  string    `gorm:"size:64;index" json:"request_id"`
	CreatedAt  time.Time `json:"created_at"`
}
