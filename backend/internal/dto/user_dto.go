package dto

// RegisterRequest 注册入参。
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=64"`
	Nickname string `json:"nickname" binding:"required,min=1,max=32"`
	Email    string `json:"email" binding:"omitempty,email"`
	Phone    string `json:"phone" binding:"omitempty,max=20"`
}

// LoginRequest 登录入参。
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UpdateProfileRequest 更新个人资料入参。
type UpdateProfileRequest struct {
	Nickname string `json:"nickname" binding:"omitempty,min=1,max=32"`
	Email    string `json:"email" binding:"omitempty,email"`
	Phone    string `json:"phone" binding:"omitempty,max=20"`
	Avatar   string `json:"avatar" binding:"omitempty,max=512"`
}

// UpdateRoleRequest 管理员修改用户角色入参。
type UpdateRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=user admin"`
}

// UserVO 用户视图对象。
type UserVO struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	Nickname    string `json:"nickname"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Avatar      string `json:"avatar"`
	Role        string `json:"role"`
	CreditScore int    `json:"credit_score"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

// LoginResponse 登录响应。
type LoginResponse struct {
	Token string  `json:"token"`
	User  UserVO  `json:"user"`
}
