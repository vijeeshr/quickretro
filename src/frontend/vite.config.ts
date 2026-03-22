import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), 'VITE_')
  const apiBase = env.VITE_API_BASE_URL || 'http://localhost:8080'

  return {
    plugins: [vue(), tailwindcss()],
    server: {
      proxy: {
        '^/(ws)': {
          target: apiBase,
          changeOrigin: true,
          ws: true,
        },
        '^/(api)': {
          target: apiBase,
          changeOrigin: true,
        },
        '/config.js': {
          target: apiBase,
          changeOrigin: true,
          secure: false,
        },
      },
    },
  }
})
