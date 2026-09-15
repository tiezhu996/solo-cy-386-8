package dto

// CartAddRequest 加入购物车入参。
type CartAddRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"omitempty,min=1,max=99"`
}

// CartUpdateRequest 更新购物车条目入参。
type CartUpdateRequest struct {
	Quantity *int  `json:"quantity" binding:"omitempty,min=1,max=99"`
	Selected *bool `json:"selected"`
}

// CartItemVO 购物车条目视图对象。
type CartItemVO struct {
	ID        uint       `json:"id"`
	ProductID uint       `json:"product_id"`
	Quantity  int        `json:"quantity"`
	Selected  bool       `json:"selected"`
	Product   *ProductVO `json:"product,omitempty"`
}

// CartResponse 购物车响应。
type CartResponse struct {
	Items []CartItemVO `json:"items"`
	Total float64      `json:"total"`
}
