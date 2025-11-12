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

// Admin user management API methods
export async function promoteUserToAdmin(userId: string) {
  try {
    const response = await apiClient.put(`/admin/users/${userId}`, { admin: true })
    return response.data
  } catch (error) {
    console.error('Failed to promote user to admin:', error)
    throw error
  }
}

export async function demoteUserFromAdmin(userId: string) {
  try {
    const response = await apiClient.put(`/admin/users/${userId}`, { admin: false })
    return response.data
  } catch (error) {
    console.error('Failed to demote user from admin:', error)
    throw error
  }
}

export async function deleteUser(userId: string) {
  try {
    const response = await apiClient.delete(`/admin/users/${userId}`)
    return response.data
  } catch (error) {
    console.error('Failed to delete user:', error)
    throw error
  }
}

// Project sharing API methods
export async function fetchProjectShares(projectId: string) {
  try {
    const response = await apiClient.get(`/projects/${projectId}/shares`)
    return response.data
  } catch (error) {
    console.error('Failed to fetch project shares:', error)
    throw error
  }
}

export async function createProjectShare(projectId: string, groupName: string) {
  try {
    const response = await apiClient.post(`/projects/${projectId}/shares`, { group_name: groupName })
    return response.data
  } catch (error) {
    console.error('Failed to create project share:', error)
    throw error
  }
}

export async function deleteProjectShare(projectId: string, shareId: number) {
  try {
    const response = await apiClient.delete(`/projects/${projectId}/shares/${shareId}`)
    return response.data
  } catch (error) {
    console.error('Failed to delete project share:', error)
    throw error
  }
}

export async function transferProjectOwnership(projectId: string, newOwnerId: string) {
  try {
    const response = await apiClient.post(`/projects/${projectId}/transfer-ownership`, { new_owner_id: newOwnerId })
    return response.data
  } catch (error) {
    console.error('Failed to transfer project ownership:', error)
    throw error
  }
}

export async function fetchUsersForOwnershipTransfer() {
  try {
    const response = await apiClient.get('/users')
    return response.data
  } catch (error) {
    console.error('Failed to fetch users for ownership transfer:', error)
    throw error
  }
}

export async function deleteProject(projectId: string) {
  try {
    const response = await apiClient.delete(`/projects/${projectId}`)
    return response.data
  } catch (error) {
    console.error('Failed to delete project:', error)
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

export async function fetchBuild(buildId: string) {
  try {
    const response = await apiClient.get(`/builds/${buildId}`)
    return response.data
  } catch (error) {
    console.error('Failed to fetch build:', error)
    throw error
  }
}

export default apiClient

// Refresh project token
export async function refreshProjectToken(projectId: string) {
  try {
    const response = await apiClient.post(`/projects/${projectId}/refresh-token`)
    return response.data
  } catch (error) {
    console.error('Failed to refresh project token:', error)
    throw error
  }
}
