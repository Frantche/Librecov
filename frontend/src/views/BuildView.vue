<template>
  <div class="build-view">
    <div class="header-section">
      <div>
        <h2>Build #{{ build?.build_num }}</h2>
        <p class="build-info">
          <span class="branch">{{ build?.branch }}</span>
          <span class="coverage" :class="getCoverageClass(build?.coverage_rate || 0)">
            {{ build?.coverage_rate.toFixed(2) }}%
          </span>
        </p>
        <p class="commit-info" v-if="build?.commit_sha">
          <code class="commit-sha">{{ build.commit_sha.substring(0, 7) }}</code>
          <span class="commit-msg">{{ build.commit_msg }}</span>
        </p>
        <p class="created-at">{{ formatDate(build?.created_at || '') }}</p>
      </div>
      <router-link :to="`/projects/${build?.project_id}`" class="btn btn-secondary">
        Back to Project
      </router-link>
    </div>

    <!-- Jobs Section -->
    <div class="jobs-section card">
      <h3>Jobs ({{ jobs.length }})</h3>
      <div v-if="loading" class="loading">Loading jobs...</div>
      <div v-else-if="jobs.length === 0" class="empty-state">
        <p>No jobs found for this build.</p>
      </div>
      <div v-else class="jobs-list">
        <div 
          v-for="job in jobs" 
          :key="job.id" 
          class="job-card"
          @click="selectedJob = selectedJob?.id === job.id ? null : job"
        >
          <div class="job-header">
            <div class="job-info">
              <span class="job-number">Job {{ job.job_number }}</span>
              <span class="job-coverage" :class="getCoverageClass(job.coverage_rate)">
                {{ job.coverage_rate.toFixed(2) }}%
              </span>
            </div>
            <span class="job-date">{{ formatDate(job.created_at) }}</span>
          </div>
          
          <!-- Job Details (expanded) -->
          <div v-if="selectedJob?.id === job.id" class="job-details">
            <h4>Files ({{ job.files?.length || 0 }})</h4>
            <div v-if="job.files && job.files.length > 0" class="files-list">
              <div 
                v-for="file in job.files" 
                :key="file.id" 
                class="file-item"
                @click.stop="showFileDetails(file)"
              >
                <span class="file-name">{{ file.name }}</span>
                <span class="file-coverage" :class="getCoverageClass(file.coverage_rate)">
                  {{ file.coverage_rate.toFixed(2) }}%
                </span>
              </div>
            </div>
            <div v-else class="empty-state">
              <p>No files in this job.</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- File Details Modal -->
    <div v-if="selectedFile" class="modal" @click.self="closeFileDetails">
      <div class="modal-content file-modal">
        <h3>{{ selectedFile.name }}</h3>
        <div class="file-stats">
          <span class="coverage" :class="getCoverageClass(selectedFile.coverage_rate)">
            Coverage: {{ selectedFile.coverage_rate.toFixed(2) }}%
          </span>
        </div>
        <div class="source-code">
          <div
            v-for="(line, index) in getFileLines(selectedFile)"
            :key="index"
            class="code-line"
            :class="getLineClass(selectedFile, index)"
          >
            <span class="line-number">{{ index + 1 }}</span>
            <span class="line-content">{{ line }}</span>
          </div>
        </div>
        <div class="modal-actions">
          <button @click="closeFileDetails" class="btn btn-primary">Close</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { fetchBuild } from '../services/api'
import type { Build, Job, JobFile } from '../types'

const route = useRoute()
const buildId = computed(() => route.params.id as string)

const build = ref<Build | null>(null)
const jobs = ref<Job[]>([])
const selectedJob = ref<Job | null>(null)
const selectedFile = ref<JobFile | null>(null)
const loading = ref(true)

const fetchBuildData = async () => {
  try {
    loading.value = true
    const data = await fetchBuild(buildId.value)
    build.value = data
    jobs.value = data.jobs || []
  } catch (error) {
    console.error('Failed to fetch build:', error)
  } finally {
    loading.value = false
  }
}

