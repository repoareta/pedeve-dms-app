import { fileURLToPath, URL } from 'node:url'
import { readFileSync, existsSync } from 'fs'
import { join } from 'path'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    vueDevTools(),
    // Plugin untuk serve static files dari /user-guideline sebelum Vue Router
    {
      name: 'user-guideline-static',
      configureServer(server) {
        server.middlewares.use((req, res, next) => {
          // Handle request ke /user-guideline
          if (req.url?.startsWith('/user-guideline')) {
            let filePath = req.url.replace('/user-guideline', '')
            if (!filePath || filePath === '/') {
              filePath = '/index.html'
            }
            
            const publicPath = join(__dirname, 'public', 'user-guideline', filePath)
            
            // Jika file ada, serve langsung
            if (existsSync(publicPath)) {
              try {
                const content = readFileSync(publicPath)
                const ext = publicPath.split('.').pop()?.toLowerCase()
                
                // Set content type berdasarkan extension
                const contentType: Record<string, string> = {
                  'html': 'text/html; charset=utf-8',
                  'js': 'application/javascript; charset=utf-8',
                  'css': 'text/css; charset=utf-8',
                  'json': 'application/json; charset=utf-8',
                  'png': 'image/png',
                  'jpg': 'image/jpeg',
                  'jpeg': 'image/jpeg',
                  'svg': 'image/svg+xml',
                  'ico': 'image/x-icon',
                  'woff': 'font/woff',
                  'woff2': 'font/woff2',
                  'ttf': 'font/ttf',
                  'map': 'application/json',
                }
                
                res.setHeader('Content-Type', contentType[ext || ''] || 'text/plain')
                res.setHeader('Cache-Control', 'no-cache')
                res.end(content)
                return
              } catch (error) {
                console.error('Error serving user-guideline file:', error)
              }
            }
            
            // Jika tidak ada file dan path adalah root, serve index.html
            if (filePath === '/index.html' || req.url === '/user-guideline' || req.url === '/user-guideline/') {
              const indexPath = join(__dirname, 'public', 'user-guideline', 'index.html')
              if (existsSync(indexPath)) {
                try {
                  const content = readFileSync(indexPath)
                  res.setHeader('Content-Type', 'text/html; charset=utf-8')
                  res.setHeader('Cache-Control', 'no-cache')
                  res.end(content)
                  return
                } catch (error) {
                  console.error('Error serving user-guideline index.html:', error)
                }
              }
            }
          }
          
          next()
        })
      },
    },
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    },
  },
  optimizeDeps: {
    include: ['mammoth', 'xlsx', 'jspdf', 'jspdf-autotable'],
  },
})
