package router

import (
	"github.com/gin-gonic/gin"
	"github.com/marketpal/marketpal/internal/handler"
	"github.com/marketpal/marketpal/internal/middleware"
)

// RegisterUploadRoutes 图片上传路由。
func RegisterUploadRoutes(api *gin.RouterGroup, h *handler.UploadHandler, secret string) {
	upload := api.Group("/upload", middleware.Auth(secret))
	{
		upload.POST("", h.Upload)
	}
}
