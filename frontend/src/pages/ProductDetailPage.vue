<template>
  <div v-loading="loading">
    <el-card v-if="product">
      <div class="detail">
        <div class="gallery">
          <el-carousel v-if="product.images && product.images.length" height="360px" trigger="click">
            <el-carousel-item v-for="(img, i) in product.images" :key="i">
              <el-image :src="img" fit="contain" class="gallery-img" />
            </el-carousel-item>
          </el-carousel>
          <div v-else class="no-img">暂无图片</div>
        </div>
        <div class="body">
          <h2>{{ product.title }}</h2>
          <div class="price-row">
            <span class="price">¥{{ formatPrice(product.price) }}</span>
            <span class="orig">原价 ¥{{ formatPrice(product.original_price) }}</span>
          </div>
          <el-descriptions :column="1" border class="desc">
            <el-descriptions-item label="成色">{{ formatCondition(product.condition) }}</el-descriptions-item>
            <el-descriptions-item label="分类">{{ formatCategory(product.category) }}</el-descriptions-item>
            <el-descriptions-item label="状态"><StatusBadge type="product" :value="product.status" /></el-descriptions-item>
            <el-descriptions-item label="浏览 / 收藏">{{ product.view_count }} / {{ product.favorite_count }}</el-descriptions-item>
            <el-descriptions-item label="发布时间">{{ product.created_at }}</el-descriptions-item>
          </el-descriptions>
          <div class="seller">
            <el-avatar :size="40">{{ (product.seller?.nickname || 'U').slice(0, 1) }}</el-avatar>
            <div class="seller-info">
              <div>{{ product.seller?.nickname || `用户${product.seller_id}` }}</div>
              <div class="credit">信用分：{{ product.seller?.credit_score ?? '-' }}</div>
            </div>
          </div>
          <div class="actions">
            <el-button type="danger" size="large" @click="addCart">加入购物车</el-button>
            <el-button type="primary" size="large" @click="buyNow">立即购买</el-button>
            <el-button size="large" @click="toggleFavorite">
              {{ product.is_favorite ? '取消收藏' : '收藏' }}
            </el-button>
            <el-button size="large" @click="contactSeller">联系卖家</el-button>
          </div>
        </div>
      </div>
      <el-divider content-position="left">商品描述</el-divider>
      <p class="desc-text">{{ product.description }}</p>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import * as productApi from '../api/product'
import { useUserStore } from '../stores/userStore'
import { useCartStore } from '../stores/cartStore'
import StatusBadge from '../components/StatusBadge.vue'
import { formatPrice, formatCondition, formatCategory } from '../utils/format'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const cartStore = useCartStore()
const product = ref<any>(null)
const loading = ref(false)

onMounted(load)

async function load() {
  loading.value = true
  try {
    const res: any = await productApi.getProduct(Number(route.params.id))
    product.value = res.data
  } finally {
    loading.value = false
  }
}

function requireLogin(): boolean {
  if (!userStore.isLoggedIn) {
    router.push('/login')
    return false
  }
  return true
}

async function addCart() {
  if (!requireLogin()) return
  await cartStore.add(product.value.id, 1)
  ElMessage.success('已加入购物车')
}

async function toggleFavorite() {
  if (!requireLogin()) return
  if (product.value.is_favorite) {
    await productApi.unfavoriteProduct(product.value.id)
    product.value.is_favorite = false
    ElMessage.success('已取消收藏')
  } else {
    await productApi.favoriteProduct(product.value.id)
    product.value.is_favorite = true
    ElMessage.success('收藏成功')
  }
}

function buyNow() {
  if (!requireLogin()) return
  router.push({ path: '/checkout', query: { product_id: product.value.id } })
}

function contactSeller() {
  if (!requireLogin()) return
  router.push({ path: '/messages', query: { peer_id: product.value.seller_id, product_id: product.value.id } })
}
</script>

<style scoped>
.detail {
  display: flex;
  gap: 24px;
}
.gallery {
  width: 420px;
  flex-shrink: 0;
}
.gallery-img {
  width: 100%;
  height: 100%;
}
.no-img {
  height: 360px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f7fa;
  color: #909399;
}
.body {
  flex: 1;
}
.price-row {
  display: flex;
  align-items: baseline;
  gap: 12px;
  margin: 8px 0 16px;
}
.price {
  font-size: 28px;
  color: #f56c6c;
  font-weight: 700;
}
.orig {
  color: #909399;
  text-decoration: line-through;
}
.desc {
  margin-top: 8px;
}
.seller {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 16px 0;
}
.credit {
  font-size: 12px;
  color: #909399;
}
.actions {
  margin-top: 8px;
}
.desc-text {
  color: #606266;
  line-height: 1.8;
  white-space: pre-wrap;
}
</style>
