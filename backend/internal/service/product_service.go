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
	"gorm.io/gorm"
)

// ProductService 商品业务服务（同时承载收藏逻辑，收藏复用商品仓储的计数方法）。
type ProductService struct {
	db           *gorm.DB
	productRepo  repository.ProductRepository
	favoriteRepo repository.FavoriteRepository
	reviewRepo   repository.ProductReviewRepository
	logger       *slog.Logger
}

// NewProductService 构造商品服务。
func NewProductService(db *gorm.DB, productRepo repository.ProductRepository, favoriteRepo repository.FavoriteRepository, reviewRepo repository.ProductReviewRepository, logger *slog.Logger) *ProductService {
	return &ProductService{db: db, productRepo: productRepo, favoriteRepo: favoriteRepo, reviewRepo: reviewRepo, logger: logger}
}

// Create 发布商品：新商品一律进入待审核并下架，同时生成第 1 轮审核记录，审核通过前大厅/搜索不可见、不可加购下单。
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
		Status:        constants.ProductStatusOffShelf,
		ReviewStatus:  constants.ProductReviewPending,
		ReviewRound:   1,
	}
	var reviewID uint
	if err := s.runTx(func(tx *gorm.DB) error {
		if err := s.productRepo.CreateTx(tx, product); err != nil {
			return fmt.Errorf("create product seller=%d: %w", sellerID, err)
		}
		review := &model.ProductReview{
			ProductID: product.ID,
			SellerID:  sellerID,
			Round:     1,
			Status:    constants.ProductReviewPending,
		}
		if err := s.reviewRepo.CreateTx(tx, review); err != nil {
			return fmt.Errorf("create pending review for product %d: %w", product.ID, err)
		}
		reviewID = review.ID
		return nil
	}); err != nil {
		return nil, err
	}
	s.logger.Info(constants.LogProductCreated, "product_id", product.ID, "seller_id", sellerID, "category", product.Category, "condition", product.Condition, "review_status", product.ReviewStatus)
	s.logger.Info(constants.LogProductReviewSubmit, "product_id", product.ID, "seller_id", sellerID, "review_id", reviewID, "round", 1, "trigger", "create")
	return product, nil
}

// Update 卖家更新商品。
// 审核通过（approved）的商品修改后立即下架并进入复审；驳回（rejected）商品修改即重新提交；
// 待审核（pending_review）期间重复提交直接拒绝，不重复改状态（幂等）。
func (s *ProductService) Update(userID, productID uint, req dto.ProductUpdateRequest) (*model.Product, error) {
	var product *model.Product
	var reviewID uint
	err := s.runTx(func(tx *gorm.DB) error {
		p, err := s.productRepo.GetByIDForUpdate(tx, productID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return util.NewAppError(constants.CodeProductNotFound, "商品更新失败：商品 id="+fmt.Sprint(productID)+" 不存在", err)
			}
			return fmt.Errorf("lock product %d for update: %w", productID, err)
		}
		if p.SellerID != userID {
			return util.NewAppError(constants.CodeForbidden, "商品更新失败：只有卖家（用户 id="+fmt.Sprint(userID)+"）可修改商品", nil)
		}
		switch p.ReviewStatus {
		case constants.ProductReviewPending:
			// 待审核中重复提交/重复修改不能重复改状态。
			s.logger.Warn(constants.LogProductReviewDuplicate, "product_id", productID, "seller_id", userID, "review_status", p.ReviewStatus)
			return util.NewAppError(constants.CodeReviewDuplicate, "商品更新失败：商品 id="+fmt.Sprint(productID)+" 正在平台审核中，请等待审核结果后再修改", nil)
		case constants.ProductReviewApproved, constants.ProductReviewRejected:
			// 允许修改并重新送审。
		default:
			return util.NewAppError(constants.CodeReviewStateInvalid, "商品更新失败：商品 id="+fmt.Sprint(productID)+" 审核状态 "+p.ReviewStatus+" 不可修改", nil)
		}
		if req.Title != nil {
			p.Title = *req.Title
		}
		if req.Description != nil {
			p.Description = *req.Description
		}
		if req.OriginalPrice != nil {
			p.OriginalPrice = *req.OriginalPrice
		}
		if req.Price != nil {
			p.Price = *req.Price
		}
		if req.Condition != nil {
			if !constants.ValidProductCondition(*req.Condition) {
				return util.NewAppError(constants.CodeBadRequest, "商品更新失败：成色 "+*req.Condition+" 非法", nil)
			}
			p.Condition = *req.Condition
		}
		if req.Category != nil {
			if !constants.ValidProductCategory(*req.Category) {
				return util.NewAppError(constants.CodeBadRequest, "商品更新失败：分类 "+*req.Category+" 非法", nil)
			}
			p.Category = *req.Category
		}
		if req.Images != nil {
			p.Images = strings.Join(req.Images, ",")
		}
		// 修改已上架/被驳回商品 → 新一轮待审核，复审期间继续下架。
		newRound := p.ReviewRound + 1
		p.ReviewRound = newRound
		p.ReviewStatus = constants.ProductReviewPending
		p.RejectReason = ""
		p.Status = constants.ProductStatusOffShelf
		if err := s.productRepo.UpdateTx(tx, p); err != nil {
			return fmt.Errorf("update product %d: %w", productID, err)
		}
		review := &model.ProductReview{
			ProductID: p.ID,
			SellerID:  p.SellerID,
			Round:     newRound,
			Status:    constants.ProductReviewPending,
		}
		if err := s.reviewRepo.CreateTx(tx, review); err != nil {
			return fmt.Errorf("resubmit review for product %d: %w", productID, err)
		}
		reviewID = review.ID
		product = p
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.logger.Info(constants.LogProductUpdated, "product_id", productID, "seller_id", userID, "status", product.Status, "review_status", product.ReviewStatus, "round", product.ReviewRound)
	s.logger.Info(constants.LogProductReviewSubmit, "product_id", productID, "seller_id", userID, "review_id", reviewID, "round", product.ReviewRound, "trigger", "update")
	return product, nil
}

// OffShelf 卖家主动下架商品（不改变审核状态；待审商品本身已下架，重复下架不改变状态）。
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
	if product.Status == constants.ProductStatusOffShelf {
		// 幂等：重复下架不重复改状态。
		return product, nil
	}
	product.Status = constants.ProductStatusOffShelf
	if err := s.productRepo.Update(product); err != nil {
		return nil, fmt.Errorf("off shelf product %d: %w", productID, err)
	}
	s.logger.Info(constants.LogProductOffShelf, "product_id", productID, "seller_id", userID, "status", product.Status)
	return product, nil
}

