<template>
  <div class="file-uploader">
    <div class="upload-area" :class="{ 'drag-over': isDragOver }" @drop.prevent="handleDrop" @dragover.prevent="isDragOver = true"
      @dragleave="isDragOver = false" @click="triggerFileInput">
      <input ref="fileInput" type="file" accept=".pdf,.docx,.doc" @change="handleFileSelect" style="display: none" />

      <div class="upload-icon">
        <svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 24 24" fill="none"
          stroke="currentColor" stroke-width="2">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
          <polyline points="17 8 12 3 7 8" />
          <line x1="12" y1="3" x2="12" y2="15" />
        </svg>
      </div>

      <div class="upload-text">
        <p class="upload-title">点击或拖拽文件到此处上传</p>
        <p class="upload-subtitle">支持 PDF、Word 文档，最大 10MB</p>
      </div>

      <div v-if="selectedFile" class="selected-file">
        <span>已选择: {{ selectedFile.name }}</span>
        <span class="file-size">({{ formatFileSize(selectedFile.size) }})</span>
      </div>
    </div>

    <div v-if="error" class="error-message">
      {{ error }}
    </div>

    <div v-if="selectedFile" class="upload-actions">
      <button class="btn btn-secondary" @click="clearFile">取消</button>
      <button class="btn btn-primary" @click="uploadFile" :disabled="isUploading">
        <span v-if="isUploading">上传中...</span>
        <span v-else>开始上传</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useImportStore } from '@/stores/import'

const importStore = useImportStore()

const fileInput = ref<HTMLInputElement>()
const selectedFile = ref<File | null>(null)
const isDragOver = ref(false)
const isUploading = ref(false)
const error = ref<string | null>(null)

const emit = defineEmits<{
  uploaded: [jobId: number]
}>()

function triggerFileInput() {
  fileInput.value?.click()
}

function handleFileSelect(event: Event) {
  const target = event.target as HTMLInputElement
  if (target.files && target.files[0]) {
    selectFile(target.files[0])
  }
}

function handleDrop(event: DragEvent) {
  isDragOver.value = false
  if (event.dataTransfer?.files && event.dataTransfer.files[0]) {
    selectFile(event.dataTransfer.files[0])
  }
}

function selectFile(file: File) {
  error.value = null

  // Validate file type
  const validTypes = ['.pdf', '.docx', '.doc']
  const fileExt = '.' + file.name.split('.').pop()?.toLowerCase()
  if (!validTypes.includes(fileExt)) {
    error.value = '仅支持 PDF 和 Word 文档'
    return
  }

  // Validate file size (10MB)
  const maxSize = 10 * 1024 * 1024
  if (file.size > maxSize) {
    error.value = '文件大小不能超过 10MB'
    return
  }

  selectedFile.value = file
}

function clearFile() {
  selectedFile.value = null
  error.value = null
  if (fileInput.value) {
    fileInput.value.value = ''
  }
}

async function uploadFile() {
  if (!selectedFile.value) return

  isUploading.value = true
  error.value = null

  const job = await importStore.uploadFile(selectedFile.value)

  if (job) {
    emit('uploaded', job.id)
    clearFile()
  } else {
    error.value = importStore.error || '上传失败'
  }

  isUploading.value = false
}

function formatFileSize(bytes: number): string {
  return importStore.formatFileSize(bytes)
}
</script>

<style scoped>
.file-uploader {
  width: 100%;
}

.upload-area {
  border: 2px dashed #cbd5e0;
  border-radius: 8px;
  padding: 40px;
  text-align: center;
  cursor: pointer;
  transition: all 0.3s;
  background: #f7fafc;
}

.upload-area:hover {
  border-color: #4299e1;
  background: #ebf8ff;
}

.upload-area.drag-over {
  border-color: #3182ce;
  background: #bee3f8;
}

.upload-icon {
  color: #4299e1;
  margin-bottom: 16px;
}

.upload-text {
  margin-bottom: 16px;
}

.upload-title {
  font-size: 16px;
  font-weight: 500;
  color: #2d3748;
  margin-bottom: 8px;
}

.upload-subtitle {
  font-size: 14px;
  color: #718096;
}

.selected-file {
  margin-top: 16px;
  padding: 12px;
  background: white;
  border-radius: 4px;
  border: 1px solid #e2e8f0;
}

.file-size {
  color: #718096;
  font-size: 12px;
  margin-left: 8px;
}

.error-message {
  margin-top: 16px;
  padding: 12px;
  background: #fff5f5;
  color: #c53030;
  border-radius: 4px;
  border: 1px solid #feb2b2;
}

.upload-actions {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.btn {
  padding: 8px 16px;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  border: none;
  transition: all 0.2s;
}

.btn-primary {
  background: #4299e1;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #3182ce;
}

.btn-primary:disabled {
  background: #a0aec0;
  cursor: not-allowed;
}

.btn-secondary {
  background: #e2e8f0;
  color: #2d3748;
}

.btn-secondary:hover {
  background: #cbd5e0;
}
</style>
