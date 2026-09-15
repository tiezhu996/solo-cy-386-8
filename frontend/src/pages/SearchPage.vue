<template>
  <div>
    <ProductFilterBar @search="onSearch" />
    <div v-loading="loading">
      <div class="list" v-if="products.length">
        <ProductCard v-for="p in products" :key="p.id" :product="p" />
      </div>
      <EmptyState v-if="!loading && products.length === 0" description="没有找到符合条件的商品" />
    </div>
    <div class="pager">
      <el-pagination
        background
        layout="prev, pager, next, total"
        :total="total"
        :page-size="pageSize"
        :current-page="page"
        @current-change="loadPage"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import * as productApi from '../api/product'
import ProductCard from '../components/ProductCard.vue'
import ProductFilterBar from '../components/ProductFilterBar.vue'
import EmptyState from '../components/EmptyState.vue'

const route = useRoute()
const products = ref<any[]>([])
const loading = ref(false)
const total = ref(0)
const page = ref(1)
const pageSize = 12
const params = reactive<Record<string, unknown>>({ keyword: (route.query.keyword as string) || '' })

onMounted(() => load(1))

async function load(p: number) {
  loading.value = true
  try {
    const clean: Record<string, unknown> = {}
    Object.entries(params).forEach(([k, v]) => {
      if (v !== '' && v !== undefined && v !== null) clean[k] = v
    })
    const res: any = await productApi.listProducts({ ...clean, page: p, page_size: pageSize })
    products.value = res.data.list || []
    total.value = Number(res.data.total || 0)
    page.value = p
  } finally {
    loading.value = false
  }
}

function onSearch(payload: Record<string, unknown>) {
  Object.assign(params, payload)
  load(1)
}

function loadPage(p: number) {
  load(p)
}
</script>

<style scoped>
.list {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}
.pager {
  margin-top: 16px;
  display: flex;
  justify-content: center;
}
</style>
