import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import { canAccessBrandRoute, ensureBrandingLoaded } from './lib/brand'
import 'element-plus/theme-chalk/dark/css-vars.css'
import './styles/base.css'

router.beforeEach(async (to) => {
  await ensureBrandingLoaded()
  if (canAccessBrandRoute(to.path)) {
    return true
  }
  return '/login'
})

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.mount('#app')
