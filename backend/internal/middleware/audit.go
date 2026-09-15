package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/marketpal/marketpal/internal/service"
)

// Audit 操作审计中间件：请求完成后异步写入审计日志。
func Audit(auditService *service.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		// 仅记录写操作，避免日志膨胀。
		if c.Request.Method == "GET" || c.Request.Method == "OPTIONS" {
			return
		}
		userID, _ := c.Get(ctxUserID)
		username, _ := c.Get(ctxUsername)
		requestID, _ := c.Get("request_id")
		uid, _ := userID.(uint)
		uname, _ := username.(string)
		rid, _ := requestID.(string)
		_ = auditService.Create(uid, uname, c.Request.Method, c.FullPath(), "", c.Request.URL.Path, c.ClientIP(), rid)
	}
}
