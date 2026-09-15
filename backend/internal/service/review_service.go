package service

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/marketpal/marketpal/internal/constants"
	"github.com/marketpal/marketpal/internal/dto"
	"github.com/marketpal/marketpal/internal/model"
	"github.com/marketpal/marketpal/internal/repository"
	"github.com/marketpal/marketpal/internal/util"
	"gorm.io/gorm"
)

// ReviewService 评价业务服务：交易完成后互评，评价影响信用积分。
type ReviewService struct {
	reviewRepo  repository.ReviewRepository
	orderRepo   repository.OrderRepository
	userService *UserService
	db          *gorm.DB
	logger      *slog.Logger
}

// NewReviewService 构造评价服务。
func NewReviewService(db *gorm.DB, reviewRepo repository.ReviewRepository, orderRepo repository.OrderRepository, userService *UserService, logger *slog.Logger) *ReviewService {
	return &ReviewService{db: db, reviewRepo: reviewRepo, orderRepo: orderRepo, userService: userService, logger: logger}
}

// Create 评价（事务）：订单必须已完成 → 不能评价自己 → 唯一性 → 更新信用积分。
func (s *ReviewService) Create(reviewerID uint, req dto.ReviewCreateRequest) (*model.Review, error) {
	var review *model.Review
	err := s.db.Transaction(func(tx *gorm.DB) error {
		order, err := s.orderRepo.GetByIDForUpdate(tx, req.OrderID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return utilAppError(constants.CodeOrderNotFound, "评价失败：订单 id="+fmt.Sprint(req.OrderID)+" 不存在", err)
			}
			return fmt.Errorf("lock order %d for review: %w", req.OrderID, err)
		}
		if order.Status != constants.OrderStatusCompleted {
			return utilAppError(constants.CodeOrderStateInvalid, "评价失败：订单 "+order.OrderNo+" 状态为 "+order.Status+"，交易完成后才能评价", nil)
		}
		revieweeID := order.SellerID
		if reviewerID == order.SellerID {
			revieweeID = order.BuyerID
		}
		if reviewerID == revieweeID {
			return utilAppError(constants.CodeCannotSelfReview, "评价失败：不能评价自己（用户 id="+fmt.Sprint(reviewerID)+"）", nil)
		}
		if _, err := s.reviewRepo.GetByOrderID(req.OrderID); err == nil {
			return utilAppError(constants.CodeReviewExists, "评价失败：订单 id="+fmt.Sprint(req.OrderID)+" 已有评价记录", nil)
		} else if !errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("check review order %d: %w", req.OrderID, err)
		}
		review = &model.Review{
			OrderID:    order.ID,
			ProductID:  order.ProductID,
			ReviewerID: reviewerID,
			RevieweeID: revieweeID,
			Rating:     req.Rating,
			Content:    req.Content,
		}
		if err := s.reviewRepo.CreateWithTx(tx, review); err != nil {
			return fmt.Errorf("create review order=%d reviewer=%d: %w", req.OrderID, reviewerID, err)
		}
		delta := creditDelta(req.Rating)
		if err := s.userService.UpdateCredit(tx, revieweeID, delta); err != nil {
			return fmt.Errorf("update credit of user %d delta=%d: %w", revieweeID, delta, err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.logger.Info(constants.LogReviewCreated, "review_id", review.ID, "order_id", req.OrderID, "reviewer_id", reviewerID, "reviewee_id", review.RevieweeID, "rating", review.Rating)
	return review, nil
}

// creditDelta 评价等级 → 信用积分增减（好评 +5，中评 0，差评 -5）。
func creditDelta(rating string) int {
	switch rating {
	case constants.ReviewRatingGood:
		return 5
	case constants.ReviewRatingBad:
		return -5
	default:
		return 0
	}
}

// ListByUser 查询某用户相关评价（复用同一 service 方法，个人中心/卖家详情共用）。
func (s *ReviewService) ListByUser(userID uint, page, pageSize int) (*dto.ReviewListResponse, error) {
	p := util.NormalizePage(page, pageSize)
	list, total, err := s.reviewRepo.ListByUser(userID, p.Page, p.PageSize)
	if err != nil {
		return nil, fmt.Errorf("list reviews user=%d: %w", userID, err)
	}
	res := &dto.ReviewListResponse{List: []dto.ReviewVO{}, Total: total, Page: p.Page, Size: p.PageSize}
	for i := range list {
		vo := dto.ReviewVO{
			ID:         list[i].ID,
			OrderID:    list[i].OrderID,
			ProductID:  list[i].ProductID,
			ReviewerID: list[i].ReviewerID,
			RevieweeID: list[i].RevieweeID,
			Rating:     list[i].Rating,
			Content:    list[i].Content,
			CreatedAt:  util.FormatTime(list[i].CreatedAt),
		}
		if list[i].Reviewer != nil {
			vo.Reviewer = dto.FromUser(list[i].Reviewer)
		}
		res.List = append(res.List, vo)
	}
	return res, nil
}

// ListByProduct 查询某商品评价（商品详情页复用）。
func (s *ReviewService) ListByProduct(productID uint, page, pageSize int) (*dto.ReviewListResponse, error) {
	p := util.NormalizePage(page, pageSize)
	list, total, err := s.reviewRepo.ListByProduct(productID, p.Page, p.PageSize)
	if err != nil {
		return nil, fmt.Errorf("list product reviews product=%d: %w", productID, err)
	}
	res := &dto.ReviewListResponse{List: []dto.ReviewVO{}, Total: total, Page: p.Page, Size: p.PageSize}
	for i := range list {
		vo := dto.ReviewVO{
			ID:         list[i].ID,
			OrderID:    list[i].OrderID,
			ProductID:  list[i].ProductID,
			ReviewerID: list[i].ReviewerID,
			RevieweeID: list[i].RevieweeID,
			Rating:     list[i].Rating,
			Content:    list[i].Content,
			CreatedAt:  util.FormatTime(list[i].CreatedAt),
		}
		if list[i].Reviewer != nil {
			vo.Reviewer = dto.FromUser(list[i].Reviewer)
		}
		res.List = append(res.List, vo)
	}
	return res, nil
}
