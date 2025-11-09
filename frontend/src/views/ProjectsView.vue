<template>
  <div class="projects-view">
    <div class="header-section">
      <h2>Projects</h2>
      <button v-if="authStore.isAuthenticated" @click="showCreateModal = true" class="btn btn-primary">
        New Project
      </button>
    </div>

    <div v-if="loading" class="loading">Loading projects...</div>

    <div v-else-if="projects.length === 0" class="empty-state">
      <p>No projects found. Create your first project to get started.</p>
    </div>

    <div v-else class="projects-grid">
      <div v-for="project in projects" :key="project.id" class="project-card card">
        <router-link :to="`/projects/${project.id}`" class="project-link">
          <h3>{{ project.name }}</h3>
          <div class="coverage-badge" :class="getCoverageClass(project.coverage_rate)">
            {{ project.coverage_rate.toFixed(1) }}%
          </div>
          <div class="project-info">
            <span>Branch: {{ project.current_branch || 'main' }}</span>
          </div>
        </router-link>
      </div>
    </div>

    <!-- Create Project Modal - simplified for now -->
    <div v-if="showCreateModal" class="modal" @click.self="showCreateModal = false">
      <div class="modal-content">
        <h3>Create New Project</h3>
        <form @submit.prevent="createProject">
          <input
            v-model="newProject.name"
            type="text"
            placeholder="Project Name"
            required
            class="form-input"
          />
          <input
            v-model="newProject.current_branch"
            type="text"
            placeholder="Main Branch (optional)"
            class="form-input"
          />
          <div class="modal-actions">
            <button type="button" @click="showCreateModal = false" class="btn btn-secondary">
              Cancel
            </button>
            <button type="submit" class="btn btn-primary">Create</button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { apiClient } from '../services/api'
import type { Project } from '../types'

const authStore = useAuthStore()
const projects = ref<Project[]>([])
const loading = ref(true)
const showCreateModal = ref(false)
const newProject = ref({
  name: '',
  current_branch: 'main',
})

const fetchProjects = async () => {
  try {
    loading.value = true
    const response = await apiClient.get('/projects')
    projects.value = response.data
  } catch (error) {
    console.error('Failed to fetch projects:', error)
  } finally {
    loading.value = false
  }
}

const createProject = async () => {
  try {
    await apiClient.post('/projects', newProject.value)
    showCreateModal.value = false
    newProject.value = { name: '', current_branch: 'main' }
    await fetchProjects()
  } catch (error) {
    console.error('Failed to create project:', error)
  }
}

const getCoverageClass = (coverage: number) => {
  if (coverage >= 80) return 'high'
  if (coverage >= 60) return 'medium'
  return 'low'
}

onMounted(() => {
  if (authStore.isAuthenticated) {
    fetchProjects()
  } else {
    loading.value = false
  }
})
</script>

<style scoped>
.projects-view {
  padding: 1rem 0;
}

.header-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.projects-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1.5rem;
}

.project-card {
  transition: transform 0.2s;
}

.project-card:hover {
  transform: translateY(-4px);
}

.project-link {
  text-decoration: none;
  color: inherit;
}

.coverage-badge {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  border-radius: 12px;
  font-weight: bold;
  margin: 0.5rem 0;
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

.project-info {
  color: #7f8c8d;
  font-size: 0.9rem;
  margin-top: 0.5rem;
}

.loading, .empty-state {
  text-align: center;
  padding: 3rem;
  color: #7f8c8d;
}

.modal {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  padding: 2rem;
  border-radius: 8px;
  min-width: 400px;
}

.form-input {
  width: 100%;
  padding: 0.5rem;
  margin: 0.5rem 0;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.modal-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  margin-top: 1rem;
}
</style>
