<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { message } from 'ant-design-vue'

const router = useRouter()
const authStore = useAuthStore()

const email = ref('')
const password = ref('')
const loading = ref(false)

const handleLogin = async () => {
  try {
    loading.value = true
    await authStore.login(email.value, password.value)
    message.success('Login berhasil!')
    // Wait a bit before redirect to ensure state is saved
    setTimeout(() => {
      router.push('/dashboard')
    }, 100)
  } catch (error: any) {
    console.error('Login error:', error)
    // Extract error message from various possible locations
    const errorMessage = 
      error?.response?.data?.message || 
      error?.response?.data?.Message || 
      authStore.error || 
      error?.message || 
      'Email atau password salah'
    
    message.error({
      content: errorMessage,
      duration: 5,
    })
  } finally {
    loading.value = false
  }
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
          <h2 class="login-title">Login</h2>
          <p class="login-subtitle">Login to your account</p>

          <a-form 
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
