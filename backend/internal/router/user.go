package router

import (
	"github.com/gin-gonic/gin"
	"github.com/marketpal/marketpal/internal/handler"
	"github.com/marketpal/marketpal/internal/middleware"
)

// RegisterUserRoutes 用户模块路由。
func RegisterUserRoutes(api *gin.RouterGroup, h *handler.UserHandler, secret string) {
	auth := api.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
	}
	users := api.Group("/users", middleware.Auth(secret))
	{
		users.GET("/me", h.Profile)
		users.PUT("/me", h.UpdateProfile)
	}
	admin := api.Group("/users", middleware.Auth(secret), middleware.RBAC("admin"))
	{
		admin.GET("", h.ListUsers)
		admin.PUT("/:id/role", h.UpdateRole)
	}
}
