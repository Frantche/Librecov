import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User, AuthConfig } from '../types'
import { apiClient, fetchAuthConfig } from '../services/api'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const token = ref<string | null>(localStorage.getItem('token'))
  const authConfig = ref<AuthConfig | null>(null)

  const isAuthenticated = computed(() => !!user.value && !!token.value)
  const isOIDCEnabled = computed(() => authConfig.value?.oidc_enabled ?? false)

  const loadAuthConfig = async () => {
    authConfig.value = await fetchAuthConfig()
  }

  const setToken = (newToken: string) => {
    token.value = newToken
    localStorage.setItem('token', newToken)
    apiClient.defaults.headers.common['Authorization'] = `Bearer ${newToken}`
  }

  const clearToken = () => {
    token.value = null
    localStorage.removeItem('token')
    delete apiClient.defaults.headers.common['Authorization']
  }

  const fetchUser = async () => {
    if (!token.value) return

    try {
      const response = await apiClient.get('/auth/me')
      user.value = response.data
    } catch (error) {
      console.error('Failed to fetch user:', error)
      clearToken()
    }
  }

  const loginWithOIDC = () => {
    // Redirect to backend OIDC login endpoint
    window.location.href = '/auth/login'
  }

  const logout = async () => {
    try {
      await apiClient.post('/auth/logout')
    } catch (error) {
      console.error('Logout error:', error)
    } finally {
      user.value = null
      clearToken()
      window.location.href = '/'
    }
  }

  // Initialize auth state
  const initialize = async () => {
    await loadAuthConfig()
    if (token.value) {
      apiClient.defaults.headers.common['Authorization'] = `Bearer ${token.value}`
      await fetchUser()
    }
  }

  // Auto-initialize
  initialize()

  return {
    user,
    token,
    authConfig,
    isAuthenticated,
    isOIDCEnabled,
    setToken,
    clearToken,
    fetchUser,
    loginWithOIDC,
    logout,
    loadAuthConfig,
  }
})
