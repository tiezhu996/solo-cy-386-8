// 路由与守卫：未登录访问受保护页跳转登录，管理员页校验角色。
import { createRouter, createWebHistory } from 'vue-router'
import HomePage from '../pages/HomePage.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'home', component: HomePage },
    { path: '/search', name: 'search', component: () => import('../pages/SearchPage.vue') },
    { path: '/products/:id', name: 'product-detail', component: () => import('../pages/ProductDetailPage.vue') },
    { path: '/products/create', name: 'product-create', component: () => import('../pages/ProductCreatePage.vue'), meta: { requiresAuth: true } },
    { path: '/cart', name: 'cart', component: () => import('../pages/CartPage.vue'), meta: { requiresAuth: true } },
    { path: '/checkout', name: 'checkout', component: () => import('../pages/CheckoutPage.vue'), meta: { requiresAuth: true } },
    { path: '/orders', name: 'orders', component: () => import('../pages/OrdersPage.vue'), meta: { requiresAuth: true } },
    { path: '/messages', name: 'messages', component: () => import('../pages/MessagesPage.vue'), meta: { requiresAuth: true } },
    { path: '/profile', name: 'profile', component: () => import('../pages/ProfilePage.vue'), meta: { requiresAuth: true } },
    { path: '/admin/audits', name: 'admin-audits', component: () => import('../pages/AdminAuditPage.vue'), meta: { requiresAuth: true, requiresAdmin: true } },
    { path: '/login', name: 'login', component: () => import('../pages/LoginPage.vue') },
    { path: '/register', name: 'register', component: () => import('../pages/RegisterPage.vue') }
  ]
})

router.beforeEach((to) => {
  const token = localStorage.getItem('marketpal_token')
  if (to.meta.requiresAuth && !token) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }
  if (to.meta.requiresAdmin) {
    const user = JSON.parse(localStorage.getItem('marketpal_user') || 'null')
    if (!user || user.role !== 'admin') {
      return { path: '/' }
    }
  }
  return true
})

export default router
