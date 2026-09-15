package service

import (
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/marketpal/marketpal/internal/constants"
	"github.com/marketpal/marketpal/internal/dto"
	"github.com/marketpal/marketpal/internal/model"
	"github.com/marketpal/marketpal/internal/repository"
	"github.com/marketpal/marketpal/internal/util"
	"gorm.io/gorm"
)

// OrderService 订单业务服务：状态机与多步写事务（SELECT ... FOR UPDATE 防超卖）。
type OrderService struct {
	orderRepo   repository.OrderRepository
	productRepo repository.ProductRepository
	addressRepo repository.AddressRepository
	cartRepo    repository.CartRepository
	db          *gorm.DB
	logger      *slog.Logger
}

// NewOrderService 构造订单服务。
func NewOrderService(db *gorm.DB, orderRepo repository.OrderRepository, productRepo repository.ProductRepository, addressRepo repository.AddressRepository, cartRepo repository.CartRepository, logger *slog.Logger) *OrderService {
	return &OrderService{db: db, orderRepo: orderRepo, productRepo: productRepo, addressRepo: addressRepo, cartRepo: cartRepo, logger: logger}
}

// Create 下单（事务）：锁商品 → 校验在售 → 校验地址归属 → 创建订单 → 商品置为已售。
func (s *OrderService) Create(buyerID uint, req dto.OrderCreateRequest) (*model.Order, error) {
	quantity := req.Quantity
	if quantity < 1 {
		quantity = 1
	}
	var order *model.Order
	err := s.db.Transaction(func(tx *gorm.DB) error {
		product, err := s.productRepo.GetByIDForUpdate(tx, req.ProductID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return utilAppError(constants.CodeProductNotFound, "下单失败：商品 id="+fmt.Sprint(req.ProductID)+" 不存在", err)
			}
			return fmt.Errorf("lock product %d: %w", req.ProductID, err)
		}
		if product.Status != constants.ProductStatusOnSale {
			return utilAppError(constants.CodeProductSold, "下单失败：商品 id="+fmt.Sprint(req.ProductID)+" 当前状态为 "+product.Status+"，无法购买", nil)
		}
		if product.SellerID == buyerID {
			return utilAppError(constants.CodeConflict, "下单失败：不能购买自己发布的商品（卖家 id="+fmt.Sprint(buyerID)+"）", nil)
		}
		addr, err := s.addressRepo.GetByIDAndUserTx(tx, req.AddressID, buyerID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return utilAppError(constants.CodeAddressNotFound, "下单失败：收货地址 id="+fmt.Sprint(req.AddressID)+" 不存在或不属于买家 id="+fmt.Sprint(buyerID), err)
			}
			return fmt.Errorf("get address %d for order: %w", req.AddressID, err)
		}
		_ = addr
		order = &model.Order{
			OrderNo:    genOrderNo(),
			BuyerID:    buyerID,
			SellerID:   product.SellerID,
			ProductID:  product.ID,
			AddressID:  req.AddressID,
			Quantity:   quantity,
			TotalPrice: product.Price * float64(quantity),
			Status:     constants.OrderStatusPendingPayment,
			Remark:     req.Remark,
		}
		if err := s.orderRepo.CreateWithTx(tx, order); err != nil {
			return fmt.Errorf("create order buyer=%d product=%d: %w", buyerID, req.ProductID, err)
		}
		if err := s.productRepo.UpdateStatusForUpdate(tx, product.ID, constants.ProductStatusSold); err != nil {
			return fmt.Errorf("mark product %d sold: %w", product.ID, err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.logger.Info(constants.LogOrderCreated, "order_no", order.OrderNo, "buyer_id", buyerID, "seller_id", order.SellerID, "product_id", req.ProductID, "status", order.Status)
	return s.orderRepo.GetByID(order.ID)
}

// Pay 买家付款：待付款 → 待发货。
func (s *OrderService) Pay(buyerID, orderID uint) (*model.Order, error) {
	var order *model.Order
	err := s.db.Transaction(func(tx *gorm.DB) error {
		o, err := s.orderRepo.GetByIDForUpdate(tx, orderID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return utilAppError(constants.CodeOrderNotFound, "订单支付失败：订单 id="+fmt.Sprint(orderID)+" 不存在", err)
			}
			return fmt.Errorf("lock order %d: %w", orderID, err)
		}
		if o.BuyerID != buyerID {
			return utilAppError(constants.CodeNotOrderOwner, "订单支付失败：买家 id="+fmt.Sprint(buyerID)+" 无权操作订单 "+o.OrderNo, nil)
		}
		if !canTransition(o.Status, constants.OrderStatusPendingShipment) {
			return utilAppError(constants.CodeOrderStateInvalid, "订单支付失败：订单 "+o.OrderNo+" 状态 "+o.Status+" 不可流转到 "+constants.OrderStatusPendingShipment, nil)
		}
		now := time.Now()
		if err := s.orderRepo.UpdateStatusForUpdate(tx, orderID, constants.OrderStatusPendingShipment, map[string]interface{}{"paid_at": &now}); err != nil {
			return fmt.Errorf("pay order %d: %w", orderID, err)
		}
		o.Status = constants.OrderStatusPendingShipment
		order = o
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.logger.Info(constants.LogOrderPaid, "order_no", order.OrderNo, "buyer_id", buyerID, "status", order.Status)
	return order, nil
}

// Ship 卖家发货：待发货 → 已发货。
func (s *OrderService) Ship(sellerID, orderID uint) (*model.Order, error) {
	var order *model.Order
	err := s.db.Transaction(func(tx *gorm.DB) error {
		o, err := s.orderRepo.GetByIDForUpdate(tx, orderID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return utilAppError(constants.CodeOrderNotFound, "订单发货失败：订单 id="+fmt.Sprint(orderID)+" 不存在", err)
			}
			return fmt.Errorf("lock order %d: %w", orderID, err)
		}
		if o.SellerID != sellerID {
			return utilAppError(constants.CodeNotOrderOwner, "订单发货失败：卖家 id="+fmt.Sprint(sellerID)+" 无权操作订单 "+o.OrderNo, nil)
		}
		if !canTransition(o.Status, constants.OrderStatusShipped) {
			return utilAppError(constants.CodeOrderStateInvalid, "订单发货失败：订单 "+o.OrderNo+" 状态 "+o.Status+" 不可流转到 "+constants.OrderStatusShipped, nil)
		}
		now := time.Now()
		if err := s.orderRepo.UpdateStatusForUpdate(tx, orderID, constants.OrderStatusShipped, map[string]interface{}{"shipped_at": &now}); err != nil {
			return fmt.Errorf("ship order %d: %w", orderID, err)
		}
		o.Status = constants.OrderStatusShipped
		order = o
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.logger.Info(constants.LogOrderShipped, "order_no", order.OrderNo, "seller_id", sellerID, "status", order.Status)
	return order, nil
}

// Receive 买家确认收货：已发货 → 已收货。
func (s *OrderService) Receive(buyerID, orderID uint) (*model.Order, error) {
	var order *model.Order
	err := s.db.Transaction(func(tx *gorm.DB) error {
		o, err := s.orderRepo.GetByIDForUpdate(tx, orderID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return utilAppError(constants.CodeOrderNotFound, "确认收货失败：订单 id="+fmt.Sprint(orderID)+" 不存在", err)
			}
			return fmt.Errorf("lock order %d: %w", orderID, err)
		}
		if o.BuyerID != buyerID {
			return utilAppError(constants.CodeNotOrderOwner, "确认收货失败：买家 id="+fmt.Sprint(buyerID)+" 无权操作订单 "+o.OrderNo, nil)
		}
		if !canTransition(o.Status, constants.OrderStatusReceived) {
			return utilAppError(constants.CodeOrderStateInvalid, "确认收货失败：订单 "+o.OrderNo+" 状态 "+o.Status+" 不可流转到 "+constants.OrderStatusReceived, nil)
		}
		now := time.Now()
		if err := s.orderRepo.UpdateStatusForUpdate(tx, orderID, constants.OrderStatusReceived, map[string]interface{}{"received_at": &now}); err != nil {
			return fmt.Errorf("receive order %d: %w", orderID, err)
		}
		o.Status = constants.OrderStatusReceived
		order = o
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.logger.Info(constants.LogOrderReceived, "order_no", order.OrderNo, "buyer_id", buyerID, "status", order.Status)
	return order, nil
}

// Complete 交易完成：已收货 → 已完成（评价前置条件）。
func (s *OrderService) Complete(buyerID, orderID uint) (*model.Order, error) {
	var order *model.Order
	err := s.db.Transaction(func(tx *gorm.DB) error {
		o, err := s.orderRepo.GetByIDForUpdate(tx, orderID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return utilAppError(constants.CodeOrderNotFound, "完成交易失败：订单 id="+fmt.Sprint(orderID)+" 不存在", err)
			}
			return fmt.Errorf("lock order %d: %w", orderID, err)
		}
		if o.BuyerID != buyerID {
			return utilAppError(constants.CodeNotOrderOwner, "完成交易失败：买家 id="+fmt.Sprint(buyerID)+" 无权操作订单 "+o.OrderNo, nil)
		}
		if !canTransition(o.Status, constants.OrderStatusCompleted) {
			return utilAppError(constants.CodeOrderStateInvalid, "完成交易失败：订单 "+o.OrderNo+" 状态 "+o.Status+" 不可流转到 "+constants.OrderStatusCompleted, nil)
		}
		now := time.Now()
		if err := s.orderRepo.UpdateStatusForUpdate(tx, orderID, constants.OrderStatusCompleted, map[string]interface{}{"completed_at": &now}); err != nil {
			return fmt.Errorf("complete order %d: %w", orderID, err)
		}
		o.Status = constants.OrderStatusCompleted
		order = o
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.logger.Info(constants.LogOrderCompleted, "order_no", order.OrderNo, "buyer_id", buyerID, "status", order.Status)
	return order, nil
}

// Cancel 取消订单：买家取消待付款，卖家取消待发货。
func (s *OrderService) Cancel(userID, orderID uint, role string) (*model.Order, error) {
	var order *model.Order
	err := s.db.Transaction(func(tx *gorm.DB) error {
		o, err := s.orderRepo.GetByIDForUpdate(tx, orderID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return utilAppError(constants.CodeOrderNotFound, "取消订单失败：订单 id="+fmt.Sprint(orderID)+" 不存在", err)
			}
			return fmt.Errorf("lock order %d: %w", orderID, err)
		}
		if role == "buyer" && o.BuyerID != userID {
			return utilAppError(constants.CodeNotOrderOwner, "取消订单失败：买家 id="+fmt.Sprint(userID)+" 无权操作订单 "+o.OrderNo, nil)
		}
		if role == "seller" && o.SellerID != userID {
			return utilAppError(constants.CodeNotOrderOwner, "取消订单失败：卖家 id="+fmt.Sprint(userID)+" 无权操作订单 "+o.OrderNo, nil)
		}
		if !canTransition(o.Status, constants.OrderStatusCancelled) {
			return utilAppError(constants.CodeOrderStateInvalid, "取消订单失败：订单 "+o.OrderNo+" 状态 "+o.Status+" 不可取消", nil)
		}
		now := time.Now()
		if err := s.orderRepo.UpdateStatusForUpdate(tx, orderID, constants.OrderStatusCancelled, map[string]interface{}{"cancelled_at": &now}); err != nil {
			return fmt.Errorf("cancel order %d: %w", orderID, err)
		}
		// 订单取消后商品重新上架。
		o.Status = constants.OrderStatusCancelled
		if err := s.productRepo.UpdateStatusForUpdate(tx, o.ProductID, constants.ProductStatusOnSale); err != nil {
			return fmt.Errorf("restore product %d on sale: %w", o.ProductID, err)
		}
		order = o
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.logger.Info(constants.LogOrderCancelled, "order_no", order.OrderNo, "operator", userID, "status", order.Status)
	return order, nil
}

// List 订单列表（buyer/seller 双视角，复用同一 service 方法）。
func (s *OrderService) List(userID uint, q dto.OrderQuery) (*dto.OrderListResponse, error) {
	page, pageSize := q.Page, q.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	var orders []model.Order
	var total int64
	var err error
	if q.Role == "seller" {
		orders, total, err = s.orderRepo.ListBySeller(userID, q.Status, page, pageSize)
	} else {
		orders, total, err = s.orderRepo.ListByBuyer(userID, q.Status, page, pageSize)
	}
	if err != nil {
		return nil, fmt.Errorf("list orders user=%d role=%s: %w", userID, q.Role, err)
	}
	list := make([]dto.OrderVO, 0, len(orders))
	for i := range orders {
		list = append(list, toOrderVO(&orders[i]))
	}
	return &dto.OrderListResponse{List: list, Total: total, Page: page, Size: pageSize}, nil
}

// GetDetail 订单详情（买家/卖家均可查看）。
func (s *OrderService) GetDetail(userID, orderID uint) (*model.Order, error) {
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, utilAppError(constants.CodeOrderNotFound, "订单详情查询失败：订单 id="+fmt.Sprint(orderID)+" 不存在", err)
		}
		return nil, fmt.Errorf("get order detail %d: %w", orderID, err)
	}
	if order.BuyerID != userID && order.SellerID != userID {
		return nil, utilAppError(constants.CodeNotOrderOwner, "订单详情查询失败：用户 id="+fmt.Sprint(userID)+" 非订单 "+order.OrderNo+" 的归属人", nil)
	}
	return order, nil
}

// canTransition 状态机校验（与 constants.OrderStatusTransitions 对应）。
func canTransition(from, to string) bool {
	for _, s := range constants.OrderStatusTransitions[to] {
		if s == from {
			return true
		}
	}
	return false
}

// genOrderNo 生成订单号：yyyyMMddHHmmss + 6 位随机数。
func genOrderNo() string {
	return fmt.Sprintf("%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// toOrderVO model.Order → 视图对象。
func toOrderVO(o *model.Order) dto.OrderVO {
	vo := dto.OrderVO{
		ID:         o.ID,
		OrderNo:    o.OrderNo,
		BuyerID:    o.BuyerID,
		SellerID:   o.SellerID,
		ProductID:  o.ProductID,
		AddressID:  o.AddressID,
		Quantity:   o.Quantity,
		TotalPrice: o.TotalPrice,
		Status:     o.Status,
		Remark:     o.Remark,
		CreatedAt:  util.FormatTime(o.CreatedAt),
	}
	if o.PaidAt != nil {
		t := util.FormatTime(*o.PaidAt)
		vo.PaidAt = &t
	}
	if o.ShippedAt != nil {
		t := util.FormatTime(*o.ShippedAt)
		vo.ShippedAt = &t
	}
	if o.ReceivedAt != nil {
		t := util.FormatTime(*o.ReceivedAt)
		vo.ReceivedAt = &t
	}
	if o.Product != nil {
		prod := dto.FromProduct(o.Product, false)
		vo.Product = &prod
	}
	if o.Address != nil {
		addr := dto.FromAddress(o.Address)
		vo.Address = &addr
	}
	if o.Buyer != nil {
		vo.Buyer = dto.FromUser(o.Buyer)
	}
	if o.Seller != nil {
		vo.Seller = dto.FromUser(o.Seller)
	}
	return vo
}
