import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type { PaginatedResult } from '@/api/client'

// Import job types
export type ImportJobType = 'FILE_UPLOAD' | 'TEXT_INPUT' | 'TEMPLATE_IMPORT' | 'ADMIN_LLM_GENERATE'
export type ImportJobStatus = 'PENDING' | 'RUNNING' | 'NEEDS_CONFIRMATION' | 'SUCCEEDED' | 'FAILED'

// Import result summary
export interface ImportResultSummary {
  total_rows: number
  success_rows: number
  failed_rows: number
  skipped_rows: number
  warnings?: string[]
  created_quotes?: number
}

// Import job
export interface ImportJob {
  id: number
  type: ImportJobType
  status: ImportJobStatus
  file_name?: string
  file_hash?: string
  file_size?: number
  file_path?: string
  raw_text?: string
  idempotency_key?: string
  model_version?: string
  prompt_version?: string
  result_summary?: ImportResultSummary
  error_message?: string
  started_at?: string
  completed_at?: string
  duration_ms?: number
  created_at: string
  created_by: number
}

// Pagination params
export interface PaginationParams {
  page?: number
  page_size?: number
}

// Paginated result
export interface ImportPaginatedResult<T> {
  items: T[]
  page: number
  page_size: number
  total: number
  total_page: number
}

// List jobs params
export interface ListJobsParams extends PaginationParams {
  status?: ImportJobStatus
  type?: ImportJobType
}

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'

export const useImportStore = defineStore('import', () => {
  // State
  const jobs = ref<ImportJob[]>([])
  const currentJob = ref<ImportJob | null>(null)
  const total = ref(0)
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // Helper to get auth headers
  function getAuthHeaders(): HeadersInit {
    const token = localStorage.getItem('cruise_access_token')
    return {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    }
  }

  // Computed
  const pendingJobs = computed(() => jobs.value.filter(j => j.status === 'PENDING'))
  const runningJobs = computed(() => jobs.value.filter(j => j.status === 'RUNNING'))
  const completedJobs = computed(() => jobs.value.filter(j => j.status === 'SUCCEEDED'))
  const failedJobs = computed(() => jobs.value.filter(j => j.status === 'FAILED'))

  // Actions
  async function uploadFile(file: File): Promise<ImportJob | null> {
    isLoading.value = true
    error.value = null

    try {
      const formData = new FormData()
      formData.append('file', file)

      const token = localStorage.getItem('cruise_access_token')
      const response = await fetch(`${API_BASE_URL}/import/upload`, {
        method: 'POST',
        headers: {
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        body: formData,
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.message || 'Failed to upload file')
      }

      const data = await response.json()
      const job = data.data as ImportJob
      
      // Add to jobs list
      jobs.value.unshift(job)
      currentJob.value = job

      return job
    } catch (err: any) {
      error.value = err.message || 'Failed to upload file'
      return null
    } finally {
      isLoading.value = false
    }
  }

  async function listJobs(params?: ListJobsParams): Promise<ImportPaginatedResult<ImportJob> | null> {
    isLoading.value = true
    error.value = null

    try {
      const url = new URL(`${API_BASE_URL}/import/jobs`)
      if (params) {
        Object.entries(params).forEach(([key, value]) => {
          if (value !== undefined && value !== null) {
            url.searchParams.append(key, String(value))
          }
        })
      }

      const response = await fetch(url.toString(), {
        method: 'GET',
        headers: getAuthHeaders(),
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.message || 'Failed to list jobs')
      }

      const data = await response.json()
      jobs.value = data.data as ImportJob[]
      total.value = data.pagination.total

      return {
        items: jobs.value,
        ...data.pagination,
      }
    } catch (err: any) {
      error.value = err.message || 'Failed to list jobs'
      return null
    } finally {
      isLoading.value = false
    }
  }

  async function getJob(id: number): Promise<ImportJob | null> {
    isLoading.value = true
    error.value = null

    try {
      const response = await fetch(`${API_BASE_URL}/import/jobs/${id}`, {
        method: 'GET',
        headers: getAuthHeaders(),
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.message || 'Failed to get job')
      }

      const data = await response.json()
      const job = data.data as ImportJob
      
      currentJob.value = job

      // Update in jobs list if exists
      const index = jobs.value.findIndex(j => j.id === id)
      if (index !== -1) {
        jobs.value[index] = job
      }

      return job
    } catch (err: any) {
      error.value = err.message || 'Failed to get job'
      return null
    } finally {
      isLoading.value = false
    }
  }

  async function retryJob(id: number): Promise<boolean> {
    isLoading.value = true
    error.value = null

    try {
      const response = await fetch(`${API_BASE_URL}/import/jobs/${id}/retry`, {
        method: 'POST',
        headers: getAuthHeaders(),
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.message || 'Failed to retry job')
      }

      const data = await response.json()
      const job = data.data as ImportJob
      
      currentJob.value = job

      // Update in jobs list
      const index = jobs.value.findIndex(j => j.id === id)
      if (index !== -1) {
        jobs.value[index] = job
      }

      return true
    } catch (err: any) {
      error.value = err.message || 'Failed to retry job'
      return false
    } finally {
      isLoading.value = false
    }
  }

  // Polling for job status updates
  function startPolling(jobId: number, intervalMs: number = 3000): () => void {
    const intervalId = setInterval(async () => {
      await getJob(jobId)
      
      // Stop polling if job is completed or failed
      if (currentJob.value && 
          (currentJob.value.status === 'SUCCEEDED' || currentJob.value.status === 'FAILED')) {
        stopPolling()
      }
    }, intervalMs)

    // Return stop function
    const stopPolling = () => clearInterval(intervalId)
    return stopPolling
  }

  // Format file size
  function formatFileSize(bytes?: number): string {
    if (!bytes) return '-'
    
    const kb = bytes / 1024
    const mb = kb / 1024
    
    if (mb >= 1) {
      return `${mb.toFixed(2)} MB`
    } else if (kb >= 1) {
      return `${kb.toFixed(2)} KB`
    } else {
      return `${bytes} B`
    }
  }

  // Format duration
  function formatDuration(ms?: number): string {
    if (!ms) return '-'
    
    const seconds = Math.floor(ms / 1000)
    const minutes = Math.floor(seconds / 60)
    
    if (minutes > 0) {
      const remainingSeconds = seconds % 60
      return `${minutes}m ${remainingSeconds}s`
    } else {
      return `${seconds}s`
    }
  }

  // Get status color
  function getStatusColor(status: ImportJobStatus): string {
    switch (status) {
      case 'PENDING': return 'warning'
      case 'RUNNING': return 'info'
      case 'SUCCEEDED': return 'success'
      case 'FAILED': return 'danger'
      case 'NEEDS_CONFIRMATION': return 'warning'
      default: return 'secondary'
    }
  }

  // Get status text
  function getStatusText(status: ImportJobStatus): string {
    switch (status) {
      case 'PENDING': return '等待处理'
      case 'RUNNING': return '处理中'
      case 'SUCCEEDED': return '成功'
      case 'FAILED': return '失败'
      case 'NEEDS_CONFIRMATION': return '需要确认'
      default: return status
    }
  }

  return {
    // State
    jobs,
    currentJob,
    total,
    isLoading,
    error,
    
    // Computed
    pendingJobs,
    runningJobs,
    completedJobs,
    failedJobs,
    
    // Actions
    uploadFile,
    listJobs,
    getJob,
    retryJob,
    startPolling,
    
    // Helpers
    formatFileSize,
    formatDuration,
    getStatusColor,
    getStatusText,
  }
})
