package dto

// MessageSendRequest 发送私信入参。
type MessageSendRequest struct {
	ReceiverID uint   `json:"receiver_id" binding:"required"`
	ProductID  uint   `json:"product_id" binding:"omitempty"`
	Content    string `json:"content" binding:"required,min=1,max=1000"`
}

// ConversationVO 会话视图对象。
type ConversationVO struct {
	PeerID      uint   `json:"peer_id"`
	PeerName    string `json:"peer_name"`
	PeerAvatar  string `json:"peer_avatar"`
	ProductID   uint   `json:"product_id,omitempty"`
	ProductName string `json:"product_name,omitempty"`
	LastContent string `json:"last_content"`
	LastTime    string `json:"last_time"`
	UnreadCount int    `json:"unread_count"`
}

// MessageVO 消息视图对象。
type MessageVO struct {
	ID         uint   `json:"id"`
	SenderID   uint   `json:"sender_id"`
	ReceiverID uint   `json:"receiver_id"`
	ProductID  uint   `json:"product_id"`
	Content    string `json:"content"`
	IsRead     bool   `json:"is_read"`
	CreatedAt  string `json:"created_at"`
}
