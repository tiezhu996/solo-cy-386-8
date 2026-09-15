<template>
  <div class="auth-wrap">
    <el-card class="auth-card">
      <h2>登录 MarketPal</h2>
      <el-form :model="form" label-width="0" @submit.prevent>
        <el-form-item>
          <el-input v-model="form.username" placeholder="用户名" size="large" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="form.password" type="password" placeholder="密码" size="large" show-password @keyup.enter="submit" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" size="large" class="full" :loading="loading" @click="submit">登录</el-button>
        </el-form-item>
      </el-form>
      <div class="tips">
        还没有账号？<router-link to="/register">立即注册</router-link>
        <div class="hint">演示账号：admin / admin123</div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '../stores/userStore'

const router = useRouter()
const route = useRoute()
const store = useUserStore()
const loading = ref(false)
const form = reactive({ username: '', password: '' })

async function submit() {
  if (!form.username || !form.password) {
    ElMessage.warning('请输入用户名和密码')
    return
  }
  loading.value = true
  try {
    await store.login(form.username, form.password)
    ElMessage.success('登录成功')
    router.push((route.query.redirect as string) || '/')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-wrap {
  display: flex;
  justify-content: center;
  padding-top: 60px;
}
.auth-card {
  width: 400px;
}
.auth-card h2 {
  text-align: center;
  color: #303133;
}
.full {
  width: 100%;
}
.tips {
  text-align: center;
  color: #909399;
  font-size: 13px;
}
.hint {
  margin-top: 8px;
  color: #c0c4cc;
}
</style>
