package dto

// OrderCreateRequest 下单入参。
type OrderCreateRequest struct {
	ProductID uint   `json:"product_id" binding:"required"`
	AddressID uint   `json:"address_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"omitempty,min=1,max=99"`
	Remark    string `json:"remark" binding:"omitempty,max=255"`
}

// OrderActionRequest 订单状态流转入参（发货等）。
type OrderActionRequest struct {
	Remark string `json:"remark" binding:"omitempty,max=255"`
}

// OrderQuery 订单列表查询入参。
type OrderQuery struct {
	Status   string `form:"status" binding:"omitempty,oneof=pending_payment pending_shipment shipped received completed cancelled"`
	Role     string `form:"role" binding:"omitempty,oneof=buyer seller"`
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=50"`
}

// OrderVO 订单视图对象。
type OrderVO struct {
	ID         uint        `json:"id"`
	OrderNo    string      `json:"order_no"`
	BuyerID    uint        `json:"buyer_id"`
	SellerID   uint        `json:"seller_id"`
	ProductID  uint        `json:"product_id"`
	AddressID  uint        `json:"address_id"`
	Quantity   int         `json:"quantity"`
	TotalPrice float64     `json:"total_price"`
	Status     string      `json:"status"`
	Remark     string      `json:"remark"`
	PaidAt     *string     `json:"paid_at,omitempty"`
	ShippedAt  *string     `json:"shipped_at,omitempty"`
	ReceivedAt *string     `json:"received_at,omitempty"`
	CreatedAt  string      `json:"created_at"`
	Product    *ProductVO  `json:"product,omitempty"`
	Address    *AddressVO  `json:"address,omitempty"`
	Buyer      *UserVO     `json:"buyer,omitempty"`
	Seller     *UserVO     `json:"seller,omitempty"`
}

// OrderListResponse 订单列表分页响应。
type OrderListResponse struct {
	List  []OrderVO `json:"list"`
	Total int64     `json:"total"`
	Page  int       `json:"page"`
	Size  int       `json:"page_size"`
}
