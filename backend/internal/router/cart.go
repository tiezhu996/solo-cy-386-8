package router

import (
	"github.com/gin-gonic/gin"
	"github.com/marketpal/marketpal/internal/handler"
	"github.com/marketpal/marketpal/internal/middleware"
)

// RegisterCartRoutes 购物车模块路由。
func RegisterCartRoutes(api *gin.RouterGroup, h *handler.CartHandler, secret string) {
	cart := api.Group("/cart", middleware.Auth(secret))
	{
		cart.GET("", h.List)
		cart.POST("", h.Add)
		cart.PUT("/:id", h.Update)
		cart.DELETE("/:id", h.Delete)
	}
}
