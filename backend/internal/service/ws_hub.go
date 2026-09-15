package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// WSMessage 通过 WebSocket 推送给前端的消息载荷。
type WSMessage struct {
	Type      string `json:"type"`
	MessageID uint   `json:"message_id,omitempty"`
	SenderID  uint   `json:"sender_id"`
	ReceiverID uint  `json:"receiver_id"`
	ProductID uint   `json:"product_id,omitempty"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	Unread    int64  `json:"unread,omitempty"`
}

// Client 单个 WebSocket 连接。
type Client struct {
	UserID uint
	Conn   *websocket.Conn
	Send   chan []byte
}

// Hub 管理全部 WebSocket 连接，并通过 Redis pub/sub 跨进程广播。
type Hub struct {
	mu      sync.RWMutex
	clients map[uint]map[*Client]bool
	rdb     *redis.Client
	logger  *slog.Logger
}

// NewHub 构造 Hub。
func NewHub(rdb *redis.Client, logger *slog.Logger) *Hub {
	return &Hub{
		clients: make(map[uint]map[*Client]bool),
		rdb:     rdb,
		logger:  logger,
	}
}

// Register 注册连接并启动该用户的 Redis 订阅转发。
func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	if h.clients[client.UserID] == nil {
		h.clients[client.UserID] = make(map[*Client]bool)
	}
	h.clients[client.UserID][client] = true
	h.mu.Unlock()
	h.logger.Info("websocket connected", "user_id", client.UserID)
	go h.subscribe(client.UserID)
}

// Unregister 注销连接。
func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	if m, ok := h.clients[client.UserID]; ok {
		delete(m, client)
		if len(m) == 0 {
			delete(h.clients, client.UserID)
		}
	}
	h.mu.Unlock()
	h.logger.Info("websocket disconnected", "user_id", client.UserID)
}

// BroadcastToUser 向指定用户的所有在线连接推送消息。
func (h *Hub) BroadcastToUser(userID uint, payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients[userID] {
		select {
		case c.Send <- payload:
		default:
			// 发送缓冲满则丢弃，等待客户端重新拉取。
		}
	}
}

// Publish 发布消息到 Redis 通道，供该用户的所有订阅实例转发。
func (h *Hub) Publish(userID uint, payload []byte) {
	if h.rdb == nil {
		h.BroadcastToUser(userID, payload)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := h.rdb.Publish(ctx, channelOf(userID), payload).Err(); err != nil {
		h.logger.Warn("websocket publish failed", "receiver_id", userID, "err", err)
		h.BroadcastToUser(userID, payload)
	}
}

func (h *Hub) subscribe(userID uint) {
	if h.rdb == nil {
		return
	}
	ctx := context.Background()
	sub := h.rdb.Subscribe(ctx, channelOf(userID))
	defer sub.Close()
	ch := sub.Channel()
	for msg := range ch {
		h.BroadcastToUser(userID, []byte(msg.Payload))
	}
}

func channelOf(userID uint) string {
	return "marketpal:msg:user:" + strconv.FormatUint(uint64(userID), 10)
}

// MarshalWSMessage 序列化 WebSocket 消息。
func MarshalWSMessage(m *WSMessage) ([]byte, error) {
	return json.Marshal(m)
}