const showFileDetails = (file: JobFile) => {
  selectedFile.value = file
}

const closeFileDetails = () => {
  selectedFile.value = null
}

const getFileLines = (file: JobFile) => {
  return file.source.split('\n')
}

const getLineClass = (file: JobFile, lineIndex: number) => {
  try {
    const coverageData = JSON.parse(file.coverage || '[]')
    const coverage = coverageData[lineIndex]

    if (coverage === null) return '' // No coverage needed
    if (coverage === 0) return 'uncovered' // Not covered
    if (coverage > 0) return 'covered' // Covered
  } catch (e) {
    console.error('Error parsing coverage data:', e)
  }

  return ''
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
  fetchBuildData()
})
</script>

<style scoped>
.build-view {
  padding: 1rem 0;
}

.header-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.build-info {
  display: flex;
  gap: 1rem;
  align-items: center;
  margin: 0.5rem 0;
}

.branch {
  background: #e8f4f8;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.875rem;
}

.coverage {
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-weight: bold;
  font-size: 0.875rem;
}

.commit-info {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  color: #666;
  font-size: 0.875rem;
  margin: 0.5rem 0;
}

.commit-sha {
  background: #f8f9fa;
  padding: 0.125rem 0.25rem;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
}

.created-at {
  color: #666;
  font-size: 0.875rem;
  margin: 0.25rem 0;
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

.loading,
.empty-state {
  text-align: center;
  padding: 2rem;
  color: #7f8c8d;
}

.jobs-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.job-card {
  border: 1px solid #ddd;
  border-radius: 8px;
  padding: 1rem;
  cursor: pointer;
  transition: all 0.2s;
}

.job-card:hover {
  border-color: #3498db;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.job-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.job-info {
  display: flex;
  gap: 1rem;
  align-items: center;
}

.job-number {
  font-weight: bold;
  font-size: 1.1rem;
}

.job-coverage {
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-weight: bold;
  font-size: 0.875rem;
}

.job-date {
  color: #666;
  font-size: 0.875rem;
}

.job-details {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid #ddd;
}

.job-details h4 {
  margin-top: 0;
  margin-bottom: 1rem;
}

.files-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.file-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem;
  background: #f8f9fa;
  border-radius: 4px;
  cursor: pointer;
  transition: background 0.2s;
}

.file-item:hover {
  background: #e9ecef;
}

.file-name {
  flex: 1;
  font-family: 'Courier New', monospace;
  font-size: 0.875rem;
  word-break: break-all;
}

.file-coverage {
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-weight: bold;
  font-size: 0.75rem;
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
  min-width: 800px;
  max-width: 90%;
  max-height: 90vh;
  overflow-y: auto;
}

.file-modal {
  min-width: 1000px;
}

.modal-content h3 {
  margin-top: 0;
  word-break: break-all;
}

.file-stats {
  margin: 1rem 0;
}

.source-code {
  background: #f8f9fa;
  padding: 1rem;
  border-radius: 4px;
  overflow-x: auto;
  max-height: 600px;
  overflow-y: auto;
}

.code-line {
  display: flex;
  border-bottom: 1px solid #eee;
  min-height: 1.4em;
}

.code-line:last-child {
  border-bottom: none;
}

.line-number {
  display: inline-block;
  width: 50px;
  padding: 0 0.5rem;
  text-align: right;
  color: #666;
  border-right: 1px solid #ddd;
  background: #f8f9fa;
  user-select: none;
  flex-shrink: 0;
}

.line-content {
  flex: 1;
  padding: 0 0.5rem;
  white-space: pre;
  overflow-x: auto;
}

.code-line.covered {
  background-color: rgba(25, 135, 84, 0.08);
}

.code-line.uncovered {
  background-color: rgba(220, 53, 69, 0.08);
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
</style>
