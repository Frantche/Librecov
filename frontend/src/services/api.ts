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

// Admin API methods
export async function fetchAllUsers() {
  try {
    const response = await apiClient.get('/admin/users')
    return response.data
  } catch (error) {
    console.error('Failed to fetch all users:', error)
    throw error
  }
}

export async function fetchAllProjects() {
  try {
    const response = await apiClient.get('/admin/projects')
    return response.data
  } catch (error) {
    console.error('Failed to fetch all projects:', error)
    throw error
  }
}

// Project sharing API methods
export async function fetchProjectShares(projectId: number) {
  try {
    const response = await apiClient.get(`/projects/${projectId}/shares`)
    return response.data
  } catch (error) {
    console.error('Failed to fetch project shares:', error)
    throw error
  }
}

export async function createProjectShare(projectId: number, groupName: string) {
  try {
    const response = await apiClient.post(`/projects/${projectId}/shares`, { group_name: groupName })
    return response.data
  } catch (error) {
    console.error('Failed to create project share:', error)
    throw error
  }
}

export async function deleteProjectShare(projectId: number, shareId: number) {
  try {
    const response = await apiClient.delete(`/projects/${projectId}/shares/${shareId}`)
    return response.data
  } catch (error) {
    console.error('Failed to delete project share:', error)
    throw error
  }
}

// Get user groups
export async function fetchUserGroups() {
  try {
    const response = await axios.get('/auth/groups', { withCredentials: true })
    return response.data.groups || []
  } catch (error) {
    console.error('Failed to fetch user groups:', error)
    return []
  }
}

export default apiClient
