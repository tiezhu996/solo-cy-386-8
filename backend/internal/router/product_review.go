package router

import (
	"github.com/gin-gonic/gin"
	"github.com/marketpal/marketpal/internal/handler"
	"github.com/marketpal/marketpal/internal/middleware"
)

// RegisterProductReviewRoutes 商品审核模块路由。
// 待审队列与审核决策仅管理员；单商品审核历史卖家本人或管理员可读（在 service 内鉴权）。
func RegisterProductReviewRoutes(api *gin.RouterGroup, h *handler.ProductReviewHandler, secret string) {
	admin := api.Group("/product-reviews", middleware.Auth(secret), middleware.RBAC("admin"))
	{
		admin.GET("", h.List)
		admin.POST("/:id/decision", h.Decide)
	}
	authed := api.Group("/products", middleware.Auth(secret))
	{
		authed.GET("/:id/reviews", h.History)
	}
}
