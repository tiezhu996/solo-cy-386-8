import request from '../utils/request'
import type { CartItemVO } from './types'

export function getCart() {
  return request.get('/cart')
}

export function addToCart(productId: number, quantity = 1) {
  return request.post('/cart', { product_id: productId, quantity })
}

export function updateCartItem(id: number, data: { quantity?: number; selected?: boolean }) {
  return request.put(`/cart/${id}`, data)
}

export function removeCartItem(id: number) {
  return request.delete(`/cart/${id}`)
}

export type { CartItemVO }
