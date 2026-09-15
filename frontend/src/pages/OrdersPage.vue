<template>
  <el-card>
    <h2>我的订单</h2>
    <el-tabs v-model="tab" @tab-change="load(1)">
      <el-tab-pane label="全部" name="all" />
      <el-tab-pane label="待付款" name="pending_payment" />
      <el-tab-pane label="待发货" name="pending_shipment" />
      <el-tab-pane label="已发货" name="shipped" />
      <el-tab-pane label="已收货" name="received" />
      <el-tab-pane label="已完成" name="completed" />
      <el-tab-pane label="已取消" name="cancelled" />
    </el-tabs>

    <div v-loading="loading">
      <el-card v-for="o in orders" :key="o.id" class="order-card" shadow="never">
        <div class="order-head">
          <span class="no">订单号：{{ o.order_no }}</span>
          <StatusBadge type="order" :value="o.status" />
        </div>
        <div class="order-body" @click="$router.push(`/products/${o.product_id}`)">
          <el-image v-if="o.product?.images?.length" :src="o.product.images[0]" fit="cover" class="thumb" />
          <div class="info">
            <div class="title">{{ o.product?.title }}</div>
            <div class="sub">{{ o.product?.condition_text || '' }} × {{ o.quantity }}</div>
          </div>
          <div class="amount">¥{{ formatPrice(o.total_price) }}</div>
        </div>
        <div class="order-actions">
          <template v-if="o.status === 'pending_payment'">
            <el-button type="primary" size="small" @click="pay(o.id)">去付款</el-button>
            <el-button size="small" @click="cancel(o.id)">取消订单</el-button>
          </template>
          <template v-else-if="o.status === 'pending_shipment'">
            <el-button size="small" disabled>等待卖家发货</el-button>
          </template>
          <template v-else-if="o.status === 'shipped'">
            <el-button type="primary" size="small" @click="receive(o.id)">确认收货</el-button>
          </template>
          <template v-else-if="o.status === 'received'">
            <el-button type="primary" size="small" @click="complete(o.id)">完成交易</el-button>
          </template>
          <template v-else-if="o.status === 'completed'">
            <el-button size="small" @click="review(o)">评价</el-button>
          </template>
        </div>
      </el-card>
      <EmptyState v-if="!loading && orders.length === 0" description="暂无相关订单" />
    </div>

    <div class="pager">
      <el-pagination background layout="prev, pager, next, total" :total="total" :page-size="pageSize" :current-page="page" @current-change="load" />
    </div>
  </el-card>

  <el-dialog v-model="reviewDialog" title="交易评价" width="420px">
    <el-form label-width="60px">
      <el-form-item label="评价">
        <el-radio-group v-model="reviewForm.rating">
          <el-radio value="good">好评</el-radio>
          <el-radio value="neutral">中评</el-radio>
          <el-radio value="bad">差评</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="内容">
        <el-input v-model="reviewForm.content" type="textarea" :rows="3" maxlength="500" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="reviewDialog = false">取消</el-button>
      <el-button type="primary" @click="submitReview">提交</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import * as orderApi from '../api/order'
import * as reviewApi from '../api/review'
import StatusBadge from '../components/StatusBadge.vue'
import EmptyState from '../components/EmptyState.vue'
import { formatPrice } from '../utils/format'

const tab = ref('all')
const orders = ref<any[]>([])
const loading = ref(false)
const total = ref(0)
const page = ref(1)
const pageSize = 10
const reviewDialog = ref(false)
const reviewForm = ref<{ order_id: number; rating: string; content: string }>({ order_id: 0, rating: 'good', content: '' })

onMounted(() => load(1))

async function load(p: number) {
  loading.value = true
  try {
    const params: Record<string, unknown> = { page: p, page_size: pageSize }
    if (tab.value !== 'all') params.status = tab.value
    const res: any = await orderApi.listOrders(params)
    orders.value = (res.data.list || []).map((o: any) => ({
      ...o,
      product: o.product ? { ...o.product, condition_text: '' } : undefined
    }))
    total.value = Number(res.data.total || 0)
    page.value = p
  } finally {
    loading.value = false
  }
}

async function pay(id: number) {
  await orderApi.payOrder(id)
  ElMessage.success('支付成功')
  load(page.value)
}

async function cancel(id: number) {
  await orderApi.cancelOrder(id)
  ElMessage.success('订单已取消')
  load(page.value)
}

async function receive(id: number) {
  await orderApi.receiveOrder(id)
  ElMessage.success('已确认收货')
  load(page.value)
}

async function complete(id: number) {
  await orderApi.completeOrder(id)
  ElMessage.success('交易完成，可以评价啦')
  load(page.value)
}

function review(o: any) {
  reviewForm.value = { order_id: o.id, rating: 'good', content: '' }
  reviewDialog.value = true
}

async function submitReview() {
  await reviewApi.createReview(reviewForm.value)
  reviewDialog.value = false
  ElMessage.success('评价成功')
}
</script>

<style scoped>
.order-card {
  margin-bottom: 12px;
}
.order-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 8px;
  border-bottom: 1px dashed #ebeef5;
}
.no {
  color: #909399;
  font-size: 13px;
}
.order-body {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 12px 0;
  cursor: pointer;
}
.thumb {
  width: 72px;
  height: 72px;
  border-radius: 6px;
}
.info {
  flex: 1;
}
.title {
  font-weight: 600;
  color: #303133;
}
.sub {
  color: #909399;
  font-size: 13px;
  margin-top: 4px;
}
.amount {
  color: #f56c6c;
  font-weight: 700;
  font-size: 18px;
}
.order-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
.pager {
  display: flex;
  justify-content: center;
  margin-top: 16px;
}
</style>
