package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/marketpal/marketpal/internal/middleware"
	"github.com/marketpal/marketpal/internal/service"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WSHandler WebSocket 实时消息处理器。
type WSHandler struct {
	hub    *service.Hub
	logger *slog.Logger
}

// NewWSHandler 构造 WebSocket 处理器。
func NewWSHandler(hub *service.Hub, logger *slog.Logger) *WSHandler {
	return &WSHandler{hub: hub, logger: logger}
}

// Handle GET /api/v1/ws 升级为 WebSocket 连接。
func (h *WSHandler) Handle(c *gin.Context) {
	userID := middleware.GetUserID(c)
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Warn("websocket upgrade failed", "user_id", userID, "err", err)
		return
	}
	client := &service.Client{
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 64),
	}
	h.hub.Register(client)
	defer h.hub.Unregister(client)

	// 读协程：处理 ping 与关闭。
	go func() {
		defer conn.Close()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()

	// 写协程：转发实时消息。
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case msg := <-client.Send:
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// parsePeerID 辅助解析（保留供扩展）。
func parsePeerID(s string) uint {
	n, _ := strconv.ParseUint(s, 10, 64)
	return uint(n)
}
