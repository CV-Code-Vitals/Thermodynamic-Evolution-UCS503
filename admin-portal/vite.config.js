import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), 'VITE_')
  const apiBase = process.env.VITE_API_BASE || env.VITE_API_BASE || ''

  if (mode === 'production') {
    if (!apiBase) {
      throw new Error('VITE_API_BASE must be set to the deployed backend HTTPS URL before building for production.')
    }

    const localProxy = apiBase === '/api' && process.env.VITE_ALLOW_RELATIVE_API_BASE === 'true'
    if (!localProxy) {
      let apiURL
      try {
        apiURL = new URL(apiBase)
      } catch {
        throw new Error('VITE_API_BASE must be an absolute HTTPS URL, not a relative path such as /api.')
      }

      if (apiURL.protocol !== 'https:') {
        throw new Error('VITE_API_BASE must be an absolute HTTPS URL, not a relative path such as /api.')
      }
    }
  }

  return {
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
  }
})
