<template>
  <div class="auth-wrap">
    <el-card class="auth-card">
      <h2>注册 MarketPal</h2>
      <el-form :model="form" label-width="0" @submit.prevent>
        <el-form-item><el-input v-model="form.username" placeholder="用户名（3-32位）" size="large" /></el-form-item>
        <el-form-item><el-input v-model="form.nickname" placeholder="昵称" size="large" /></el-form-item>
        <el-form-item><el-input v-model="form.password" type="password" placeholder="密码（至少6位）" size="large" show-password /></el-form-item>
        <el-form-item><el-input v-model="form.email" placeholder="邮箱（选填）" size="large" /></el-form-item>
        <el-form-item><el-input v-model="form.phone" placeholder="手机号（选填）" size="large" /></el-form-item>
        <el-form-item>
          <el-button type="primary" size="large" class="full" :loading="loading" @click="submit">注册</el-button>
        </el-form-item>
      </el-form>
      <div class="tips">已有账号？<router-link to="/login">去登录</router-link></div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '../stores/userStore'

const router = useRouter()
const store = useUserStore()
const loading = ref(false)
const form = reactive({ username: '', nickname: '', password: '', email: '', phone: '' })

async function submit() {
  if (!form.username || !form.nickname || !form.password) {
    ElMessage.warning('请填写用户名、昵称和密码')
    return
  }
  loading.value = true
  try {
    await store.register({ ...form })
    ElMessage.success('注册成功，请登录')
    router.push('/login')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-wrap {
  display: flex;
  justify-content: center;
  padding-top: 40px;
}
.auth-card {
  width: 400px;
}
.auth-card h2 {
  text-align: center;
}
.full {
  width: 100%;
}
.tips {
  text-align: center;
  color: #909399;
  font-size: 13px;
}
</style>
