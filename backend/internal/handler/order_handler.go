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

// OrderHandler 订单 HTTP 处理器。
type OrderHandler struct {
	svc *service.OrderService
}

// NewOrderHandler 构造订单处理器。
func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

// Create POST /api/v1/orders
func (h *OrderHandler) Create(c *gin.Context) {
	var req dto.OrderCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "下单失败：参数校验不通过 "+err.Error())
		return
	}
	order, err := h.svc.Create(middleware.GetUserID(c), req)
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OKMessage(c, "下单成功，请尽快付款", orderVO(order))
}

// Pay POST /api/v1/orders/:id/pay
func (h *OrderHandler) Pay(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "订单支付失败：订单 id 参数非法")
		return
	}
	order, err := h.svc.Pay(middleware.GetUserID(c), uint(id))
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OKMessage(c, "支付成功", orderVO(order))
}

// Ship POST /api/v1/orders/:id/ship
func (h *OrderHandler) Ship(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "订单发货失败：订单 id 参数非法")
		return
	}
	order, err := h.svc.Ship(middleware.GetUserID(c), uint(id))
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OKMessage(c, "已发货", orderVO(order))
}

// Receive POST /api/v1/orders/:id/receive
func (h *OrderHandler) Receive(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "确认收货失败：订单 id 参数非法")
		return
	}
	order, err := h.svc.Receive(middleware.GetUserID(c), uint(id))
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OKMessage(c, "已确认收货", orderVO(order))
}

// Complete POST /api/v1/orders/:id/complete
func (h *OrderHandler) Complete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "完成交易失败：订单 id 参数非法")
		return
	}
	order, err := h.svc.Complete(middleware.GetUserID(c), uint(id))
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OKMessage(c, "交易完成", orderVO(order))
}

// Cancel POST /api/v1/orders/:id/cancel
func (h *OrderHandler) Cancel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "取消订单失败：订单 id 参数非法")
		return
	}
	order, err := h.svc.Cancel(middleware.GetUserID(c), uint(id), "buyer")
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OKMessage(c, "订单已取消", orderVO(order))
}

// CancelBySeller POST /api/v1/orders/:id/cancel (role=seller)
func (h *OrderHandler) CancelBySeller(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "取消订单失败：订单 id 参数非法")
		return
	}
	order, err := h.svc.Cancel(middleware.GetUserID(c), uint(id), "seller")
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OKMessage(c, "订单已取消", orderVO(order))
}

// List GET /api/v1/orders
func (h *OrderHandler) List(c *gin.Context) {
	var q dto.OrderQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "订单列表失败：查询参数不合法 "+err.Error())
		return
	}
	res, err := h.svc.List(middleware.GetUserID(c), q)
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OK(c, res)
}

// Detail GET /api/v1/orders/:id
func (h *OrderHandler) Detail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "订单详情失败：订单 id 参数非法")
		return
	}
	order, err := h.svc.GetDetail(middleware.GetUserID(c), uint(id))
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OK(c, orderVO(order))
}
