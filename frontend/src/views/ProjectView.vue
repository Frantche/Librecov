<template>
  <div class="project-view">
    <div class="header-section">
      <div>
        <h2>{{ project?.name || 'Project Details' }}</h2>
        <p class="project-id">ID: {{ projectId }}</p>
      </div>
      <router-link :to="`/projects/${projectId}/settings`" class="btn btn-secondary">
        Project Settings
      </router-link>
    </div>

    <!-- Project Badge -->
    <div v-if="project" class="badge-section card">
      <h3>Project Badge</h3>
      <p class="description">Use this badge to display coverage status in your README</p>
      <div class="badge-container">
        <img :src="badgeUrl" :alt="`${project.name} coverage badge`" class="badge-img" />
        <button @click="copyBadgeMarkdown" class="btn btn-sm btn-primary">
          {{ badgeCopied ? 'Copied!' : 'Copy Markdown' }}
        </button>
      </div>
      <div class="badge-code">
        <code>{{ badgeMarkdown }}</code>
      </div>
    </div>

    <!-- Project Token Info (for shared users and owners) -->
    <div v-if="project" class="token-info-section card">
      <h3>Project Token</h3>
      <p class="description">Use this token to upload coverage reports</p>
      <div class="token-actions">
        <button @click="showTokenModal = true" class="btn btn-primary">
          View Token Info
        </button>
        <button @click="handleRefreshToken" class="btn btn-secondary" :disabled="refreshing">
          {{ refreshing ? 'Refreshing...' : 'Refresh Token' }}
        </button>
      </div>
      <p class="warning-text">⚠️ Refreshing the token will invalidate the current token</p>
    </div>

    <!-- Builds Section -->
    <div class="builds-section card">
      <h3>Builds</h3>
      <div v-if="loadingBuilds" class="loading">Loading builds...</div>
      <div v-else-if="builds.length === 0" class="empty-state">
        <p>No builds yet. Upload coverage data to see builds here.</p>
      </div>
      <div v-else class="builds-list">
        <div 
          v-for="build in builds" 
          :key="build.id" 
          class="build-card"
          @click="selectedBuild = selectedBuild?.id === build.id ? null : build"
        >
          <div class="build-header">
            <div class="build-info">
              <span class="build-number">#{{ build.build_num }}</span>
              <span class="build-branch">{{ build.branch }}</span>
              <span class="build-coverage" :class="getCoverageClass(build.coverage_rate)">
                {{ build.coverage_rate.toFixed(2) }}%
              </span>
            </div>
            <span class="build-date">{{ formatDate(build.created_at) }}</span>
          </div>
          <div v-if="build.commit_sha" class="build-commit">
            <code class="commit-sha">{{ build.commit_sha.substring(0, 7) }}</code>
            <span class="commit-msg">{{ build.commit_msg }}</span>
          </div>
          
          <!-- Build Details (expanded) -->
          <div v-if="selectedBuild?.id === build.id" class="build-details">
            <h4>Build Details</h4>
            <div class="detail-row">
              <strong>Build Number:</strong> {{ build.build_num }}
            </div>
            <div class="detail-row">
              <strong>Branch:</strong> {{ build.branch || 'N/A' }}
            </div>
            <div class="detail-row">
              <strong>Commit SHA:</strong> <code>{{ build.commit_sha || 'N/A' }}</code>
            </div>
            <div class="detail-row">
              <strong>Coverage Rate:</strong> {{ build.coverage_rate.toFixed(2) }}%
            </div>
            <div class="detail-row">
              <strong>Created:</strong> {{ formatDate(build.created_at) }}
            </div>
            <div v-if="build.jobs && build.jobs.length > 0" class="jobs-section">
              <h5>Jobs ({{ build.jobs.length }})</h5>
              <div v-for="job in build.jobs" :key="job.id" class="job-item">
                <span>Job {{ job.job_number }}</span>
                <span class="job-coverage">{{ job.coverage_rate.toFixed(2) }}%</span>
              </div>
            </div>
            <div class="build-actions">
              <router-link :to="`/builds/${build.id}`" class="btn btn-primary">
                View Full Details
              </router-link>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Token Info Modal -->
    <div v-if="showTokenModal" class="modal" @click.self="showTokenModal = false">
      <div class="modal-content">
        <h3>Project Token Information</h3>
        <div class="info-box">
          <p><strong>Project:</strong> {{ project?.name }}</p>
          <p><strong>Token:</strong></p>
          <div class="token-display">
            <code>{{ project?.token || 'Loading...' }}</code>
            <button @click="copyToken" class="btn btn-sm btn-secondary">
              {{ tokenCopied ? 'Copied!' : 'Copy' }}
            </button>
          </div>
          <div class="usage-example">
            <h4>Usage Example:</h4>
            <pre><code>curl -X POST {{ baseUrl }}/upload/v2 \
  -H "Content-Type: application/json" \
  -d '{
    "repo_token": "{{ project?.token }}",
    "source_files": [...]
  }'</code></pre>
          </div>
        </div>
        <div class="modal-actions">
          <button @click="showTokenModal = false" class="btn btn-primary">Close</button>
        </div>
      </div>
    </div>

    <!-- Refresh Token Confirmation Modal -->
    <div v-if="showRefreshConfirm" class="modal" @click.self="showRefreshConfirm = false">
      <div class="modal-content">
        <h3>Refresh Project Token?</h3>
        <div class="warning-box">
          <p><strong>⚠️ Warning:</strong> This will generate a new token and invalidate the current one.</p>
          <p>Any integrations using the current token will stop working until you update them with the new token.</p>
          <p>Are you sure you want to continue?</p>
        </div>
        <div class="modal-actions">
          <button @click="showRefreshConfirm = false" class="btn btn-secondary">Cancel</button>
          <button @click="confirmRefreshToken" class="btn btn-danger">Yes, Refresh Token</button>
        </div>
      </div>
    </div>

    <!-- New Token Modal -->
    <div v-if="newToken" class="modal" @click.self="closeNewTokenModal">
      <div class="modal-content">
        <h3>Token Refreshed Successfully!</h3>
        <div class="success-box">
          <p><strong>Important:</strong> Copy this token now. You won't be able to see it again!</p>
          <div class="token-display">
            <code>{{ newToken }}</code>
            <button @click="copyNewToken" class="btn btn-sm btn-secondary">
              {{ newTokenCopied ? 'Copied!' : 'Copy' }}
            </button>
          </div>
        </div>
        <div class="modal-actions">
          <button @click="closeNewTokenModal" class="btn btn-primary">I've Saved My Token</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { apiClient, refreshProjectToken } from '../services/api'
