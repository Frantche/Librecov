import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User, AuthConfig } from '../types'
import { fetchAuthConfig } from '../services/api'
import axios from 'axios'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const authConfig = ref<AuthConfig | null>(null)
  const refreshInterval = ref<number | null>(null)

  const isAuthenticated = computed(() => !!user.value)
  const isOIDCEnabled = computed(() => authConfig.value?.oidc_enabled ?? false)

  const loadAuthConfig = async () => {
    authConfig.value = await fetchAuthConfig()
  }

  const fetchUser = async () => {
    try {
      const response = await axios.get('/auth/me', { withCredentials: true })
      user.value = response.data
    } catch (error) {
      console.error('Failed to fetch user:', error)
      user.value = null
    }
  }

  const loginWithOIDC = async () => {
    window.location.href = '/auth/login'
  }

  const refreshSession = async () => {
    try {
      // Use axios directly to avoid /api/v1 prefix
      const response = await axios.post('/auth/refresh', {}, { withCredentials: true })
      user.value = response.data.user
      return true
    } catch (error) {
      console.error('Failed to refresh session:', error)
      user.value = null
      return false
    }
  }

  const startSessionRefresh = () => {
    if (refreshInterval.value) {
      clearInterval(refreshInterval.value)
    }
    refreshInterval.value = window.setInterval(async () => {
      const success = await refreshSession()
      if (!success) {
        stopSessionRefresh()
      }
    }, 15 * 60 * 1000)
  }

  const stopSessionRefresh = () => {
    if (refreshInterval.value) {
      clearInterval(refreshInterval.value)
      refreshInterval.value = null
    }
  }

  const logout = async () => {
    try {
      await axios.post('/auth/logout', {}, { withCredentials: true })
    } catch (error) {
      console.error('Logout error:', error)
    } finally {
      user.value = null
      stopSessionRefresh()
      window.location.href = '/'
    }
  }

  const initialize = async () => {
    await loadAuthConfig()
    const success = await refreshSession()
    if (success) {
      startSessionRefresh()
    }
  }

  return {
    user,
    authConfig,
    isAuthenticated,
    isOIDCEnabled,
    fetchUser,
    loginWithOIDC,
    refreshSession,
    logout,
    loadAuthConfig,
    initialize,
    startSessionRefresh,
    stopSessionRefresh,
  }
})
