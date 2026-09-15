package service

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/marketpal/marketpal/internal/constants"
	"github.com/marketpal/marketpal/internal/dto"
	"github.com/marketpal/marketpal/internal/model"
	"github.com/marketpal/marketpal/internal/repository"
)

// CartService 购物车业务服务。
type CartService struct {
	cartRepo    repository.CartRepository
	productRepo repository.ProductRepository
	logger      *slog.Logger
}

// NewCartService 构造购物车服务。
func NewCartService(cartRepo repository.CartRepository, productRepo repository.ProductRepository, logger *slog.Logger) *CartService {
	return &CartService{cartRepo: cartRepo, productRepo: productRepo, logger: logger}
}

// Add 加入购物车（已存在则累加数量）。
func (s *CartService) Add(userID uint, req dto.CartAddRequest) (*model.CartItem, error) {
	product, err := s.productRepo.GetByID(req.ProductID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, utilAppError(constants.CodeProductNotFound, "加入购物车失败：商品 id="+fmt.Sprint(req.ProductID)+" 不存在", err)
		}
		return nil, fmt.Errorf("get product %d for cart: %w", req.ProductID, err)
	}
	if product.SellerID == userID {
		return nil, utilAppError(constants.CodeConflict, "加入购物车失败：不能购买自己发布的商品（卖家 id="+fmt.Sprint(userID)+"）", nil)
	}
	if product.Status != constants.ProductStatusOnSale {
		return nil, utilAppError(constants.CodeProductOffShelf, "加入购物车失败：商品 id="+fmt.Sprint(req.ProductID)+" 当前状态为 "+product.Status+" 不可购买", nil)
	}
	qty := req.Quantity
	if qty < 1 {
		qty = 1
	}
	if existing, err := s.cartRepo.GetByUserAndProduct(userID, req.ProductID); err == nil {
		existing.Quantity += qty
		if err := s.cartRepo.Update(existing); err != nil {
			return nil, fmt.Errorf("update cart item %d: %w", existing.ID, err)
		}
		s.logger.Info(constants.LogCartUpdated, "user_id", userID, "cart_item_id", existing.ID, "quantity", existing.Quantity)
		return existing, nil
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("get cart user=%d product=%d: %w", userID, req.ProductID, err)
	}
	item := &model.CartItem{UserID: userID, ProductID: req.ProductID, Quantity: qty, Selected: true}
	if err := s.cartRepo.Create(item); err != nil {
		return nil, fmt.Errorf("create cart item user=%d product=%d: %w", userID, req.ProductID, err)
	}
	s.logger.Info(constants.LogCartAdded, "user_id", userID, "product_id", req.ProductID, "quantity", qty)
	return item, nil
}

// List 购物车列表（含商品信息与合计金额）。
func (s *CartService) List(userID uint) (*dto.CartResponse, error) {
	items, err := s.cartRepo.ListByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("list cart user=%d: %w", userID, err)
	}
	res := &dto.CartResponse{Items: []dto.CartItemVO{}, Total: 0}
	for i := range items {
		if items[i].Product == nil {
			continue
		}
		vo := dto.CartItemVO{
			ID:        items[i].ID,
			ProductID: items[i].ProductID,
			Quantity:  items[i].Quantity,
			Selected:  items[i].Selected,
		}
		prod := toProductVO(items[i].Product, false)
		vo.Product = &prod
		res.Items = append(res.Items, vo)
		if items[i].Selected {
			res.Total += items[i].Product.Price * float64(items[i].Quantity)
		}
	}
	return res, nil
}

// Update 更新购物车条目数量/选中态。
func (s *CartService) Update(userID, itemID uint, req dto.CartUpdateRequest) (*model.CartItem, error) {
	item, err := s.cartRepo.GetByID(itemID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, utilAppError(constants.CodeCartItemNotFound, "购物车更新失败：条目 id="+fmt.Sprint(itemID)+" 不存在", err)
		}
		return nil, fmt.Errorf("get cart item %d: %w", itemID, err)
	}
	if item.UserID != userID {
		return nil, utilAppError(constants.CodeForbidden, "购物车更新失败：条目 id="+fmt.Sprint(itemID)+" 不属于用户 id="+fmt.Sprint(userID), nil)
	}
	if req.Quantity != nil {
		item.Quantity = *req.Quantity
	}
	if req.Selected != nil {
		item.Selected = *req.Selected
	}
	if err := s.cartRepo.Update(item); err != nil {
		return nil, fmt.Errorf("update cart item %d: %w", itemID, err)
	}
	s.logger.Info(constants.LogCartUpdated, "user_id", userID, "cart_item_id", itemID, "quantity", item.Quantity)
	return item, nil
}

// Delete 删除购物车条目。
func (s *CartService) Delete(userID, itemID uint) error {
	if err := s.cartRepo.Delete(itemID, userID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return utilAppError(constants.CodeCartItemNotFound, "购物车删除失败：条目 id="+fmt.Sprint(itemID)+" 不存在或不属于用户 id="+fmt.Sprint(userID), err)
		}
		return fmt.Errorf("delete cart item %d: %w", itemID, err)
	}
	s.logger.Info(constants.LogCartRemoved, "user_id", userID, "cart_item_id", itemID)
	return nil
}

// Clear 清空购物车（下单成功后复用）。
func (s *CartService) Clear(userID uint) error {
	if err := s.cartRepo.Clear(userID); err != nil {
		return fmt.Errorf("clear cart user=%d: %w", userID, err)
	}
	return nil
}
