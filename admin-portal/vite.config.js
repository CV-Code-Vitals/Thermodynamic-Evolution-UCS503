import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig(({ mode }) => ({
  plugins: [react()],
  // Relative asset URLs keep this app portable: it can be served at
  // /admin-portal/ on GitHub Pages or from the local static server.
  base: mode === 'development' ? '/' : './',
  server: {
    host: '0.0.0.0',
    port: 5173,
    strictPort: true,
    proxy: {
      '/api': {
        target: process.env.VITE_DEV_API_TARGET || 'http://localhost:8080',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ''),
      },
      '/uploads': {
        target: process.env.VITE_DEV_API_TARGET || 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
}))
