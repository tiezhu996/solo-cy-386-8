<template>
  <el-card>
    <div class="head">
      <h2>商品平台审核</h2>
      <el-radio-group v-model="statusFilter" @change="load(1)">
        <el-radio-button label="pending_review">待审核</el-radio-button>
        <el-radio-button label="approved">已通过</el-radio-button>
        <el-radio-button label="rejected">已驳回</el-radio-button>
        <el-radio-button label="">全部</el-radio-button>
      </el-radio-group>
    </div>

    <DataTable
      :data="reviews"
      :loading="loading"
      :total="total"
      :page="page"
      :page-size="pageSize"
      show-pager
      @page-change="load"
    >
      <el-table-column prop="id" label="记录ID" width="80" />
      <el-table-column label="商品" min-width="200">
        <template #default="{ row }">
          <router-link :to="`/products/${row.product_id}`" class="prod-link">
            {{ row.product?.title || `商品#${row.product_id}` }}
          </router-link>
          <div class="sub">¥{{ formatPrice(row.product?.price) }} · 第{{ row.round }}轮 · 卖家#{{ row.seller_id }}</div>
        </template>
      </el-table-column>
      <el-table-column label="审核状态" width="110">
        <template #default="{ row }"><StatusBadge type="review" :value="row.status" /></template>
      </el-table-column>
      <el-table-column label="驳回原因" min-width="160">
        <template #default="{ row }">{{ row.reason || '—' }}</template>
      </el-table-column>
      <el-table-column label="审核员" width="90">
        <template #default="{ row }">{{ row.reviewer_id || '—' }}</template>
      </el-table-column>
      <el-table-column prop="created_at" label="送审时间" width="170" />
      <el-table-column prop="reviewed_at" label="审核时间" width="170">
        <template #default="{ row }">{{ row.reviewed_at || '—' }}</template>
      </el-table-column>
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <template v-if="row.status === 'pending_review'">
            <el-button link type="success" :loading="actingId === row.id" @click="approve(row)">通过</el-button>
            <el-button link type="danger" @click="openReject(row)">驳回</el-button>
          </template>
          <span v-else class="done">已处理</span>
        </template>
      </el-table-column>
    </DataTable>

    <el-dialog v-model="rejectVisible" title="驳回商品" width="460px">
      <el-alert
        :title="`确认驳回「${current?.product?.title || '商品#' + current?.product_id}」？驳回后卖家可修改并重新提交。`"
        type="warning"
        :closable="false"
        class="reject-tip"
      />
      <el-input
        v-model="rejectReason"
        type="textarea"
        :rows="4"
        maxlength="500"
        show-word-limit
        placeholder="请填写驳回原因（必填），将展示给卖家"
      />
      <template #footer>
        <el-button @click="rejectVisible = false">取消</el-button>
        <el-button type="danger" :loading="actingId === current?.id" @click="confirmReject">确认驳回</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import * as reviewApi from '../api/productReview'
import DataTable from '../components/DataTable.vue'
import StatusBadge from '../components/StatusBadge.vue'
import { formatPrice } from '../utils/format'

const reviews = ref<any[]>([])
const loading = ref(false)
const total = ref(0)
const page = ref(1)
const pageSize = 15
const statusFilter = ref('pending_review')
const actingId = ref<number | null>(null)

const rejectVisible = ref(false)
const current = ref<any>(null)
const rejectReason = ref('')

onMounted(() => load(1))

async function load(p: number) {
  loading.value = true
  try {
    const params: Record<string, unknown> = { page: p, page_size: pageSize }
    if (statusFilter.value) params.status = statusFilter.value
    const res: any = await reviewApi.listProductReviews(params)
    reviews.value = res.data.list || []
    total.value = Number(res.data.total || 0)
    page.value = p
  } finally {
    loading.value = false
  }
}

async function approve(row: any) {
  actingId.value = row.id
  try {
    await reviewApi.decideProductReview(row.id, { action: 'approve' })
    ElMessage.success('已审核通过，商品已上架')
    await load(page.value)
  } finally {
    actingId.value = null
  }
}

function openReject(row: any) {
  current.value = row
  rejectReason.value = ''
  rejectVisible.value = true
}

async function confirmReject() {
  if (!rejectReason.value.trim()) {
    ElMessage.warning('驳回必须填写原因')
    return
  }
  if (!current.value) return
  actingId.value = current.value.id
  try {
    await reviewApi.decideProductReview(current.value.id, { action: 'reject', reason: rejectReason.value.trim() })
    ElMessage.success('已驳回')
    rejectVisible.value = false
    await load(page.value)
  } finally {
    actingId.value = null
  }
}
</script>

<style scoped>
.head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.head h2 {
  margin: 0;
}
.prod-link {
  color: #409eff;
  text-decoration: none;
}
.sub {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}
.done {
  color: #909399;
  font-size: 12px;
}
.reject-tip {
  margin-bottom: 12px;
}
</style>
