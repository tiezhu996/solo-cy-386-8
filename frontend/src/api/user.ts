import request from '../utils/request'
import type { PageResult, UserVO } from './types'

export function register(data: { username: string; password: string; nickname: string; email?: string; phone?: string }) {
  return request.post('/auth/register', data)
}

export function login(data: { username: string; password: string }) {
  return request.post('/auth/login', data)
}

export function getProfile() {
  return request.get('/users/me')
}

export function updateProfile(data: { nickname?: string; email?: string; phone?: string; avatar?: string }) {
  return request.put('/users/me', data)
}

export function listUsers(params: { page?: number; page_size?: number } = {}) {
  return request.get('/users', { params })
}

export function updateRole(id: number, role: string) {
  return request.put(`/users/${id}/role`, { role })
}

export type { UserVO }
