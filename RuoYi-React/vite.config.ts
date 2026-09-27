import { fileURLToPath, URL } from 'node:url'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'
import { createSvgIconsPlugin } from 'vite-plugin-svg-icons'

// 代理目标与 RuoYi-Vue3 vite.config.js 一致：三版后端同监听 8080（互斥运行）
const baseUrl = 'http://localhost:8080'

export default defineConfig({
  plugins: [
    react(),
    createSvgIconsPlugin({
      // 与 RuoYi-Vue3 同源复制（DB 菜单 meta.icon 存精灵名，symbolId 格式必须一致）
      iconDirs: [fileURLToPath(new URL('./src/assets/icons/svg', import.meta.url))],
      symbolId: 'icon-[dir]-[name]',
    }),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    // 默认 8090；PORT 环境变量可覆盖（多会话并行开发时避免端口冲突）
    port: process.env.PORT ? Number(process.env.PORT) : 8090,
    host: true,
    open: true,
    proxy: {
      '/dev-api': {
        target: baseUrl,
        changeOrigin: true,
        rewrite: (p) => p.replace(/^\/dev-api/, ''),
      },
      // springdoc 文档透传，不剥离前缀（与 RuoYi-Vue3 一致）
      '^/v3/api-docs/(.*)': {
        target: baseUrl,
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
    chunkSizeWarningLimit: 2000,
    rollupOptions: {
      output: {
        chunkFileNames: 'static/js/[name]-[hash].js',
        entryFileNames: 'static/js/[name]-[hash].js',
        assetFileNames: 'static/[ext]/[name]-[hash].[ext]',
      },
    },
  },
})
