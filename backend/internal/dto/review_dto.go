package dto

// ReviewCreateRequest 创建评价入参。
type ReviewCreateRequest struct {
	OrderID   uint   `json:"order_id" binding:"required"`
	Rating    string `json:"rating" binding:"required,oneof=good neutral bad"`
	Content   string `json:"content" binding:"omitempty,max=500"`
}

// ReviewQuery 评价列表查询入参。
type ReviewQuery struct {
	UserID    uint `form:"user_id"`
	ProductID uint `form:"product_id"`
	Page      int  `form:"page" binding:"omitempty,min=1"`
	PageSize  int  `form:"page_size" binding:"omitempty,min=1,max=50"`
}

// ReviewVO 评价视图对象。
type ReviewVO struct {
	ID         uint   `json:"id"`
	OrderID    uint   `json:"order_id"`
	ProductID  uint   `json:"product_id"`
	ReviewerID uint   `json:"reviewer_id"`
	RevieweeID uint   `json:"reviewee_id"`
	Rating     string `json:"rating"`
	Content    string `json:"content"`
	CreatedAt  string `json:"created_at"`
	Reviewer   *UserVO `json:"reviewer,omitempty"`
}

// ReviewListResponse 评价列表分页响应。
type ReviewListResponse struct {
	List  []ReviewVO `json:"list"`
	Total int64      `json:"total"`
	Page  int        `json:"page"`
	Size  int        `json:"page_size"`
}
