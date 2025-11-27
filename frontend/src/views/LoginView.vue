<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { message } from 'ant-design-vue'
import { Icon as IconifyIcon } from '@iconify/vue'

const router = useRouter()
const authStore = useAuthStore()

// Clear auth state saat masuk ke halaman login
// Memastikan state benar-benar clean saat redirect dari protected route
onMounted(() => {
  // Clear state lokal untuk memastikan kondisi seperti belum login
  // Cookie akan tetap ada di browser, tapi state aplikasi di-reset
  authStore.clearAuthState()
})

const email = ref('')
const password = ref('')
const loading = ref(false)
const requires2FA = ref(false)
const twoFACode = ref('')

const handleLogin = async () => {
  try {
    loading.value = true
    const response = await authStore.login(email.value, password.value)
    
    // Cek apakah 2FA diperlukan
    if (response.requires_2fa) {
      requires2FA.value = true
      message.info(response.message || 'Masukkan kode 2FA dari authenticator app Anda')
      return
    }
    
    // Login sukses normal
    message.success('Login berhasil!')
    setTimeout(() => {
      router.push('/dashboard')
    }, 100)
  } catch (error: unknown) {
    console.error('Login error:', error)
    const axiosError = error as { response?: { status?: number; data?: { requires_2fa?: boolean; message?: string; Message?: string } }; message?: string }
    
    // Cek apakah error karena 2FA diperlukan
    if (axiosError.response?.status === 200 && axiosError.response?.data?.requires_2fa) {
      requires2FA.value = true
      message.info('Masukkan kode 2FA dari authenticator app Anda')
      return
    }
    
    // Ekstrak pesan error dari berbagai lokasi yang mungkin
    const errorMessage = 
      axiosError.response?.data?.message || 
      axiosError.response?.data?.Message || 
      authStore.error || 
      axiosError.message || 
      'Email atau password salah'
    
    message.error({
      content: errorMessage,
      duration: 5,
    })
  } finally {
    loading.value = false
  }
}

const handleVerify2FA = async () => {
  if (!twoFACode.value || twoFACode.value.length !== 6) {
    message.error('Kode harus 6 digit')
    return
  }

  try {
    loading.value = true
    await authStore.loginWith2FA(email.value, password.value, twoFACode.value)
    message.success('Login berhasil!')
    setTimeout(() => {
      router.push('/dashboard')
    }, 100)
  } catch (error: unknown) {
    console.error('2FA verification error:', error)
    const axiosError = error as { response?: { data?: { message?: string; Message?: string } }; message?: string }
    const errorMessage = 
      axiosError.response?.data?.message || 
      axiosError.response?.data?.Message || 
      authStore.error || 
      axiosError.message || 
      'Kode 2FA tidak valid'
    
    message.error({
      content: errorMessage,
      duration: 5,
    })
  } finally {
    loading.value = false
  }
}

const handleBackToLogin = () => {
  requires2FA.value = false
  twoFACode.value = ''
  password.value = ''
  // Hapus state auth yang ada saat kembali ke login
  authStore.logout()
}
</script>

<template>
  <div class="login-page">
    <!-- Left Side - Login Form -->
    <div class="login-left">
      <div class="login-content">
        <!-- Logo -->
        <div class="logo-container">
          <img src="/logo.png" alt="Pertamina Logo" class="logo-image" />
        </div>

        <!-- Login Form -->
        <div class="login-form-container">
          <h2 class="login-title">{{ requires2FA ? 'Verifikasi 2FA' : 'Login' }}</h2>
          <p class="login-subtitle">
            {{ requires2FA ? 'Masukkan kode 6 digit dari authenticator app Anda' : 'Login to your account' }}
          </p>

          <!-- 2FA Verification Form -->
          <a-form 
            v-if="requires2FA"
            :model="{ twoFACode }" 
            @finish="handleVerify2FA" 
            layout="vertical" 
            class="login-form"
            :validate-trigger="['submit']"
          >
            <a-form-item 
              label="Kode 2FA" 
              name="twoFACode"
              :rules="[
                { required: true, message: 'Kode 2FA wajib diisi' },
                { len: 6, message: 'Kode harus 6 digit', trigger: 'blur' }
              ]">
              <a-input 
                v-model:value="twoFACode" 
                placeholder="Masukkan kode 6 digit"
                size="large" 
                :maxlength="6"
                autocomplete="one-time-code"
                :disabled="loading">
                <template #prefix>
                  <IconifyIcon icon="mdi:shield-lock" width="18" />
                </template>
              </a-input>
            </a-form-item>

            <a-form-item>
              <a-button type="primary" html-type="submit" block size="large" :loading="loading" class="login-button">
                Verifikasi
              </a-button>
            </a-form-item>

            <a-form-item>
              <a-button type="link" block @click="handleBackToLogin" :disabled="loading">
                Kembali ke Login
              </a-button>
            </a-form-item>
          </a-form>

          <!-- Normal Login Form -->
          <a-form 
            v-else
            :model="{ email, password }" 
            @finish="handleLogin" 
            layout="vertical" 
            class="login-form"
            :validate-trigger="['submit']"
          >
             <a-form-item 
               label="Email" 
               name="email"
               :rules="[
                 { required: true, message: 'Email wajib diisi' },
                 { type: 'email', message: 'Email tidak valid', trigger: 'blur' }
               ]">
               <a-input 
                 v-model:value="email" 
                 placeholder="mail@example.com" 
                 size="large" 
                 autocomplete="email"
                 :disabled="loading" />
             </a-form-item>

             <a-form-item 
               label="Password" 
               name="password"
               :rules="[{ required: true, message: 'Password wajib diisi', trigger: 'blur' }]">
               <a-input-password 
                 v-model:value="password" 
                 placeholder="Min. 8 Character" 
                 size="large"
                 autocomplete="current-password"
                 :disabled="loading" />
             </a-form-item>

            <a-form-item>
              <a-button type="primary" html-type="submit" block size="large" :loading="loading" class="login-button">
                Login
              </a-button>
            </a-form-item>
          </a-form>
        </div>
      </div>
    </div>

    <!-- Right Side - Image -->
    <div class="login-right">
      <img src="/imgLogin.png" alt="Login Background" class="login-image" />
    </div>
  </div>
</template>
