package handler

import (
	"github.com/marketpal/marketpal/internal/dto"
	"github.com/marketpal/marketpal/internal/model"
	"github.com/marketpal/marketpal/internal/util"
)

// orderVO 订单实体转视图对象（handler 层复用 service 返回的 model）。
func orderVO(o *model.Order) *dto.OrderVO {
	vo := &dto.OrderVO{
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
