import { defineConfig } from 'vite';
export default defineConfig({
  base: '/static/dist/',
  build: {
    outDir: '../static/dist',
    emptyOutDir: true,
  },
});
