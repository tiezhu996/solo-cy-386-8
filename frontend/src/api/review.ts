import request from '../utils/request'
import type { PageResult, ReviewVO } from './types'

export function createReview(data: { order_id: number; rating: string; content?: string }) {
  return request.post('/reviews', data)
}

export function listReviews(params: Record<string, unknown> = {}) {
  return request.get('/reviews', { params })
}

export type { ReviewVO, PageResult }
