<template>
  <div class="cabin-type-selector">
    <label v-if="label" class="form-label">{{ label }}</label>
    <select 
      v-model="selectedValue" 
      class="form-select"
      :class="{ 'is-invalid': error }"
      :disabled="disabled || isLoading"
      @change="handleChange"
    >
      <option value="">{{ placeholder || '请选择房型' }}</option>
      <optgroup v-for="category in groupedCabinTypes" :key="category.id" :label="category.name">
        <option 
          v-for="cabinType in category.types" 
          :key="cabinType.id" 
          :value="cabinType.id"
        >
          {{ cabinType.name }} {{ cabinType.code ? `(${cabinType.code})` : '' }}
        </option>
      </optgroup>
    </select>
    <div v-if="isLoading" class="form-text text-muted">
      <span class="spinner-border spinner-border-sm me-1"></span>
      加载中...
    </div>
    <div v-if="error" class="invalid-feedback d-block">{{ error }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'

export interface CabinType {
  id: number
  shipId: number
  categoryId: number
  name: string
  code?: string
  categoryName?: string
}

interface Props {
  modelValue?: number | null
  label?: string
  placeholder?: string
  shipId?: number
  disabled?: boolean
  error?: string
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: null,
  shipId: undefined,
})

const emit = defineEmits<{
  'update:modelValue': [value: number | null]
  'change': [cabinType: CabinType | null]
}>()

const cabinTypes = ref<CabinType[]>([])
const isLoading = ref(false)
const selectedValue = ref<number | string>(props.modelValue || '')

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'

const groupedCabinTypes = computed(() => {
  const groups = new Map<number, { id: number; name: string; types: CabinType[] }>()
  
  cabinTypes.value.forEach(ct => {
    if (!groups.has(ct.categoryId)) {
      groups.set(ct.categoryId, {
        id: ct.categoryId,
        name: ct.categoryName || '其他',
        types: []
      })
    }
    groups.get(ct.categoryId)?.types.push(ct)
  })
  
  return Array.from(groups.values())
})

function getAuthHeaders(): HeadersInit {
  const token = localStorage.getItem('cruise_access_token')
  return {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  }
}

async function loadCabinTypes() {
  isLoading.value = true
  try {
    const url = new URL(`${API_BASE_URL}/cabin-types`)
    url.searchParams.append('page', '1')
    url.searchParams.append('page_size', '200')
    url.searchParams.append('enabled_only', 'true')
    
    if (props.shipId) {
      url.searchParams.append('ship_id', String(props.shipId))
    }

    const response = await fetch(url.toString(), {
      headers: getAuthHeaders(),
    })

    if (!response.ok) throw new Error('Failed to load cabin types')

    const data = await response.json()
    cabinTypes.value = data.items || []
  } catch (err) {
    console.error('Failed to load cabin types:', err)
    cabinTypes.value = []
  } finally {
    isLoading.value = false
  }
}

function handleChange() {
  const value = selectedValue.value
  if (value === '') {
    emit('update:modelValue', null)
    emit('change', null)
  } else {
    const numValue = Number(value)
    emit('update:modelValue', numValue)
    
    const cabinType = cabinTypes.value.find(ct => ct.id === numValue)
    emit('change', cabinType || null)
  }
}

watch(() => props.modelValue, (newValue) => {
  selectedValue.value = newValue || ''
})

watch(() => props.shipId, () => {
  selectedValue.value = ''
  emit('update:modelValue', null)
  loadCabinTypes()
})

onMounted(() => {
  loadCabinTypes()
})
</script>
