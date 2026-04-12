import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const proxyTarget = env.VITE_DEV_API_PROXY_TARGET || env.VITE_API_BASE_URL || 'http://localhost:8080'

  return {
    plugins: [
      vue(),
      Components({
        dts: 'src/components.d.ts',
        resolvers: [ElementPlusResolver({ importStyle: 'css' })],
      }),
    ],
    server: {
      proxy: {
        '/api': {
          target: proxyTarget,
          changeOrigin: true,
        },
        '/healthz': {
          target: proxyTarget,
          changeOrigin: true,
        },
      },
    },
  }
})
