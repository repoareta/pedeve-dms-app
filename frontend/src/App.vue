<script setup lang="ts">
import { RouterView } from 'vue-router'
import { onMounted } from 'vue'
import { ConfigProvider } from 'ant-design-vue'
import { getCSRFToken } from './api/client'

// Inisialisasi token CSRF saat app mount
// Handle connection errors dengan graceful (tidak throw error)
onMounted(async () => {
  try {
    await getCSRFToken()
  } catch {
    // Ignore errors saat mount - token akan di-fetch lagi saat diperlukan
    // Server mungkin belum ready, tidak perlu throw error
  }
})

const theme = {
  token: {
    colorPrimary: '#035CAB',
    colorSuccess: '#52c41a',
    colorWarning: '#faad14',
    colorError: '#DB241B',
    colorInfo: '#035CAB',
    colorText: '#333333',
    colorTextSecondary: '#666666',
    borderRadius: 6,
  },
}
</script>

<template>
  <ConfigProvider :theme="theme">
    <RouterView />
  </ConfigProvider>
</template>

<style scoped>
header {
  line-height: 1.5;
  background: white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  padding: 1rem 0;
}

.wrapper {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 2rem;
}

nav {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

nav a {
  display: inline-block;
  padding: 0.5rem 1rem;
  text-decoration: none;
  color: #2c3e50;
  font-weight: 500;
  transition: color 0.3s;
}

nav a:hover {
  color: #667eea;
}

nav a.router-link-exact-active {
  color: #667eea;
  border-bottom: 2px solid #667eea;
}

.auth-section {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.user-info {
  color: #666;
  font-size: 0.9rem;
}

.logout-btn {
  padding: 0.5rem 1rem;
  background: #dc3545;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.9rem;
  transition: background 0.3s;
}

.logout-btn:hover {
  background: #c82333;
}
</style>
