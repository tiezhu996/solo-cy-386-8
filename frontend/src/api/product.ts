import request from '../utils/request'
import type { PageResult, ProductVO } from './types'

export function listProducts(params: Record<string, unknown>) {
  return request.get('/products', { params })
}

export function getProduct(id: number) {
  return request.get(`/products/${id}`)
}

export function createProduct(data: Record<string, unknown>) {
  return request.post('/products', data)
}

export function updateProduct(id: number, data: Record<string, unknown>) {
  return request.put(`/products/${id}`, data)
}

export function offShelfProduct(id: number) {
  return request.post(`/products/${id}/off-shelf`)
}

export function myProducts(params: Record<string, unknown>) {
  return request.get('/products/mine', { params })
}

export function favoriteProduct(id: number) {
  return request.post(`/products/${id}/favorite`)
}

export function unfavoriteProduct(id: number) {
  return request.delete(`/products/${id}/favorite`)
}

export function myFavorites(params: Record<string, unknown>) {
  return request.get('/favorites', { params })
}

export function uploadImage(file: File) {
  const form = new FormData()
  form.append('file', file)
  return request.post('/upload', form, { headers: { 'Content-Type': 'multipart/form-data' } })
}

export type { ProductVO }
