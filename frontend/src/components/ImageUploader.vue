<template>
  <div class="uploader">
    <el-upload
      :http-request="doUpload"
      list-type="picture-card"
      :file-list="fileList"
      :limit="9"
      accept="image/jpeg,image/png,image/webp"
      @remove="onRemove"
    >
      <el-icon><Plus /></el-icon>
    </el-upload>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { uploadImage } from '../api/product'

const props = defineProps<{ modelValue: string[] }>()
const emit = defineEmits<{ (e: 'update:modelValue', v: string[]): void }>()

const fileList = ref<any[]>(props.modelValue.map((url) => ({ name: url.split('/').pop(), url })))

watch(
  () => props.modelValue,
  (v) => {
    fileList.value = v.map((url) => ({ name: url.split('/').pop(), url }))
  }
)

async function doUpload(options: any) {
  try {
    const res: any = await uploadImage(options.file)
    const url = res.data.url
    const urls = [...props.modelValue, url]
    emit('update:modelValue', urls)
    ElMessage.success('图片上传成功')
  } catch {
    ElMessage.error('图片上传失败')
  }
}

function onRemove(file: any) {
  const urls = props.modelValue.filter((u) => u !== file.url)
  emit('update:modelValue', urls)
}
</script>
