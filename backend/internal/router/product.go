package router

import (
	"github.com/gin-gonic/gin"
	"github.com/marketpal/marketpal/internal/handler"
	"github.com/marketpal/marketpal/internal/middleware"
)

// RegisterProductRoutes 商品模块路由。
func RegisterProductRoutes(api *gin.RouterGroup, h *handler.ProductHandler, secret string) {
	public := api.Group("/products")
	{
		public.GET("", h.List)
		public.GET("/:id", h.Detail)
	}
	authed := api.Group("/products", middleware.Auth(secret))
	{
		authed.POST("", h.Create)
		authed.PUT("/:id", h.Update)
		authed.POST("/:id/off-shelf", h.OffShelf)
		authed.POST("/:id/favorite", h.Favorite)
		authed.DELETE("/:id/favorite", h.Unfavorite)
		authed.GET("/mine", h.MyProducts)
	}
	favs := api.Group("/favorites", middleware.Auth(secret))
	{
		favs.GET("", h.Favorites)
	}
}
