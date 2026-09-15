import request from '../utils/request'
import type { AddressVO } from './types'

export function listAddresses() {
  return request.get('/addresses')
}

export function createAddress(data: Partial<AddressVO>) {
  return request.post('/addresses', data)
}

export function updateAddress(id: number, data: Partial<AddressVO>) {
  return request.put(`/addresses/${id}`, data)
}

export function deleteAddress(id: number) {
  return request.delete(`/addresses/${id}`)
}
