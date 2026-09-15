package service

import (
	"fmt"
	"log/slog"

	"github.com/marketpal/marketpal/internal/constants"
	"github.com/marketpal/marketpal/internal/dto"
	"github.com/marketpal/marketpal/internal/model"
	"github.com/marketpal/marketpal/internal/repository"
	"github.com/marketpal/marketpal/internal/util"
)

// MessageService 站内私信业务服务（发送后经 Hub 实时推送）。
type MessageService struct {
	repo    repository.MessageRepository
	hub     *Hub
	logger  *slog.Logger
}

// NewMessageService 构造私信服务。
func NewMessageService(repo repository.MessageRepository, hub *Hub, logger *slog.Logger) *MessageService {
	return &MessageService{repo: repo, hub: hub, logger: logger}
}

// Send 发送私信：落库 → Redis 发布 → 返回消息。
func (s *MessageService) Send(senderID uint, req dto.MessageSendRequest) (*model.Message, error) {
	if senderID == req.ReceiverID {
		return nil, utilAppError(constants.CodeBadRequest, "私信发送失败：不能给自己（用户 id="+fmt.Sprint(senderID)+"）发送消息", nil)
	}
	msg := &model.Message{
		SenderID:   senderID,
		ReceiverID: req.ReceiverID,
		ProductID:  req.ProductID,
		Content:    req.Content,
	}
	if err := s.repo.Create(msg); err != nil {
		return nil, fmt.Errorf("create message sender=%d receiver=%d: %w", senderID, req.ReceiverID, err)
	}
	payload, err := MarshalWSMessage(&WSMessage{
		Type:       "new_message",
		MessageID:  msg.ID,
		SenderID:   senderID,
		ReceiverID: req.ReceiverID,
		ProductID:  req.ProductID,
		Content:    req.Content,
		CreatedAt:  util.FormatTime(msg.CreatedAt),
	})
	if err != nil {
		s.logger.Warn("marshal ws message failed", "message_id", msg.ID, "err", err)
	} else {
		s.hub.Publish(req.ReceiverID, payload)
	}
	s.logger.Info(constants.LogMessageSent, "message_id", msg.ID, "sender_id", senderID, "receiver_id", req.ReceiverID, "product_id", req.ProductID)
	return msg, nil
}

// ListConversation 会话消息列表。
func (s *MessageService) ListConversation(userID, peerID uint, page, pageSize int) ([]model.Message, int64, error) {
	p := util.NormalizePage(page, pageSize)
	msgs, total, err := s.repo.ListConversation(userID, peerID, p.Page, p.PageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("list conversation user=%d peer=%d: %w", userID, peerID, err)
	}
	return msgs, total, nil
}

// ListConversations 会话列表（聚合最近一条消息与未读数）。
func (s *MessageService) ListConversations(userID uint) ([]dto.ConversationVO, error) {
	msgs, err := s.repo.ListPeers(userID)
	if err != nil {
		return nil, fmt.Errorf("list peers user=%d: %w", userID, err)
	}
	conversations := map[uint]*dto.ConversationVO{}
	var order []uint
	for _, m := range msgs {
		peerID := m.SenderID
		if m.SenderID == userID {
			peerID = m.ReceiverID
		}
		if _, ok := conversations[peerID]; !ok {
			peerName := fmt.Sprintf("用户%d", peerID)
			if m.Sender != nil && m.Sender.ID != userID {
				peerName = m.Sender.Nickname
			}
			if m.Receiver != nil && m.Receiver.ID != userID {
				peerName = m.Receiver.Nickname
			}
			conversations[peerID] = &dto.ConversationVO{
				PeerID:    peerID,
				PeerName:  peerName,
				LastContent: m.Content,
				LastTime:  util.FormatTime(m.CreatedAt),
			}
			if m.ProductID > 0 {
				conversations[peerID].ProductID = m.ProductID
				if m.Product != nil {
					conversations[peerID].ProductName = m.Product.Title
				}
			}
			order = append(order, peerID)
		}
	}
	// 逐会话未读统计。
	for _, m := range msgs {
		if m.ReceiverID == userID && !m.IsRead {
			peerID := m.SenderID
			conversations[peerID].UnreadCount++
		}
	}
	result := make([]dto.ConversationVO, 0, len(order))
	for _, peerID := range order {
		result = append(result, *conversations[peerID])
	}
	return result, nil
}

// MarkRead 标记与某用户的会话为已读。
func (s *MessageService) MarkRead(userID, peerID uint) error {
	count, err := s.repo.MarkRead(userID, peerID)
	if err != nil {
		return fmt.Errorf("mark read user=%d peer=%d: %w", userID, peerID, err)
	}
	s.logger.Info(constants.LogMessageRead, "user_id", userID, "count", count)
	return nil
}

// UnreadCount 未读消息总数。
func (s *MessageService) UnreadCount(userID uint) (int64, error) {
	count, err := s.repo.UnreadCount(userID)
	if err != nil {
		return 0, fmt.Errorf("unread count user=%d: %w", userID, err)
	}
	return count, nil
}
