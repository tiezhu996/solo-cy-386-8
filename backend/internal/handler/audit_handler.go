package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/marketpal/marketpal/internal/dto"
	"github.com/marketpal/marketpal/internal/service"
	"github.com/marketpal/marketpal/internal/util"
)

// AuditHandler 审计日志 HTTP 处理器。
type AuditHandler struct {
	svc *service.AuditService
}

// NewAuditHandler 构造审计处理器。
func NewAuditHandler(svc *service.AuditService) *AuditHandler {
	return &AuditHandler{svc: svc}
}

// List GET /api/v1/audits (admin)
func (h *AuditHandler) List(c *gin.Context) {
	var q dto.AuditQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		util.Fail(c, 400, 40000, "审计日志查询失败：查询参数不合法 "+err.Error())
		return
	}
	res, err := h.svc.List(q)
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OK(c, res)
}
