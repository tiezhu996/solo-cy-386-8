// useAuth.ts 鉴权 hook：路由守卫与按钮显隐共用。
import { computed } from 'vue'
import { useUserStore } from '../stores/userStore'

export function useAuth() {
  const store = useUserStore()
  const isLoggedIn = computed(() => store.isLoggedIn)
  const isAdmin = computed(() => store.isAdmin)
  const user = computed(() => store.user)

  function requireLogin(): boolean {
    if (!store.isLoggedIn) {
      window.location.href = '/login'
      return false
    }
    return true
  }

  return { isLoggedIn, isAdmin, user, requireLogin }
}