import type { Project, Build } from '../types'

const route = useRoute()
const router = useRouter()
const projectId = computed(() => route.params.id as string)

const project = ref<Project | null>(null)
const builds = ref<Build[]>([])
const selectedBuild = ref<Build | null>(null)
const loadingBuilds = ref(true)
const showTokenModal = ref(false)
const showRefreshConfirm = ref(false)
const refreshing = ref(false)
const newToken = ref<string | null>(null)
const tokenCopied = ref(false)
const newTokenCopied = ref(false)
const badgeCopied = ref(false)

const baseUrl = computed(() => window.location.origin)
const badgeUrl = computed(() => `${baseUrl.value}/projects/${projectId.value}/badge.svg`)
const badgeMarkdown = computed(() => 
  `![Coverage](${badgeUrl.value})`
)

const fetchProject = async () => {
  try {
    const response = await apiClient.get(`/projects/${projectId.value}`)
    project.value = response.data
  } catch (error: any) {
    if (error.response?.status === 404) {
      router.replace('/')
      return
    }
    console.error('Failed to fetch project:', error)
  }
}

const fetchBuilds = async () => {
  try {
    loadingBuilds.value = true
    const response = await apiClient.get(`/projects/${projectId.value}/builds`)
    builds.value = response.data
  } catch (error) {
    console.error('Failed to fetch builds:', error)
  } finally {
    loadingBuilds.value = false
  }
}

const handleRefreshToken = () => {
  showRefreshConfirm.value = true
}

const confirmRefreshToken = async () => {
  try {
    refreshing.value = true
    showRefreshConfirm.value = false
    const response = await refreshProjectToken(projectId.value)
    newToken.value = response.token
    // Update project token in memory
    if (project.value) {
      project.value.token = response.token
    }
  } catch (error) {
    console.error('Failed to refresh token:', error)
    alert('Failed to refresh token. Please try again.')
  } finally {
    refreshing.value = false
  }
}

const closeNewTokenModal = () => {
  newToken.value = null
  newTokenCopied.value = false
}

const copyToken = async () => {
  if (project.value?.token) {
    try {
      await navigator.clipboard.writeText(project.value.token)
      tokenCopied.value = true
      setTimeout(() => {
        tokenCopied.value = false
      }, 2000)
    } catch (error) {
      console.error('Failed to copy token:', error)
    }
  }
}

