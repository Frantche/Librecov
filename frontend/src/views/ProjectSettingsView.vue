<template>
  <div class="project-settings-view">
    <div class="header-section">
      <h2>{{ project?.name }} - Settings</h2>
      <div class="header-actions">
        <button 
          v-if="isProjectOwner" 
          @click="showTransferModal = true" 
          class="btn btn-warning"
        >
          Transfer Ownership
        </button>
        <button 
          v-if="canDeleteProject" 
          @click="confirmDeleteProject" 
          class="btn btn-danger"
        >
          Delete Project
        </button>
        <router-link :to="`/projects/${projectId}`" class="btn btn-secondary">
          Back to Project
        </router-link>
      </div>
    </div>

    <!-- Project Sharing Section -->
    <div class="settings-section">
      <h3>Project Sharing</h3>
      <p class="description">
        Share this project with groups. Only groups from your OIDC token are available for sharing.
      </p>

      <button @click="showShareModal = true" class="btn btn-primary mb-2">
        Share with Group
      </button>

      <div v-if="loadingShares" class="loading">Loading shares...</div>

      <div v-else-if="shares.length === 0" class="empty-state">
        <p>This project is not shared with any groups.</p>
      </div>

      <div v-else class="shares-list">
        <div v-for="share in shares" :key="share.id" class="share-card card">
          <div class="share-header">
            <h4>{{ share.group_name }}</h4>
            <div class="share-actions">
              <span 
                class="membership-indicator"
                :class="{ 'member-indicator': share.is_user_member, 'non-member-indicator': !share.is_user_member }"
              >
                {{ share.is_user_member ? 'You are a member' : 'You are not a member' }}
              </span>
              <button 
                v-if="isProjectOwner" 
                @click="deleteShare(share.id)" 
                class="btn btn-danger btn-sm"
              >
                Remove
              </button>
            </div>
          </div>
          <div class="share-info">
            <p><strong>Shared:</strong> {{ formatDate(share.created_at) }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Share Project Modal -->
    <div v-if="showShareModal" class="modal" @click.self="showShareModal = false">
      <div class="modal-content">
        <h3>Share Project with Group</h3>
        <form @submit.prevent="createShare">
          <div class="form-group">
            <label for="group-name">Group Name</label>
            <select
              id="group-name"
              v-model="selectedGroup"
              required
              class="form-input"
            >
              <option value="">Select a group...</option>
              <option v-for="group in userGroups" :key="group" :value="group">
                {{ group }}
              </option>
            </select>
            <small>Only groups from your OIDC token are available for selection.</small>
          </div>
          <div v-if="userGroups.length === 0" class="warning-box">
            <p><strong>No groups available.</strong> You don't have any groups in your OIDC token.</p>
          </div>
          <div class="modal-actions">
            <button type="button" @click="showShareModal = false" class="btn btn-secondary">
              Cancel
            </button>
            <button type="submit" class="btn btn-primary" :disabled="userGroups.length === 0">
              Share Project
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Transfer Ownership Modal -->
    <div v-if="showTransferModal" class="modal" @click.self="showTransferModal = false">
      <div class="modal-content">
        <h3>Transfer Project Ownership</h3>
        <div class="warning-box">
          <p><strong>Warning:</strong> Transferring ownership will give the new owner full control over this project, including the ability to delete it and manage its shares. This action cannot be undone.</p>
        </div>
        <form @submit.prevent="transferOwnership">
          <div class="form-group">
            <label for="new-owner">New Owner</label>
            <select
              id="new-owner"
              v-model="selectedNewOwner"
              required
              class="form-input"
            >
              <option value="">Select a new owner...</option>
              <option 
                v-for="user in allUsers" 
                :key="user.id" 
                :value="user.id.toString()"
                :disabled="user.id === project?.user_id"
              >
                {{ user.name }} ({{ user.email }})
                <span v-if="user.id === project?.user_id"> - Current Owner</span>
              </option>
            </select>
            <small>Select the user who will become the new owner of this project.</small>
          </div>
          <div class="modal-actions">
            <button type="button" @click="showTransferModal = false" class="btn btn-secondary">
              Cancel
            </button>
            <button type="submit" class="btn btn-danger" :disabled="!selectedNewOwner">
              Transfer Ownership
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { apiClient, fetchProjectShares, createProjectShare, deleteProjectShare, fetchUserGroups, fetchUsersForOwnershipTransfer, transferProjectOwnership, deleteProject } from '../services/api'
import { useAuthStore } from '../stores/auth'
import type { Project, ProjectShare, User } from '../types'

const route = useRoute()
const projectId = computed(() => route.params.id as string)
const authStore = useAuthStore()

const project = ref<Project | null>(null)
const shares = ref<ProjectShare[]>([])
const userGroups = ref<string[]>([])
const allUsers = ref<User[]>([])
const loadingShares = ref(true)
const showShareModal = ref(false)
const showTransferModal = ref(false)
const selectedGroup = ref('')
const selectedNewOwner = ref('')

const isProjectOwner = computed(() => {
  return project.value && authStore.user && project.value.user_id === authStore.user.id
})

const canDeleteProject = computed(() => {
  return isProjectOwner.value || (authStore.user && authStore.user.admin)
})

const fetchProject = async () => {
  try {
    const response = await apiClient.get(`/projects/${projectId.value}`)
    project.value = response.data
  } catch (error) {
    console.error('Failed to fetch project:', error)
  }
}

const fetchShares = async () => {
  try {
    loadingShares.value = true
    shares.value = await fetchProjectShares(projectId.value)
  } catch (error) {
    console.error('Failed to fetch shares:', error)
  } finally {
    loadingShares.value = false
  }
}

const loadUserGroups = async () => {
  try {
    userGroups.value = await fetchUserGroups()
  } catch (error) {
    console.error('Failed to fetch user groups:', error)
  }
}

const loadAllUsers = async () => {
  try {
    allUsers.value = await fetchUsersForOwnershipTransfer()
  } catch (error) {
    console.error('Failed to fetch all users:', error)
  }
}

const createShare = async () => {
  if (!selectedGroup.value) {
    return
  }

  try {
    await createProjectShare(projectId.value, selectedGroup.value)
    showShareModal.value = false
    selectedGroup.value = ''
    await fetchShares()
  } catch (error: any) {
    console.error('Failed to create share:', error)
    const errorMsg = error.response?.data?.error || 'Failed to create share. Please try again.'
    alert(errorMsg)
  }
}

const deleteShare = async (shareId: number) => {
  if (!confirm('Are you sure you want to remove this group share?')) {
    return
  }

  try {
    await deleteProjectShare(projectId.value, shareId)
    await fetchShares()
  } catch (error) {
    console.error('Failed to delete share:', error)
    alert('Failed to delete share. Please try again.')
  }
}

const transferOwnership = async () => {
  if (!selectedNewOwner.value) {
    return
  }

  if (!confirm('Are you sure you want to transfer ownership of this project? This action cannot be undone.')) {
    return
  }

  try {
    await transferProjectOwnership(projectId.value, selectedNewOwner.value)
    showTransferModal.value = false
    selectedNewOwner.value = ''
    await fetchProject() // Refresh project data to show new owner
    alert('Ownership transferred successfully!')
  } catch (error: any) {
    console.error('Failed to transfer ownership:', error)
    const errorMsg = error.response?.data?.error || 'Failed to transfer ownership. Please try again.'
    alert(errorMsg)
  }
}

const confirmDeleteProject = async () => {
  if (!project.value) return

  const message = `Are you sure you want to delete the project "${project.value.name}"? This action cannot be undone and will permanently delete all project data including builds, jobs, and coverage information.`

  if (!confirm(message)) {
    return
  }

  try {
    await deleteProject(projectId.value)
    alert('Project deleted successfully!')
    // Navigate back to projects list
    window.location.href = '/projects'
  } catch (error: any) {
    console.error('Failed to delete project:', error)
    const errorMsg = error.response?.data?.error || 'Failed to delete project. Please try again.'
    alert(errorMsg)
  }
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

onMounted(() => {
  fetchProject()
  fetchShares()
  loadUserGroups()
  loadAllUsers()
})
</script>

<style scoped>
.project-settings-view {
  padding: 1rem 0;
}

.header-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.header-actions {
  display: flex;
  gap: 1rem;
  align-items: center;
}

.settings-section {
  background: white;
  padding: 2rem;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  margin-bottom: 2rem;
}

.settings-section h3 {
  margin-top: 0;
}

.description {
  color: #666;
  margin-bottom: 1.5rem;
}

.mb-2 {
  margin-bottom: 1rem;
}

.loading,
.empty-state {
  text-align: center;
  padding: 2rem;
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
  min-width: 500px;
  max-width: 90%;
  max-height: 90vh;
  overflow-y: auto;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: bold;
}

.form-group small {
  display: block;
  margin-top: 0.25rem;
  color: #666;
  font-size: 0.875rem;
}

.form-input {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.modal-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  margin-top: 1.5rem;
}

.success-box {
  background: #d4edda;
  border: 1px solid #c3e6cb;
  padding: 1rem;
  border-radius: 4px;
  margin: 1rem 0;
}

.success-box p {
  margin: 0.5rem 0;
}

.token-display code {
  flex: 1;
  word-break: break-all;
  font-family: 'Courier New', monospace;
  font-size: 0.9rem;
}

.token-name {
  margin-top: 0.5rem;
}

.usage-example {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid #c3e6cb;
}

.usage-example h4 {
  margin-top: 0;
  margin-bottom: 0.5rem;
}

.usage-example pre {
  background: #f8f9fa;
  padding: 1rem;
  border-radius: 4px;
  overflow-x: auto;
}

.usage-example code {
  font-family: 'Courier New', monospace;
  font-size: 0.875rem;
}

.btn {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 1rem;
  text-decoration: none;
  display: inline-block;
}

.btn-primary {
  background: #3498db;
  color: white;
}

.btn-primary:hover {
  background: #2980b9;
}

.btn-secondary {
  background: #95a5a6;
  color: white;
}

.btn-secondary:hover {
  background: #7f8c8d;
}

.btn-danger {
  background: #e74c3c;
  color: white;
}

.btn-danger:hover {
  background: #c0392b;
}

.btn-warning {
  background: #f39c12;
  color: white;
}

.btn-warning:hover {
  background: #e67e22;
}

.btn-sm {
  padding: 0.25rem 0.75rem;
  font-size: 0.875rem;
}

.card {
  background: white;
  border: 1px solid #ddd;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.shares-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  margin-top: 1rem;
}

.share-card {
  padding: 1.5rem;
}

.share-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.share-header h4 {
  margin: 0;
}

.share-actions {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.membership-indicator {
  font-size: 0.75rem;
  padding: 0.25rem 0.5rem;
  border-radius: 12px;
  font-weight: 500;
}

.member-indicator {
  background: #e8f5e8;
  color: #2e7d32;
  border: 1px solid #4caf50;
}

.non-member-indicator {
  background: #fff3e0;
  color: #ef6c00;
  border: 1px solid #ff9800;
}

.share-info p {
  margin: 0.5rem 0;
  color: #666;
}

.warning-box {
  background: #fff3cd;
  border: 1px solid #ffeeba;
  padding: 1rem;
  border-radius: 4px;
  margin: 1rem 0;
  color: #856404;
}

.warning-box p {
  margin: 0;
}
</style>
