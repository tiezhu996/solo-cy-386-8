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

// AddressHandler 收货地址 HTTP 处理器。
type AddressHandler struct {
	svc *service.AddressService
}

// NewAddressHandler 构造收货地址处理器。
func NewAddressHandler(svc *service.AddressService) *AddressHandler {
	return &AddressHandler{svc: svc}
}

// Create POST /api/v1/addresses
func (h *AddressHandler) Create(c *gin.Context) {
	var req dto.AddressCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "地址保存失败：参数校验不通过 "+err.Error())
		return
	}
	addr, err := h.svc.Create(middleware.GetUserID(c), req)
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OKMessage(c, "收货地址已保存", dto.FromAddress(addr))
}

// Update PUT /api/v1/addresses/:id
func (h *AddressHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "地址更新失败：地址 id 参数非法")
		return
	}
	var req dto.AddressUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "地址更新失败：参数校验不通过 "+err.Error())
		return
	}
	addr, err := h.svc.Update(middleware.GetUserID(c), uint(id), req)
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OKMessage(c, "收货地址已更新", dto.FromAddress(addr))
}

// Delete DELETE /api/v1/addresses/:id
func (h *AddressHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "地址删除失败：地址 id 参数非法")
		return
	}
	if err := h.svc.Delete(middleware.GetUserID(c), uint(id)); err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OKMessage(c, "收货地址已删除", nil)
}

// List GET /api/v1/addresses
func (h *AddressHandler) List(c *gin.Context) {
	list, err := h.svc.List(middleware.GetUserID(c))
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	items := make([]dto.AddressVO, 0, len(list))
	for i := range list {
		items = append(items, dto.FromAddress(&list[i]))
	}
	util.OK(c, items)
}
