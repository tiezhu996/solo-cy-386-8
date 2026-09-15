<template>
  <header class="header">
    <div class="inner">
      <div class="logo" @click="$router.push('/')">MarketPal</div>
      <nav class="nav">
        <router-link to="/">首页</router-link>
        <router-link to="/search">逛逛</router-link>
        <router-link v-if="userStore.isLoggedIn" to="/products/create">发布闲置</router-link>
        <router-link v-if="userStore.isLoggedIn" to="/orders">我的订单</router-link>
        <router-link v-if="userStore.isLoggedIn" to="/messages">
          私信
          <el-badge v-if="messageStore.unread > 0" :value="messageStore.unread" class="badge" />
        </router-link>
        <router-link v-if="userStore.isAdmin" to="/admin/audits">审计日志</router-link>
      </nav>
      <div class="actions">
        <el-badge :value="cartStore.count" :hidden="cartStore.count === 0">
          <el-button text @click="$router.push('/cart')">购物车</el-button>
        </el-badge>
        <template v-if="userStore.isLoggedIn">
          <el-dropdown @command="onCommand">
            <span class="user">{{ userStore.user?.nickname || userStore.user?.username }}</span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">个人中心</el-dropdown-item>
                <el-dropdown-item command="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
        <el-button v-else type="primary" size="small" @click="$router.push('/login')">登录</el-button>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/userStore'
import { useCartStore } from '../stores/cartStore'
import { useMessageStore } from '../stores/messageStore'

const router = useRouter()
const userStore = useUserStore()
const cartStore = useCartStore()
const messageStore = useMessageStore()

onMounted(() => {
  if (userStore.isLoggedIn) {
    cartStore.load().catch(() => {})
    messageStore.loadUnread()
  }
})

function onCommand(cmd: string) {
  if (cmd === 'profile') {
    router.push('/profile')
  } else if (cmd === 'logout') {
    userStore.logout()
    cartStore.reset()
    router.push('/')
  }
}
</script>

<style scoped>
.header {
  background: #fff;
  border-bottom: 1px solid #ebeef5;
  position: sticky;
  top: 0;
  z-index: 100;
}
.inner {
  width: 1200px;
  max-width: 96%;
  margin: 0 auto;
  height: 60px;
  display: flex;
  align-items: center;
  gap: 24px;
}
.logo {
  font-size: 20px;
  font-weight: 700;
  color: #409eff;
  cursor: pointer;
}
.nav {
  display: flex;
  gap: 18px;
  flex: 1;
}
.nav a {
  color: #303133;
  text-decoration: none;
  font-size: 14px;
}
.nav a.router-link-active {
  color: #409eff;
  font-weight: 600;
}
.actions {
  display: flex;
  align-items: center;
  gap: 12px;
}
.user {
  cursor: pointer;
  color: #409eff;
}
.badge {
  margin-left: 4px;
}
</style>
