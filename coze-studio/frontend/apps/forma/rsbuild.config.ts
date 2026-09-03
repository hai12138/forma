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
      {
        // Coze SessionAuth passport (login / logout / register) — same-origin to Forma UI.
        context: ['/api/passport'],
        target: 'http://localhost:8888/',
        changeOrigin: true,
        onProxyRes(proxyRes, req) {
          // Ensure logout always expires HttpOnly session_key for the Forma UI origin.
          if (req.url && req.url.includes('/logout')) {
            const cleared = 'session_key=; Max-Age=0; Path=/; HttpOnly; SameSite=Lax';
            const existing = proxyRes.headers['set-cookie'];
            proxyRes.headers['set-cookie'] = Array.isArray(existing)
              ? [...existing, cleared]
              : existing
                ? [existing, cleared]
                : [cleared];
          }
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
