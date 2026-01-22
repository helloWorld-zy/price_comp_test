<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const username = ref('')
const password = ref('')
const isLoading = ref(false)
const error = ref('')

const handleSubmit = async () => {
  if (!username.value || !password.value) {
    error.value = '请输入用户名和密码'
    return
  }

  isLoading.value = true
  error.value = ''

  try {
    const success = await authStore.login(username.value, password.value)
    if (success) {
      // Redirect to intended page or dashboard
      const redirect = route.query.redirect as string
      router.push(redirect || '/dashboard')
    } else {
      error.value = authStore.error || '登录失败'
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : '登录失败'
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="login-container d-flex align-items-center justify-content-center min-vh-100 bg-light">
    <div class="card shadow-sm" style="width: 100%; max-width: 400px">
      <div class="card-body p-4">
        <div class="text-center mb-4">
          <h4 class="mb-1">邮轮价格对比系统</h4>
          <p class="text-muted">请登录您的账号</p>
        </div>

        <form @submit.prevent="handleSubmit">
          <div class="mb-3">
            <label for="username" class="form-label">用户名</label>
            <input
              id="username"
              v-model="username"
              type="text"
              class="form-control"
              placeholder="请输入用户名"
              :disabled="isLoading"
              autocomplete="username"
            />
          </div>

          <div class="mb-3">
            <label for="password" class="form-label">密码</label>
            <input
              id="password"
              v-model="password"
              type="password"
              class="form-control"
              placeholder="请输入密码"
              :disabled="isLoading"
              autocomplete="current-password"
            />
          </div>

          <div v-if="error" class="alert alert-danger py-2" role="alert">
            {{ error }}
          </div>

          <button type="submit" class="btn btn-primary w-100" :disabled="isLoading">
            <span v-if="isLoading">
              <span class="spinner-border spinner-border-sm me-2" role="status"></span>
              登录中...
            </span>
            <span v-else>登录</span>
          </button>
        </form>

        <div class="mt-4 text-center text-muted small">
          <p class="mb-0">默认管理员账号: admin / admin123</p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-container {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.card {
  border: none;
  border-radius: 1rem;
}
</style>
