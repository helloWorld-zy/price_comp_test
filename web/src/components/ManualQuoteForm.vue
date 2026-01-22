<template>
  <div class="manual-quote-form card">
    <div class="card-body">
      <h5 class="card-title">手动录入报价</h5>
      
      <form @submit.prevent="handleSubmit">
        <SailingSelector
          v-model="form.sailingId"
          label="选择航次 *"
          :error="errors.sailingId"
          @change="handleSailingChange"
        />

        <CabinTypeSelector
          v-model="form.cabinTypeId"
          label="选择房型 *"
          :ship-id="selectedShipId"
          :error="errors.cabinTypeId"
        />

        <div class="row">
          <div class="col-md-6">
            <label class="form-label">价格 *</label>
            <input
              v-model="form.price"
              type="number"
              step="0.01"
              class="form-control"
              :class="{ 'is-invalid': errors.price }"
              placeholder="请输入价格"
              required
            />
            <div v-if="errors.price" class="invalid-feedback">{{ errors.price }}</div>
          </div>

          <div class="col-md-3">
            <label class="form-label">币种</label>
            <select v-model="form.currency" class="form-select">
              <option value="CNY">CNY</option>
              <option value="USD">USD</option>
              <option value="EUR">EUR</option>
            </select>
          </div>

          <div class="col-md-3">
            <label class="form-label">计价口径 *</label>
            <select v-model="form.pricingUnit" class="form-select" required>
              <option value="PER_PERSON">每人</option>
              <option value="PER_CABIN">每间</option>
              <option value="TOTAL">总价</option>
            </select>
          </div>
        </div>

        <div class="row mt-3">
          <div class="col-md-6">
            <label class="form-label">适用人数</label>
            <input v-model.number="form.guestCount" type="number" class="form-control" placeholder="如: 2" />
          </div>

          <div class="col-md-6">
            <label class="form-label">舱房数量</label>
            <input v-model.number="form.cabinQuantity" type="number" class="form-control" placeholder="如: 10" />
          </div>
        </div>

        <div class="mt-3">
          <label class="form-label">有效期</label>
          <input v-model="form.validUntil" type="date" class="form-control" />
        </div>

        <div class="mt-3">
          <label class="form-label">适用条件</label>
          <textarea v-model="form.conditions" class="form-control" rows="2" placeholder="如: 提前60天预订"></textarea>
        </div>

        <div class="mt-3">
          <label class="form-label">促销信息</label>
          <input v-model="form.promotion" type="text" class="form-control" placeholder="如: 早鸟优惠" />
        </div>

        <div class="mt-3">
          <label class="form-label">备注</label>
          <textarea v-model="form.notes" class="form-control" rows="2"></textarea>
        </div>

        <div class="mt-4">
          <button type="submit" class="btn btn-primary" :disabled="isSubmitting">
            <span v-if="isSubmitting" class="spinner-border spinner-border-sm me-1"></span>
            {{ isSubmitting ? '提交中...' : '提交报价' }}
          </button>
          <button type="button" class="btn btn-secondary ms-2" @click="handleReset">重置</button>
        </div>

        <div v-if="submitError" class="alert alert-danger mt-3">{{ submitError }}</div>
        <div v-if="submitSuccess" class="alert alert-success mt-3">报价提交成功！</div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import SailingSelector from '@/components/SailingSelector.vue'
import CabinTypeSelector from '@/components/CabinTypeSelector.vue'
import { useQuoteStore } from '@/stores/quote'
import type { Sailing } from '@/components/SailingSelector.vue'

const quoteStore = useQuoteStore()

const form = reactive({
  sailingId: null as number | null,
  cabinTypeId: null as number | null,
  price: '',
  currency: 'CNY',
  pricingUnit: 'PER_PERSON' as 'PER_PERSON' | 'PER_CABIN' | 'TOTAL',
  conditions: '',
  guestCount: undefined as number | undefined,
  promotion: '',
  cabinQuantity: undefined as number | undefined,
  validUntil: '',
  notes: '',
})

const selectedShipId = ref<number | undefined>(undefined)
const errors = reactive<Record<string, string>>({})
const isSubmitting = ref(false)
const submitError = ref('')
const submitSuccess = ref(false)

function handleSailingChange(sailing: Sailing | null) {
  selectedShipId.value = sailing?.shipId
  form.cabinTypeId = null
}

function validate(): boolean {
  errors.sailingId = ''
  errors.cabinTypeId = ''
  errors.price = ''

  let isValid = true

  if (!form.sailingId) {
    errors.sailingId = '请选择航次'
    isValid = false
  }

  if (!form.cabinTypeId) {
    errors.cabinTypeId = '请选择房型'
    isValid = false
  }

  if (!form.price || parseFloat(form.price) <= 0) {
    errors.price = '请输入有效价格'
    isValid = false
  }

  return isValid
}

async function handleSubmit() {
  submitError.value = ''
  submitSuccess.value = false

  if (!validate()) {
    return
  }

  isSubmitting.value = true

  try {
    const result = await quoteStore.createQuote({
      sailingId: form.sailingId!,
      cabinTypeId: form.cabinTypeId!,
      price: form.price,
      currency: form.currency,
      pricingUnit: form.pricingUnit,
      conditions: form.conditions,
      guestCount: form.guestCount,
      promotion: form.promotion,
      cabinQuantity: form.cabinQuantity,
      validUntil: form.validUntil || undefined,
      notes: form.notes,
    })

    if (result) {
      submitSuccess.value = true
      handleReset()
    } else {
      submitError.value = quoteStore.error || '提交失败'
    }
  } catch (err) {
    submitError.value = err instanceof Error ? err.message : '提交失败'
  } finally {
    isSubmitting.value = false
  }
}

function handleReset() {
  form.sailingId = null
  form.cabinTypeId = null
  form.price = ''
  form.currency = 'CNY'
  form.pricingUnit = 'PER_PERSON'
  form.conditions = ''
  form.guestCount = undefined
  form.promotion = ''
  form.cabinQuantity = undefined
  form.validUntil = ''
  form.notes = ''
  selectedShipId.value = undefined
  Object.keys(errors).forEach(key => errors[key] = '')
}
</script>
