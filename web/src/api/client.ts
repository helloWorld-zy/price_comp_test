/**
 * API Client for cruise price comparison tool
 * Handles HTTP requests with JWT authentication
 */

import { useAuthStore } from '@/stores/auth'

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'

export interface ApiError {
  error: string
  message: string
  validation?: Record<string, string>
}

export interface PaginatedResult<T> {
  items: T[]
  total: number
  page: number
  pageSize: number
  totalPages: number
}

export interface PaginationParams {
  page?: number
  pageSize?: number
  sortBy?: string
  sortOrder?: 'asc' | 'desc'
}

class ApiClient {
  private baseUrl: string

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl
  }

  private getHeaders(): HeadersInit {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    }

    // Get token from auth store
    const authStore = useAuthStore()
    if (authStore.accessToken) {
      headers['Authorization'] = `Bearer ${authStore.accessToken}`
    }

    return headers
  }

  private async handleResponse<T>(response: Response): Promise<T> {
    if (response.status === 401) {
      // Token expired, try to refresh
      const authStore = useAuthStore()
      if (authStore.refreshToken) {
        try {
          await authStore.refreshAccessToken()
          // Retry the request would need to be handled by the caller
        } catch {
          await authStore.logout()
          throw new Error('Session expired. Please login again.')
        }
      }
      throw new Error('Unauthorized')
    }

    if (!response.ok) {
      let errorData: ApiError
      try {
        errorData = await response.json()
      } catch {
        errorData = {
          error: 'ERR_UNKNOWN',
          message: `Request failed with status ${response.status}`,
        }
      }
      throw errorData
    }

    // Handle empty responses (204 No Content)
    if (response.status === 204) {
      return {} as T
    }

    return response.json()
  }

  async get<T>(endpoint: string, params?: Record<string, unknown>): Promise<T> {
    const url = new URL(`${this.baseUrl}${endpoint}`)
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          url.searchParams.append(key, String(value))
        }
      })
    }

    const response = await fetch(url.toString(), {
      method: 'GET',
      headers: this.getHeaders(),
    })

    return this.handleResponse<T>(response)
  }

  async post<T>(endpoint: string, data?: unknown): Promise<T> {
    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      method: 'POST',
      headers: this.getHeaders(),
      body: data ? JSON.stringify(data) : undefined,
    })

    return this.handleResponse<T>(response)
  }

  async put<T>(endpoint: string, data: unknown): Promise<T> {
    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      method: 'PUT',
      headers: this.getHeaders(),
      body: JSON.stringify(data),
    })

    return this.handleResponse<T>(response)
  }

  async delete<T>(endpoint: string): Promise<T> {
    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      method: 'DELETE',
      headers: this.getHeaders(),
    })

    return this.handleResponse<T>(response)
  }
}

// Singleton instance
export const apiClient = new ApiClient(API_BASE_URL)

// Auth API
export const authApi = {
  login: (username: string, password: string) =>
    apiClient.post<{
      access_token: string
      refresh_token: string
      expires_at: number
      user: {
        id: number
        username: string
        role: string
        supplier_id?: number
        supplier_name?: string
        status: string
      }
    }>('/auth/login', { username, password }),

  refresh: (refreshToken: string) =>
    apiClient.post<{
      access_token: string
      refresh_token: string
      expires_at: number
    }>('/auth/refresh', { refresh_token: refreshToken }),

  getCurrentUser: () =>
    apiClient.get<{
      id: number
      username: string
      role: string
      supplier_id?: number
      supplier_name?: string
      status: string
    }>('/auth/me'),

  changePassword: (oldPassword: string, newPassword: string) =>
    apiClient.put<{ success: boolean }>('/auth/password', {
      old_password: oldPassword,
      new_password: newPassword,
    }),

  logout: () => apiClient.post<{ success: boolean }>('/auth/logout'),
}

// Cruise Line API
export interface CruiseLine {
  id: number
  name: string
  logo_url?: string
  aliases?: string[]
  status: string
  created_at: string
  updated_at: string
}

export const cruiseLineApi = {
  list: (params?: PaginationParams & { status?: string }) =>
    apiClient.get<PaginatedResult<CruiseLine>>('/cruise-lines', params as Record<string, unknown>),

  get: (id: number) => apiClient.get<CruiseLine>(`/cruise-lines/${id}`),

  create: (data: Omit<CruiseLine, 'id' | 'created_at' | 'updated_at' | 'status'>) =>
    apiClient.post<CruiseLine>('/admin/cruise-lines', data),

  update: (id: number, data: Partial<CruiseLine>) =>
    apiClient.put<CruiseLine>(`/admin/cruise-lines/${id}`, data),

  delete: (id: number) => apiClient.delete(`/admin/cruise-lines/${id}`),
}

