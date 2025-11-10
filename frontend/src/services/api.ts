import axios from 'axios'
import type { AuthConfig } from '../types'

export const apiClient = axios.create({
  baseURL: '/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
})

// Add response interceptor for error handling
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
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

export default apiClient
