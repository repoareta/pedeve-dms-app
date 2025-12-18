<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()

const goHome = () => {
  router.push('/subsidiaries')
}
</script>

<template>
  <div class="not-found-container">
    <div class="not-found-content">
      <h1 class="error-code">404</h1>
      <h2 class="error-title">Halaman Tidak Ditemukan</h2>
      <p class="error-description">Halaman yang Anda cari tidak ditemukan.</p>
      <a-button type="primary" size="large" @click="goHome">
        Kembali ke Subsidiaries
      </a-button>
    </div>
  </div>
</template>

<style scoped lang="scss">
.not-found-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
}

.not-found-content {
  text-align: center;
  max-width: 600px;
  width: 100%;
}

.error-code {
  font-size: 120px;
  font-weight: 700;
  color: #1890ff;
  margin: 0;
  line-height: 1;
}

.error-title {
  font-size: 32px;
  font-weight: 600;
  color: #2c3e50;
  margin: 20px 0 10px 0;
}

.error-description {
  font-size: 16px;
  color: #666;
  margin-bottom: 30px;
}

@media (max-width: 768px) {
  .error-code {
    font-size: 80px;
  }

  .error-title {
    font-size: 24px;
  }
}
</style>

