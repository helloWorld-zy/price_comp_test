<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { shipApi, cruiseLineApi } from '@/api/client'
import type { Ship, CruiseLine, PaginatedResult, ApiError } from '@/api/client'

const items = ref<Ship[]>([])
const cruiseLines = ref<CruiseLine[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)
const isLoading = ref(false)
const error = ref<string | null>(null)
const filterCruiseLineId = ref<number | undefined>(undefined)

// Modal state
const showModal = ref(false)
const modalMode = ref<'create' | 'edit'>('create')
const editingItem = ref<Ship | null>(null)
const form = ref({
  cruise_line_id: 0,
  name: '',
  imo: '',
  aliases: '',
  status: 'ACTIVE',
})
const formError = ref<string | null>(null)
const isSaving = ref(false)

// Delete confirmation
const showDeleteModal = ref(false)
const deletingItem = ref<Ship | null>(null)
const isDeleting = ref(false)

const totalPages = computed(() => Math.ceil(total.value / pageSize.value))

const getCruiseLineName = (cruiseLineId: number) => {
  const cl = cruiseLines.value.find((c) => c.id === cruiseLineId)
  return cl?.name || '-'
}

const loadCruiseLines = async () => {
  try {
    const result = await cruiseLineApi.list({ page: 1, pageSize: 100 })
    cruiseLines.value = result.items
  } catch (err) {
    console.error('Failed to load cruise lines:', err)
  }
}

const loadItems = async () => {
  isLoading.value = true
  error.value = null

  try {
    const result: PaginatedResult<Ship> = await shipApi.list({
      page: currentPage.value,
      pageSize: pageSize.value,
      cruise_line_id: filterCruiseLineId.value,
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
  form.value = { cruise_line_id: cruiseLines.value[0]?.id || 0, name: '', imo: '', aliases: '', status: 'ACTIVE' }
  formError.value = null
  showModal.value = true
}

const openEditModal = (item: Ship) => {
  modalMode.value = 'edit'
  editingItem.value = item
  form.value = {
    cruise_line_id: item.cruise_line_id,
    name: item.name,
    imo: item.imo || '',
    aliases: item.aliases?.join(', ') || '',
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
  if (!form.value.name.trim()) {
    formError.value = '请输入名称'
    return
  }
  if (!form.value.cruise_line_id) {
    formError.value = '请选择邮轮公司'
    return
  }

  isSaving.value = true
  formError.value = null

  try {
    const data = {
      cruise_line_id: form.value.cruise_line_id,
      name: form.value.name.trim(),
      imo: form.value.imo.trim() || undefined,
      aliases: form.value.aliases
        ? form.value.aliases.split(',').map((s) => s.trim()).filter(Boolean)
        : undefined,
      status: form.value.status,
    }

    if (modalMode.value === 'create') {
      await shipApi.create(data)
    } else if (editingItem.value) {
      await shipApi.update(editingItem.value.id, data)
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

const openDeleteModal = (item: Ship) => {
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
    await shipApi.delete(deletingItem.value.id)
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

onMounted(async () => {
  await loadCruiseLines()
  await loadItems()
})
</script>

<template>
  <div>
    <div class="d-flex justify-content-between align-items-center mb-4">
      <h4 class="mb-0">邮轮管理</h4>
      <button class="btn btn-primary" @click="openCreateModal">
        <i class="bi bi-plus-circle me-2"></i>添加邮轮
      </button>
    </div>

    <!-- Filters -->
    <div class="card mb-4">
      <div class="card-body">
        <div class="row g-3">
          <div class="col-md-4">
            <label class="form-label">按邮轮公司筛选</label>
            <select v-model="filterCruiseLineId" class="form-select" @change="handleFilterChange">
              <option :value="undefined">全部</option>
              <option v-for="cl in cruiseLines" :key="cl.id" :value="cl.id">{{ cl.name }}</option>
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
                <th>名称</th>
                <th>邮轮公司</th>
                <th>IMO</th>
                <th>别名</th>
                <th>状态</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in items" :key="item.id">
                <td>{{ item.id }}</td>
                <td>{{ item.name }}</td>
                <td>{{ getCruiseLineName(item.cruise_line_id) }}</td>
                <td>{{ item.imo || '-' }}</td>
                <td>
                  <span v-if="item.aliases?.length">{{ item.aliases.join(', ') }}</span>
                  <span v-else class="text-muted">-</span>
                </td>
                <td>
                  <span
                    class="badge"
                    :class="item.status === 'ACTIVE' ? 'bg-success' : 'bg-secondary'"
                  >
                    {{ item.status === 'ACTIVE' ? '启用' : '禁用' }}
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
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ modalMode === 'create' ? '添加邮轮' : '编辑邮轮' }}</h5>
            <button type="button" class="btn-close" @click="closeModal"></button>
          </div>
          <div class="modal-body">
            <div v-if="formError" class="alert alert-danger">{{ formError }}</div>
            <div class="mb-3">
              <label class="form-label">邮轮公司 <span class="text-danger">*</span></label>
              <select v-model="form.cruise_line_id" class="form-select">
                <option v-for="cl in cruiseLines" :key="cl.id" :value="cl.id">{{ cl.name }}</option>
              </select>
            </div>
            <div class="mb-3">
              <label class="form-label">名称 <span class="text-danger">*</span></label>
              <input v-model="form.name" type="text" class="form-control" placeholder="例如：海洋光谱号" />
            </div>
            <div class="mb-3">
              <label class="form-label">IMO 编号</label>
              <input v-model="form.imo" type="text" class="form-control" placeholder="例如：9780450" />
            </div>
            <div class="mb-3">
              <label class="form-label">别名</label>
              <input v-model="form.aliases" type="text" class="form-control" placeholder="用逗号分隔" />
              <div class="form-text">多个别名用逗号分隔</div>
            </div>
            <div v-if="modalMode === 'edit'" class="mb-3">
              <label class="form-label">状态</label>
              <select v-model="form.status" class="form-select">
                <option value="ACTIVE">启用</option>
                <option value="INACTIVE">禁用</option>
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
            <p>确定要删除邮轮 <strong>{{ deletingItem?.name }}</strong> 吗？此操作不可撤销。</p>
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
