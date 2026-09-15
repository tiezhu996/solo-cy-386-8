package dto

// AddressCreateRequest 创建收货地址入参。
type AddressCreateRequest struct {
	ReceiverName string `json:"receiver_name" binding:"required,min=1,max=32"`
	Phone        string `json:"phone" binding:"required,max=20"`
	Province     string `json:"province" binding:"required,max=32"`
	City         string `json:"city" binding:"required,max=32"`
	District     string `json:"district" binding:"omitempty,max=32"`
	Detail       string `json:"detail" binding:"required,max=255"`
	IsDefault    bool   `json:"is_default"`
}

// AddressUpdateRequest 更新收货地址入参。
type AddressUpdateRequest struct {
	ReceiverName *string `json:"receiver_name" binding:"omitempty,min=1,max=32"`
	Phone        *string `json:"phone" binding:"omitempty,max=20"`
	Province     *string `json:"province" binding:"omitempty,max=32"`
	City         *string `json:"city" binding:"omitempty,max=32"`
	District     *string `json:"district" binding:"omitempty,max=32"`
	Detail       *string `json:"detail" binding:"omitempty,max=255"`
	IsDefault    *bool   `json:"is_default"`
}

// AddressVO 收货地址视图对象。
type AddressVO struct {
	ID           uint   `json:"id"`
	ReceiverName string `json:"receiver_name"`
	Phone        string `json:"phone"`
	Province     string `json:"province"`
	City         string `json:"city"`
	District     string `json:"district"`
	Detail       string `json:"detail"`
	IsDefault    bool   `json:"is_default"`
}
