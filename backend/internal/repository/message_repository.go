package repository

import (
	"errors"
	"fmt"

	"github.com/marketpal/marketpal/internal/model"
	"gorm.io/gorm"
)

// MessageRepository 私信仓储接口。
type MessageRepository interface {
	Create(msg *model.Message) error
	ListConversation(userID, peerID uint, page, pageSize int) ([]model.Message, int64, error)
	ListPeers(userID uint) ([]model.Message, error)
	MarkRead(userID, peerID uint) (int64, error)
	UnreadCount(userID uint) (int64, error)
}

type messageRepo struct {
	db *gorm.DB
}

// NewMessageRepository 构造私信仓储。
func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepo{db: db}
}

func (r *messageRepo) Create(msg *model.Message) error {
	if err := r.db.Create(msg).Error; err != nil {
		return fmt.Errorf("create message: %w", err)
	}
	return nil
}

func (r *messageRepo) ListConversation(userID, peerID uint, page, pageSize int) ([]model.Message, int64, error) {
	var msgs []model.Message
	var total int64
	q := r.db.Model(&model.Message{}).
		Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)", userID, peerID, peerID, userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count conversation: %w", err)
	}
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&msgs).Error; err != nil {
		return nil, 0, fmt.Errorf("list conversation: %w", err)
	}
	return msgs, total, nil
}

func (r *messageRepo) ListPeers(userID uint) ([]model.Message, error) {
	// 取出与我相关的全部消息（按 id 倒序），由 service 聚合成会话列表。
	var msgs []model.Message
	if err := r.db.Preload("Product").Preload("Sender").Preload("Receiver").
		Where("sender_id = ? OR receiver_id = ?", userID, userID).
		Order("id DESC").Find(&msgs).Error; err != nil {
		return nil, fmt.Errorf("list peer messages user=%d: %w", userID, err)
	}
	return msgs, nil
}

func (r *messageRepo) MarkRead(userID, peerID uint) (int64, error) {
	res := r.db.Model(&model.Message{}).
		Where("receiver_id = ? AND sender_id = ? AND is_read = ?", userID, peerID, false).
		Update("is_read", true)
	if res.Error != nil {
		return 0, fmt.Errorf("mark messages read user=%d peer=%d: %w", userID, peerID, res.Error)
	}
	return res.RowsAffected, nil
}

func (r *messageRepo) UnreadCount(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Message{}).Where("receiver_id = ? AND is_read = ?", userID, false).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("count unread messages user=%d: %w", userID, err)
	}
	return count, nil
}

// ErrMessageNotFound 兼容哨兵错误。
var ErrMessageNotFound = errors.New("message not found")
