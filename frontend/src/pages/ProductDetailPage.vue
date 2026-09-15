<template>
  <div v-loading="loading">
    <el-card v-if="product">
      <!-- 审核状态横幅：仅卖家/管理员可见未通过商品详情 -->
      <el-alert
        v-if="product.review_status !== 'approved'"
        :title="reviewBanner.title"
        :type="reviewBanner.type"
        :closable="false"
        show-icon
        class="review-banner"
      >
        <template v-if="product.review_status === 'rejected'">
          <div>驳回原因：{{ product.reject_reason || '未填写' }}</div>
          <div>修改商品后将自动重新提交审核。</div>
        </template>
        <template v-else>
          <div>商品正在平台审核中，审核通过后才会在大厅、搜索中展示并允许加购、下单。</div>
        </template>
      </el-alert>

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
            <el-descriptions-item label="售卖状态"><StatusBadge type="product" :value="product.status" /></el-descriptions-item>
            <el-descriptions-item label="审核状态">
              <StatusBadge type="review" :value="product.review_status" />
              <span v-if="product.review_round > 1" class="round">第 {{ product.review_round }} 轮</span>
            </el-descriptions-item>
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
            <!-- 审核通过前不能加购或下单 -->
            <template v-if="canBuy">
              <el-button type="danger" size="large" @click="addCart">加入购物车</el-button>
              <el-button type="primary" size="large" @click="buyNow">立即购买</el-button>
            </template>
            <el-tag v-else type="warning" size="large" effect="plain" class="locked-tag">
              {{ product.review_status === 'rejected' ? '商品已被驳回，暂不可购买' : '审核通过前不可购买' }}
            </el-tag>
            <el-button v-if="canFavorite" size="large" @click="toggleFavorite">
              {{ product.is_favorite ? '取消收藏' : '收藏' }}
            </el-button>
            <el-button v-if="isOwner && product.review_status === 'rejected'" size="large" type="warning" plain @click="editProduct">修改重提</el-button>
            <el-button v-if="!isOwner" size="large" @click="contactSeller">联系卖家</el-button>
            <el-button v-if="canViewHistory" size="large" @click="showHistory = true">审核记录</el-button>
          </div>
        </div>
      </div>
      <el-divider content-position="left">商品描述</el-divider>
      <p class="desc-text">{{ product.description }}</p>
    </el-card>

    <!-- 审核记录回读 -->
    <el-dialog v-model="showHistory" title="审核记录" width="640px">
      <el-table :data="history" v-loading="historyLoading" size="small">
        <el-table-column prop="round" label="轮次" width="70" />
        <el-table-column label="结果" width="100">
          <template #default="{ row }"><StatusBadge type="review" :value="row.status" /></template>
        </el-table-column>
        <el-table-column prop="reason" label="原因">
          <template #default="{ row }">{{ row.reason || '—' }}</template>
        </el-table-column>
        <el-table-column prop="reviewer_id" label="审核员" width="90">
          <template #default="{ row }">{{ row.reviewer_id || '—' }}</template>
        </el-table-column>
        <el-table-column prop="reviewed_at" label="审核时间" width="160">
          <template #default="{ row }">{{ row.reviewed_at || '待审核' }}</template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import * as productApi from '../api/product'
import * as reviewApi from '../api/productReview'
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
const showHistory = ref(false)
const history = ref<any[]>([])
const historyLoading = ref(false)

onMounted(load)

const isOwner = computed(() => userStore.user?.id === product.value?.seller_id)
const isAdmin = computed(() => userStore.isAdmin)
const canBuy = computed(
  () => product.value && product.value.review_status === 'approved' && product.value.status === 'on_sale' && !isOwner.value
)
const canFavorite = computed(
  () => !isOwner.value && (product.value?.review_status === 'approved' || product.value?.is_favorite)
)
const canViewHistory = computed(() => product.value && (isOwner.value || isAdmin.value))
const reviewBanner = computed(() => {
  if (product.value?.review_status === 'rejected') {
    return { title: '该商品审核未通过，请按驳回原因修改后重新提交', type: 'error' as const }
  }
  return { title: '该商品正在等待平台审核', type: 'warning' as const }
})

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

function editProduct() {
  router.push(`/products/${product.value.id}/edit`)
}

function contactSeller() {
  if (!requireLogin()) return
  router.push({ path: '/messages', query: { peer_id: product.value.seller_id, product_id: product.value.id } })
}

async function loadHistory() {
  historyLoading.value = true
  try {
    const res: any = await reviewApi.productReviewHistory(product.value.id)
    history.value = res.data.list || []
  } finally {
    historyLoading.value = false
  }
}

watch(showHistory, (v) => {
  if (v) loadHistory()
})
</script>

<style scoped>
.review-banner {
  margin-bottom: 16px;
}
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
.round {
  margin-left: 8px;
  color: #909399;
  font-size: 12px;
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
.locked-tag {
  margin-right: 12px;
}
.desc-text {
  color: #606266;
  line-height: 1.8;
  white-space: pre-wrap;
}
</style>
