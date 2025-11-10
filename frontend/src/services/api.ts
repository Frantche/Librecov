import axios, { AxiosResponse } from 'axios'
import type { AuthConfig } from '../types'

export const apiClient = axios.create({
  baseURL: '/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true, // Include cookies in requests
})

// Add response interceptor for error handling
apiClient.interceptors.response.use(
  (response: AxiosResponse) => response,
  (error: any) => {
    if (error.response?.status === 401) {
      // Redirect to login on unauthorized
      localStorage.removeItem('token')
      window.location.href = '/auth/login'
    }
    return Promise.reject(error)
  }
)

// Fetch authentication configuration from the server
export async function fetchAuthConfig(): Promise<AuthConfig> {
  try {
    const response = await axios.get('/auth/config')
    return response.data
  } catch (error) {
    console.error('Failed to fetch auth config:', error)
    // Return default config if fetch fails
    return {
      oidc_enabled: false,
    }
  }
}

export async function refreshSession() {
  try {
    const response = await apiClient.post('/auth/refresh')
    return response.data
  } catch (error) {
    console.error('Failed to refresh session:', error)
    throw error
  }
}

export default apiClient
