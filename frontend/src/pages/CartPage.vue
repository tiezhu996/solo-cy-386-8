<template>
  <el-card v-loading="loading">
    <h2>购物车</h2>
    <template v-if="items.length">
      <el-table :data="items">
        <el-table-column label="选择" width="60">
          <template #default="{ row }">
            <el-checkbox :model-value="row.selected" @change="(v: boolean) => update(row, { selected: v })" />
          </template>
        </el-table-column>
        <el-table-column label="商品">
          <template #default="{ row }">
            <div class="prod" @click="$router.push(`/products/${row.product_id}`)">
              <el-image v-if="row.product?.images?.length" :src="row.product.images[0]" fit="cover" class="thumb" />
              <div class="title">{{ row.product?.title }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="单价" width="120">
          <template #default="{ row }">¥{{ formatPrice(row.product?.price) }}</template>
        </el-table-column>
        <el-table-column label="数量" width="160">
          <template #default="{ row }">
            <el-input-number :model-value="row.quantity" :min="1" :max="99" size="small" @change="(v: number) => update(row, { quantity: v })" />
          </template>
        </el-table-column>
        <el-table-column label="小计" width="120">
          <template #default="{ row }">¥{{ formatPrice(Number(row.product?.price || 0) * row.quantity) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-button link type="danger" @click="remove(row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="footer">
        <span>已选合计：<b class="total">¥{{ formatPrice(cartStore.total) }}</b></span>
        <el-button type="danger" size="large" :disabled="!selectedFirst" @click="goCheckout">去结算</el-button>
      </div>
    </template>
    <EmptyState v-else description="购物车还是空的" action-text="去逛逛" @action="$router.push('/search')" />
  </el-card>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useCartStore } from '../stores/cartStore'
import EmptyState from '../components/EmptyState.vue'
import { formatPrice } from '../utils/format'

const router = useRouter()
const cartStore = useCartStore()
const loading = ref(false)
const items = computed(() => cartStore.items)
const selectedFirst = computed(() => cartStore.items.find((it) => it.selected))

function goCheckout() {
  if (!selectedFirst.value) {
    ElMessage.warning('请先勾选要结算的商品')
    return
  }
  router.push({ path: '/checkout', query: { product_id: selectedFirst.value.product_id } })
}

onMounted(async () => {
  loading.value = true
  try {
    await cartStore.load()
  } finally {
    loading.value = false
  }
})

async function update(row: any, data: { quantity?: number; selected?: boolean }) {
  await cartStore.update(row.id, data)
}

async function remove(id: number) {
  await cartStore.remove(id)
  ElMessage.success('已移除')
}
</script>

<style scoped>
.prod {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
}
.thumb {
  width: 56px;
  height: 56px;
  border-radius: 4px;
}
.title {
  color: #303133;
}
.footer {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 20px;
  margin-top: 16px;
}
.total {
  color: #f56c6c;
  font-size: 20px;
}
</style>
