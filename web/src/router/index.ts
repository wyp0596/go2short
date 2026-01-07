import { createRouter, createWebHistory } from 'vue-router'
import { getToken } from '../api'

const router = createRouter({
  history: createWebHistory('/admin/'),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/Login.vue'),
    },
    {
      path: '/',
      name: 'dashboard',
      component: () => import('../views/Dashboard.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/links',
      name: 'links',
      component: () => import('../views/Links.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/links/:code/stats',
      name: 'linkStats',
      component: () => import('../views/LinkStats.vue'),
      meta: { requiresAuth: true },
    },
  ],
})

router.beforeEach((to, _from, next) => {
  const token = getToken()
  if (to.meta.requiresAuth && !token) {
    next('/login')
  } else if (to.name === 'login' && token) {
    next('/')
  } else {
    next()
  }
})

export default router
