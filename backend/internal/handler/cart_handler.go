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

// CartHandler 购物车 HTTP 处理器。
type CartHandler struct {
	svc *service.CartService
}

// NewCartHandler 构造购物车处理器。
func NewCartHandler(svc *service.CartService) *CartHandler {
	return &CartHandler{svc: svc}
}

// Add POST /api/v1/cart
func (h *CartHandler) Add(c *gin.Context) {
	var req dto.CartAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "加入购物车失败：参数校验不通过 "+err.Error())
		return
	}
	item, err := h.svc.Add(middleware.GetUserID(c), req)
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OKMessage(c, "已加入购物车", gin.H{"cart_item_id": item.ID})
}

// List GET /api/v1/cart
func (h *CartHandler) List(c *gin.Context) {
	res, err := h.svc.List(middleware.GetUserID(c))
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OK(c, res)
}

// Update PUT /api/v1/cart/:id
func (h *CartHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "购物车更新失败：条目 id 参数非法")
		return
	}
	var req dto.CartUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "购物车更新失败：参数校验不通过 "+err.Error())
		return
	}
	item, err := h.svc.Update(middleware.GetUserID(c), uint(id), req)
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OKMessage(c, "购物车已更新", gin.H{"cart_item_id": item.ID, "quantity": item.Quantity, "selected": item.Selected})
}

// Delete DELETE /api/v1/cart/:id
func (h *CartHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "购物车删除失败：条目 id 参数非法")
		return
	}
	if err := h.svc.Delete(middleware.GetUserID(c), uint(id)); err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OKMessage(c, "已移除购物车条目", nil)
}
