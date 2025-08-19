import { createApp } from 'vue'
import { createPinia } from 'pinia'
import Antd from 'ant-design-vue'
import 'ant-design-vue/dist/reset.css'

import App from './App.vue'
import router from './router'
import './styles/index.css'
import { initStorageMigration } from './utils/migration'
import { useSiteStore } from './stores/site'

// 在开发环境中可以添加调试代码
if (import.meta.env.DEV) {
  console.log('EIAM Platform - Development Mode')
}

// 初始化存储迁移
initStorageMigration()

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)
app.use(Antd)

// 在应用启动后加载站点配置
app.mount('#app')

// 初始化公开站点信息（不需要认证）
const siteStore = useSiteStore()
siteStore.loadPublicSiteInfo().catch(error => {
  console.warn('Failed to load public site info:', error)
  // 失败不影响应用启动，使用默认值
})
