import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'
import fs from 'fs'

function goTemplatePlugin() {
  return {
    name: 'go-template-plugin',
    closeBundle() {
      const frontendDistDir = path.resolve(__dirname, 'dist')
      const sourceDistDir = path.resolve(__dirname, '..', 'source', 'dist')
      const indexPath = path.join(frontendDistDir, 'index.html')
      const basePath = path.join(sourceDistDir, 'base.gohtml')

      // Copy assets from frontend/dist to source/dist
      const frontendAssetsDir = path.join(frontendDistDir, 'assets')
      const sourceAssetsDir = path.join(sourceDistDir, 'assets')

      // Ensure source assets directory exists
      if (!fs.existsSync(sourceAssetsDir)) {
        fs.mkdirSync(sourceAssetsDir, { recursive: true })
      }

      // Copy all asset files
      if (fs.existsSync(frontendAssetsDir)) {
        const files = fs.readdirSync(frontendAssetsDir)
        for (const file of files) {
          fs.copyFileSync(
            path.join(frontendAssetsDir, file),
            path.join(sourceAssetsDir, file)
          )
        }
        console.log(`Copied ${files.length} asset files to source/dist/assets`)
      }

      // Copy index.html
      fs.copyFileSync(indexPath, path.join(sourceDistDir, 'index.html'))

      const indexHtml = fs.readFileSync(indexPath, 'utf-8')

      const cssMatch = indexHtml.match(/<link[^>]*rel="stylesheet"[^>]*href="([^"]+)"[^>]*>/g) || []
      const preloadMatch = indexHtml.match(/<link[^>]*rel="modulepreload"[^>]*href="([^"]+)"[^>]*>/g) || []
      const jsMatch = indexHtml.match(/<script[^>]*src="([^"]+)"[^>]*><\/script>/g) || []

      const cssLinks = cssMatch.map(m => m.replace(/href="\//, 'href="')).join('\n    ')
      const preloadLinks = preloadMatch.map(m => m.replace(/href="\//, 'href="')).join('\n    ')
      const jsScripts = jsMatch.map(m => m.replace(/src="\//, 'src="')).join('\n')

      let baseGohtml = fs.readFileSync(basePath, 'utf-8')
      baseGohtml = baseGohtml.replace(
        /(<link rel="modulepreload"[^>]*>[\s\S]*?)?<link rel="stylesheet" crossorigin href="assets\/index-[^"]+\.css">/,
        `${preloadLinks}\n    ${cssLinks}`
      )
      baseGohtml = baseGohtml.replace(
        /<script type="module" crossorigin src="assets\/index-[^"]+\.js"><\/script>/,
        jsScripts
      )

      fs.writeFileSync(basePath, baseGohtml)
      console.log('Updated base.gohtml with Vite asset references')
    }
  }
}

export default defineConfig({
  plugins: [vue(), goTemplatePlugin()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  css: {
    preprocessorOptions: {
      scss: {
        additionalData: ``,
      },
    },
  },
  server: {
    port: 8888,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        cookieDomainRewrite: 'localhost',
      },
      '/oauth': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['vue', 'vue-router', 'pinia', 'axios'],
        },
      },
    },
  },
})
