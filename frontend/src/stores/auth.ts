import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User } from '../types'
import { apiClient } from '../services/api'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const token = ref<string | null>(localStorage.getItem('token'))

  const isAuthenticated = computed(() => !!user.value && !!token.value)

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
  if (token.value) {
    apiClient.defaults.headers.common['Authorization'] = `Bearer ${token.value}`
    fetchUser()
  }

  return {
    user,
    token,
    isAuthenticated,
    setToken,
    clearToken,
    fetchUser,
    logout,
  }
})
