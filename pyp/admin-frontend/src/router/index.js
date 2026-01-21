import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/Login.vue'),
      meta: { requiresAuth: false }
    },
    {
      path: '/',
      component: () => import('@/layouts/MainLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          name: 'Dashboard',
          component: () => import('@/views/Dashboard.vue')
        },
        {
          path: 'users',
          name: 'Users',
          component: () => import('@/views/Users.vue')
        },
        {
          path: 'collision-codes',
          name: 'CollisionCodes',
          component: () => import('@/views/CollisionCodes.vue')
        },
        {
          path: 'keywords',
          name: 'Keywords',
          component: () => import('@/views/Keywords.vue')
        },
        {
          path: 'forbidden-keywords',
          name: 'ForbiddenKeywords',
          component: () => import('@/views/ForbiddenKeywords.vue')
        },
        {
          path: 'records',
          name: 'Records',
          component: () => import('@/views/Records.vue')
        },
        {
          path: 'admins',
          name: 'Admins',
          component: () => import('@/views/Admins.vue')
        },
        {
          path: 'audit-settings',
          name: 'AuditSettings',
          component: () => import('@/views/AuditSettings.vue')
        },
        {
          path: 'audit-list',
          name: 'AuditList',
          component: () => import('@/views/AuditList.vue')
        },
        {
          path: 'email-logs',
          name: 'EmailLogs',
          component: () => import('@/views/EmailLogs.vue')
        }
      ]
    }
  ]
})

router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next('/login')
  } else if (to.path === '/login' && authStore.isAuthenticated) {
    next('/')
  } else {
    next()
  }
})

export default router
