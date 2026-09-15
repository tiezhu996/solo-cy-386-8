package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/marketpal/marketpal/internal/dto"
	"github.com/marketpal/marketpal/internal/middleware"
	"github.com/marketpal/marketpal/internal/service"
	"github.com/marketpal/marketpal/internal/util"
)

// MessageHandler 私信 HTTP 处理器。
type MessageHandler struct {
	svc *service.MessageService
}

// NewMessageHandler 构造私信处理器。
func NewMessageHandler(svc *service.MessageService) *MessageHandler {
	return &MessageHandler{svc: svc}
}

// Send POST /api/v1/messages
func (h *MessageHandler) Send(c *gin.Context) {
	var req dto.MessageSendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "私信发送失败：参数校验不通过 "+err.Error())
		return
	}
	msg, err := h.svc.Send(middleware.GetUserID(c), req)
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OKMessage(c, "消息已发送", gin.H{"message_id": msg.ID})
}

// Conversations GET /api/v1/messages/conversations
func (h *MessageHandler) Conversations(c *gin.Context) {
	list, err := h.svc.ListConversations(middleware.GetUserID(c))
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OK(c, list)
}

// Conversation GET /api/v1/messages/conversations/:peerId
func (h *MessageHandler) Conversation(c *gin.Context) {
	peerID, err := strconv.ParseUint(c.Param("peerId"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "会话查询失败：peer id 参数非法")
		return
	}
	page := parseInt(c.DefaultQuery("page", "1"), 1)
	pageSize := parseInt(c.DefaultQuery("page_size", "20"), 20)
	msgs, total, err := h.svc.ListConversation(middleware.GetUserID(c), uint(peerID), page, pageSize)
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	list := make([]dto.MessageVO, 0, len(msgs))
	for i := range msgs {
		list = append(list, dto.MessageVO{
			ID:         msgs[i].ID,
			SenderID:   msgs[i].SenderID,
			ReceiverID: msgs[i].ReceiverID,
			ProductID:  msgs[i].ProductID,
			Content:    msgs[i].Content,
			IsRead:     msgs[i].IsRead,
			CreatedAt:  util.FormatTime(msgs[i].CreatedAt),
		})
	}
	util.OK(c, gin.H{"list": list, "total": total, "page": page, "page_size": pageSize})
}

// MarkRead PUT /api/v1/messages/conversations/:peerId/read
func (h *MessageHandler) MarkRead(c *gin.Context) {
	peerID, err := strconv.ParseUint(c.Param("peerId"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "已读标记失败：peer id 参数非法")
		return
	}
	if err := h.svc.MarkRead(middleware.GetUserID(c), uint(peerID)); err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OKMessage(c, "消息已标记已读", nil)
}

// UnreadCount GET /api/v1/messages/unread-count
func (h *MessageHandler) UnreadCount(c *gin.Context) {
	count, err := h.svc.UnreadCount(middleware.GetUserID(c))
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OK(c, gin.H{"unread": count})
}
