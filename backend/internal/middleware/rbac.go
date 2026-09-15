package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/marketpal/marketpal/internal/constants"
	"github.com/marketpal/marketpal/internal/util"
)

// RBAC 权限中间件：仅允许指定角色访问。
func RBAC(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get(ctxRole)
		roleStr, _ := role.(string)
		for _, r := range roles {
			if r == roleStr {
				c.Next()
				return
			}
		}
		util.Fail(c, 403, constants.CodeForbidden, "权限不足：当前角色 "+roleStr+" 无权访问该资源")
	}
}
