<template>
  <div class="product-card" @click="$router.push(`/products/${product.id}`)">
    <div class="cover">
      <el-image v-if="product.images && product.images.length" :src="product.images[0]" fit="cover" class="img" />
      <div v-else class="img placeholder">暂无图片</div>
      <el-tag class="status-tag" :type="product.status === 'on_sale' ? 'success' : 'info'" size="small">
        {{ formatProductStatus(product.status) }}
      </el-tag>
    </div>
    <div class="info">
      <div class="title">{{ product.title }}</div>
      <div class="meta">
        <span class="price">¥{{ formatPrice(product.price) }}</span>
        <span class="cond">{{ formatCondition(product.condition) }}</span>
      </div>
      <div class="sub">
        <span>{{ formatCategory(product.category) }}</span>
        <span>{{ product.view_count }} 浏览 · {{ product.favorite_count }} 收藏</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ProductVO } from '../api/types'
import { formatPrice, formatCondition, formatCategory, formatProductStatus } from '../utils/format'

defineProps<{ product: ProductVO }>()
</script>

<style scoped>
.product-card {
  border: 1px solid #ebeef5;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  transition: box-shadow 0.2s;
  background: #fff;
}
.product-card:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
}
.cover {
  position: relative;
  height: 180px;
}
.img {
  width: 100%;
  height: 100%;
}
.placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f7fa;
  color: #909399;
}
.status-tag {
  position: absolute;
  top: 8px;
  left: 8px;
}
.info {
  padding: 10px 12px;
}
.title {
  font-size: 14px;
  color: #303133;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 6px;
}
.price {
  color: #f56c6c;
  font-weight: 600;
  font-size: 16px;
}
.cond {
  font-size: 12px;
  color: #909399;
}
.sub {
  display: flex;
  justify-content: space-between;
  margin-top: 6px;
  font-size: 12px;
  color: #909399;
}
</style>
