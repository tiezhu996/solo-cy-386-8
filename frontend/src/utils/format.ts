// format.ts 前端格式化工具（与后端 util/formatters.go 对应）。
export function formatTime(value?: string | null): string {
  if (!value) return '-'
  return value
}

export function formatPrice(value?: number | null): string {
  if (value === undefined || value === null) return '0.00'
  return Number(value).toFixed(2)
}

export function formatCondition(text: string): string {
  const map: Record<string, string> = {
    brand_new: '全新',
    almost_new: '几乎全新',
    lightly_used: '轻微使用',
    obviously_used: '明显使用'
  }
  return map[text] ?? text
}

export function formatCategory(text: string): string {
  const map: Record<string, string> = {
    digital: '数码',
    clothing: '服饰',
    books: '图书',
    home: '家居',
    sports: '运动',
    other: '其他'
  }
  return map[text] ?? text
}

export function formatOrderStatus(text: string): string {
  const map: Record<string, string> = {
    pending_payment: '待付款',
    pending_shipment: '待发货',
    shipped: '已发货',
    received: '已收货',
    completed: '已完成',
    cancelled: '已取消'
  }
  return map[text] ?? text
}

export function formatProductStatus(text: string): string {
  const map: Record<string, string> = {
    on_sale: '在售',
    sold: '已售出',
    off_shelf: '已下架'
  }
  return map[text] ?? text
}

export function formatRating(text: string): string {
  const map: Record<string, string> = {
    good: '好评',
    neutral: '中评',
    bad: '差评'
  }
  return map[text] ?? text
}

export function formatRole(text: string): string {
  const map: Record<string, string> = {
    admin: '管理员',
    user: '普通用户'
  }
  return map[text] ?? text
}
