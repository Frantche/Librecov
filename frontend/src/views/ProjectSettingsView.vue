<template>
  <div class="project-settings-view">
    <div class="header-section">
      <h2>{{ project?.name }} - Settings</h2>
      <router-link :to="`/projects/${projectId}`" class="btn btn-secondary">
        Back to Project
      </router-link>
    </div>

    <div class="settings-section">
      <h3>Project API Tokens</h3>
      <p class="description">
        Create project-specific tokens for uploading coverage data. These tokens can only be used
        for this project.
      </p>

      <button @click="showCreateModal = true" class="btn btn-primary mb-2">
        Create New Token
      </button>

      <div v-if="loadingTokens" class="loading">Loading tokens...</div>

      <div v-else-if="tokens.length === 0" class="empty-state">
        <p>No project tokens found. Create your first token to get started.</p>
      </div>

      <div v-else class="tokens-list">
        <div v-for="token in tokens" :key="token.id" class="token-card card">
          <div class="token-header">
            <h4>{{ token.name }}</h4>
            <button @click="deleteToken(token.id)" class="btn btn-danger btn-sm">
              Delete
            </button>
          </div>
          <div class="token-info">
            <p><strong>Created:</strong> {{ formatDate(token.created_at) }}</p>
            <p v-if="token.last_used">
              <strong>Last Used:</strong> {{ formatDate(token.last_used) }}
            </p>
            <p v-else><strong>Last Used:</strong> Never</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Create Token Modal -->
    <div v-if="showCreateModal" class="modal" @click.self="showCreateModal = false">
      <div class="modal-content">
        <h3>Create New Project Token</h3>
        <form @submit.prevent="createToken">
          <div class="form-group">
            <label for="token-name">Token Name</label>
            <input
              id="token-name"
              v-model="newTokenName"
              type="text"
              placeholder="e.g., GitHub Actions"
              required
              class="form-input"
            />
            <small>Choose a descriptive name to identify where this token is used.</small>
          </div>
          <div class="modal-actions">
            <button type="button" @click="showCreateModal = false" class="btn btn-secondary">
              Cancel
            </button>
            <button type="submit" class="btn btn-primary">Create Token</button>
          </div>
        </form>
      </div>
    </div>

    <!-- Token Created Modal -->
    <div v-if="newlyCreatedToken" class="modal" @click.self="closeTokenModal">
      <div class="modal-content">
        <h3>Token Created Successfully!</h3>
        <div class="success-box">
          <p><strong>Important:</strong> Copy this token now. You won't be able to see it again!</p>
          <div class="token-display">
            <code>{{ newlyCreatedToken.token }}</code>
            <button @click="copyToken" class="btn btn-sm btn-secondary">
              {{ copied ? 'Copied!' : 'Copy' }}
            </button>
          </div>
          <p class="token-name"><strong>Token Name:</strong> {{ newlyCreatedToken.name }}</p>
          
          <div class="usage-example">
            <h4>Usage Example:</h4>
            <pre><code>curl -X POST {{ baseUrl }}/upload/v2 \
  -H "Authorization: Bearer {{ newlyCreatedToken.token }}" \
  -F "json_file=@coverage.json"</code></pre>
          </div>
        </div>
        <div class="modal-actions">
          <button @click="closeTokenModal" class="btn btn-primary">I've Saved My Token</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { apiClient } from '../services/api'
import type { Project } from '../types'

interface Token {
  id: number
  name: string
  token?: string
  created_at: string
  last_used?: string
}

const route = useRoute()
const projectId = computed(() => route.params.id as string)

const project = ref<Project | null>(null)
const tokens = ref<Token[]>([])
const loadingTokens = ref(true)
const showCreateModal = ref(false)
const newTokenName = ref('')
const newlyCreatedToken = ref<Token | null>(null)
const copied = ref(false)

const baseUrl = computed(() => window.location.origin)

const fetchProject = async () => {
  try {
    const response = await apiClient.get(`/projects/${projectId.value}`)
    project.value = response.data
  } catch (error) {
    console.error('Failed to fetch project:', error)
  }
}

const fetchTokens = async () => {
  try {
    loadingTokens.value = true
    const response = await apiClient.get(`/projects/${projectId.value}/tokens`)
    tokens.value = response.data
  } catch (error) {
    console.error('Failed to fetch tokens:', error)
  } finally {
    loadingTokens.value = false
  }
}

const createToken = async () => {
  try {
    const response = await apiClient.post(`/projects/${projectId.value}/tokens`, {
      name: newTokenName.value,
    })
    newlyCreatedToken.value = response.data
    showCreateModal.value = false
    newTokenName.value = ''
    await fetchTokens()
  } catch (error) {
    console.error('Failed to create token:', error)
    alert('Failed to create token. Please try again.')
  }
}

const deleteToken = async (tokenId: number) => {
  if (!confirm('Are you sure you want to delete this token? This action cannot be undone.')) {
    return
  }

  try {
    await apiClient.delete(`/projects/${projectId.value}/tokens/${tokenId}`)
    await fetchTokens()
  } catch (error) {
    console.error('Failed to delete token:', error)
    alert('Failed to delete token. Please try again.')
  }
}

const copyToken = async () => {
  if (newlyCreatedToken.value?.token) {
    try {
      await navigator.clipboard.writeText(newlyCreatedToken.value.token)
      copied.value = true
      setTimeout(() => {
        copied.value = false
      }, 2000)
    } catch (error) {
      console.error('Failed to copy token:', error)
    }
  }
}

const closeTokenModal = () => {
  newlyCreatedToken.value = null
  copied.value = false
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
  fetchTokens()
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

.tokens-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  margin-top: 1rem;
}

.token-card {
  padding: 1.5rem;
}

.token-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.token-header h4 {
  margin: 0;
}

.token-info p {
  margin: 0.5rem 0;
  color: #666;
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

.token-display {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  background: #fff;
  padding: 1rem;
  border-radius: 4px;
  margin: 1rem 0;
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
</style>
