package dto

// ProductCreateRequest 商品发布入参。
type ProductCreateRequest struct {
	Title         string   `json:"title" binding:"required,min=2,max=128"`
	Description   string   `json:"description" binding:"required,min=5"`
	OriginalPrice float64  `json:"original_price" binding:"required,gte=0"`
	Price         float64  `json:"price" binding:"required,gt=0"`
	Condition     string   `json:"condition" binding:"required,oneof=brand_new almost_new lightly_used obviously_used"`
	Category      string   `json:"category" binding:"required,oneof=digital clothing books home sports other"`
	Images        []string `json:"images" binding:"omitempty,max=9"`
}

// ProductUpdateRequest 商品更新入参。
type ProductUpdateRequest struct {
	Title         *string  `json:"title" binding:"omitempty,min=2,max=128"`
	Description   *string  `json:"description" binding:"omitempty,min=5"`
	OriginalPrice *float64 `json:"original_price" binding:"omitempty,gte=0"`
	Price         *float64 `json:"price" binding:"omitempty,gt=0"`
	Condition     *string  `json:"condition" binding:"omitempty,oneof=brand_new almost_new lightly_used obviously_used"`
	Category      *string  `json:"category" binding:"omitempty,oneof=digital clothing books home sports other"`
	Images        []string `json:"images" binding:"omitempty,max=9"`
}

// ProductQuery 商品列表查询入参（支持关键词/分类/价格区间/成色/排序）。
type ProductQuery struct {
	Keyword    string  `form:"keyword"`
	Category   string  `form:"category" binding:"omitempty,oneof=digital clothing books home sports other"`
	Condition  string  `form:"condition" binding:"omitempty,oneof=brand_new almost_new lightly_used obviously_used"`
	MinPrice   float64 `form:"min_price" binding:"omitempty,gte=0"`
	MaxPrice   float64 `form:"max_price" binding:"omitempty,gte=0"`
	SortBy     string  `form:"sort_by" binding:"omitempty,oneof=price price_desc time time_desc"`
	Status     string  `form:"status" binding:"omitempty,oneof=on_sale sold off_shelf"`
	Page       int     `form:"page" binding:"omitempty,min=1"`
	PageSize   int     `form:"page_size" binding:"omitempty,min=1,max=50"`
}

// ProductVO 商品视图对象。
type ProductVO struct {
	ID            uint      `json:"id"`
	SellerID      uint      `json:"seller_id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	OriginalPrice float64   `json:"original_price"`
	Price         float64   `json:"price"`
	Condition     string    `json:"condition"`
	Category      string    `json:"category"`
	Images        []string  `json:"images"`
	Status        string    `json:"status"`
	ViewCount     int       `json:"view_count"`
	FavoriteCount int       `json:"favorite_count"`
	CreatedAt     string    `json:"created_at"`
	Seller        *UserVO   `json:"seller,omitempty"`
	IsFavorite    bool      `json:"is_favorite"`
}

// ProductListResponse 商品列表分页响应。
type ProductListResponse struct {
	List  []ProductVO `json:"list"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"page_size"`
}
