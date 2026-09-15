package service

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/marketpal/marketpal/internal/constants"
	"github.com/marketpal/marketpal/internal/dto"
	"github.com/marketpal/marketpal/internal/model"
	"github.com/marketpal/marketpal/internal/repository"
	"github.com/marketpal/marketpal/internal/util"
)

// ProductService 商品业务服务（同时承载收藏逻辑，收藏复用商品仓储的计数方法）。
type ProductService struct {
	productRepo  repository.ProductRepository
	favoriteRepo repository.FavoriteRepository
	logger       *slog.Logger
}

// NewProductService 构造商品服务。
func NewProductService(productRepo repository.ProductRepository, favoriteRepo repository.FavoriteRepository, logger *slog.Logger) *ProductService {
	return &ProductService{productRepo: productRepo, favoriteRepo: favoriteRepo, logger: logger}
}

// Create 发布商品。
func (s *ProductService) Create(sellerID uint, req dto.ProductCreateRequest) (*model.Product, error) {
	if !constants.ValidProductCondition(req.Condition) {
		return nil, util.NewAppError(constants.CodeBadRequest, "商品发布失败：成色 "+req.Condition+" 非法", nil)
	}
	if !constants.ValidProductCategory(req.Category) {
		return nil, util.NewAppError(constants.CodeBadRequest, "商品发布失败：分类 "+req.Category+" 非法", nil)
	}
	product := &model.Product{
		SellerID:      sellerID,
		Title:         req.Title,
		Description:   req.Description,
		OriginalPrice: req.OriginalPrice,
		Price:         req.Price,
		Condition:     req.Condition,
		Category:      req.Category,
		Images:        strings.Join(req.Images, ","),
		Status:        constants.ProductStatusOnSale,
	}
	if err := s.productRepo.Create(product); err != nil {
		return nil, fmt.Errorf("create product seller=%d: %w", sellerID, err)
	}
	s.logger.Info(constants.LogProductCreated, "product_id", product.ID, "seller_id", sellerID, "category", product.Category, "condition", product.Condition)
	return product, nil
}

// Update 卖家更新商品。
func (s *ProductService) Update(userID, productID uint, req dto.ProductUpdateRequest) (*model.Product, error) {
	product, err := s.productRepo.GetByID(productID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeProductNotFound, "商品更新失败：商品 id="+fmt.Sprint(productID)+" 不存在", err)
		}
		return nil, fmt.Errorf("get product %d for update: %w", productID, err)
	}
	if product.SellerID != userID {
		return nil, util.NewAppError(constants.CodeForbidden, "商品更新失败：只有卖家（用户 id="+fmt.Sprint(userID)+"）可修改商品", nil)
	}
	if req.Title != nil {
		product.Title = *req.Title
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.OriginalPrice != nil {
		product.OriginalPrice = *req.OriginalPrice
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.Condition != nil {
		if !constants.ValidProductCondition(*req.Condition) {
			return nil, util.NewAppError(constants.CodeBadRequest, "商品更新失败：成色 "+*req.Condition+" 非法", nil)
		}
		product.Condition = *req.Condition
	}
	if req.Category != nil {
		if !constants.ValidProductCategory(*req.Category) {
			return nil, util.NewAppError(constants.CodeBadRequest, "商品更新失败：分类 "+*req.Category+" 非法", nil)
		}
		product.Category = *req.Category
	}
	if req.Images != nil {
		product.Images = strings.Join(req.Images, ",")
	}
	if err := s.productRepo.Update(product); err != nil {
		return nil, fmt.Errorf("update product %d: %w", productID, err)
	}
	s.logger.Info(constants.LogProductUpdated, "product_id", productID, "seller_id", userID, "status", product.Status)
	return product, nil
}

// OffShelf 卖家下架商品。
func (s *ProductService) OffShelf(userID, productID uint) (*model.Product, error) {
	product, err := s.productRepo.GetByID(productID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeProductNotFound, "商品下架失败：商品 id="+fmt.Sprint(productID)+" 不存在", err)
		}
		return nil, fmt.Errorf("get product %d for off shelf: %w", productID, err)
	}
	if product.SellerID != userID {
		return nil, util.NewAppError(constants.CodeForbidden, "商品下架失败：只有卖家（用户 id="+fmt.Sprint(userID)+"）可下架商品", nil)
	}
	product.Status = constants.ProductStatusOffShelf
	if err := s.productRepo.Update(product); err != nil {
		return nil, fmt.Errorf("off shelf product %d: %w", productID, err)
	}
	s.logger.Info(constants.LogProductOffShelf, "product_id", productID, "seller_id", userID, "status", product.Status)
	return product, nil
}

// GetDetail 商品详情（自增浏览量）。
func (s *ProductService) GetDetail(productID uint) (*model.Product, error) {
	product, err := s.productRepo.GetByID(productID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeProductNotFound, "商品详情查询失败：商品 id="+fmt.Sprint(productID)+" 不存在", err)
		}
		return nil, fmt.Errorf("get product detail %d: %w", productID, err)
	}
	if err := s.productRepo.IncrViewCount(productID); err != nil {
		s.logger.Warn("incr view count failed", "product_id", productID, "err", err)
	}
	s.logger.Info(constants.LogProductViewed, "product_id", productID, "view_count", product.ViewCount+1)
	return product, nil
}

