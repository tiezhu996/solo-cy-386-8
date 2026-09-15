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

// ProductHandler 商品 HTTP 处理器。
type ProductHandler struct {
	svc *service.ProductService
}

// NewProductHandler 构造商品处理器。
func NewProductHandler(svc *service.ProductService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

// Create POST /api/v1/products
func (h *ProductHandler) Create(c *gin.Context) {
	var req dto.ProductCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "商品发布失败：参数校验不通过 "+err.Error())
		return
	}
	product, err := h.svc.Create(middleware.GetUserID(c), req)
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OKMessage(c, "商品发布成功", dto.FromProduct(product, false))
}

// Update PUT /api/v1/products/:id
func (h *ProductHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "商品更新失败：商品 id 参数非法")
		return
	}
	var req dto.ProductUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "商品更新失败：参数校验不通过 "+err.Error())
		return
	}
	product, err := h.svc.Update(middleware.GetUserID(c), uint(id), req)
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OKMessage(c, "商品已更新", dto.FromProduct(product, false))
}

// OffShelf POST /api/v1/products/:id/off-shelf
func (h *ProductHandler) OffShelf(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "商品下架失败：商品 id 参数非法")
		return
	}
	product, err := h.svc.OffShelf(middleware.GetUserID(c), uint(id))
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OKMessage(c, "商品已下架", dto.FromProduct(product, false))
}

// Detail GET /api/v1/products/:id
func (h *ProductHandler) Detail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "商品详情失败：商品 id 参数非法")
		return
	}
	product, err := h.svc.GetDetail(uint(id))
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OK(c, dto.FromProduct(product, false))
}

// List GET /api/v1/products
func (h *ProductHandler) List(c *gin.Context) {
	var q dto.ProductQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "商品列表失败：查询参数不合法 "+err.Error())
		return
	}
	res, err := h.svc.List(q, 0)
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OK(c, res)
}

// MyProducts GET /api/v1/products/mine
func (h *ProductHandler) MyProducts(c *gin.Context) {
	page := parseInt(c.DefaultQuery("page", "1"), 1)
	pageSize := parseInt(c.DefaultQuery("page_size", "10"), 10)
	res, err := h.svc.ListBySeller(middleware.GetUserID(c), page, pageSize)
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OK(c, res)
}

// Favorite POST /api/v1/products/:id/favorite
func (h *ProductHandler) Favorite(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "收藏失败：商品 id 参数非法")
		return
	}
	if err := h.svc.Favorite(middleware.GetUserID(c), uint(id)); err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OKMessage(c, "收藏成功", nil)
}

// Unfavorite DELETE /api/v1/products/:id/favorite
func (h *ProductHandler) Unfavorite(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "取消收藏失败：商品 id 参数非法")
		return
	}
	if err := h.svc.Unfavorite(middleware.GetUserID(c), uint(id)); err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OKMessage(c, "已取消收藏", nil)
}

// Favorites GET /api/v1/favorites
func (h *ProductHandler) Favorites(c *gin.Context) {
	page := parseInt(c.DefaultQuery("page", "1"), 1)
	pageSize := parseInt(c.DefaultQuery("page_size", "10"), 10)
	res, err := h.svc.ListFavorites(middleware.GetUserID(c), page, pageSize)
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OK(c, res)
}
