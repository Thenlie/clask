import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// Build output is embedded into the Go binary (internal/server/static.go).
// In dev, `clask serve --no-open` provides the API and Vite proxies to it.
export default defineConfig({
  plugins: [react()],
  build: {
    outDir: '../internal/server/dist',
    emptyOutDir: true,
  },
  server: {
    proxy: {
      '/api': 'http://127.0.0.1:7373',
    },
  },
})
