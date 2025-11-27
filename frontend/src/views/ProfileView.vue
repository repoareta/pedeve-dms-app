<template>
  <div class="profile-layout">
    <DashboardHeader @logout="handleLogout" />

    <div class="profile-content">
      <a-card class="profile-card">
        <template #title>
          <div class="card-title">
            <IconifyIcon icon="mdi:account" width="24" style="margin-right: 8px;" />
            Profile Saya
          </div>
        </template>

        <a-tabs v-model:activeKey="activeTab" type="card">
          <!-- Profile Information Tab -->
          <a-tab-pane key="info" tab="Informasi Profil">
            <a-descriptions :column="1" bordered>
              <a-descriptions-item label="Nama Pengguna">
                {{ user?.username }}
              </a-descriptions-item>
              <a-descriptions-item label="Email">
                {{ user?.email }}
              </a-descriptions-item>
              <a-descriptions-item label="Peran">
                <a-tag :color="getRoleColor(user?.role)">
                  {{ user?.role || '-' }}
                </a-tag>
              </a-descriptions-item>
              <a-descriptions-item label="Tanggal Dibuat">
                {{ formatDate(user?.created_at) }}
              </a-descriptions-item>
              <a-descriptions-item label="Terakhir Diupdate">
                {{ formatDate(user?.updated_at) }}
              </a-descriptions-item>
            </a-descriptions>
          </a-tab-pane>

          <!-- Edit Email Tab -->
          <a-tab-pane key="email" tab="Ubah Email">
            <a-form
              :model="emailForm"
              :rules="emailRules"
              layout="vertical"
              @finish="handleUpdateEmail"
            >
              <a-form-item label="Email Saat Ini">
                <a-input :value="user?.email" disabled />
              </a-form-item>
              <a-form-item label="Email Baru" name="email">
                <a-input
                  v-model:value="emailForm.email"
                  placeholder="Masukkan email baru"
                  type="email"
                />
              </a-form-item>
              <a-form-item>
                <a-button type="primary" html-type="submit" :loading="emailLoading">
                  <IconifyIcon icon="mdi:content-save" width="16" style="margin-right: 8px;" />
                  Simpan Email
                </a-button>
                <a-button style="margin-left: 8px;" @click="resetEmailForm">
                  Batal
                </a-button>
              </a-form-item>
            </a-form>
          </a-tab-pane>

          <!-- Change Password Tab -->
          <a-tab-pane key="password" tab="Ubah Password">
            <a-form
              :model="passwordForm"
              :rules="passwordRules"
              layout="vertical"
              @finish="handleChangePassword"
            >
              <a-form-item label="Password Lama" name="oldPassword">
                <a-input-password
                  v-model:value="passwordForm.oldPassword"
                  placeholder="Masukkan password lama"
                />
              </a-form-item>
              <a-form-item label="Password Baru" name="newPassword">
                <a-input-password
                  v-model:value="passwordForm.newPassword"
                  placeholder="Masukkan password baru (min 8 karakter)"
                />
              </a-form-item>
              <a-form-item label="Konfirmasi Password Baru" name="confirmPassword">
                <a-input-password
                  v-model:value="passwordForm.confirmPassword"
                  placeholder="Konfirmasi password baru"
                />
              </a-form-item>
              <a-form-item>
                <a-button type="primary" html-type="submit" :loading="passwordLoading">
                  <IconifyIcon icon="mdi:lock-reset" width="16" style="margin-right: 8px;" />
                  Ubah Password
                </a-button>
                <a-button style="margin-left: 8px;" @click="resetPasswordForm">
                  Batal
                </a-button>
              </a-form-item>
            </a-form>
          </a-tab-pane>
        </a-tabs>
      </a-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { message } from 'ant-design-vue'
import DashboardHeader from '../components/DashboardHeader.vue'
import { Icon as IconifyIcon } from '@iconify/vue'
import { authApi } from '../api/auth'

const router = useRouter()
const authStore = useAuthStore()

const activeTab = ref('info')
const emailLoading = ref(false)
const passwordLoading = ref(false)

const user = computed(() => authStore.user)

const emailForm = ref({
  email: '',
})

const passwordForm = ref({
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
})

// Validation rules
const emailRules = {
  email: [
    { required: true, message: 'Email wajib diisi', trigger: 'blur' },
    { type: 'email', message: 'Format email tidak valid', trigger: 'blur' },
  ],
}

const passwordRules = {
  oldPassword: [
    { required: true, message: 'Password lama wajib diisi', trigger: 'blur' },
  ],
  newPassword: [
    { required: true, message: 'Password baru wajib diisi', trigger: 'blur' },
    { min: 8, message: 'Password minimal 8 karakter', trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: 'Konfirmasi password wajib diisi', trigger: 'blur' },
    {
      validator: (_rule: unknown, value: string) => {
        if (value !== passwordForm.value.newPassword) {
          return Promise.reject('Konfirmasi password tidak cocok')
        }
        return Promise.resolve()
      },
      trigger: 'blur',
    },
  ],
}

const handleLogout = async () => {
  await authStore.logout()
  router.push('/login')
}

const handleUpdateEmail = async () => {
  if (!emailForm.value.email) {
    message.error('Email wajib diisi')
    return
  }

  emailLoading.value = true
  try {
    const updatedUser = await authApi.updateEmail(emailForm.value.email)
    authStore.user = updatedUser
    localStorage.setItem('auth_user', JSON.stringify(updatedUser))
    message.success('Email berhasil diupdate')
    resetEmailForm()
  } catch (error: unknown) {
    message.error(error.response?.data?.message || 'Gagal mengupdate email')
  } finally {
    emailLoading.value = false
  }
}

const handleChangePassword = async () => {
  if (passwordForm.value.newPassword !== passwordForm.value.confirmPassword) {
    message.error('Konfirmasi password tidak cocok')
    return
  }

  passwordLoading.value = true
  try {
    await authApi.changePassword(
      passwordForm.value.oldPassword,
      passwordForm.value.newPassword
    )
    message.success('Password berhasil diubah')
    resetPasswordForm()
  } catch (error: unknown) {
    message.error(error.response?.data?.message || 'Gagal mengubah password')
  } finally {
    passwordLoading.value = false
  }
}

const resetEmailForm = () => {
  emailForm.value.email = ''
}

const resetPasswordForm = () => {
  passwordForm.value.oldPassword = ''
  passwordForm.value.newPassword = ''
  passwordForm.value.confirmPassword = ''
}

const formatDate = (dateString?: string) => {
  if (!dateString) return '-'
  const date = new Date(dateString)
  return date.toLocaleString('id-ID', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const getRoleColor = (role?: string) => {
  switch (role?.toLowerCase()) {
    case 'superadmin':
      return 'red'
    case 'admin':
      return 'blue'
    case 'manager':
      return 'green'
    case 'staff':
      return 'orange'
    default:
      return 'default'
  }
}

onMounted(async () => {
  // Fetch latest profile data
  try {
    await authStore.fetchProfile()
  } catch (error) {
    console.error('Failed to fetch profile:', error)
  }
})
</script>

<style scoped>
.profile-layout {
  min-height: 100vh;
  background: #f5f5f5;
}

.profile-content {
  max-width: 1000px;
  margin: 0 auto;
  padding: 24px;
}

.profile-card {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  border-radius: 8px;
}

.card-title {
  display: flex;
  align-items: center;
  font-size: 18px;
  font-weight: 600;
}

:deep(.ant-descriptions-item-label) {
  font-weight: 600;
  width: 200px;
}

:deep(.ant-form-item) {
  margin-bottom: 24px;
}
</style>

