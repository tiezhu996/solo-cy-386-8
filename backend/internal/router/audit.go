package router

import (
	"github.com/gin-gonic/gin"
	"github.com/marketpal/marketpal/internal/handler"
	"github.com/marketpal/marketpal/internal/middleware"
)

// RegisterAuditRoutes 审计日志模块路由（管理员）。
func RegisterAuditRoutes(api *gin.RouterGroup, h *handler.AuditHandler, secret string) {
	audits := api.Group("/audits", middleware.Auth(secret), middleware.RBAC("admin"))
	{
		audits.GET("", h.List)
	}
}
