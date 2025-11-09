import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('../views/ProjectsView.vue'),
    },
    {
      path: '/projects/:id',
      name: 'project',
      component: () => import('../views/ProjectView.vue'),
    },
    {
      path: '/builds/:id',
      name: 'build',
      component: () => import('../views/BuildView.vue'),
    },
    {
      path: '/jobs/:id',
      name: 'job',
      component: () => import('../views/JobView.vue'),
    },
    {
      path: '/admin',
      name: 'admin',
      component: () => import('../views/AdminView.vue'),
      meta: { requiresAuth: true, requiresAdmin: true },
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import('../views/SettingsView.vue'),
      meta: { requiresAuth: true },
    },
  ],
})

router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    // Redirect to login
    window.location.href = '/auth/login'
    return
  }

  if (to.meta.requiresAdmin && !authStore.user?.admin) {
    // Redirect to home if not admin
    next({ name: 'home' })
    return
  }

  next()
})

export default router
