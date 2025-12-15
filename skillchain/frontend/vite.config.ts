import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import path from 'path'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  define: {
    // For compatibility with certain libraries like CosmJS
    global: 'globalThis',
  },
  server: {
    port: 3000,
    proxy: {
        // Proxy to backend API
      '/api': {
        target: 'http://localhost:1317',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ''),
      },
    },
  },
})