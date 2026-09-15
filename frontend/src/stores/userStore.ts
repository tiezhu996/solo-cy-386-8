// userStore.ts 用户状态管理（登录态、个人资料、角色）。
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as userApi from '../api/user'
import type { UserVO } from '../api/types'

export const useUserStore = defineStore('user', () => {
  const token = ref<string>(localStorage.getItem('marketpal_token') || '')
  const user = ref<UserVO | null>(JSON.parse(localStorage.getItem('marketpal_user') || 'null'))

  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.role === 'admin')

  function setAuth(t: string, u: UserVO) {
    token.value = t
    user.value = u
    localStorage.setItem('marketpal_token', t)
    localStorage.setItem('marketpal_user', JSON.stringify(u))
  }

  async function login(username: string, password: string) {
    const res: any = await userApi.login({ username, password })
    setAuth(res.data.token, res.data.user)
    return res.data.user
  }

  async function register(payload: { username: string; password: string; nickname: string; email?: string; phone?: string }) {
    const res: any = await userApi.register(payload)
    return res
  }

  async function fetchProfile() {
    const res: any = await userApi.getProfile()
    user.value = res.data
    localStorage.setItem('marketpal_user', JSON.stringify(res.data))
    return res.data
  }

  async function updateProfile(payload: { nickname?: string; email?: string; phone?: string; avatar?: string }) {
    const res: any = await userApi.updateProfile(payload)
    user.value = res.data
    localStorage.setItem('marketpal_user', JSON.stringify(res.data))
    return res.data
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('marketpal_token')
    localStorage.removeItem('marketpal_user')
  }

  return { token, user, isLoggedIn, isAdmin, login, register, fetchProfile, updateProfile, logout }
})
