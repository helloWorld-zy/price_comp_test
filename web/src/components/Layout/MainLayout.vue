<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const isAdmin = computed(() => authStore.isAdmin)
const username = computed(() => authStore.user?.username || '')

const menuItems = computed(() => {
  const items = [{ name: '控制台', path: '/dashboard', icon: 'bi-speedometer2' }]

  if (isAdmin.value) {
    items.push(
      { name: '邮轮公司', path: '/admin/cruise-lines', icon: 'bi-building' },
      { name: '邮轮', path: '/admin/ships', icon: 'bi-tsunami' },
      { name: '航次', path: '/admin/sailings', icon: 'bi-calendar-event' },
      { name: '房型', path: '/admin/cabin-types', icon: 'bi-door-open' },
      { name: '供应商', path: '/admin/suppliers', icon: 'bi-people' }
    )
  }

  return items
})

const isActive = (path: string) => {
  return route.path === path
}

const handleLogout = async () => {
  await authStore.logout()
  router.push('/login')
}
</script>

<template>
  <div class="d-flex" style="min-height: 100vh">
    <!-- Sidebar -->
    <nav
      class="d-flex flex-column bg-dark text-white p-3"
      style="width: 240px; min-height: 100vh"
    >
      <div class="mb-4">
        <h5 class="text-center">邮轮价格对比</h5>
      </div>

      <ul class="nav nav-pills flex-column mb-auto">
        <li v-for="item in menuItems" :key="item.path" class="nav-item">
          <router-link
            :to="item.path"
            class="nav-link"
            :class="{ active: isActive(item.path), 'text-white': !isActive(item.path) }"
          >
            <i :class="['bi', item.icon, 'me-2']"></i>
            {{ item.name }}
          </router-link>
        </li>
      </ul>

      <hr />

      <div class="dropdown">
        <a
          href="#"
          class="d-flex align-items-center text-white text-decoration-none dropdown-toggle"
          data-bs-toggle="dropdown"
        >
          <i class="bi bi-person-circle me-2 fs-5"></i>
          <strong>{{ username }}</strong>
        </a>
        <ul class="dropdown-menu dropdown-menu-dark text-small shadow">
          <li>
            <span class="dropdown-item-text text-muted">
              {{ isAdmin ? '管理员' : '供应商' }}
            </span>
          </li>
          <li><hr class="dropdown-divider" /></li>
          <li>
            <a class="dropdown-item" href="#" @click.prevent="handleLogout">
              <i class="bi bi-box-arrow-right me-2"></i>退出登录
            </a>
          </li>
        </ul>
      </div>
    </nav>

    <!-- Main content -->
    <main class="flex-grow-1 bg-light">
      <div class="p-4">
        <router-view />
      </div>
    </main>
  </div>
</template>

<style scoped>
.nav-link {
  border-radius: 0.375rem;
}

.nav-link:hover:not(.active) {
  background-color: rgba(255, 255, 255, 0.1);
}

.nav-link.active {
  background-color: var(--bs-primary);
}
</style>
