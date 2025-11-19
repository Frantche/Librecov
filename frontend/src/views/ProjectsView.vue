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
        
        <!-- Shared Groups Section -->
        <div v-if="project.shares && project.shares.length > 0" class="shared-groups">
          <div class="shared-label">Shared with:</div>
          <div class="group-tags">
            <span 
              v-for="share in project.shares" 
              :key="share.id"
              class="group-tag"
              :class="{ 'member-group': share.is_user_member, 'non-member-group': !share.is_user_member }"
            >
              {{ share.group_name }}
            </span>
          </div>
        </div>
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
import { ref, onMounted, watch } from 'vue'
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

watch(() => authStore.isAuthenticated, (isAuth) => {
  if (isAuth) {
    fetchProjects()
  } else {
    projects.value = []
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
  gap: 1rem;
  flex-wrap: wrap;
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

/* Shared Groups Section */
.shared-groups {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid #eee;
}

.shared-label {
  font-size: 0.875rem;
  color: #666;
  margin-bottom: 0.5rem;
  font-weight: 500;
}

.group-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.group-tag {
  display: inline-block;
  padding: 0.25rem 0.5rem;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 500;
  border: 1px solid;
}

.member-group {
  background: #e8f5e8;
  color: #2e7d32;
  border-color: #4caf50;
}

.non-member-group {
  background: #fff3e0;
  color: #ef6c00;
  border-color: #ff9800;
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
  padding: 1rem;
}

.modal-content {
  background: white;
  padding: 2rem;
  border-radius: 8px;
  min-width: 400px;
  max-width: 100%;
  width: 100%;
  max-width: 500px;
}

.form-input {
  width: 100%;
  padding: 0.5rem;
  margin: 0.5rem 0;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 1rem;
}

.modal-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  margin-top: 1rem;
  flex-wrap: wrap;
}

/* Tablet breakpoint */
@media (max-width: 768px) {
  .projects-grid {
    grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
    gap: 1rem;
  }
  
  .header-section {
    margin-bottom: 1.5rem;
  }
  
  .header-section h2 {
    font-size: 1.5rem;
  }
}

/* Mobile breakpoint */
@media (max-width: 640px) {
  .projects-view {
    padding: 0.75rem 0;
  }
  
  .header-section {
    flex-direction: column;
    align-items: stretch;
  }
  
  .header-section h2 {
    font-size: 1.25rem;
    text-align: center;
  }
  
  .projects-grid {
    grid-template-columns: 1fr;
    gap: 1rem;
  }
  
  .modal {
    padding: 0.5rem;
  }
  
  .modal-content {
    padding: 1.5rem;
    min-width: 0;
  }
  
  .modal-actions {
    flex-direction: column-reverse;
  }
  
  .modal-actions button {
    width: 100%;
  }
  
  .loading, .empty-state {
    padding: 2rem 1rem;
  }
}
</style>
