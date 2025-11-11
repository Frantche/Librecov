<template>
  <div class="admin-view">
    <h2>Admin Panel</h2>
    <p>Manage users and system settings.</p>

    <!-- Users Section -->
    <div class="section">
      <h3>Users</h3>
      <div v-if="loadingUsers" class="loading">Loading users...</div>
      <div v-else-if="users.length === 0" class="empty-state">
        <p>No users found.</p>
      </div>
      <div v-else class="table-container">
        <table class="data-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>Email</th>
              <th>Name</th>
              <th>Admin</th>
              <th>Created</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="user in users" :key="user.id">
              <td>{{ user.id }}</td>
              <td>{{ user.email }}</td>
              <td>{{ user.name }}</td>
              <td>
                <span :class="['badge', user.admin ? 'badge-success' : 'badge-default']">
                  {{ user.admin ? 'Yes' : 'No' }}
                </span>
              </td>
              <td>{{ formatDate(user.created_at) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Projects Section -->
    <div class="section">
      <h3>All Projects</h3>
      <div v-if="loadingProjects" class="loading">Loading projects...</div>
      <div v-else-if="projects.length === 0" class="empty-state">
        <p>No projects found.</p>
      </div>
      <div v-else class="table-container">
        <table class="data-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>Name</th>
              <th>Owner</th>
              <th>Coverage</th>
              <th>Branch</th>
              <th>Shared With</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="project in projects" :key="project.id">
              <td>{{ project.id }}</td>
              <td>
                <router-link :to="`/projects/${project.id}`" class="link">
                  {{ project.name }}
                </router-link>
              </td>
              <td>{{ project.user?.email || 'N/A' }}</td>
              <td>
                <span :class="['coverage-badge', getCoverageClass(project.coverage_rate)]">
                  {{ project.coverage_rate.toFixed(1) }}%
                </span>
              </td>
              <td>{{ project.current_branch || 'main' }}</td>
              <td>
                <span v-if="project.shares && project.shares.length > 0" class="shares-info">
                  {{ project.shares.map(s => s.group_name).join(', ') }}
                </span>
                <span v-else class="text-muted">None</span>
              </td>
              <td>
                <router-link :to="`/projects/${project.id}`" class="btn btn-sm btn-primary">
                  View
                </router-link>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { fetchAllUsers, fetchAllProjects } from '../services/api'
import type { User, Project } from '../types'

const users = ref<User[]>([])
const projects = ref<Project[]>([])
const loadingUsers = ref(true)
const loadingProjects = ref(true)

const loadUsers = async () => {
  try {
    loadingUsers.value = true
    users.value = await fetchAllUsers()
  } catch (error) {
    console.error('Failed to load users:', error)
  } finally {
    loadingUsers.value = false
  }
}

const loadProjects = async () => {
  try {
    loadingProjects.value = true
    projects.value = await fetchAllProjects()
  } catch (error) {
    console.error('Failed to load projects:', error)
  } finally {
    loadingProjects.value = false
  }
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString()
}

const getCoverageClass = (coverage: number) => {
  if (coverage >= 80) return 'high'
  if (coverage >= 60) return 'medium'
  return 'low'
}

onMounted(() => {
  loadUsers()
  loadProjects()
})
</script>

<style scoped>
.admin-view {
  padding: 1rem 0;
}

.section {
  margin-bottom: 3rem;
}

.section h3 {
  margin-bottom: 1rem;
  padding-bottom: 0.5rem;
  border-bottom: 2px solid #e0e0e0;
}

.loading, .empty-state {
  text-align: center;
  padding: 2rem;
  color: #7f8c8d;
}

.table-container {
  overflow-x: auto;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  background: white;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  border-radius: 4px;
}

.data-table thead {
  background: #f8f9fa;
}

.data-table th,
.data-table td {
  padding: 0.75rem 1rem;
  text-align: left;
  border-bottom: 1px solid #e0e0e0;
}

.data-table th {
  font-weight: 600;
  color: #2c3e50;
}

.data-table tbody tr:hover {
  background: #f8f9fa;
}

.badge {
  display: inline-block;
  padding: 0.25rem 0.5rem;
  border-radius: 12px;
  font-size: 0.85rem;
  font-weight: 600;
}

.badge-success {
  background: #d4edda;
  color: #155724;
}

.badge-default {
  background: #e2e3e5;
  color: #383d41;
}

.coverage-badge {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  border-radius: 12px;
  font-weight: bold;
  font-size: 0.9rem;
}

.coverage-badge.high {
  background: #2ecc71;
  color: white;
}

.coverage-badge.medium {
  background: #f39c12;
  color: white;
}

.coverage-badge.low {
  background: #e74c3c;
  color: white;
}

.link {
  color: #3498db;
  text-decoration: none;
}

.link:hover {
  text-decoration: underline;
}

.shares-info {
  font-size: 0.9rem;
  color: #555;
}

.text-muted {
  color: #95a5a6;
  font-style: italic;
}

.btn-sm {
  padding: 0.375rem 0.75rem;
  font-size: 0.875rem;
}
</style>
