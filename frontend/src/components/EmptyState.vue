<template>
  <div class="empty-state">
    <el-empty :description="description">
      <el-button v-if="actionText" type="primary" @click="$emit('action')">{{ actionText }}</el-button>
    </el-empty>
  </div>
</template>

<script setup lang="ts">
defineProps<{ description?: string; actionText?: string }>()
defineEmits<{ (e: 'action'): void }>()
</script>

<style scoped>
.empty-state {
  padding: 40px 0;
}
</style>
