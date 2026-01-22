<template>
  <div class="sailing-selector">
    <label v-if="label" class="form-label">{{ label }}</label>
    <select 
      v-model="selectedValue" 
      class="form-select"
      :class="{ 'is-invalid': error }"
      :disabled="disabled || isLoading"
      @change="handleChange"
    >
      <option value="">{{ placeholder || '请选择航次' }}</option>
      <option 
        v-for="sailing in sailings" 
        :key="sailing.id" 
        :value="sailing.id"
      >
        {{ formatSailingOption(sailing) }}
      </option>
    </select>
    <div v-if="isLoading" class="form-text text-muted">
      <span class="spinner-border spinner-border-sm me-1"></span>
      加载中...
    </div>
    <div v-if="error" class="invalid-feedback d-block">{{ error }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'

export interface Sailing {
  id: number
  shipId: number
  sailingCode?: string
  departureDate: string
  returnDate: string
  nights: number
  route: string
  shipName?: string
  cruiseLineName?: string
  status: string
}

interface Props {
  modelValue?: number | null
  label?: string
  placeholder?: string
  shipId?: number
  disabled?: boolean
  error?: string
}

interface Emits {
  (e: 'update:modelValue', value: number | null): void
  (e: 'change', sailing: Sailing | null): void
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: null,
  label: '',
  placeholder: '',
  shipId: undefined,
  disabled: false,
  error: '',
})

const emit = defineEmits<Emits>()

const sailings = ref<Sailing[]>([])
const isLoading = ref(false)
const selectedValue = ref<number | string>(props.modelValue || '')

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'

function getAuthHeaders(): HeadersInit {
  const token = localStorage.getItem('cruise_access_token')
  return {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  }
}

async function loadSailings() {
  isLoading.value = true
  try {
    const url = new URL(`${API_BASE_URL}/sailings`)
    url.searchParams.append('page', '1')
    url.searchParams.append('page_size', '100')
    url.searchParams.append('status', 'ACTIVE')
    
    if (props.shipId) {
      url.searchParams.append('ship_id', String(props.shipId))
    }

    const response = await fetch(url.toString(), {
      headers: getAuthHeaders(),
    })

    if (!response.ok) {
      throw new Error('Failed to load sailings')
    }

    const data = await response.json()
    sailings.value = data.items || []
  } catch (err) {
    console.error('Failed to load sailings:', err)
    sailings.value = []
  } finally {
    isLoading.value = false
  }
}

function formatSailingOption(sailing: Sailing): string {
  const parts = []
  
  if (sailing.shipName) {
    parts.push(sailing.shipName)
  }
  
  if (sailing.sailingCode) {
    parts.push(`(${sailing.sailingCode})`)
  }
  
  parts.push(`${sailing.departureDate} - ${sailing.nights}晚`)
  parts.push(sailing.route)
  
  return parts.join(' ')
}

function handleChange() {
  const value = selectedValue.value
  if (value === '') {
    emit('update:modelValue', null)
    emit('change', null)
  } else {
    const numValue = Number(value)
    emit('update:modelValue', numValue)
    
    const sailing = sailings.value.find(s => s.id === numValue)
    emit('change', sailing || null)
  }
}

watch(() => props.modelValue, (newValue) => {
  selectedValue.value = newValue || ''
})

watch(() => props.shipId, () => {
  selectedValue.value = ''
  emit('update:modelValue', null)
  loadSailings()
})

onMounted(() => {
  loadSailings()
})
</script>

<style scoped>
.sailing-selector {
  margin-bottom: 1rem;
}
</style>
