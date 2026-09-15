<template>
  <div>
    <el-table :data="data" v-loading="loading" stripe>
      <slot />
    </el-table>
    <div class="pager" v-if="showPager">
      <el-pagination
        background
        layout="prev, pager, next, total"
        :total="total"
        :page-size="pageSize"
        :current-page="page"
        @current-change="$emit('page-change', $event)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  data: unknown[]
  loading?: boolean
  total?: number
  page?: number
  pageSize?: number
  showPager?: boolean
}>()
defineEmits<{ (e: 'page-change', page: number): void }>()
</script>

<style scoped>
.pager {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
