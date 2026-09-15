package repository

import (
	"fmt"

	"github.com/marketpal/marketpal/internal/model"
	"gorm.io/gorm"
)

// AuditRepository 审计日志仓储接口。
type AuditRepository interface {
	Create(log *model.AuditLog) error
	List(query map[string]interface{}, page, pageSize int) ([]model.AuditLog, int64, error)
}

type auditRepo struct {
	db *gorm.DB
}

// NewAuditRepository 构造审计日志仓储。
func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &auditRepo{db: db}
}

func (r *auditRepo) Create(log *model.AuditLog) error {
	if err := r.db.Create(log).Error; err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}
	return nil
}

func (r *auditRepo) List(query map[string]interface{}, page, pageSize int) ([]model.AuditLog, int64, error) {
	var list []model.AuditLog
	var total int64
	q := r.db.Model(&model.AuditLog{})
	if v, ok := query["module"]; ok && v != "" {
		q = q.Where("module = ?", v)
	}
	if v, ok := query["action"]; ok && v != "" {
		q = q.Where("action = ?", v)
	}
	if v, ok := query["user_id"]; ok && v != "" {
		q = q.Where("user_id = ?", v)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count audit logs: %w", err)
	}
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("list audit logs: %w", err)
	}
	return list, total, nil
}