// Ship API
export interface Ship {
  id: number
  cruise_line_id: number
  name: string
  imo?: string
  aliases?: string[]
  status: string
  created_at: string
  updated_at: string
}

export const shipApi = {
  list: (params?: PaginationParams & { cruise_line_id?: number; status?: string }) =>
    apiClient.get<PaginatedResult<Ship>>('/ships', params as Record<string, unknown>),

  get: (id: number) => apiClient.get<Ship>(`/ships/${id}`),

  getCabinTypes: (shipId: number) => apiClient.get<CabinType[]>(`/ships/${shipId}/cabin-types`),

  create: (data: Omit<Ship, 'id' | 'created_at' | 'updated_at' | 'status'>) =>
    apiClient.post<Ship>('/admin/ships', data),

  update: (id: number, data: Partial<Ship>) => apiClient.put<Ship>(`/admin/ships/${id}`, data),

  delete: (id: number) => apiClient.delete(`/admin/ships/${id}`),
}

// Cabin Category API
export interface CabinCategory {
  id: number
  name: string
  sort_order: number
}

export const cabinCategoryApi = {
  list: () => apiClient.get<CabinCategory[]>('/cabin-categories'),

  create: (data: Omit<CabinCategory, 'id'>) =>
    apiClient.post<CabinCategory>('/admin/cabin-categories', data),

  update: (id: number, data: Partial<CabinCategory>) =>
    apiClient.put<CabinCategory>(`/admin/cabin-categories/${id}`, data),

  delete: (id: number) => apiClient.delete(`/admin/cabin-categories/${id}`),
}

// Cabin Type API
export interface CabinType {
  id: number
  ship_id: number
  category_id: number
  name: string
  code?: string
  is_enabled: boolean
  created_at: string
  updated_at: string
}

export const cabinTypeApi = {
  list: (params?: PaginationParams & { ship_id?: number; category_id?: number; enabled_only?: boolean }) =>
    apiClient.get<PaginatedResult<CabinType>>('/cabin-types', params as Record<string, unknown>),

  get: (id: number) => apiClient.get<CabinType>(`/cabin-types/${id}`),

  create: (data: Omit<CabinType, 'id' | 'created_at' | 'updated_at' | 'is_enabled'>) =>
    apiClient.post<CabinType>('/admin/cabin-types', data),

  update: (id: number, data: Partial<CabinType>) =>
    apiClient.put<CabinType>(`/admin/cabin-types/${id}`, data),

  delete: (id: number) => apiClient.delete(`/admin/cabin-types/${id}`),
}

// Sailing API
export interface Sailing {
  id: number
  ship_id: number
  departure_date: string
  return_date?: string
  duration_days: number
  itinerary: string
  departure_port?: string
  status: string
  created_at: string
  updated_at: string
}

export const sailingApi = {
  list: (params?: PaginationParams & { ship_id?: number; status?: string }) =>
    apiClient.get<PaginatedResult<Sailing>>('/sailings', params as Record<string, unknown>),

  get: (id: number) => apiClient.get<Sailing>(`/sailings/${id}`),

  create: (data: Omit<Sailing, 'id' | 'created_at' | 'updated_at' | 'status'>) =>
    apiClient.post<Sailing>('/admin/sailings', data),

  update: (id: number, data: Partial<Sailing>) =>
    apiClient.put<Sailing>(`/admin/sailings/${id}`, data),

  delete: (id: number) => apiClient.delete(`/admin/sailings/${id}`),
}

// Supplier API
export interface Supplier {
  id: number
  name: string
  contact?: string
  aliases?: string[]
  status: string
  created_at: string
  updated_at: string
}

export const supplierApi = {
  list: (params?: PaginationParams & { status?: string }) =>
    apiClient.get<PaginatedResult<Supplier>>('/suppliers', params as Record<string, unknown>),

  get: (id: number) => apiClient.get<Supplier>(`/suppliers/${id}`),

  create: (data: Omit<Supplier, 'id' | 'created_at' | 'updated_at' | 'status'>) =>
    apiClient.post<Supplier>('/admin/suppliers', data),

  update: (id: number, data: Partial<Supplier>) =>
    apiClient.put<Supplier>(`/admin/suppliers/${id}`, data),

  delete: (id: number) => apiClient.delete(`/admin/suppliers/${id}`),
}

export default apiClient
