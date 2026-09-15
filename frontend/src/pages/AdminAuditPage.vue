<template>
  <el-card>
    <h2>操作审计日志</h2>
    <div class="filters">
      <el-input v-model="module" placeholder="模块" clearable class="f" />
      <el-input v-model="action" placeholder="动作" clearable class="f" />
      <el-button type="primary" @click="load(1)">查询</el-button>
    </div>
    <DataTable :data="logs" :loading="loading" :total="total" :page="page" :page-size="pageSize" show-pager @page-change="load">
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="username" label="用户" width="120" />
      <el-table-column prop="module" label="模块" width="100" />
      <el-table-column prop="action" label="动作" width="140" />
      <el-table-column prop="path" label="路径" />
      <el-table-column prop="ip" label="IP" width="130" />
      <el-table-column prop="created_at" label="时间" width="170" />
    </DataTable>
  </el-card>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import * as auditApi from '../api/audit'
import DataTable from '../components/DataTable.vue'

const logs = ref<any[]>([])
const loading = ref(false)
const total = ref(0)
const page = ref(1)
const pageSize = 15
const module = ref('')
const action = ref('')

onMounted(() => load(1))

async function load(p: number) {
  loading.value = true
  try {
    const params: Record<string, unknown> = { page: p, page_size: pageSize }
    if (module.value) params.module = module.value
    if (action.value) params.action = action.value
    const res: any = await auditApi.listAudits(params)
    logs.value = (res.data.list || []).map((l: any) => ({ ...l, path: l.detail }))
    total.value = Number(res.data.total || 0)
    page.value = p
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.filters {
  display: flex;
  gap: 10px;
  margin-bottom: 12px;
}
.f {
  width: 180px;
}
</style>
