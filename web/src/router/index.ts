import { createRouter, createWebHistory } from 'vue-router'
import { getToken } from '../api'

const router = createRouter({
  history: createWebHistory('/'),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('../views/Home.vue'),
    },
    {
      path: '/docs',
      name: 'docs',
      component: () => import('../views/ApiDocs.vue'),
    },
    {
      path: '/admin/login',
      name: 'login',
      component: () => import('../views/Login.vue'),
    },
    {
      path: '/admin',
      name: 'dashboard',
      component: () => import('../views/Dashboard.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/admin/links',
      name: 'links',
      component: () => import('../views/Links.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/admin/links/:code/stats',
      name: 'linkStats',
      component: () => import('../views/LinkStats.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/admin/tokens',
      name: 'tokens',
      component: () => import('../views/Tokens.vue'),
      meta: { requiresAuth: true },
    },
  ],
})

router.beforeEach((to, _from, next) => {
  const token = getToken()
  if (to.meta.requiresAuth && !token) {
    next('/admin/login')
  } else if (to.name === 'login' && token) {
    next('/admin')
  } else {
    next()
  }
})

export default router
