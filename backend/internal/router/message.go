package router

import (
	"github.com/gin-gonic/gin"
	"github.com/marketpal/marketpal/internal/handler"
	"github.com/marketpal/marketpal/internal/middleware"
)

// RegisterMessageRoutes 私信模块路由（含 WebSocket）。
func RegisterMessageRoutes(api *gin.RouterGroup, h *handler.MessageHandler, ws *handler.WSHandler, secret string) {
	msgs := api.Group("/messages", middleware.Auth(secret))
	{
		msgs.POST("", h.Send)
		msgs.GET("/conversations", h.Conversations)
		msgs.GET("/conversations/:peerId", h.Conversation)
		msgs.PUT("/conversations/:peerId/read", h.MarkRead)
		msgs.GET("/unread-count", h.UnreadCount)
	}
	api.GET("/ws", middleware.Auth(secret), ws.Handle)
}
