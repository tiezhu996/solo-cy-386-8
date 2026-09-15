import request from '../utils/request'
import type { OrderVO, PageResult } from './types'

export function createOrder(data: { product_id: number; address_id: number; quantity?: number; remark?: string }) {
  return request.post('/orders', data)
}

export function listOrders(params: Record<string, unknown>) {
  return request.get('/orders', { params })
}

export function getOrder(id: number) {
  return request.get(`/orders/${id}`)
}

export function payOrder(id: number) {
  return request.post(`/orders/${id}/pay`)
}

export function shipOrder(id: number) {
  return request.post(`/orders/${id}/ship`)
}

export function receiveOrder(id: number) {
  return request.post(`/orders/${id}/receive`)
}

export function completeOrder(id: number) {
  return request.post(`/orders/${id}/complete`)
}

export function cancelOrder(id: number) {
  return request.post(`/orders/${id}/cancel`)
}

export type { OrderVO, PageResult }
