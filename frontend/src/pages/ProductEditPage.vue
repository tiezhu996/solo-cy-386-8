<template>
  <el-card class="wrap" v-loading="loading">
    <h2>修改商品并重新提交审核</h2>
    <el-alert
      v-if="product?.review_status === 'rejected'"
      :title="`上次审核被驳回：${product.reject_reason || '未填写原因'}`"
      type="error"
      :closable="false"
      show-icon
      class="reject-alert"
    />
    <el-alert
      v-else
      title="商品修改提交后将立即下架并重新进入平台审核，复审期间不可购买。"
      type="warning"
      :closable="false"
      show-icon
      class="reject-alert"
    />
    <el-form :model="form" label-width="90px">
      <el-form-item label="商品名称" required>
        <el-input v-model="form.title" maxlength="128" show-word-limit />
      </el-form-item>
      <el-form-item label="描述" required>
        <el-input v-model="form.description" type="textarea" :rows="4" />
      </el-form-item>
      <el-form-item label="分类" required>
        <el-select v-model="form.category">
          <el-option v-for="(text, key) in ProductCategoryText" :key="key" :label="text" :value="key" />
        </el-select>
      </el-form-item>
      <el-form-item label="成色" required>
        <el-select v-model="form.condition">
          <el-option v-for="(text, key) in ProductConditionText" :key="key" :label="text" :value="key" />
        </el-select>
      </el-form-item>
      <el-form-item label="原价" required>
        <el-input-number v-model="form.original_price" :min="0" :precision="2" />
      </el-form-item>
      <el-form-item label="售价" required>
        <el-input-number v-model="form.price" :min="0.01" :precision="2" />
      </el-form-item>
      <el-form-item label="商品图片">
        <ImageUploader v-model="form.images" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :loading="submitting" @click="submit">提交审核</el-button>
        <el-button @click="$router.back()">取消</el-button>
      </el-form-item>
    </el-form>
  </el-card>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import * as productApi from '../api/product'
import ImageUploader from '../components/ImageUploader.vue'
import { ProductCategoryText, ProductConditionText } from '../constants'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const submitting = ref(false)
const product = ref<any>(null)
const form = reactive({
  title: '',
  description: '',
  category: '',
  condition: '',
  original_price: 0,
  price: 0,
  images: [] as string[]
})

onMounted(load)

async function load() {
  loading.value = true
  try {
    const res: any = await productApi.getProduct(Number(route.params.id))
    product.value = res.data
    Object.assign(form, {
      title: res.data.title,
      description: res.data.description,
      category: res.data.category,
      condition: res.data.condition,
      original_price: Number(res.data.original_price),
      price: Number(res.data.price),
      images: res.data.images || []
    })
  } finally {
    loading.value = false
  }
}

async function submit() {
  if (!form.title || !form.description || !form.category || !form.condition) {
    ElMessage.warning('请填写完整商品信息')
    return
  }
  if (form.price <= 0) {
    ElMessage.warning('售价必须大于 0')
    return
  }
  submitting.value = true
  try {
    await productApi.updateProduct(Number(route.params.id), { ...form })
    ElMessage.success('已重新提交审核')
    router.push(`/products/${route.params.id}`)
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.wrap {
  max-width: 760px;
  margin: 0 auto;
}
.reject-alert {
  margin-bottom: 16px;
}
</style>
