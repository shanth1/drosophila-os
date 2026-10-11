import { defineConfig } from 'vite';

export default defineConfig({
  server: {
    proxy: {
      '/api': {
        target: process.env.DROSOPHILA_API_URL ?? 'http://127.0.0.1:8080',
        ws: true,
        // Preserve Host so the backend's same-origin WebSocket check also works
        // for the browser's Vite origin. Do not disable backend origin checks.
        changeOrigin: false,
      },
    },
  },
});
