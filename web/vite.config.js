import { defineConfig } from 'vite'
import { viteSingleFile } from 'vite-plugin-singlefile'

export default defineConfig({
  plugins: [viteSingleFile()],
  build: {
    outDir: '../internal/server/assets',
    emptyOutDir: true,
    assetsInlineLimit: 100000000,
  },
  server: {
    port: 3000,
    proxy: {
      '/api': 'http://localhost:41111',
      '/ws': {
        target: 'ws://localhost:41111',
        ws: true,
      },
    },
  },
})