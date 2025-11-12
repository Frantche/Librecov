<template>
  <div class="job-view">
    <div class="job-header">
      <h2>Job Coverage Details</h2>
      <div class="job-info">
        <p><strong>Job ID:</strong> {{ job?.id }}</p>
        <p><strong>Job Number:</strong> {{ job?.job_number }}</p>
        <p><strong>Coverage Rate:</strong> {{ job?.coverage_rate?.toFixed(1) }}%</p>
        <p><strong>Created:</strong> {{ formatDate(job?.created_at) }}</p>
      </div>
    </div>

    <div v-if="loading" class="loading">
      Loading job files...
    </div>

    <div v-else-if="error" class="error">
      Error loading job files: {{ error }}
    </div>

    <div v-else class="files-section">
      <h3>Files ({{ files.length }})</h3>
      <div class="files-list">
        <div
          v-for="file in files"
          :key="file.id"
          class="file-item"
          @click="selectFile(file)"
          :class="{ selected: selectedFile?.id === file.id }"
        >
          <div class="file-name">{{ file.name }}</div>
          <div class="file-coverage">{{ file.coverage_rate.toFixed(1) }}%</div>
        </div>
      </div>

      <div v-if="selectedFile" class="file-viewer">
        <h4>{{ selectedFile.name }}</h4>
        <div class="source-code">
          <div
            v-for="(line, index) in sourceLines"
            :key="index"
            class="code-line"
            :class="getLineClass(index)"
          >
            <span class="line-number">{{ index + 1 }}</span>
            <span class="line-content">{{ line }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import type { Job, JobFile } from '../types'
import { fetchJob, fetchJobFiles } from '../services/api'

const route = useRoute()
const job = ref<Job | null>(null)
const files = ref<JobFile[]>([])
const selectedFile = ref<JobFile | null>(null)
const loading = ref(false)
const error = ref<string | null>(null)

const sourceLines = computed(() => {
  if (!selectedFile.value?.source) return []
  return selectedFile.value.source.split('\n')
})

const getLineClass = (lineIndex: number) => {
  if (!selectedFile.value?.coverage) return ''

  try {
    const coverageData = JSON.parse(selectedFile.value.coverage)
    const coverage = coverageData[lineIndex]

    if (coverage === null) return '' // No coverage needed
    if (coverage === 0) return 'uncovered' // Not covered
    if (coverage > 0) return 'covered' // Covered
  } catch (e) {
    console.error('Error parsing coverage data:', e)
  }

  return ''
}

const formatDate = (dateString?: string) => {
  if (!dateString) return ''
  return new Date(dateString).toLocaleString()
}

const selectFile = async (file: JobFile) => {
  selectedFile.value = file
}

onMounted(async () => {
  const jobId = route.params.id as string
  if (!jobId) return

  loading.value = true
  error.value = null

  try {
    job.value = await fetchJob(jobId)
    files.value = await fetchJobFiles(jobId)
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Unknown error'
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.job-view {
  padding: 1rem;
  max-width: 1200px;
  margin: 0 auto;
}

.job-header {
  margin-bottom: 2rem;
  padding: 1rem;
  background: #f8f9fa;
  border-radius: 8px;
}

.job-info {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
  margin-top: 1rem;
}

.job-info p {
  margin: 0.5rem 0;
}

.files-section {
  margin-top: 2rem;
}

.files-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 0.5rem;
  margin-bottom: 2rem;
}

.file-item {
  padding: 0.75rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.2s;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.file-item:hover {
  background-color: #f8f9fa;
}

.file-item.selected {
  border-color: #007bff;
  background-color: #e7f3ff;
}

.file-name {
  font-family: monospace;
  font-size: 0.9rem;
  flex: 1;
}

.file-coverage {
  font-weight: bold;
  color: #28a745;
}

.file-viewer {
  border: 1px solid #ddd;
  border-radius: 8px;
  overflow: hidden;
}

.file-viewer h4 {
  margin: 0;
  padding: 1rem;
  background: #f8f9fa;
  border-bottom: 1px solid #ddd;
}

.source-code {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 0.85rem;
  line-height: 1.4;
  background: #f8f9fa;
  max-height: 600px;
  overflow-y: auto;
}

.code-line {
  display: flex;
  border-bottom: 1px solid #eee;
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
}

.line-content {
  flex: 1;
  padding: 0 0.5rem;
  white-space: pre;
  overflow-x: auto;
}

.code-line.covered {
  background-color: #d4edda;
}

.code-line.uncovered {
  background-color: #f8d7da;
}

.loading, .error {
  text-align: center;
  padding: 2rem;
  font-size: 1.1rem;
}

.error {
  color: #dc3545;
  background: #f8d7da;
  border: 1px solid #f5c6cb;
  border-radius: 4px;
}
</style>
