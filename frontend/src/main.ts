import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'
import 'virtual:uno.css'
import './style.css'
import { initTheme } from './composables/useTheme'

// 在挂载前同步初始化主题，避免首帧闪烁（FOUC）
initTheme()

createApp(App).use(router).use(ElementPlus).mount('#app')
