package middleware

import "github.com/gin-gonic/gin"

// GetUserID 从上下文读取当前登录用户 ID。
func GetUserID(c *gin.Context) uint {
	v, _ := c.Get(ctxUserID)
	id, _ := v.(uint)
	return id
}

// GetUsername 从上下文读取当前登录用户名。
func GetUsername(c *gin.Context) string {
	v, _ := c.Get(ctxUsername)
	s, _ := v.(string)
	return s
}

// GetRole 从上下文读取当前登录用户角色。
func GetRole(c *gin.Context) string {
	v, _ := c.Get(ctxRole)
	s, _ := v.(string)
	return s
}
