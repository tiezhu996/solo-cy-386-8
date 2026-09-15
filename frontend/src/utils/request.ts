// request.ts axios 封装：统一响应拦截、JWT 注入、错误码提示（对应后端 error_handler）。
import axios from 'axios'
import { ElMessage } from 'element-plus'

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 15000
})

request.interceptors.request.use((config) => {
  const token = localStorage.getItem('marketpal_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

request.interceptors.response.use(
  (response) => {
    const data = response.data
    if (data && typeof data === 'object' && 'code' in data && data.code !== 0) {
      ElMessage.error(data.message || '请求失败')
      return Promise.reject(new Error(data.message || '请求失败'))
    }
    return data
  },
  (error) => {
    const res = error.response
    if (res && res.status === 401) {
      localStorage.removeItem('marketpal_token')
      localStorage.removeItem('marketpal_user')
      if (window.location.pathname !== '/login') {
        window.location.href = '/login'
      }
      ElMessage.error('登录已过期，请重新登录')
    } else if (res && res.data && res.data.message) {
      ElMessage.error(res.data.message)
    } else {
      ElMessage.error('网络异常，请稍后重试')
    }
    return Promise.reject(error)
  }
)

export default request
