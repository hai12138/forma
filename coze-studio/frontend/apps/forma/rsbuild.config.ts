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
        onProxyRes(proxyRes) {
          // Strip illegal Domain=host:port so browsers accept Forma logout Set-Cookie.
          const cookies = proxyRes.headers['set-cookie'];
          if (cookies) {
            const list = Array.isArray(cookies) ? cookies : [cookies];
            proxyRes.headers['set-cookie'] = list.map(c =>
              String(c).replace(/;\s*domain=[^;]*/gi, ''),
            );
          }
        },
      },
      {
        // Coze SessionAuth passport (login / register) — same-origin to Forma UI.
        // Logout uses Forma-owned /api/forma/v1/auth/logout — do not patch Coze core.
        context: ['/api/passport'],
        target: 'http://localhost:8888/',
        changeOrigin: true,
        onProxyRes(proxyRes) {
          // Strip illegal Domain=host:port from Set-Cookie so browsers accept session_key.
          const cookies = proxyRes.headers['set-cookie'];
          if (cookies) {
            const list = Array.isArray(cookies) ? cookies : [cookies];
            proxyRes.headers['set-cookie'] = list.map(c =>
              String(c).replace(/;\s*domain=[^;]*/gi, ''),
            );
          }
        },
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
