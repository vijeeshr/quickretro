import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      '^/(ws)': {
        target: 'http://localhost:8080/',
        changeOrigin: true,
        ws: true,
      },      
      '^/(api)': {
        target: 'http://localhost:8080/',
        changeOrigin: true,
      },
    }
  }
})