const copyNewToken = async () => {
  if (newToken.value) {
    try {
      await navigator.clipboard.writeText(newToken.value)
      newTokenCopied.value = true
      setTimeout(() => {
        newTokenCopied.value = false
      }, 2000)
    } catch (error) {
      console.error('Failed to copy token:', error)
    }
  }
}

const copyBadgeMarkdown = async () => {
  try {
    await navigator.clipboard.writeText(badgeMarkdown.value)
    badgeCopied.value = true
    setTimeout(() => {
      badgeCopied.value = false
    }, 2000)
  } catch (error) {
    console.error('Failed to copy badge markdown:', error)
  }
}

const getCoverageClass = (rate: number) => {
  if (rate >= 80) return 'coverage-high'
  if (rate >= 60) return 'coverage-medium'
  return 'coverage-low'
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

onMounted(() => {
  fetchProject()
  fetchBuilds()
})
</script>

<style scoped>
.project-view {
  padding: 1rem 0;
}

.header-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.project-id {
  color: #666;
  font-size: 0.875rem;
  margin-top: 0.25rem;
}

.card {
  background: white;
  padding: 2rem;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  margin-bottom: 2rem;
}

.card h3 {
  margin-top: 0;
}

.description {
  color: #666;
  margin-bottom: 1rem;
}

/* Badge Section */
.badge-container {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1rem;
}

.badge-img {
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 0.5rem;
  background: #f8f9fa;
}

.badge-code {
  background: #f8f9fa;
  padding: 1rem;
  border-radius: 4px;
  overflow-x: auto;
}

.badge-code code {
  font-family: 'Courier New', monospace;
  font-size: 0.875rem;
}

/* Token Info Section */
.token-actions {
  display: flex;
  gap: 1rem;
  margin-bottom: 1rem;
}

.warning-text {
  color: #856404;
  font-size: 0.875rem;
  margin-top: 0.5rem;
}

/* Builds Section */
.loading,
.empty-state {
  text-align: center;
  padding: 2rem;
  color: #7f8c8d;
}

.builds-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.build-card {
  border: 1px solid #ddd;
  border-radius: 8px;
  padding: 1rem;
  cursor: pointer;
  transition: all 0.2s;
}

.build-card:hover {
  border-color: #3498db;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.build-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.build-info {
  display: flex;
  gap: 1rem;
  align-items: center;
}

.build-number {
  font-weight: bold;
  font-size: 1.1rem;
}

.build-branch {
  background: #e8f4f8;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.875rem;
}

.build-coverage {
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-weight: bold;
  font-size: 0.875rem;
}

.coverage-high {
  background: #d4edda;
  color: #155724;
}

.coverage-medium {
  background: #fff3cd;
  color: #856404;
}

.coverage-low {
  background: #f8d7da;
  color: #721c24;
}

.build-date {
  color: #666;
  font-size: 0.875rem;
}

.build-commit {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  color: #666;
  font-size: 0.875rem;
}

.commit-sha {
  background: #f8f9fa;
  padding: 0.125rem 0.25rem;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
}

.commit-msg {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.build-details {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid #ddd;
}

.build-details h4 {
  margin-top: 0;
  margin-bottom: 1rem;
}

.detail-row {
  margin-bottom: 0.5rem;
  font-size: 0.9rem;
}

.detail-row strong {
  display: inline-block;
  min-width: 120px;
}

.jobs-section {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid #eee;
}

.jobs-section h5 {
  margin-top: 0;
  margin-bottom: 0.5rem;
}

.job-item {
  display: flex;
  justify-content: space-between;
  padding: 0.5rem;
  background: #f8f9fa;
  border-radius: 4px;
  margin-bottom: 0.25rem;
  font-size: 0.875rem;
}

.build-actions {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid #eee;
  display: flex;
  justify-content: flex-end;
}

/* Modal Styles */
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

.modal-content h3 {
  margin-top: 0;
}

.info-box,
.success-box {
  background: #d4edda;
  border: 1px solid #c3e6cb;
  padding: 1rem;
  border-radius: 4px;
  margin: 1rem 0;
}

.warning-box {
  background: #fff3cd;
  border: 1px solid #ffeeba;
  padding: 1rem;
  border-radius: 4px;
  margin: 1rem 0;
  color: #856404;
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

.modal-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  margin-top: 1.5rem;
}

/* Button Styles */
.btn {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 1rem;
  text-decoration: none;
  display: inline-block;
  transition: background 0.2s;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-primary {
  background: #3498db;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #2980b9;
}

.btn-secondary {
  background: #95a5a6;
  color: white;
}

.btn-secondary:hover:not(:disabled) {
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