// GetDetail 商品详情（自增浏览量）。审核未通过商品仅卖家本人/管理员可看，其他访问者按不存在处理，
// 使审核结果与大厅、搜索、详情同时生效。
func (s *ProductService) GetDetail(productID, viewerID uint, viewerRole string) (*model.Product, error) {
	product, err := s.productRepo.GetByID(productID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeProductNotFound, "商品详情查询失败：商品 id="+fmt.Sprint(productID)+" 不存在", err)
		}
		return nil, fmt.Errorf("get product detail %d: %w", productID, err)
	}
	if product.ReviewStatus != constants.ProductReviewApproved {
		isOwner := viewerID > 0 && viewerID == product.SellerID
		isAdmin := viewerRole == constants.UserRoleAdmin
		if !isOwner && !isAdmin {
			return nil, util.NewAppError(constants.CodeProductNotFound, "商品详情查询失败：商品 id="+fmt.Sprint(productID)+" 尚未通过平台审核", nil)
		}
	}
	if err := s.productRepo.IncrViewCount(productID); err != nil {
		s.logger.Warn("incr view count failed", "product_id", productID, "err", err)
	}
	s.logger.Info(constants.LogProductViewed, "product_id", productID, "view_count", product.ViewCount+1)
	return product, nil
}

// List 商品大厅/搜索：仅返回审核通过且在售商品（审核结果与大厅、搜索同时生效）。
func (s *ProductService) List(q dto.ProductQuery, viewerID uint) (*dto.ProductListResponse, error) {
	page, pageSize := q.Page, q.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	status := q.Status
	if status == "" {
		status = constants.ProductStatusOnSale
	}
	query := map[string]interface{}{
		"keyword":       q.Keyword,
		"category":      q.Category,
		"condition":     q.Condition,
		"min_price":     q.MinPrice,
		"max_price":     q.MaxPrice,
		"status":        status,
		"review_status": constants.ProductReviewApproved,
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

// ListBySeller 我的发布（卖家视角，含待审/驳回商品与审核字段）。
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

// Favorite 收藏商品（仅审核通过商品可被收藏；已有收藏不受审核影响）。
func (s *ProductService) Favorite(userID, productID uint) error {
	product, err := s.productRepo.GetByID(productID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(constants.CodeProductNotFound, "收藏失败：商品 id="+fmt.Sprint(productID)+" 不存在", err)
		}
		return fmt.Errorf("get product %d for favorite: %w", productID, err)
	}
	if product.ReviewStatus != constants.ProductReviewApproved {
		return util.NewAppError(constants.CodeProductReviewing, "收藏失败：商品 id="+fmt.Sprint(productID)+" 尚未通过平台审核，暂不可收藏", nil)
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

// runTx 运行数据库事务；db 为空（单元测试内存仓储装配）时以 nil tx 直接执行。
func (s *ProductService) runTx(fn func(tx *gorm.DB) error) error {
	if s.db == nil {
		return fn(nil)
	}
	return s.db.Transaction(fn)
}
