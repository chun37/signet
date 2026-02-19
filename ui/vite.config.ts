import path from 'path'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    proxy: {
      '/chain': 'http://localhost:8080',
      '/peers': 'http://localhost:8080',
      '/transaction': 'http://localhost:8080',
      '/block': 'http://localhost:8080',
      '/register': 'http://localhost:8080',
      '/info': 'http://localhost:8080',
    },
  },
})
