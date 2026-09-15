package service

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/marketpal/marketpal/internal/constants"
	"github.com/marketpal/marketpal/internal/dto"
	"github.com/marketpal/marketpal/internal/model"
	"github.com/marketpal/marketpal/internal/repository"
	"github.com/marketpal/marketpal/internal/util"
	"gorm.io/gorm"
)

// ProductReviewService 商品平台审核服务：卖家送审、管理员通过/驳回、审核记录回读。
// 审核是商品状态机的一部分，所有多步写均在事务内对 products 行 SELECT ... FOR UPDATE 后进行。
type ProductReviewService struct {
	db          *gorm.DB
	reviewRepo  repository.ProductReviewRepository
	productRepo repository.ProductRepository
	logger      *slog.Logger
}

// NewProductReviewService 构造商品审核服务。
func NewProductReviewService(db *gorm.DB, reviewRepo repository.ProductReviewRepository, productRepo repository.ProductRepository, logger *slog.Logger) *ProductReviewService {
	return &ProductReviewService{db: db, reviewRepo: reviewRepo, productRepo: productRepo, logger: logger}
}

// Decide 管理员对一轮待审核记录给出结果。
// 并发安全：事务内先锁商品行，再条件更新审核记录（仅 pending_review 可改），
// 同一商品的两个并发审核事务只有一个能写入结果，重复审核返回 CodeReviewStateInvalid，不重复改状态。
func (s *ProductReviewService) Decide(reviewerID uint, reviewID uint, req dto.ProductReviewDecisionRequest) (*model.ProductReview, error) {
	if req.Action == "reject" && strings.TrimSpace(req.Reason) == "" {
		return nil, util.NewAppError(constants.CodeReviewReasonEmpty, "商品审核驳回失败：管理员 id="+fmt.Sprint(reviewerID)+" 必须填写驳回原因", nil)
	}
	reason := ""
	if req.Action == "reject" {
		reason = strings.TrimSpace(req.Reason)
	}
	var decided *model.ProductReview
	err := s.runTx(func(tx *gorm.DB) error {
		rv, err := s.reviewRepo.GetByIDForUpdate(tx, reviewID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return util.NewAppError(constants.CodeReviewNotFound, "商品审核失败：审核记录 id="+fmt.Sprint(reviewID)+" 不存在", err)
			}
			return fmt.Errorf("lock product review %d: %w", reviewID, err)
		}
		// 锁定同一商品行：与下单/修改/取消订单互斥，保证“同一商品并发审核只能产生一个结果”。
		product, err := s.productRepo.GetByIDForUpdate(tx, rv.ProductID)
		if err != nil {
			return fmt.Errorf("lock product %d for review: %w", rv.ProductID, err)
		}
		if rv.Status != constants.ProductReviewPending {
			s.logger.Warn(constants.LogInvalidState, "entity", "product_review", "id", reviewID, "from", rv.Status, "to", req.Action, "operator", reviewerID)
			return util.NewAppError(constants.CodeReviewStateInvalid,
				"商品审核失败：商品 id="+fmt.Sprint(rv.ProductID)+" 第"+fmt.Sprint(rv.Round)+"轮审核已结束（"+rv.Status+"），管理员 id="+fmt.Sprint(reviewerID)+" 不能重复审核", nil)
		}
		targetStatus := constants.ProductReviewApproved
		if req.Action == "reject" {
			targetStatus = constants.ProductReviewRejected
		}
		now := time.Now()
		// 条件更新：即使绕过内存判断，数据库层也保证只有一个并发事务能写结果。
		if err := s.reviewRepo.DecideForUpdate(tx, reviewID, reviewerID, targetStatus, reason, &now); err != nil {
			if errors.Is(err, repository.ErrReviewAlreadyDecided) {
				return util.NewAppError(constants.CodeReviewStateInvalid,
					"商品审核失败：审核记录 id="+fmt.Sprint(reviewID)+" 已被其他管理员处理，重复审核无效", err)
			}
			return fmt.Errorf("decide product review %d: %w", reviewID, err)
		}
		rv.Status = targetStatus
		rv.Reason = reason
		rv.ReviewerID = reviewerID
		rv.ReviewedAt = &now
		updates := map[string]interface{}{
			"review_status": targetStatus,
			"reject_reason": reason,
		}
		if targetStatus == constants.ProductReviewApproved {
			// 审核通过 → 重新上架可售（卖家主动下架的商品在送审时已为 off_shelf，按通过即上架处理）。
			updates["status"] = constants.ProductStatusOnSale
		} else {
			// 驳回 → 继续下架，等待卖家修改重提。
			updates["status"] = constants.ProductStatusOffShelf
		}
		if err := s.productRepo.UpdateReviewForUpdate(tx, product.ID, updates); err != nil {
			return fmt.Errorf("apply review result to product %d: %w", product.ID, err)
		}
		decided = rv
		return nil
	})
	if err != nil {
		return nil, err
	}
	if req.Action == "reject" {
		s.logger.Info(constants.LogProductReviewReject, "product_id", decided.ProductID, "review_id", decided.ID, "reviewer_id", reviewerID, "round", decided.Round, "reason", reason)
	} else {
		s.logger.Info(constants.LogProductReviewApprove, "product_id", decided.ProductID, "review_id", decided.ID, "reviewer_id", reviewerID, "round", decided.Round)
	}
	return decided, nil
}

