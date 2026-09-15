import request from '../utils/request'
import type { ConversationVO, MessageVO } from './types'

export function sendMessage(data: { receiver_id: number; product_id?: number; content: string }) {
  return request.post('/messages', data)
}

export function listConversations() {
  return request.get('/messages/conversations')
}

export function listConversation(peerId: number, params: Record<string, unknown> = {}) {
  return request.get(`/messages/conversations/${peerId}`, { params })
}

export function markRead(peerId: number) {
  return request.put(`/messages/conversations/${peerId}/read`)
}

export function unreadCount() {
  return request.get('/messages/unread-count')
}

export type { ConversationVO, MessageVO }
