package router

import (
	"github.com/gin-gonic/gin"
	"github.com/marketpal/marketpal/internal/handler"
	"github.com/marketpal/marketpal/internal/middleware"
)

// RegisterAddressRoutes 收货地址模块路由。
func RegisterAddressRoutes(api *gin.RouterGroup, h *handler.AddressHandler, secret string) {
	addr := api.Group("/addresses", middleware.Auth(secret))
	{
		addr.GET("", h.List)
		addr.POST("", h.Create)
		addr.PUT("/:id", h.Update)
		addr.DELETE("/:id", h.Delete)
	}
}
