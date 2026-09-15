package router

import (
	"github.com/gin-gonic/gin"
	"github.com/marketpal/marketpal/internal/handler"
	"github.com/marketpal/marketpal/internal/middleware"
)

// RegisterReviewRoutes 评价模块路由。
func RegisterReviewRoutes(api *gin.RouterGroup, h *handler.ReviewHandler, secret string) {
	api.GET("/reviews", h.List)
	authed := api.Group("/reviews", middleware.Auth(secret))
	{
		authed.POST("", h.Create)
	}
}
