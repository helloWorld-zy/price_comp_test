/**
 * Vue Router Configuration
 */
import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

// Lazy load views
const Login = () => import('@/pages/Login.vue')
const AdminDashboard = () => import('@/pages/admin/Dashboard.vue')
const CruiseLines = () => import('@/pages/admin/CruiseLines.vue')
const Ships = () => import('@/pages/admin/Ships.vue')
const Sailings = () => import('@/pages/admin/Sailings.vue')
const CabinTypes = () => import('@/pages/admin/CabinTypes.vue')
const Suppliers = () => import('@/pages/admin/Suppliers.vue')

// Layouts
const MainLayout = () => import('@/components/Layout/MainLayout.vue')

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: Login,
    meta: { requiresAuth: false, title: '登录' },
  },
  {
    path: '/',
    component: MainLayout,
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        redirect: '/dashboard',
      },
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: AdminDashboard,
        meta: { title: '控制台', requiresAdmin: false },
      },
      // Admin routes
      {
        path: 'admin/cruise-lines',
        name: 'CruiseLines',
        component: CruiseLines,
        meta: { title: '邮轮公司管理', requiresAdmin: true },
      },
      {
        path: 'admin/ships',
        name: 'Ships',
        component: Ships,
        meta: { title: '邮轮管理', requiresAdmin: true },
      },
      {
        path: 'admin/sailings',
        name: 'Sailings',
        component: Sailings,
        meta: { title: '航次管理', requiresAdmin: true },
      },
      {
        path: 'admin/cabin-types',
        name: 'CabinTypes',
        component: CabinTypes,
        meta: { title: '房型管理', requiresAdmin: true },
      },
      {
        path: 'admin/suppliers',
        name: 'Suppliers',
        component: Suppliers,
        meta: { title: '供应商管理', requiresAdmin: true },
      },
    ],
  },
  // Catch all - redirect to dashboard or login
  {
    path: '/:pathMatch(.*)*',
    redirect: '/',
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

// Navigation guard
router.beforeEach(async (to, _from, next) => {
  const authStore = useAuthStore()

  // Update page title
  document.title = to.meta.title ? `${to.meta.title} - 邮轮价格对比` : '邮轮价格对比'

  // Check if route requires authentication
  const requiresAuth = to.matched.some((record) => record.meta.requiresAuth !== false)
  const requiresAdmin = to.matched.some((record) => record.meta.requiresAdmin === true)

  if (requiresAuth) {
    // Check if user is authenticated
    if (!authStore.isAuthenticated) {
      // Try to restore session
      const isValid = await authStore.checkAuth()
      if (!isValid) {
        // Redirect to login
        next({
          path: '/login',
          query: { redirect: to.fullPath },
        })
        return
      }
    }

    // Check admin requirement
    if (requiresAdmin && !authStore.isAdmin) {
      // Not authorized - redirect to dashboard
      next({ path: '/dashboard' })
      return
    }
  }

  // If going to login page while already authenticated
  if (to.path === '/login' && authStore.isAuthenticated) {
    next({ path: '/dashboard' })
    return
  }

  next()
})

export default router
