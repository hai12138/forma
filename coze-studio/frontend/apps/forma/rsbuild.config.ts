import { defineConfig } from '@coze-arch/rsbuild-config';

export default defineConfig({
  server: {
    port: 3001,
    strictPort: false,
    proxy: [
      {
        context: ['/api/forma'],
        target: 'http://localhost:8888/',
        changeOrigin: true,
      },
    ],
  },
  html: {
    title: 'Forma',
    template: './index.html',
  },
  source: {
    entry: {
      index: './src/index.tsx',
    },
  },
});
