<template>
  <div style="display: flex; justify-content: center; align-items: center; height: 100vh;">
    <a-spin size="large" />
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { userApi } from '../api/userManagement'
import { message } from 'ant-design-vue'

const router = useRouter()

onMounted(async () => {
  try {
    // Ambil companies user
    const companies = await userApi.getMyCompanies()
    
    if (companies.length === 0) {
      // Tidak ada companies yang di-assign, redirect ke subsidiaries list
      message.warning('Anda belum di-assign ke perusahaan')
      router.replace({ name: 'subsidiaries' })
    } else if (companies.length === 1) {
      // Single company, redirect to detail page
      router.replace({ name: 'subsidiary-detail', params: { id: companies[0]!.company.id } })
    } else {
      // Multiple companies, redirect ke subsidiaries list
      router.replace({ name: 'subsidiaries' })
    }
  } catch (error) {
    console.error('Error loading user companies:', error)
    message.error('Gagal memuat data perusahaan')
    // Kalau error, redirect ke subsidiaries list
    router.replace({ name: 'subsidiaries' })
  }
})
</script>
