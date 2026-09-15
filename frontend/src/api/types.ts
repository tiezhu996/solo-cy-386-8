// 全局类型定义。
export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

export interface UserVO {
  id: number
  username: string
  nickname: string
  email?: string
  phone?: string
  avatar?: string
  role: string
  credit_score: number
  status: string
  created_at: string
}

export interface ProductVO {
  id: number
  seller_id: number
  title: string
  description: string
  original_price: number
  price: number
  condition: string
  category: string
  images: string[]
  status: string
  view_count: number
  favorite_count: number
  created_at: string
  seller?: UserVO
  is_favorite?: boolean
}

export interface AddressVO {
  id: number
  receiver_name: string
  phone: string
  province: string
  city: string
  district?: string
  detail: string
  is_default: boolean
}

export interface OrderVO {
  id: number
  order_no: string
  buyer_id: number
  seller_id: number
  product_id: number
  address_id: number
  quantity: number
  total_price: number
  status: string
  remark?: string
  created_at: string
  product?: ProductVO
  address?: AddressVO
  buyer?: UserVO
  seller?: UserVO
}

export interface ConversationVO {
  peer_id: number
  peer_name: string
  peer_avatar?: string
  product_id?: number
  product_name?: string
  last_content: string
  last_time: string
  unread_count: number
}

export interface MessageVO {
  id: number
  sender_id: number
  receiver_id: number
  product_id: number
  content: string
  is_read: boolean
  created_at: string
}

export interface ReviewVO {
  id: number
  order_id: number
  product_id: number
  reviewer_id: number
  reviewee_id: number
  rating: string
  content: string
  created_at: string
  reviewer?: UserVO
}

export interface CartItemVO {
  id: number
  product_id: number
  quantity: number
  selected: boolean
  product?: ProductVO
}

export interface AuditVO {
  id: number
  user_id: number
  username: string
  module: string
  action: string
  resource_id: string
  detail: string
  ip: string
  request_id: string
  created_at: string
}
