<template>
  <div class="role-dashboard admin-dashboard">
    <a-card class="welcome-card">
      <div class="welcome-content">
        <div class="welcome-icon">
          <IconifyIcon icon="mdi:account-cog" width="64" height="64" />
        </div>
        <h1 class="welcome-title">Selamat Datang, {{ user?.username }}!</h1>
        <p class="welcome-role">
          <a-tag color="blue" size="large">
            <IconifyIcon icon="mdi:shield-account" width="16" style="margin-right: 8px;" />
            Role: {{ user?.role || 'Admin' }}
          </a-tag>
        </p>
        <p class="welcome-message">
          Anda login sebagai <strong>Administrator</strong>. Dashboard ini khusus untuk administrator dengan akses penuh untuk mengelola data perusahaan Anda.
        </p>
        <div class="welcome-info">
          <a-descriptions :column="1" bordered size="small">
            <a-descriptions-item label="Username">
              {{ user?.username }}
            </a-descriptions-item>
            <a-descriptions-item label="Email">
              {{ user?.email }}
            </a-descriptions-item>
            <a-descriptions-item label="Role">
              {{ user?.role || 'Admin' }}
            </a-descriptions-item>
          </a-descriptions>
        </div>
      </div>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAuthStore } from '../stores/auth'
import { Icon as IconifyIcon } from '@iconify/vue'

const authStore = useAuthStore()
const user = computed(() => authStore.user)
</script>

<style scoped>
.role-dashboard {
  padding: 24px;
  min-height: calc(100vh - 200px);
}

.welcome-card {
  max-width: 800px;
  margin: 0 auto;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  border-radius: 8px;
}

.welcome-content {
  text-align: center;
  padding: 40px 20px;
}

.welcome-icon {
  margin-bottom: 24px;
  color: #1890ff;
}

.welcome-title {
  font-size: 32px;
  font-weight: 600;
  color: #333;
  margin-bottom: 16px;
}

.welcome-role {
  margin-bottom: 24px;
}

.welcome-message {
  font-size: 16px;
  color: #666;
  line-height: 1.6;
  margin-bottom: 32px;
  max-width: 600px;
  margin-left: auto;
  margin-right: auto;
}

.welcome-info {
  margin-top: 32px;
  text-align: left;
  max-width: 500px;
  margin-left: auto;
  margin-right: auto;
}

.admin-dashboard .welcome-icon {
  color: #1890ff;
}
</style>

