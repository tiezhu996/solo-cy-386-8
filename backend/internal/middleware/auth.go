package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/marketpal/marketpal/internal/constants"
	"github.com/marketpal/marketpal/internal/util"
)

const (
	ctxUserID   = "auth_user_id"
	ctxUsername = "auth_username"
	ctxRole     = "auth_role"
)

// Auth JWT 认证中间件。
func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		tokenStr := c.Query("token")
		if header != "" && strings.HasPrefix(header, "Bearer ") {
			tokenStr = strings.TrimPrefix(header, "Bearer ")
		}
		if tokenStr == "" {
			util.Fail(c, 401, constants.CodeUnauthorized, "认证失败：缺少 Bearer Token")
			return
		}
		claims, err := util.ParseToken(secret, tokenStr)
		if err != nil {
			util.Fail(c, 401, constants.CodeUnauthorized, "认证失败：Token 无效或已过期")
			return
		}
		c.Set(ctxUserID, claims.UserID)
		c.Set(ctxUsername, claims.Username)
		c.Set(ctxRole, claims.Role)
		c.Next()
	}
}

// OptionalAuth 可选认证：携带合法 Token 时注入用户上下文，未携带或 Token 非法也放行（用于商品详情等公开页）。
func OptionalAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		tokenStr := c.Query("token")
		if header != "" && strings.HasPrefix(header, "Bearer ") {
			tokenStr = strings.TrimPrefix(header, "Bearer ")
		}
		if tokenStr != "" {
			if claims, err := util.ParseToken(secret, tokenStr); err == nil {
				c.Set(ctxUserID, claims.UserID)
				c.Set(ctxUsername, claims.Username)
				c.Set(ctxRole, claims.Role)
			}
		}
		c.Next()
	}
}
