<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { cruiseLineApi, shipApi, sailingApi, supplierApi } from '@/api/client'

const authStore = useAuthStore()

const stats = ref({
  cruiseLines: 0,
  ships: 0,
  sailings: 0,
  suppliers: 0,
})

const isLoading = ref(true)
const error = ref<string | null>(null)

const isAdmin = computed(() => authStore.isAdmin)
const welcomeMessage = computed(() => {
  const hour = new Date().getHours()
  if (hour < 12) return '上午好'
  if (hour < 18) return '下午好'
  return '晚上好'
})

const loadStats = async () => {
  if (!isAdmin.value) return

  isLoading.value = true
  error.value = null

  try {
    const [cruiseLines, ships, sailings, suppliers] = await Promise.all([
      cruiseLineApi.list({ page: 1, pageSize: 1 }),
      shipApi.list({ page: 1, pageSize: 1 }),
      sailingApi.list({ page: 1, pageSize: 1 }),
      supplierApi.list({ page: 1, pageSize: 1 }),
    ])

    stats.value = {
      cruiseLines: cruiseLines.total,
      ships: ships.total,
      sailings: sailings.total,
      suppliers: suppliers.total,
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载数据失败'
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  loadStats()
})
</script>

<template>
  <div>
    <div class="d-flex justify-content-between align-items-center mb-4">
      <div>
        <h4 class="mb-1">{{ welcomeMessage }}，{{ authStore.user?.username }}</h4>
        <p class="text-muted mb-0">欢迎使用邮轮价格对比系统</p>
      </div>
    </div>

    <!-- Admin Stats -->
    <div v-if="isAdmin" class="row g-4">
      <div class="col-md-6 col-lg-3">
        <div class="card bg-primary text-white">
          <div class="card-body">
            <div class="d-flex justify-content-between align-items-center">
              <div>
                <h6 class="card-subtitle mb-2 opacity-75">邮轮公司</h6>
                <h2 class="card-title mb-0">
                  <span v-if="isLoading" class="spinner-border spinner-border-sm"></span>
                  <span v-else>{{ stats.cruiseLines }}</span>
                </h2>
              </div>
              <i class="bi bi-building fs-1 opacity-50"></i>
            </div>
          </div>
          <router-link to="/admin/cruise-lines" class="card-footer text-white text-decoration-none d-flex justify-content-between align-items-center">
            <span>查看详情</span>
            <i class="bi bi-arrow-right"></i>
          </router-link>
        </div>
      </div>

      <div class="col-md-6 col-lg-3">
        <div class="card bg-success text-white">
          <div class="card-body">
            <div class="d-flex justify-content-between align-items-center">
              <div>
                <h6 class="card-subtitle mb-2 opacity-75">邮轮</h6>
                <h2 class="card-title mb-0">
                  <span v-if="isLoading" class="spinner-border spinner-border-sm"></span>
                  <span v-else>{{ stats.ships }}</span>
                </h2>
              </div>
              <i class="bi bi-tsunami fs-1 opacity-50"></i>
            </div>
          </div>
          <router-link to="/admin/ships" class="card-footer text-white text-decoration-none d-flex justify-content-between align-items-center">
            <span>查看详情</span>
            <i class="bi bi-arrow-right"></i>
          </router-link>
        </div>
      </div>

      <div class="col-md-6 col-lg-3">
        <div class="card bg-info text-white">
          <div class="card-body">
            <div class="d-flex justify-content-between align-items-center">
              <div>
                <h6 class="card-subtitle mb-2 opacity-75">航次</h6>
                <h2 class="card-title mb-0">
                  <span v-if="isLoading" class="spinner-border spinner-border-sm"></span>
                  <span v-else>{{ stats.sailings }}</span>
                </h2>
              </div>
              <i class="bi bi-calendar-event fs-1 opacity-50"></i>
            </div>
          </div>
          <router-link to="/admin/sailings" class="card-footer text-white text-decoration-none d-flex justify-content-between align-items-center">
            <span>查看详情</span>
            <i class="bi bi-arrow-right"></i>
          </router-link>
        </div>
      </div>

      <div class="col-md-6 col-lg-3">
        <div class="card bg-warning text-dark">
          <div class="card-body">
            <div class="d-flex justify-content-between align-items-center">
              <div>
                <h6 class="card-subtitle mb-2 opacity-75">供应商</h6>
                <h2 class="card-title mb-0">
                  <span v-if="isLoading" class="spinner-border spinner-border-sm"></span>
                  <span v-else>{{ stats.suppliers }}</span>
                </h2>
              </div>
              <i class="bi bi-people fs-1 opacity-50"></i>
            </div>
          </div>
          <router-link to="/admin/suppliers" class="card-footer text-dark text-decoration-none d-flex justify-content-between align-items-center">
            <span>查看详情</span>
            <i class="bi bi-arrow-right"></i>
          </router-link>
        </div>
      </div>
    </div>

    <!-- Error Alert -->
    <div v-if="error" class="alert alert-danger mt-4" role="alert">
      <i class="bi bi-exclamation-triangle me-2"></i>
      {{ error }}
      <button type="button" class="btn btn-link" @click="loadStats">重试</button>
    </div>

    <!-- Vendor Dashboard -->
    <div v-if="!isAdmin" class="card">
      <div class="card-body text-center py-5">
        <i class="bi bi-clipboard-data fs-1 text-muted"></i>
        <h5 class="mt-3">供应商控制台</h5>
        <p class="text-muted">您可以在这里管理报价信息</p>
        <p class="text-muted small">更多功能即将推出...</p>
      </div>
    </div>

    <!-- Quick Actions for Admin -->
    <div v-if="isAdmin" class="mt-4">
      <h5 class="mb-3">快捷操作</h5>
      <div class="row g-3">
        <div class="col-auto">
          <router-link to="/admin/cruise-lines" class="btn btn-outline-primary">
            <i class="bi bi-plus-circle me-2"></i>添加邮轮公司
          </router-link>
        </div>
        <div class="col-auto">
          <router-link to="/admin/ships" class="btn btn-outline-primary">
            <i class="bi bi-plus-circle me-2"></i>添加邮轮
          </router-link>
        </div>
        <div class="col-auto">
          <router-link to="/admin/sailings" class="btn btn-outline-primary">
            <i class="bi bi-plus-circle me-2"></i>添加航次
          </router-link>
        </div>
        <div class="col-auto">
          <router-link to="/admin/suppliers" class="btn btn-outline-primary">
            <i class="bi bi-plus-circle me-2"></i>添加供应商
          </router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.card-footer {
  background-color: rgba(0, 0, 0, 0.1);
  border-top: 1px solid rgba(0, 0, 0, 0.1);
}
</style>
