<template>
  <div>
    <div class="hero">
      <h1>发现身边的闲置好物</h1>
      <p>C2C 本地生活市场 · 二手 / 集市 / 服务</p>
      <el-input v-model="keyword" placeholder="搜索商品名称或描述" size="large" class="hero-search" @keyup.enter="goSearch">
        <template #append><el-button @click="goSearch">搜索</el-button></template>
      </el-input>
    </div>
    <div class="section">
      <div class="section-head">
        <h2>最新上架</h2>
        <router-link to="/search">查看全部 →</router-link>
      </div>
      <div v-loading="loading" class="waterfall">
        <ProductCard v-for="p in products" :key="p.id" :product="p" class="item" />
      </div>
      <EmptyState v-if="!loading && products.length === 0" description="还没有商品，快去发布第一件闲置吧" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import * as productApi from '../api/product'
import ProductCard from '../components/ProductCard.vue'
import EmptyState from '../components/EmptyState.vue'

const router = useRouter()
const products = ref<any[]>([])
const loading = ref(false)
const keyword = ref('')

onMounted(async () => {
  loading.value = true
  try {
    const res: any = await productApi.listProducts({ page: 1, page_size: 12, sort_by: 'time_desc' })
    products.value = res.data.list || []
  } finally {
    loading.value = false
  }
})

function goSearch() {
  router.push({ path: '/search', query: keyword.value ? { keyword: keyword.value } : {} })
}
</script>

<style scoped>
.hero {
  text-align: center;
  padding: 48px 0 32px;
  background: linear-gradient(135deg, #e8f1ff, #f5f7fa);
  border-radius: 12px;
}
.hero h1 {
  margin: 0 0 8px;
  color: #303133;
}
.hero p {
  color: #909399;
  margin: 0 0 24px;
}
.hero-search {
  width: 480px;
  max-width: 90%;
}
.section {
  margin-top: 24px;
}
.section-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.section-head h2 {
  margin: 0;
}
.section-head a {
  color: #409eff;
  text-decoration: none;
}
.waterfall {
  column-count: 4;
  column-gap: 16px;
}
.item {
  break-inside: avoid;
  margin-bottom: 16px;
}
</style>
