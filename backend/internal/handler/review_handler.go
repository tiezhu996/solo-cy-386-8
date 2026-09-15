package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/marketpal/marketpal/internal/dto"
	"github.com/marketpal/marketpal/internal/middleware"
	"github.com/marketpal/marketpal/internal/service"
	"github.com/marketpal/marketpal/internal/util"
)

// ReviewHandler 评价 HTTP 处理器。
type ReviewHandler struct {
	svc *service.ReviewService
}

// NewReviewHandler 构造评价处理器。
func NewReviewHandler(svc *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{svc: svc}
}

// Create POST /api/v1/reviews
func (h *ReviewHandler) Create(c *gin.Context) {
	var req dto.ReviewCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "评价失败：参数校验不通过 "+err.Error())
		return
	}
	review, err := h.svc.Create(middleware.GetUserID(c), req)
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OKMessage(c, "评价成功", gin.H{"review_id": review.ID, "rating": review.Rating})
}

// List GET /api/v1/reviews
func (h *ReviewHandler) List(c *gin.Context) {
	var q dto.ReviewQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "评价列表失败：查询参数不合法 "+err.Error())
		return
	}
	page := q.Page
	if page < 1 {
		page = 1
	}
	pageSize := q.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	var res *dto.ReviewListResponse
	var err error
	if q.ProductID > 0 {
		res, err = h.svc.ListByProduct(q.ProductID, page, pageSize)
	} else if q.UserID > 0 {
		res, err = h.svc.ListByUser(q.UserID, page, pageSize)
	} else {
		res, err = h.svc.ListByUser(middleware.GetUserID(c), page, pageSize)
	}
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OK(c, res)
}
