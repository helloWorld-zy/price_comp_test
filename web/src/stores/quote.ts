/**
 * Quote Store - Manages price quote state
 */
import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { PaginatedResult, PaginationParams } from '@/api/client'

export interface PriceQuote {
  id: number
  sailingId: number
  cabinTypeId: number
  supplierId: number
  // Denormalized fields
  sailingCode?: string
  shipName?: string
  cruiseLineName?: string
  departureDate?: string
  cabinTypeName?: string
  cabinCategoryName?: string
  supplierName?: string
  // Price info
  price: string
  currency: string
  pricingUnit: 'PER_PERSON' | 'PER_CABIN' | 'TOTAL'
  conditions?: string
  guestCount?: number
  promotion?: string
  cabinQuantity?: number
  validUntil?: string
  notes?: string
  // Source tracking
  source: string
  sourceRef?: string
  importJobId?: number
  status: 'ACTIVE' | 'VOIDED' | 'CORRECTED'
  createdAt: string
  createdBy: number
}

export interface CreateQuoteInput {
  sailingId: number
  cabinTypeId: number
  price: string
  currency: string
  pricingUnit: 'PER_PERSON' | 'PER_CABIN' | 'TOTAL'
  conditions?: string
  guestCount?: number
  promotion?: string
  cabinQuantity?: number
  validUntil?: string
  notes?: string
  idempotencyKey?: string
}

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'

export const useQuoteStore = defineStore('quote', () => {
  const quotes = ref<PriceQuote[]>([])
  const currentQuote = ref<PriceQuote | null>(null)
  const total = ref(0)
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  function getAuthHeaders(): HeadersInit {
    const token = localStorage.getItem('cruise_access_token')
    return {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    }
  }

  async function createQuote(input: CreateQuoteInput): Promise<PriceQuote | null> {
    isLoading.value = true
    error.value = null

    try {
      const response = await fetch(`${API_BASE_URL}/quotes`, {
        method: 'POST',
        headers: getAuthHeaders(),
        body: JSON.stringify({
          sailing_id: input.sailingId,
          cabin_type_id: input.cabinTypeId,
          price: input.price,
          currency: input.currency || 'CNY',
          pricing_unit: input.pricingUnit,
          conditions: input.conditions,
          guest_count: input.guestCount,
          promotion: input.promotion,
          cabin_quantity: input.cabinQuantity,
          valid_until: input.validUntil,
          notes: input.notes,
          idempotency_key: input.idempotencyKey,
        }),
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.message || 'Failed to create quote')
      }

      const quote = await response.json()
      currentQuote.value = transformQuote(quote)
      
      return currentQuote.value
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to create quote'
      return null
    } finally {
      isLoading.value = false
    }
  }

  async function listQuotes(params?: PaginationParams & {
    sailingId?: number
    cabinTypeId?: number
    supplierId?: number
    status?: string
  }): Promise<PaginatedResult<PriceQuote> | null> {
    isLoading.value = true
    error.value = null

    try {
      const url = new URL(`${API_BASE_URL}/quotes`)
      if (params) {
        Object.entries(params).forEach(([key, value]) => {
          if (value !== undefined && value !== null) {
            url.searchParams.append(key, String(value))
          }
        })
      }

      const response = await fetch(url.toString(), {
        method: 'GET',
        headers: getAuthHeaders(),
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.message || 'Failed to list quotes')
      }

      const data = await response.json()
      quotes.value = data.items.map(transformQuote)
      total.value = data.total

      return {
        items: quotes.value,
        total: data.total,
        page: data.page,
        pageSize: data.page_size || data.pageSize,
        totalPages: data.total_pages || data.totalPages,
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to list quotes'
      return null
    } finally {
      isLoading.value = false
    }
  }

  async function getQuote(id: number): Promise<PriceQuote | null> {
    isLoading.value = true
    error.value = null

    try {
      const response = await fetch(`${API_BASE_URL}/quotes/${id}`, {
        method: 'GET',
        headers: getAuthHeaders(),
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.message || 'Failed to get quote')
      }

      const quote = await response.json()
      currentQuote.value = transformQuote(quote)
      
      return currentQuote.value
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to get quote'
      return null
    } finally {
      isLoading.value = false
    }
  }

  async function voidQuote(id: number, reason?: string): Promise<boolean> {
    isLoading.value = true
    error.value = null

    try {
      const response = await fetch(`${API_BASE_URL}/quotes/${id}/void`, {
        method: 'PUT',
        headers: getAuthHeaders(),
        body: JSON.stringify({ reason }),
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.message || 'Failed to void quote')
      }

      const quote = await response.json()
      currentQuote.value = transformQuote(quote)
      
      // Update in list if present
      const index = quotes.value.findIndex(q => q.id === id)
      if (index !== -1) {
        quotes.value[index] = currentQuote.value
      }

      return true
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to void quote'
      return false
    } finally {
      isLoading.value = false
    }
  }

  function transformQuote(data: any): PriceQuote {
    return {
      id: data.id,
      sailingId: data.sailing_id,
      cabinTypeId: data.cabin_type_id,
      supplierId: data.supplier_id,
      sailingCode: data.sailing_code,
      shipName: data.ship_name,
      cruiseLineName: data.cruise_line_name,
      departureDate: data.departure_date,
      cabinTypeName: data.cabin_type_name,
      cabinCategoryName: data.cabin_category_name,
      supplierName: data.supplier_name,
      price: data.price,
      currency: data.currency,
      pricingUnit: data.pricing_unit,
      conditions: data.conditions,
      guestCount: data.guest_count,
      promotion: data.promotion,
      cabinQuantity: data.cabin_quantity,
      validUntil: data.valid_until,
      notes: data.notes,
      source: data.source,
      sourceRef: data.source_ref,
      importJobId: data.import_job_id,
      status: data.status,
      createdAt: data.created_at,
      createdBy: data.created_by,
    }
  }

  return {
    // State
    quotes,
    currentQuote,
    total,
    isLoading,
    error,
    // Actions
    createQuote,
    listQuotes,
    getQuote,
    voidQuote,
  }
})
