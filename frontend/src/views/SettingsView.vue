<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { message } from 'ant-design-vue'
import DashboardHeader from '../components/DashboardHeader.vue'
import { Icon as IconifyIcon } from '@iconify/vue'

const router = useRouter()
const authStore = useAuthStore()

const loading = ref(false)
const is2FAEnabled = ref(false)
const setupStep = ref<'idle' | 'generate' | 'verify' | 'success'>('idle')
const qrCode = ref<string>('')
const secret = ref<string>('')
const twoFACode = ref('')
const backupCodes = ref<string[]>([])

const handleLogout = () => {
  authStore.logout()
  router.push('/login')
}

const check2FAStatus = async () => {
  try {
    loading.value = true
    const status = await authStore.get2FAStatus()
    is2FAEnabled.value = status.enabled
    if (!status.enabled) {
      setupStep.value = 'idle'
    }
  } catch (error: any) {
    console.error('Failed to get 2FA status:', error)
  } finally {
    loading.value = false
  }
}

const handleEnable2FA = async () => {
  try {
    loading.value = true
    const response = await authStore.generate2FA()
    qrCode.value = response.qr_code
    secret.value = response.secret
    setupStep.value = 'generate'
    message.success('QR Code berhasil di-generate. Silakan scan dengan authenticator app Anda.')
  } catch (error: any) {
    console.error('Error generating 2FA:', error)
    const errorMessage = 
      error?.response?.data?.message || 
      error?.response?.data?.Message ||
      error?.message ||
      authStore.error ||
      'Gagal generate QR Code. Pastikan Anda sudah login dan coba lagi.'
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
    const response = await authStore.verify2FA(twoFACode.value)
    backupCodes.value = response.backup_codes || []
    is2FAEnabled.value = true
    setupStep.value = 'success'
    message.success('2FA berhasil diaktifkan!')
  } catch (error: any) {
    const errorMessage = error?.response?.data?.message || 'Kode verifikasi tidak valid'
    message.error(errorMessage)
  } finally {
    loading.value = false
  }
}

const handleCancelSetup = () => {
  setupStep.value = 'idle'
  qrCode.value = ''
  secret.value = ''
  twoFACode.value = ''
}

const copySecret = () => {
  navigator.clipboard.writeText(secret.value)
  message.success('Secret berhasil di-copy!')
}

const copyBackupCodes = () => {
  const codesText = backupCodes.value.join('\n')
  navigator.clipboard.writeText(codesText)
  message.success('Backup codes berhasil di-copy!')
}

