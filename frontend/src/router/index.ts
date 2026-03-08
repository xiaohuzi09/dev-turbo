import { createRouter, createWebHashHistory } from 'vue-router'
import Layout from '../views/system/Layout.vue'
import Home from '../views/home/Index.vue'

const routes = [
  {
    path: '/',
    component: Layout,
    children: [
      {
        path: '',
        name: 'Home',
        component: Home
      },
      {
        path: '/key',
        name: 'Key',
        component: () => import('@/views/key/Index.vue')
      }
    ]
  }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

export default router
