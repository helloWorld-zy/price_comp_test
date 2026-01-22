/**
 * Authentication Store - Manages user authentication state
 */
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface User {
  id: number
  username: string
  role: 'ADMIN' | 'VENDOR'
  supplierId?: number
  supplierName?: string
  status: string
}

export interface AuthState {
  user: User | null
  accessToken: string | null
  refreshToken: string | null
  expiresAt: number | null
}

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'

const STORAGE_KEYS = {
  ACCESS_TOKEN: 'cruise_access_token',
  REFRESH_TOKEN: 'cruise_refresh_token',
  USER: 'cruise_user',
  EXPIRES_AT: 'cruise_expires_at',
}

export const useAuthStore = defineStore('auth', () => {
  // State
  const user = ref<User | null>(null)
  const accessToken = ref<string | null>(null)
  const refreshToken = ref<string | null>(null)
  const expiresAt = ref<number | null>(null)
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // Getters
  const isAuthenticated = computed(() => !!accessToken.value && !!user.value)
  const isAdmin = computed(() => user.value?.role === 'ADMIN')
  const isVendor = computed(() => user.value?.role === 'VENDOR')
  const isTokenExpired = computed(() => {
    if (!expiresAt.value) return true
    return Date.now() / 1000 > expiresAt.value
  })

  // Initialize from localStorage
  function initialize() {
    const storedToken = localStorage.getItem(STORAGE_KEYS.ACCESS_TOKEN)
    const storedRefreshToken = localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN)
    const storedUser = localStorage.getItem(STORAGE_KEYS.USER)
    const storedExpiresAt = localStorage.getItem(STORAGE_KEYS.EXPIRES_AT)

    if (storedToken) {
      accessToken.value = storedToken
    }
    if (storedRefreshToken) {
      refreshToken.value = storedRefreshToken
    }
    if (storedUser) {
      try {
        user.value = JSON.parse(storedUser)
      } catch {
        user.value = null
      }
    }
    if (storedExpiresAt) {
      expiresAt.value = parseInt(storedExpiresAt, 10)
    }
  }

  // Save to localStorage
  function persistAuth() {
    if (accessToken.value) {
      localStorage.setItem(STORAGE_KEYS.ACCESS_TOKEN, accessToken.value)
    }
    if (refreshToken.value) {
      localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, refreshToken.value)
    }
    if (user.value) {
      localStorage.setItem(STORAGE_KEYS.USER, JSON.stringify(user.value))
    }
    if (expiresAt.value) {
      localStorage.setItem(STORAGE_KEYS.EXPIRES_AT, expiresAt.value.toString())
    }
  }

  // Clear from localStorage
  function clearAuth() {
    Object.values(STORAGE_KEYS).forEach((key) => {
      localStorage.removeItem(key)
    })
  }

  // Actions
  async function login(username: string, password: string): Promise<boolean> {
    isLoading.value = true
    error.value = null

    try {
      const response = await fetch(`${API_BASE_URL}/auth/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username, password }),
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.message || 'Login failed')
      }

      const data = await response.json()

      accessToken.value = data.access_token
      refreshToken.value = data.refresh_token
      expiresAt.value = data.expires_at
      user.value = {
        id: data.user.id,
        username: data.user.username,
        role: data.user.role,
        supplierId: data.user.supplier_id,
        supplierName: data.user.supplier_name,
        status: data.user.status,
      }

      persistAuth()
      return true
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Login failed'
      return false
    } finally {
      isLoading.value = false
    }
  }

  async function refreshAccessToken(): Promise<boolean> {
    if (!refreshToken.value) {
      return false
    }

    try {
      const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ refresh_token: refreshToken.value }),
      })

      if (!response.ok) {
        throw new Error('Token refresh failed')
      }

      const data = await response.json()

      accessToken.value = data.access_token
      refreshToken.value = data.refresh_token
      expiresAt.value = data.expires_at

      persistAuth()
      return true
    } catch {
      // Refresh failed, clear auth
      await logout()
      return false
    }
  }

  async function logout(): Promise<void> {
    // Try to call logout endpoint (ignore errors)
    if (accessToken.value) {
      try {
        await fetch(`${API_BASE_URL}/auth/logout`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${accessToken.value}`,
          },
        })
      } catch {
        // Ignore logout errors
      }
    }

    // Clear local state
    user.value = null
    accessToken.value = null
    refreshToken.value = null
    expiresAt.value = null
    error.value = null

    clearAuth()
  }

  async function checkAuth(): Promise<boolean> {
    if (!accessToken.value) {
      return false
    }

    // Check if token is expired
    if (isTokenExpired.value) {
      // Try to refresh
      const refreshed = await refreshAccessToken()
      if (!refreshed) {
        return false
      }
    }

    // Verify token by getting current user
    try {
      const response = await fetch(`${API_BASE_URL}/auth/me`, {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${accessToken.value}`,
        },
      })

      if (!response.ok) {
        throw new Error('Token invalid')
      }

      const userData = await response.json()
      user.value = {
        id: userData.id,
        username: userData.username,
        role: userData.role,
        supplierId: userData.supplier_id,
        supplierName: userData.supplier_name,
        status: userData.status,
      }

      return true
    } catch {
      await logout()
      return false
    }
  }

  async function changePassword(oldPassword: string, newPassword: string): Promise<boolean> {
    if (!accessToken.value) {
      return false
    }

    try {
      const response = await fetch(`${API_BASE_URL}/auth/password`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${accessToken.value}`,
        },
        body: JSON.stringify({
          old_password: oldPassword,
          new_password: newPassword,
        }),
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.message || 'Password change failed')
      }

      return true
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Password change failed'
      return false
    }
  }

  // Initialize on store creation
  initialize()

  return {
    // State
    user,
    accessToken,
    refreshToken,
    expiresAt,
    isLoading,
    error,
    // Getters
    isAuthenticated,
    isAdmin,
    isVendor,
    isTokenExpired,
    // Actions
    login,
    logout,
    refreshAccessToken,
    checkAuth,
    changePassword,
    initialize,
  }
})
