package service

import (
	"fmt"
	"log/slog"

	"github.com/marketpal/marketpal/internal/constants"
	"github.com/marketpal/marketpal/internal/dto"
	"github.com/marketpal/marketpal/internal/model"
	"github.com/marketpal/marketpal/internal/repository"
	"github.com/marketpal/marketpal/internal/util"
)

// AuditService 操作审计日志业务服务。
type AuditService struct {
	repo   repository.AuditRepository
	logger *slog.Logger
}

// NewAuditService 构造审计服务。
func NewAuditService(repo repository.AuditRepository, logger *slog.Logger) *AuditService {
	return &AuditService{repo: repo, logger: logger}
}

// Create 写入审计日志（middleware 与 service 埋点复用）。
func (s *AuditService) Create(userID uint, username, module, action, resourceID, detail, ip, requestID string) error {
	entry := &model.AuditLog{
		UserID:     userID,
		Username:   username,
		Module:     module,
		Action:     action,
		ResourceID: resourceID,
		Detail:     detail,
		IP:         ip,
		RequestID:  requestID,
	}
	if err := s.repo.Create(entry); err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}
	s.logger.Info(constants.LogAuditCreated, "user_id", userID, "module", module, "action", action)
	return nil
}

// List 分页查询审计日志（管理员）。
func (s *AuditService) List(q dto.AuditQuery) (*dto.AuditListResponse, error) {
	page, pageSize := q.Page, q.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	query := map[string]interface{}{
		"module":  q.Module,
		"action":  q.Action,
		"user_id": q.UserID,
	}
	list, total, err := s.repo.List(query, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("list audit logs: %w", err)
	}
	res := &dto.AuditListResponse{List: []dto.AuditVO{}, Total: total, Page: page, Size: pageSize}
	for i := range list {
		res.List = append(res.List, dto.AuditVO{
			ID:         list[i].ID,
			UserID:     list[i].UserID,
			Username:   list[i].Username,
			Module:     list[i].Module,
			Action:     list[i].Action,
			ResourceID: list[i].ResourceID,
			Detail:     list[i].Detail,
			IP:         list[i].IP,
			RequestID:  list[i].RequestID,
			CreatedAt:  util.FormatTime(list[i].CreatedAt),
		})
	}
	return res, nil
}
