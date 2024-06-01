import suidPlugin from '@suid/vite-plugin';
import { defineConfig } from 'vite';
import solidPlugin from 'vite-plugin-solid';

export default defineConfig({
  plugins: [
    solidPlugin(),
    suidPlugin(),
  ],
  server: {
    strictPort: true,
    host: true,
    watch: {
      usePolling: true,
    }
  },
  build: {
    target: 'esnext',
  },
});
