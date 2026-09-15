package router

import (
	"github.com/gin-gonic/gin"
	"github.com/marketpal/marketpal/internal/handler"
	"github.com/marketpal/marketpal/internal/middleware"
)

// RegisterOrderRoutes 订单模块路由。
func RegisterOrderRoutes(api *gin.RouterGroup, h *handler.OrderHandler, secret string) {
	orders := api.Group("/orders", middleware.Auth(secret))
	{
		orders.GET("", h.List)
		orders.POST("", h.Create)
		orders.GET("/:id", h.Detail)
		orders.POST("/:id/pay", h.Pay)
		orders.POST("/:id/ship", h.Ship)
		orders.POST("/:id/receive", h.Receive)
		orders.POST("/:id/complete", h.Complete)
		orders.POST("/:id/cancel", h.Cancel)
	}
	// 卖家取消（同 Cancel service 方法，role=seller）。
	seller := api.Group("/seller", middleware.Auth(secret))
	{
		seller.POST("/orders/:id/cancel", h.CancelBySeller)
	}
}
