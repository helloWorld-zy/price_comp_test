<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { sailingApi, shipApi, cruiseLineApi } from '@/api/client'
import type { Sailing, Ship, CruiseLine, PaginatedResult, ApiError } from '@/api/client'

const items = ref<Sailing[]>([])
const ships = ref<Ship[]>([])
const cruiseLines = ref<CruiseLine[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)
const isLoading = ref(false)
const error = ref<string | null>(null)
const filterShipId = ref<number | undefined>(undefined)

// Modal state
const showModal = ref(false)
const modalMode = ref<'create' | 'edit'>('create')
const editingItem = ref<Sailing | null>(null)
const form = ref({
  ship_id: 0,
  departure_date: '',
  return_date: '',
  duration_days: 7,
  itinerary: '',
  departure_port: '',
  status: 'ACTIVE',
})
const formError = ref<string | null>(null)
const isSaving = ref(false)

// Delete confirmation
const showDeleteModal = ref(false)
const deletingItem = ref<Sailing | null>(null)
const isDeleting = ref(false)

const totalPages = computed(() => Math.ceil(total.value / pageSize.value))

const getShipName = (shipId: number) => {
  const ship = ships.value.find((s) => s.id === shipId)
  return ship?.name || '-'
}

const loadReferenceData = async () => {
  try {
    const [cruiseLineResult, shipResult] = await Promise.all([
      cruiseLineApi.list({ page: 1, pageSize: 100 }),
      shipApi.list({ page: 1, pageSize: 100 }),
    ])
    cruiseLines.value = cruiseLineResult.items
    ships.value = shipResult.items
  } catch (err) {
    console.error('Failed to load reference data:', err)
  }
}

const loadItems = async () => {
  isLoading.value = true
  error.value = null

  try {
    const result: PaginatedResult<Sailing> = await sailingApi.list({
      page: currentPage.value,
      pageSize: pageSize.value,
      ship_id: filterShipId.value,
    })
    items.value = result.items
    total.value = result.total
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载数据失败'
  } finally {
    isLoading.value = false
  }
}

const openCreateModal = () => {
  modalMode.value = 'create'
  editingItem.value = null
  form.value = {
    ship_id: ships.value[0]?.id || 0,
    departure_date: '',
    return_date: '',
    duration_days: 7,
    itinerary: '',
    departure_port: '',
    status: 'ACTIVE',
  }
  formError.value = null
  showModal.value = true
}

const openEditModal = (item: Sailing) => {
  modalMode.value = 'edit'
  editingItem.value = item
  form.value = {
    ship_id: item.ship_id,
    departure_date: item.departure_date.split('T')[0],
    return_date: item.return_date?.split('T')[0] || '',
    duration_days: item.duration_days,
    itinerary: item.itinerary,
    departure_port: item.departure_port || '',
    status: item.status,
  }
  formError.value = null
  showModal.value = true
}

const closeModal = () => {
  showModal.value = false
  editingItem.value = null
}

const handleSave = async () => {
  if (!form.value.ship_id) {
    formError.value = '请选择邮轮'
    return
  }
  if (!form.value.departure_date) {
    formError.value = '请选择出发日期'
    return
  }
  if (!form.value.itinerary.trim()) {
    formError.value = '请输入航线'
    return
  }

  isSaving.value = true
  formError.value = null

  try {
    const data = {
      ship_id: form.value.ship_id,
      departure_date: form.value.departure_date,
      return_date: form.value.return_date || undefined,
      duration_days: form.value.duration_days,
      itinerary: form.value.itinerary.trim(),
      departure_port: form.value.departure_port.trim() || undefined,
      status: form.value.status,
    }

    if (modalMode.value === 'create') {
      await sailingApi.create(data)
    } else if (editingItem.value) {
      await sailingApi.update(editingItem.value.id, data)
    }

    closeModal()
    await loadItems()
  } catch (err) {
    const apiError = err as ApiError
    formError.value = apiError.message || '保存失败'
  } finally {
    isSaving.value = false
  }
}

const openDeleteModal = (item: Sailing) => {
  deletingItem.value = item
  showDeleteModal.value = true
}

const closeDeleteModal = () => {
  showDeleteModal.value = false
  deletingItem.value = null
}

const handleDelete = async () => {
  if (!deletingItem.value) return

  isDeleting.value = true
  try {
    await sailingApi.delete(deletingItem.value.id)
    closeDeleteModal()
    await loadItems()
  } catch (err) {
    const apiError = err as ApiError
    error.value = apiError.message || '删除失败'
  } finally {
    isDeleting.value = false
  }
}

const goToPage = (page: number) => {
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page
    loadItems()
  }
}

const handleFilterChange = () => {
  currentPage.value = 1
  loadItems()
}

const formatDate = (dateString: string) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleDateString('zh-CN')
}

onMounted(async () => {
  await loadReferenceData()
  await loadItems()
})
</script>

