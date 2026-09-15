package dto

// AuditQuery 审计日志查询入参。
type AuditQuery struct {
	Module   string `form:"module"`
	Action   string `form:"action"`
	UserID   uint   `form:"user_id"`
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// AuditVO 审计日志视图对象。
type AuditVO struct {
	ID         uint   `json:"id"`
	UserID     uint   `json:"user_id"`
	Username   string `json:"username"`
	Module     string `json:"module"`
	Action     string `json:"action"`
	ResourceID string `json:"resource_id"`
	Detail     string `json:"detail"`
	IP         string `json:"ip"`
	RequestID  string `json:"request_id"`
	CreatedAt  string `json:"created_at"`
}

// AuditListResponse 审计日志分页响应。
type AuditListResponse struct {
	List  []AuditVO `json:"list"`
	Total int64     `json:"total"`
	Page  int       `json:"page"`
	Size  int       `json:"page_size"`
}
