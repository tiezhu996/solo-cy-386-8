import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    host: '0.0.0.0',
    port: 5173,
    proxy: {
      '/api': { target: 'http://localhost:19406', changeOrigin: true },
      '/uploads': { target: 'http://localhost:19406', changeOrigin: true },
      '/ws': { target: 'ws://localhost:19406', ws: true }
    }
  }
})
