import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import type { NavigationGuardNext, RouteLocationNormalized } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('../views/ProjectsView.vue'),
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/LoginView.vue'),
    },
    {
      path: '/tokens',
      name: 'tokens',
      component: () => import('../views/TokensView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/projects/:id',
      name: 'project',
      component: () => import('../views/ProjectView.vue'),
    },
    {
      path: '/projects/:id/settings',
      name: 'project-settings',
      component: () => import('../views/ProjectSettingsView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/builds/:id',
      name: 'build',
      component: () => import('../views/BuildView.vue'),
      meta: { requiresAuth: true },
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

router.beforeEach(
  (to: RouteLocationNormalized, from: RouteLocationNormalized, next: NavigationGuardNext) => {
    const authStore = useAuthStore()

    if (to.meta.requiresAuth && !authStore.isAuthenticated) {
      // Redirect to login
      next({ name: 'login' })
      return
    }

    if (to.meta.requiresAdmin && !authStore.user?.admin) {
      // Redirect to home if not admin
      next({ name: 'home' })
      return
    }

    next()
  }
)

export default router
