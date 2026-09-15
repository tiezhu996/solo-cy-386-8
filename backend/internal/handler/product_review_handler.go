package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/marketpal/marketpal/internal/constants"
	"github.com/marketpal/marketpal/internal/dto"
	"github.com/marketpal/marketpal/internal/middleware"
	"github.com/marketpal/marketpal/internal/service"
	"github.com/marketpal/marketpal/internal/util"
)

// ProductReviewHandler 商品平台审核 HTTP 处理器（仅管理员可给出审核结果）。
type ProductReviewHandler struct {
	svc *service.ProductReviewService
}

// NewProductReviewHandler 构造商品审核处理器。
func NewProductReviewHandler(svc *service.ProductReviewService) *ProductReviewHandler {
	return &ProductReviewHandler{svc: svc}
}

// List GET /api/v1/product-reviews 审核记录/待审队列（管理员）。
func (h *ProductReviewHandler) List(c *gin.Context) {
	var q dto.ProductReviewQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "商品审核列表查询失败：查询参数不合法 "+err.Error())
		return
	}
	res, err := h.svc.List(middleware.GetUserID(c), q)
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OK(c, res)
}

// History GET /api/v1/products/:id/reviews 单个商品审核历史（卖家本人或管理员）。
func (h *ProductReviewHandler) History(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "商品审核记录查询失败：商品 id 参数非法")
		return
	}
	res, err := h.svc.ListByProduct(middleware.GetUserID(c), middleware.GetRole(c), uint(id))
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OK(c, res)
}

// Decide POST /api/v1/product-reviews/:id/decision 管理员通过/驳回（驳回须填原因）。
func (h *ProductReviewHandler) Decide(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "商品审核失败：审核记录 id 参数非法")
		return
	}
	var req dto.ProductReviewDecisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "商品审核失败：参数校验不通过 "+err.Error())
		return
	}
	rv, err := h.svc.Decide(middleware.GetUserID(c), uint(id), req)
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	msg := constants.MsgReviewApproved
	if req.Action == "reject" {
		msg = constants.MsgReviewRejected
	}
	util.OKMessage(c, msg, gin.H{"id": rv.ID, "product_id": rv.ProductID, "round": rv.Round, "status": rv.Status})
}
