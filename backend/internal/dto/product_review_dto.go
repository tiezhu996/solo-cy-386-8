package dto

// ProductReviewDecisionRequest 管理员审核入参：通过或驳回；驳回必须填写原因。
type ProductReviewDecisionRequest struct {
	Action string `json:"action" binding:"required,oneof=approve reject"`
	Reason string `json:"reason" binding:"omitempty,max=500"`
}

// ProductReviewQuery 审核列表查询（管理员待审队列 / 卖家按状态筛选）。
type ProductReviewQuery struct {
	Status    string `form:"status" binding:"omitempty,oneof=pending_review approved rejected"`
	ProductID uint   `form:"product_id" binding:"omitempty,min=1"`
	SellerID  uint   `form:"seller_id" binding:"omitempty,min=1"`
	Page      int    `form:"page" binding:"omitempty,min=1"`
	PageSize  int    `form:"page_size" binding:"omitempty,min=1,max=50"`
}

// ProductReviewVO 审核记录视图对象（审核记录可回读）。
type ProductReviewVO struct {
	ID         uint       `json:"id"`
	ProductID  uint       `json:"product_id"`
	SellerID   uint       `json:"seller_id"`
	Round      int        `json:"round"`
	Status     string     `json:"status"`
	Reason     string     `json:"reason"`
	ReviewerID uint       `json:"reviewer_id"`
	Reviewer   *UserVO    `json:"reviewer,omitempty"`
	Product    *ProductVO `json:"product,omitempty"`
	CreatedAt  string     `json:"created_at"`
	ReviewedAt string     `json:"reviewed_at"`
}

// ProductReviewListResponse 审核记录分页响应。
type ProductReviewListResponse struct {
	List  []ProductReviewVO `json:"list"`
	Total int64             `json:"total"`
	Page  int               `json:"page"`
	Size  int               `json:"page_size"`
}
