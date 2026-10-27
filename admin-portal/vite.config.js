import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const projectRoot = dirname(fileURLToPath(import.meta.url))
const graphAssets = ['index.html', 'graph.mjs', 'model.mjs', 'routes.mjs', 'measurements.mjs', 'account.mjs', 'scan.mjs', 'config.mjs']
const graphAssetPath = (name) => resolve(projectRoot, '../graph-visualizer', name)

function graphVisualizerPlugin(config) {
  const source = (name) => name === 'config.mjs'
    ? `export default ${JSON.stringify(config)};\n`
    : readFileSync(graphAssetPath(name), 'utf8')
  return {
    name: 'graph-visualizer',
    configureServer(server) {
      server.middlewares.use('/graph-visualizer/', (request, response, next) => {
        if (!['GET', 'HEAD'].includes(request.method)) return next()
        const name = request.url.split('?')[0].replace(/^\//, '') || 'index.html'
        if (!graphAssets.includes(name)) return next()
        response.statusCode = 200
        response.setHeader('Cache-Control', 'no-store')
        response.setHeader('Content-Type', name.endsWith('.html') ? 'text/html; charset=utf-8' : 'application/javascript')
        response.end(request.method === 'HEAD' ? undefined : source(name))
      })
    },
    generateBundle() {
      for (const name of graphAssets) {
        this.emitFile({ type: 'asset', fileName: `graph-visualizer/${name}`, source: source(name) })
      }
    },
  }
}

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), 'VITE_')
  const apiBase = process.env.VITE_API_BASE || env.VITE_API_BASE || ''
  const analysisApiBase = process.env.VITE_ANALYSIS_API_BASE || env.VITE_ANALYSIS_API_BASE || ''

  if (mode === 'production') {
    if (analysisApiBase) {
      let analysisURL
      try {
        analysisURL = new URL(analysisApiBase)
      } catch {
        throw new Error('VITE_ANALYSIS_API_BASE must be an absolute HTTPS URL.')
      }
      if (analysisURL.protocol !== 'https:' || analysisURL.username || analysisURL.password || analysisURL.search || analysisURL.hash) {
        throw new Error('VITE_ANALYSIS_API_BASE must be an absolute HTTPS URL without credentials, query, or fragment.')
      }
    }
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
    plugins: [react(), graphVisualizerPlugin({ apiBase: apiBase || '/api', analysisApiBase })],
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
        },
        '/uploads': {
          target: process.env.VITE_DEV_API_TARGET || 'http://localhost:8080',
          changeOrigin: true,
        },
      },
    },
  }
})
