<template>
  <div class="tokens-view">
    <div class="header-section">
      <h2>API Tokens</h2>
      <button @click="showCreateModal = true" class="btn btn-primary">
        Create New Token
      </button>
    </div>

    <div class="info-box">
      <p>
        <strong>API tokens allow you to authenticate to the Librecov API.</strong>
        These tokens can be used to upload coverage data from CI/CD pipelines.
      </p>
      <p>
        Use the token in your API requests with the <code>Authorization: Bearer YOUR_TOKEN</code> header.
      </p>
    </div>

    <div v-if="loading" class="loading">Loading tokens...</div>

    <div v-else-if="tokens.length === 0" class="empty-state">
      <p>No API tokens found. Create your first token to get started.</p>
    </div>

    <div v-else class="tokens-list">
      <div v-for="token in tokens" :key="token.id" class="token-card card">
        <div class="token-header">
          <h3>{{ token.name }}</h3>
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

    <!-- Create Token Modal -->
    <div v-if="showCreateModal" class="modal" @click.self="showCreateModal = false">
      <div class="modal-content">
        <h3>Create New API Token</h3>
        <form @submit.prevent="createToken">
          <div class="form-group">
            <label for="token-name">Token Name</label>
            <input
              id="token-name"
              v-model="newTokenName"
              type="text"
              placeholder="e.g., CI/CD Pipeline"
              required
              class="form-input"
            />
            <small>Choose a descriptive name to help you identify this token later.</small>
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
        </div>
        <div class="modal-actions">
          <button @click="closeTokenModal" class="btn btn-primary">I've Saved My Token</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { apiClient } from '../services/api'

interface Token {
  id: number
  name: string
  token?: string
  created_at: string
  last_used?: string
}

const tokens = ref<Token[]>([])
const loading = ref(true)
const showCreateModal = ref(false)
const newTokenName = ref('')
const newlyCreatedToken = ref<Token | null>(null)
const copied = ref(false)

const fetchTokens = async () => {
  try {
    loading.value = true
    const response = await apiClient.get('/user/tokens')
    tokens.value = response.data
  } catch (error) {
    console.error('Failed to fetch tokens:', error)
  } finally {
    loading.value = false
  }
}

const createToken = async () => {
  try {
    const response = await apiClient.post('/user/tokens', {
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

const deleteToken = async (id: number) => {
  if (!confirm('Are you sure you want to delete this token? This action cannot be undone.')) {
    return
  }

  try {
    await apiClient.delete(`/user/tokens/${id}`)
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
  fetchTokens()
})
</script>

<style scoped>
.tokens-view {
  padding: 1rem 0;
}

.header-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.info-box {
  background: #e3f2fd;
  border-left: 4px solid #2196f3;
  padding: 1rem;
  margin-bottom: 2rem;
  border-radius: 4px;
}

.info-box p {
  margin: 0.5rem 0;
}

.info-box code {
  background: #fff;
  padding: 0.2rem 0.4rem;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
}

.tokens-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
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

.token-header h3 {
  margin: 0;
}

.token-info p {
  margin: 0.5rem 0;
  color: #666;
}

.loading,
.empty-state {
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
  min-width: 500px;
  max-width: 90%;
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

.btn {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 1rem;
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
</style>