// List 审核记录列表（管理员待审队列，可按状态/商品/卖家过滤）。
func (s *ProductReviewService) List(reviewerID uint, q dto.ProductReviewQuery) (*dto.ProductReviewListResponse, error) {
	page, pageSize := q.Page, q.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	query := map[string]interface{}{
		"status":     q.Status,
		"product_id": q.ProductID,
		"seller_id":  q.SellerID,
	}
	list, total, err := s.reviewRepo.List(query, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("list product reviews: %w", err)
	}
	s.logger.Info(constants.LogProductReviewListed, "reviewer_id", reviewerID, "page", page, "page_size", pageSize)
	return &dto.ProductReviewListResponse{List: toReviewVOList(list), Total: total, Page: page, Size: pageSize}, nil
}

// ListByProduct 单个商品的审核历史回读（卖家本人或管理员）。
func (s *ProductReviewService) ListByProduct(viewerID uint, viewerRole string, productID uint) (*dto.ProductReviewListResponse, error) {
	product, err := s.productRepo.GetByID(productID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeProductNotFound, "商品审核记录查询失败：商品 id="+fmt.Sprint(productID)+" 不存在", err)
		}
		return nil, fmt.Errorf("get product %d for review history: %w", productID, err)
	}
	if viewerRole != constants.UserRoleAdmin && product.SellerID != viewerID {
		return nil, util.NewAppError(constants.CodeForbidden, "商品审核记录查询失败：仅卖家 id="+fmt.Sprint(product.SellerID)+" 或管理员可查看商品 id="+fmt.Sprint(productID)+" 的审核记录", nil)
	}
	list, err := s.reviewRepo.ListByProduct(productID)
	if err != nil {
		return nil, fmt.Errorf("list review history of product %d: %w", productID, err)
	}
	s.logger.Info(constants.LogProductReviewHistory, "product_id", productID, "operator", viewerID, "count", len(list))
	vos := toReviewVOList(list)
	return &dto.ProductReviewListResponse{List: vos, Total: int64(len(vos)), Page: 1, Size: len(vos)}, nil
}

// runTx 运行数据库事务；db 为空（单元测试内存仓储装配）时以 nil tx 直接执行。
func (s *ProductReviewService) runTx(fn func(tx *gorm.DB) error) error {
	if s.db == nil {
		return fn(nil)
	}
	return s.db.Transaction(fn)
}

// toReviewVOList model → 视图对象列表。
func toReviewVOList(list []model.ProductReview) []dto.ProductReviewVO {
	vos := make([]dto.ProductReviewVO, 0, len(list))
	for i := range list {
		rv := list[i]
		vo := dto.ProductReviewVO{
			ID:         rv.ID,
			ProductID:  rv.ProductID,
			SellerID:   rv.SellerID,
			Round:      rv.Round,
			Status:     rv.Status,
			Reason:     rv.Reason,
			ReviewerID: rv.ReviewerID,
			CreatedAt:  util.FormatTime(rv.CreatedAt),
			ReviewedAt: util.FormatTimePtr(rv.ReviewedAt),
		}
		if rv.Reviewer != nil {
			vo.Reviewer = dto.FromUser(rv.Reviewer)
		}
		if rv.Product != nil {
			p := dto.FromProduct(rv.Product, false)
			vo.Product = &p
		}
		vos = append(vos, vo)
	}
	return vos
}
