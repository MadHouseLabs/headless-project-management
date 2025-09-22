import { defineConfig } from 'vite'

export default defineConfig({
  base: '/static/',
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    rollupOptions: {
      output: {
        assetFileNames: 'css/[name].[hash].css',
      },
    },
    cssCodeSplit: false,
    manifest: true,
  },
})