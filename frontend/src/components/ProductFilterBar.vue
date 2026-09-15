<template>
  <div class="filter-bar">
    <el-input v-model="local.keyword" placeholder="搜索闲置好物" clearable class="kw" @keyup.enter="$emit('search', local)" />
    <el-select v-model="local.category" placeholder="全部分类" clearable class="sel" @change="$emit('search', local)">
      <el-option v-for="(text, key) in ProductCategoryText" :key="key" :label="text" :value="key" />
    </el-select>
    <el-select v-model="local.condition" placeholder="成色不限" clearable class="sel" @change="$emit('search', local)">
      <el-option v-for="(text, key) in ProductConditionText" :key="key" :label="text" :value="key" />
    </el-select>
    <el-input-number v-model="local.min_price" :min="0" placeholder="最低价" class="num" />
    <span class="sep">-</span>
    <el-input-number v-model="local.max_price" :min="0" placeholder="最高价" class="num" />
    <el-select v-model="local.sort_by" placeholder="默认排序" clearable class="sel" @change="$emit('search', local)">
      <el-option label="价格从低到高" value="price" />
      <el-option label="价格从高到低" value="price_desc" />
      <el-option label="最新发布" value="time_desc" />
    </el-select>
    <el-button type="primary" @click="$emit('search', local)">搜索</el-button>
  </div>
</template>

<script setup lang="ts">
import { reactive } from 'vue'
import { ProductCategoryText, ProductConditionText } from '../constants'

const props = defineProps<{ modelValue?: Record<string, unknown> }>()
defineEmits<{ (e: 'search', payload: Record<string, unknown>): void }>()

const local = reactive<Record<string, unknown>>({
  keyword: props.modelValue?.keyword ?? '',
  category: props.modelValue?.category ?? '',
  condition: props.modelValue?.condition ?? '',
  min_price: props.modelValue?.min_price ?? undefined,
  max_price: props.modelValue?.max_price ?? undefined,
  sort_by: props.modelValue?.sort_by ?? ''
})
</script>

<style scoped>
.filter-bar {
  display: flex;
  gap: 10px;
  align-items: center;
  flex-wrap: wrap;
  padding: 12px 0;
}
.kw {
  width: 240px;
}
.sel {
  width: 140px;
}
.num {
  width: 130px;
}
.sep {
  color: #909399;
}
</style>
