// 与后端 internal/constants/enums.go 对应的前端枚举（共享枚举/常量）。
export const OrderStatus = {
  PENDING_PAYMENT: 'pending_payment',
  PENDING_SHIPMENT: 'pending_shipment',
  SHIPPED: 'shipped',
  RECEIVED: 'received',
  COMPLETED: 'completed',
  CANCELLED: 'cancelled'
} as const

export const OrderStatusText: Record<string, string> = {
  [OrderStatus.PENDING_PAYMENT]: '待付款',
  [OrderStatus.PENDING_SHIPMENT]: '待发货',
  [OrderStatus.SHIPPED]: '已发货',
  [OrderStatus.RECEIVED]: '已收货',
  [OrderStatus.COMPLETED]: '已完成',
  [OrderStatus.CANCELLED]: '已取消'
}

export const OrderStatusTag: Record<string, string> = {
  [OrderStatus.PENDING_PAYMENT]: 'warning',
  [OrderStatus.PENDING_SHIPMENT]: 'primary',
  [OrderStatus.SHIPPED]: 'info',
  [OrderStatus.RECEIVED]: 'success',
  [OrderStatus.COMPLETED]: 'success',
  [OrderStatus.CANCELLED]: 'danger'
}

export const ProductCondition = {
  BRAND_NEW: 'brand_new',
  ALMOST_NEW: 'almost_new',
  LIGHTLY_USED: 'lightly_used',
  OBVIOUSLY_USED: 'obviously_used'
} as const

export const ProductConditionText: Record<string, string> = {
  [ProductCondition.BRAND_NEW]: '全新',
  [ProductCondition.ALMOST_NEW]: '几乎全新',
  [ProductCondition.LIGHTLY_USED]: '轻微使用',
  [ProductCondition.OBVIOUSLY_USED]: '明显使用'
}

export const ProductCategory = {
  DIGITAL: 'digital',
  CLOTHING: 'clothing',
  BOOKS: 'books',
  HOME: 'home',
  SPORTS: 'sports',
  OTHER: 'other'
} as const

export const ProductCategoryText: Record<string, string> = {
  [ProductCategory.DIGITAL]: '数码',
  [ProductCategory.CLOTHING]: '服饰',
  [ProductCategory.BOOKS]: '图书',
  [ProductCategory.HOME]: '家居',
  [ProductCategory.SPORTS]: '运动',
  [ProductCategory.OTHER]: '其他'
}

export const ProductStatus = {
  ON_SALE: 'on_sale',
  SOLD: 'sold',
  OFF_SHELF: 'off_shelf'
} as const

export const ProductStatusText: Record<string, string> = {
  [ProductStatus.ON_SALE]: '在售',
  [ProductStatus.SOLD]: '已售出',
  [ProductStatus.OFF_SHELF]: '已下架'
}

export const UserRole = {
  USER: 'user',
  ADMIN: 'admin'
} as const

export const UserRoleText: Record<string, string> = {
  [UserRole.USER]: '普通用户',
  [UserRole.ADMIN]: '管理员'
}

export const ReviewRating = {
  GOOD: 'good',
  NEUTRAL: 'neutral',
  BAD: 'bad'
} as const

export const ReviewRatingText: Record<string, string> = {
  [ReviewRating.GOOD]: '好评',
  [ReviewRating.NEUTRAL]: '中评',
  [ReviewRating.BAD]: '差评'
}
