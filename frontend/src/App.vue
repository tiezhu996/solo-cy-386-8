<template>
  <div class="app">
    <AppHeader />
    <main class="main">
      <router-view />
    </main>
    <footer class="footer">MarketPal 本地生活市场平台 · C2C 闲置交易</footer>
  </div>
</template>

<script setup lang="ts">
import AppHeader from './components/AppHeader.vue'
</script>

<style>
* {
  box-sizing: border-box;
}
body {
  margin: 0;
  background: #f5f7fa;
  font-family: 'Helvetica Neue', Helvetica, 'PingFang SC', 'Microsoft YaHei', Arial, sans-serif;
}
.app {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}
.main {
  flex: 1;
  width: 1200px;
  max-width: 96%;
  margin: 16px auto;
}
.footer {
  text-align: center;
  color: #909399;
  font-size: 13px;
  padding: 20px 0;
}
</style>
