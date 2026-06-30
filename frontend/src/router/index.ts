import { createRouter, createWebHashHistory } from 'vue-router'
import Layout from '@/views/system/Layout.vue'
import { TOOLS } from '@/constants/tools'

const toolRouteNames: Record<string, string> = {
  '/tools/json': 'JsonFormatter',
  '/tools/hash': 'HashTool',
  '/tools/base64': 'Base64Tool',
  '/tools/timestamp': 'TimestampTool',
  '/tools/jwt': 'JwtDecoder',
  '/tools/uuid': 'UuidGenerator',
  '/tools/url': 'UrlEncoder',
  '/tools/regex': 'RegexTester',
  '/tools/api': 'ApiTester',
}

const toolComponents: Record<string, () => Promise<any>> = {
  '/tools/json': () => import('@/views/tools/JsonFormatter.vue'),
  '/tools/hash': () => import('@/views/tools/HashTool.vue'),
  '/tools/base64': () => import('@/views/tools/Base64Tool.vue'),
  '/tools/timestamp': () => import('@/views/tools/TimestampTool.vue'),
  '/tools/jwt': () => import('@/views/tools/JwtDecoder.vue'),
  '/tools/uuid': () => import('@/views/tools/UuidGenerator.vue'),
  '/tools/url': () => import('@/views/tools/UrlEncoder.vue'),
  '/tools/regex': () => import('@/views/tools/RegexTester.vue'),
  '/tools/api': () => import('@/views/tools/ApiTester.vue'),
}

const routes = [
  {
    path: '/',
    component: Layout,
    children: [
      {
        path: '',
        name: 'Home',
        component: () => import('@/views/home/Index.vue')
      },
      {
        path: '/key',
        name: 'Key',
        component: () => import('@/views/key/Index.vue')
      },
      // ===== 开发工具（由 TOOLS 常量自动生成） =====
      ...TOOLS.map((tool) => ({
        path: tool.path,
        name: toolRouteNames[tool.path],
        component: toolComponents[tool.path],
      })),
    ]
  }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

export default router
