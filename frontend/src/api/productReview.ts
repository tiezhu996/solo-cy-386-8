import request from '../utils/request'
import type { PageResult, ProductReviewVO } from './types'

// 管理员待审队列 / 审核记录查询。
export function listProductReviews(params: Record<string, unknown>) {
  return request.get('/product-reviews', { params })
}

// 管理员通过 / 驳回；驳回必须填写原因。
export function decideProductReview(id: number, data: { action: 'approve' | 'reject'; reason?: string }) {
  return request.post(`/product-reviews/${id}/decision`, data)
}

// 单个商品的审核历史（卖家本人或管理员）。
export function productReviewHistory(productId: number) {
  return request.get(`/products/${productId}/reviews`)
}

export type { ProductReviewVO, PageResult }