const downloadBackupCodes = () => {
  const codesText = backupCodes.value.join('\n')
  const blob = new Blob([codesText], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'backup-codes.txt'
  a.click()
  URL.revokeObjectURL(url)
  message.success('Backup codes berhasil di-download!')
}

const handleDone = () => {
  setupStep.value = 'idle'
  twoFACode.value = ''
  backupCodes.value = []
}

onMounted(() => {
  check2FAStatus()
})
</script>

<template>
  <div class="settings-page">
    <DashboardHeader @logout="handleLogout" />

    <div class="settings-content">
      <div class="settings-container">
        <h1 class="settings-title">Settings</h1>

        <!-- Security Section -->
        <a-card class="security-card" :loading="loading">
          <template #title>
            <div class="card-title">
              <IconifyIcon icon="mdi:shield-lock" width="24" height="24" />
              <span>Security</span>
            </div>
          </template>

          <!-- Two-Factor Authentication -->
          <div class="security-section">
            <div class="section-header">
              <div>
                <h3 class="section-title">Two-Factor Authentication (2FA)</h3>
                <p class="section-description">
                  Tambahkan lapisan keamanan ekstra dengan authenticator app
                </p>
              </div>
              <a-tag :color="is2FAEnabled ? 'success' : 'default'">
                {{ is2FAEnabled ? 'Enabled' : 'Disabled' }}
              </a-tag>
            </div>

            <!-- Step 1: Idle / Enable Button -->
            <div v-if="setupStep === 'idle' && !is2FAEnabled" class="setup-content">
              <a-button type="primary" size="large" @click="handleEnable2FA" :loading="loading">
                <IconifyIcon icon="mdi:shield-check" width="18" style="margin-right: 8px;" />
                Enable 2FA
              </a-button>
            </div>

            <!-- Step 2: Generate QR Code -->
            <div v-if="setupStep === 'generate'" class="setup-content">
              <div class="qr-container">
                <h4 class="step-title">Scan QR Code</h4>
                <p class="step-description">
                  Buka aplikasi authenticator (Google Authenticator, Authy, dll) dan scan QR code berikut:
                </p>
                
                <div class="qr-code-wrapper">
                  <img :src="`data:image/png;base64,${qrCode}`" alt="QR Code" class="qr-code-image" />
                </div>

                <div class="manual-entry">
                  <p class="manual-entry-label">Atau masukkan secret secara manual:</p>
                  <div class="secret-display">
                    <code class="secret-code">{{ secret }}</code>
                    <a-button type="text" @click="copySecret" size="small">
                      <IconifyIcon icon="mdi:content-copy" width="16" />
                    </a-button>
                  </div>
                </div>

                <a-divider />

                <h4 class="step-title">Masukkan Kode Verifikasi</h4>
                <p class="step-description">
                  Setelah scan QR code, masukkan kode 6 digit dari authenticator app:
                </p>

                <a-input
                  v-model:value="twoFACode"
                  placeholder="Masukkan kode 6 digit"
                  size="large"
                  :maxlength="6"
                  class="code-input"
                  @keyup.enter="handleVerify2FA"
                >
                  <template #prefix>
                    <IconifyIcon icon="mdi:key" width="18" />
                  </template>
                </a-input>

                <div class="setup-actions">
                  <a-button @click="handleCancelSetup">Cancel</a-button>
                  <a-button type="primary" @click="handleVerify2FA" :loading="loading">
                    Verify & Enable
                  </a-button>
                </div>
              </div>
            </div>

            <!-- Step 3: Success with Backup Codes -->
            <div v-if="setupStep === 'success'" class="setup-content">
              <a-result
                status="success"
                title="2FA Berhasil Diaktifkan!"
                sub-title="Silakan simpan backup codes berikut untuk akses darurat"
              >
                <template #extra>
                  <div class="backup-codes-container">
                    <div class="backup-codes-header">
                      <h4>Backup Codes</h4>
                      <div class="backup-codes-actions">
                        <a-button @click="copyBackupCodes" size="small">
                          <IconifyIcon icon="mdi:content-copy" width="16" style="margin-right: 4px;" />
                          Copy
                        </a-button>
                        <a-button @click="downloadBackupCodes" size="small">
                          <IconifyIcon icon="mdi:download" width="16" style="margin-right: 4px;" />
                          Download
                        </a-button>
                      </div>
                    </div>
                    <div class="backup-codes-list">
                      <div v-for="(code, index) in backupCodes" :key="index" class="backup-code-item">
                        {{ code }}
                      </div>
                    </div>
                    <a-alert
                      message="Penting!"
                      description="Simpan backup codes ini di tempat yang aman. Anda akan membutuhkannya jika kehilangan akses ke authenticator app."
                      type="warning"
                      show-icon
                      style="margin-top: 16px;"
                    />
                    <a-button type="primary" block @click="handleDone" style="margin-top: 16px;">
                      Done
                    </a-button>
                  </div>
                </template>
              </a-result>
            </div>

            <!-- 2FA Enabled Info -->
            <div v-if="is2FAEnabled && setupStep === 'idle'" class="setup-content">
              <a-alert
                message="2FA Aktif"
                description="Two-Factor Authentication sudah aktif untuk akun Anda. Pastikan Anda memiliki akses ke authenticator app saat login."
                type="success"
                show-icon
              />
            </div>
          </div>
        </a-card>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.settings-page {
  min-height: 100vh;
  background: #f5f5f5;
}

.settings-content {
  padding: 24px;
  max-width: 1200px;
  margin: 0 auto;
}

.settings-container {
  .settings-title {
    font-size: 28px;
    font-weight: 600;
    margin-bottom: 24px;
    color: #1a1a1a;
  }
}

.security-card {
  .card-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 18px;
    font-weight: 600;
  }
}

.security-section {
  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 24px;
    padding-bottom: 16px;
    border-bottom: 1px solid #f0f0f0;
  }

  .section-title {
    font-size: 16px;
    font-weight: 600;
    margin: 0 0 4px 0;
    color: #1a1a1a;
  }

  .section-description {
    margin: 0;
    color: #666;
    font-size: 14px;
  }
}

.setup-content {
  padding: 16px 0;
}

.qr-container {
  max-width: 500px;

  .step-title {
    font-size: 16px;
    font-weight: 600;
    margin: 0 0 8px 0;
    color: #1a1a1a;
  }

  .step-description {
    color: #666;
    font-size: 14px;
    margin-bottom: 16px;
  }

  .qr-code-wrapper {
    display: flex;
    justify-content: center;
    padding: 24px;
    background: #fff;
    border: 1px solid #e8e8e8;
    border-radius: 8px;
    margin-bottom: 24px;
  }

  .qr-code-image {
    width: 200px;
    height: 200px;
  }

  .manual-entry {
    margin-bottom: 24px;

    .manual-entry-label {
      font-size: 14px;
      color: #666;
      margin-bottom: 8px;
    }

    .secret-display {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 12px;
      background: #f5f5f5;
      border: 1px solid #e8e8e8;
      border-radius: 6px;

      .secret-code {
        flex: 1;
        font-family: monospace;
        font-size: 14px;
        color: #1a1a1a;
        word-break: break-all;
      }
    }
  }

  .code-input {
    margin-bottom: 16px;
  }

  .setup-actions {
    display: flex;
    gap: 12px;
    justify-content: flex-end;
  }
}

.backup-codes-container {
  max-width: 500px;
  margin: 0 auto;

  .backup-codes-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;

    h4 {
      margin: 0;
      font-size: 16px;
      font-weight: 600;
    }

    .backup-codes-actions {
      display: flex;
      gap: 8px;
    }
  }

  .backup-codes-list {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 8px;
    padding: 16px;
    background: #f5f5f5;
    border: 1px solid #e8e8e8;
    border-radius: 6px;
    margin-bottom: 16px;

    .backup-code-item {
      font-family: monospace;
      font-size: 14px;
      padding: 8px;
      background: #fff;
      border: 1px solid #e8e8e8;
      border-radius: 4px;
      text-align: center;
    }
  }
}

@media (max-width: 768px) {
  .settings-content {
    padding: 16px;
  }

  .backup-codes-list {
    grid-template-columns: 1fr !important;
  }
}
</style>