// List 商品列表/搜索（关键词、分类、价格区间、成色、排序）。
func (s *ProductService) List(q dto.ProductQuery, viewerID uint) (*dto.ProductListResponse, error) {
	page, pageSize := q.Page, q.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	query := map[string]interface{}{
		"keyword":    q.Keyword,
		"category":   q.Category,
		"condition":  q.Condition,
		"min_price":  q.MinPrice,
		"max_price":  q.MaxPrice,
		"status":     q.Status,
	}
	if q.Status == "" {
		query["status"] = constants.ProductStatusOnSale
	}
	products, total, err := s.productRepo.List(query, q.SortBy, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	list := make([]dto.ProductVO, 0, len(products))
	for _, p := range products {
		fav := false
		if viewerID > 0 {
			fav, _ = s.favoriteRepo.Exists(viewerID, p.ID)
		}
		list = append(list, toProductVO(&p, fav))
	}
	return &dto.ProductListResponse{List: list, Total: total, Page: page, Size: pageSize}, nil
}

// ListBySeller 我的发布（卖家视角）。
func (s *ProductService) ListBySeller(sellerID uint, page, pageSize int) (*dto.ProductListResponse, error) {
	p := util.NormalizePage(page, pageSize)
	products, total, err := s.productRepo.ListBySeller(sellerID, p.Page, p.PageSize)
	if err != nil {
		return nil, fmt.Errorf("list seller products seller=%d: %w", sellerID, err)
	}
	list := make([]dto.ProductVO, 0, len(products))
	for i := range products {
		list = append(list, toProductVO(&products[i], false))
	}
	return &dto.ProductListResponse{List: list, Total: total, Page: p.Page, Size: p.PageSize}, nil
}

// ListFavorites 我的收藏。
func (s *ProductService) ListFavorites(userID uint, page, pageSize int) (*dto.ProductListResponse, error) {
	p := util.NormalizePage(page, pageSize)
	favs, total, err := s.favoriteRepo.ListByUser(userID, p.Page, p.PageSize)
	if err != nil {
		return nil, fmt.Errorf("list favorites user=%d: %w", userID, err)
	}
	list := make([]dto.ProductVO, 0, len(favs))
	for i := range favs {
		if favs[i].Product == nil {
			continue
		}
		list = append(list, toProductVO(favs[i].Product, true))
	}
	return &dto.ProductListResponse{List: list, Total: total, Page: p.Page, Size: p.PageSize}, nil
}

// Favorite 收藏商品。
func (s *ProductService) Favorite(userID, productID uint) error {
	if _, err := s.productRepo.GetByID(productID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(constants.CodeProductNotFound, "收藏失败：商品 id="+fmt.Sprint(productID)+" 不存在", err)
		}
		return fmt.Errorf("get product %d for favorite: %w", productID, err)
	}
	exists, err := s.favoriteRepo.Exists(userID, productID)
	if err != nil {
		return fmt.Errorf("check favorite user=%d product=%d: %w", userID, productID, err)
	}
	if exists {
		return util.NewAppError(constants.CodeConflict, "收藏失败：用户 id="+fmt.Sprint(userID)+" 已收藏该商品", nil)
	}
	if err := s.favoriteRepo.Create(&model.Favorite{UserID: userID, ProductID: productID}); err != nil {
		return fmt.Errorf("create favorite user=%d product=%d: %w", userID, productID, err)
	}
	if err := s.productRepo.IncrFavoriteCount(nil, productID, 1); err != nil {
		s.logger.Warn("incr favorite count failed", "product_id", productID, "err", err)
	}
	s.logger.Info(constants.LogFavoriteAdded, "user_id", userID, "product_id", productID)
	return nil
}

// Unfavorite 取消收藏。
func (s *ProductService) Unfavorite(userID, productID uint) error {
	if err := s.favoriteRepo.Delete(userID, productID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(constants.CodeNotFound, "取消收藏失败：用户 id="+fmt.Sprint(userID)+" 未收藏商品 id="+fmt.Sprint(productID), err)
		}
		return fmt.Errorf("delete favorite user=%d product=%d: %w", userID, productID, err)
	}
	if err := s.productRepo.IncrFavoriteCount(nil, productID, -1); err != nil {
		s.logger.Warn("decr favorite count failed", "product_id", productID, "err", err)
	}
	s.logger.Info(constants.LogFavoriteRemoved, "user_id", userID, "product_id", productID)
	return nil
}
// toProductVO model → 视图对象（dto 集中转换，商品/订单/购物车模块复用）。
func toProductVO(p *model.Product, isFavorite bool) dto.ProductVO {
	return dto.FromProduct(p, isFavorite)
}
