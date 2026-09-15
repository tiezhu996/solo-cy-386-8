<template>
  <el-card class="wrap">
    <h2>发布闲置商品</h2>
    <el-form :model="form" label-width="90px">
      <el-form-item label="商品名称" required>
        <el-input v-model="form.title" maxlength="128" show-word-limit placeholder="例如：iPhone 13 128G 几乎全新" />
      </el-form-item>
      <el-form-item label="描述" required>
        <el-input v-model="form.description" type="textarea" :rows="4" placeholder="成色、入手渠道、使用感受等" />
      </el-form-item>
      <el-form-item label="分类" required>
        <el-select v-model="form.category" placeholder="选择分类">
          <el-option v-for="(text, key) in ProductCategoryText" :key="key" :label="text" :value="key" />
        </el-select>
      </el-form-item>
      <el-form-item label="成色" required>
        <el-select v-model="form.condition" placeholder="选择成色">
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
        <el-button type="primary" :loading="loading" @click="submit">立即发布</el-button>
        <el-button @click="$router.push('/')">取消</el-button>
      </el-form-item>
    </el-form>
  </el-card>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import * as productApi from '../api/product'
import ImageUploader from '../components/ImageUploader.vue'
import { ProductCategoryText, ProductConditionText } from '../constants'

const router = useRouter()
const loading = ref(false)
const form = reactive({
  title: '',
  description: '',
  category: '',
  condition: '',
  original_price: 0,
  price: 0,
  images: [] as string[]
})

async function submit() {
  if (!form.title || !form.description || !form.category || !form.condition) {
    ElMessage.warning('请填写完整商品信息')
    return
  }
  if (form.price <= 0) {
    ElMessage.warning('售价必须大于 0')
    return
  }
  loading.value = true
  try {
    const res: any = await productApi.createProduct({ ...form })
    ElMessage.success('发布成功')
    router.push(`/products/${res.data.id}`)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.wrap {
  max-width: 760px;
  margin: 0 auto;
}
</style>