<template>
  <div>
    <div class="d-flex justify-content-between align-items-center mb-4">
      <h4 class="mb-0">航次管理</h4>
      <button class="btn btn-primary" @click="openCreateModal">
        <i class="bi bi-plus-circle me-2"></i>添加航次
      </button>
    </div>

    <!-- Filters -->
    <div class="card mb-4">
      <div class="card-body">
        <div class="row g-3">
          <div class="col-md-4">
            <label class="form-label">按邮轮筛选</label>
            <select v-model="filterShipId" class="form-select" @change="handleFilterChange">
              <option :value="undefined">全部</option>
              <option v-for="ship in ships" :key="ship.id" :value="ship.id">{{ ship.name }}</option>
            </select>
          </div>
        </div>
      </div>
    </div>

    <!-- Error Alert -->
    <div v-if="error" class="alert alert-danger" role="alert">
      {{ error }}
      <button type="button" class="btn btn-link" @click="loadItems">重试</button>
    </div>

    <!-- Table -->
    <div class="card">
      <div class="card-body">
        <div v-if="isLoading" class="text-center py-5">
          <div class="spinner-border text-primary" role="status">
            <span class="visually-hidden">加载中...</span>
          </div>
        </div>

        <div v-else-if="items.length === 0" class="text-center py-5 text-muted">
          <i class="bi bi-inbox fs-1"></i>
          <p class="mt-2">暂无数据</p>
        </div>

        <div v-else class="table-responsive">
          <table class="table table-hover align-middle">
            <thead>
              <tr>
                <th>ID</th>
                <th>邮轮</th>
                <th>出发日期</th>
                <th>返回日期</th>
                <th>天数</th>
                <th>航线</th>
                <th>出发港口</th>
                <th>状态</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in items" :key="item.id">
                <td>{{ item.id }}</td>
                <td>{{ getShipName(item.ship_id) }}</td>
                <td>{{ formatDate(item.departure_date) }}</td>
                <td>{{ formatDate(item.return_date || '') }}</td>
                <td>{{ item.duration_days }}</td>
                <td>
                  <span class="text-truncate d-inline-block" style="max-width: 200px" :title="item.itinerary">
                    {{ item.itinerary }}
                  </span>
                </td>
                <td>{{ item.departure_port || '-' }}</td>
                <td>
                  <span
                    class="badge"
                    :class="{
                      'bg-success': item.status === 'ACTIVE',
                      'bg-secondary': item.status === 'INACTIVE',
                      'bg-danger': item.status === 'CANCELLED',
                    }"
                  >
                    {{ item.status === 'ACTIVE' ? '有效' : item.status === 'CANCELLED' ? '已取消' : '无效' }}
                  </span>
                </td>
                <td>
                  <button class="btn btn-sm btn-outline-primary me-2" @click="openEditModal(item)">
                    <i class="bi bi-pencil"></i>
                  </button>
                  <button class="btn btn-sm btn-outline-danger" @click="openDeleteModal(item)">
                    <i class="bi bi-trash"></i>
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination -->
        <nav v-if="totalPages > 1" class="mt-3">
          <ul class="pagination justify-content-center mb-0">
            <li class="page-item" :class="{ disabled: currentPage === 1 }">
              <button class="page-link" @click="goToPage(currentPage - 1)">上一页</button>
            </li>
            <li
              v-for="page in totalPages"
              :key="page"
              class="page-item"
              :class="{ active: page === currentPage }"
            >
              <button class="page-link" @click="goToPage(page)">{{ page }}</button>
            </li>
            <li class="page-item" :class="{ disabled: currentPage === totalPages }">
              <button class="page-link" @click="goToPage(currentPage + 1)">下一页</button>
            </li>
          </ul>
        </nav>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <div v-if="showModal" class="modal fade show d-block" tabindex="-1" style="background: rgba(0,0,0,0.5)">
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ modalMode === 'create' ? '添加航次' : '编辑航次' }}</h5>
            <button type="button" class="btn-close" @click="closeModal"></button>
          </div>
          <div class="modal-body">
            <div v-if="formError" class="alert alert-danger">{{ formError }}</div>
            <div class="row">
              <div class="col-md-6 mb-3">
                <label class="form-label">邮轮 <span class="text-danger">*</span></label>
                <select v-model="form.ship_id" class="form-select">
                  <option v-for="ship in ships" :key="ship.id" :value="ship.id">{{ ship.name }}</option>
                </select>
              </div>
              <div class="col-md-6 mb-3">
                <label class="form-label">天数</label>
                <input v-model.number="form.duration_days" type="number" class="form-control" min="1" />
              </div>
            </div>
            <div class="row">
              <div class="col-md-6 mb-3">
                <label class="form-label">出发日期 <span class="text-danger">*</span></label>
                <input v-model="form.departure_date" type="date" class="form-control" />
              </div>
              <div class="col-md-6 mb-3">
                <label class="form-label">返回日期</label>
                <input v-model="form.return_date" type="date" class="form-control" />
              </div>
            </div>
            <div class="mb-3">
              <label class="form-label">航线 <span class="text-danger">*</span></label>
              <input v-model="form.itinerary" type="text" class="form-control" placeholder="例如：上海-长崎-上海" />
            </div>
            <div class="mb-3">
              <label class="form-label">出发港口</label>
              <input v-model="form.departure_port" type="text" class="form-control" placeholder="例如：上海吴淞口" />
            </div>
            <div v-if="modalMode === 'edit'" class="mb-3">
              <label class="form-label">状态</label>
              <select v-model="form.status" class="form-select">
                <option value="ACTIVE">有效</option>
                <option value="INACTIVE">无效</option>
                <option value="CANCELLED">已取消</option>
              </select>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="closeModal">取消</button>
            <button type="button" class="btn btn-primary" :disabled="isSaving" @click="handleSave">
              <span v-if="isSaving" class="spinner-border spinner-border-sm me-2"></span>
              保存
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <div v-if="showDeleteModal" class="modal fade show d-block" tabindex="-1" style="background: rgba(0,0,0,0.5)">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">确认删除</h5>
            <button type="button" class="btn-close" @click="closeDeleteModal"></button>
          </div>
          <div class="modal-body">
            <p>确定要删除航次 <strong>{{ deletingItem?.itinerary }}</strong> 吗？此操作不可撤销。</p>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="closeDeleteModal">取消</button>
            <button type="button" class="btn btn-danger" :disabled="isDeleting" @click="handleDelete">
              <span v-if="isDeleting" class="spinner-border spinner-border-sm me-2"></span>
              删除
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
