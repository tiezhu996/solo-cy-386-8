// Package constants centralizes business enums, error codes, messages and log templates.
package constants

// OrderStatus 订单状态枚举（前端 constants/order.ts 同步维护）。
const (
	OrderStatusPendingPayment  string = "pending_payment"  // 待付款
	OrderStatusPendingShipment string = "pending_shipment" // 待发货
	OrderStatusShipped         string = "shipped"          // 已发货
	OrderStatusReceived        string = "received"         // 已收货
	OrderStatusCompleted       string = "completed"        // 已完成
	OrderStatusCancelled       string = "cancelled"        // 已取消
)

// ProductCondition 商品成色枚举。
const (
	ProductConditionBrandNew      string = "brand_new"      // 全新
	ProductConditionAlmostNew     string = "almost_new"     // 几乎全新
	ProductConditionLightlyUsed   string = "lightly_used"   // 轻微使用
	ProductConditionObviouslyUsed string = "obviously_used" // 明显使用
)

// ProductCategory 商品分类枚举。
const (
	ProductCategoryDigital  string = "digital"  // 数码
	ProductCategoryClothing string = "clothing" // 服饰
	ProductCategoryBooks    string = "books"    // 图书
	ProductCategoryHome     string = "home"     // 家居
	ProductCategorySports   string = "sports"   // 运动
	ProductCategoryOther    string = "other"    // 其他
)

// ProductStatus 商品上下架状态枚举。
const (
	ProductStatusOnSale   string = "on_sale"   // 在售
	ProductStatusSold     string = "sold"      // 已售出
	ProductStatusOffShelf string = "off_shelf" // 已下架
)

// ProductReviewStatus 商品平台审核状态枚举（与前端 constants/index.ts 同步维护）。
const (
	ProductReviewPending  string = "pending_review" // 待审核（新发布 / 修改重提 / 复审）
	ProductReviewApproved string = "approved"       // 审核通过（仅通过商品可进入大厅/搜索/加购/下单）
	ProductReviewRejected string = "rejected"       // 审核驳回（卖家可修改后重新提交）
)

// UserRole 用户角色枚举。
const (
	UserRoleUser  string = "user"  // 普通用户
	UserRoleAdmin string = "admin" // 管理员
)

// ReviewRating 评价等级枚举。
const (
	ReviewRatingGood    string = "good"    // 好评
	ReviewRatingNeutral string = "neutral" // 中评
	ReviewRatingBad     string = "bad"     // 差评
)

// OrderStatusTransitions 订单状态机：允许的流转映射（新状态 → 允许的前置状态集合）。
var OrderStatusTransitions = map[string][]string{
	OrderStatusPendingPayment:  {OrderStatusPendingPayment},
	OrderStatusPendingShipment: {OrderStatusPendingPayment},
	OrderStatusShipped:         {OrderStatusPendingShipment},
	OrderStatusReceived:        {OrderStatusShipped},
	OrderStatusCompleted:       {OrderStatusReceived},
	OrderStatusCancelled:       {OrderStatusPendingPayment, OrderStatusPendingShipment},
}

// ValidOrderStatus 校验订单状态值是否合法。
func ValidOrderStatus(status string) bool {
	switch status {
	case OrderStatusPendingPayment, OrderStatusPendingShipment, OrderStatusShipped,
		OrderStatusReceived, OrderStatusCompleted, OrderStatusCancelled:
		return true
	}
	return false
}

// ValidProductCondition 校验成色值是否合法。
func ValidProductCondition(cond string) bool {
	switch cond {
	case ProductConditionBrandNew, ProductConditionAlmostNew, ProductConditionLightlyUsed, ProductConditionObviouslyUsed:
		return true
	}
	return false
}

// ValidProductCategory 校验分类值是否合法。
func ValidProductCategory(cat string) bool {
	switch cat {
	case ProductCategoryDigital, ProductCategoryClothing, ProductCategoryBooks,
		ProductCategoryHome, ProductCategorySports, ProductCategoryOther:
		return true
	}
	return false
}

// ValidProductStatus 校验商品状态值是否合法。
func ValidProductStatus(status string) bool {
	switch status {
	case ProductStatusOnSale, ProductStatusSold, ProductStatusOffShelf:
		return true
	}
	return false
}

// ProductReviewTransitions 商品审核状态机：允许的流转映射（新状态 → 允许的前置状态集合）。
// 卖家：新发布（无前置）进入待审、修改已上架（approved）重新送审、驳回（rejected）后修改重提；
// 管理员：仅能在待审（pending_review）时给出通过/驳回。待审中重复提交/重复审核不改变状态。
var ProductReviewTransitions = map[string][]string{
	ProductReviewPending:  {"", ProductReviewApproved, ProductReviewRejected},
	ProductReviewApproved: {ProductReviewPending},
	ProductReviewRejected: {ProductReviewPending},
}

// ValidProductReviewStatus 校验商品审核状态值是否合法。
func ValidProductReviewStatus(status string) bool {
	switch status {
	case ProductReviewPending, ProductReviewApproved, ProductReviewRejected:
		return true
	}
	return false
}

// CanReviewTransition 审核状态机校验（service 状态机、并发审核幂等共用）。
func CanReviewTransition(from, to string) bool {
	for _, s := range ProductReviewTransitions[to] {
		if s == from {
			return true
		}
	}
	return false
}

// ValidUserRole 校验角色值是否合法。
func ValidUserRole(role string) bool {
	return role == UserRoleUser || role == UserRoleAdmin
}

// ValidReviewRating 校验评价等级是否合法。
func ValidReviewRating(rating string) bool {
	switch rating {
	case ReviewRatingGood, ReviewRatingNeutral, ReviewRatingBad:
		return true
	}
	return false
}
