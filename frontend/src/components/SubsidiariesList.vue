<script setup lang="ts">
import { computed } from 'vue'
import { Icon as IconifyIcon } from '@iconify/vue'
import { useRouter } from 'vue-router'

interface Company {
  id: string
  name: string
  logo?: string | null
}

interface Subsidiary {
  id: string
  name: string
  rkap: string
  dividen: string
  score: string
  financialRatio?: number
  company?: Company
}

const props = defineProps<{
  subsidiaries?: Subsidiary[]
  loading?: boolean
}>()

const router = useRouter()

// Get company logo URL
const getCompanyLogo = (company?: Company): string | undefined => {
  if (!company?.logo) return undefined
const apiURL = import.meta.env.VITE_API_URL || (import.meta.env.DEV ? 'http://localhost:8080' : 'https://api-pedeve-dev.aretaamany.com')
  const baseURL = apiURL.replace(/\/api\/v1$/, '')
  return company.logo.startsWith('http') ? company.logo : `${baseURL}${company.logo}`
}

// Get company initial (first letter of each word)
const getCompanyInitial = (name: string): string => {
  const trimmed = name.trim()
  if (!trimmed) return '??'

  const words = trimmed.split(/\s+/).filter(w => w.length > 0)
  if (words.length >= 2) {
    return (words[0]![0]! + words[1]![0]!).toUpperCase()
  }
  return trimmed.substring(0, 2).toUpperCase()
}

// Get icon color based on company name
const getIconColor = (name: string): string => {
  const colors: string[] = [
    '#FF6B6B', '#4ECDC4', '#45B7D1', '#FFA07A', '#98D8C8',
    '#F7DC6F', '#BB8FCE', '#85C1E2', '#F8B739', '#52BE80'
  ]
  if (!name) return colors[0]!
  
  const hash = name.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0)
  return colors[Math.abs(hash) % colors.length] || colors[0]!
}

const handleSubsidiaryClick = (sub: Subsidiary) => {
  router.push(`/subsidiaries/${sub.id}`)
}

const subsidiariesList = computed(() => {
  return props.subsidiaries || []
})
</script>

<template>
  <a-card class="subsidiaries-card" title="Underperforming Subsidiaries" :bordered="false" :loading="loading">
    <div v-if="subsidiariesList.length > 0" class="subsidiaries-list">
      <div 
        v-for="sub in subsidiariesList" 
        :key="sub.id" 
        class="subsidiary-item"
        @click="handleSubsidiaryClick(sub)"
        style="cursor: pointer;"
      >
        <div class="subsidiary-left">
          <div class="subsidiary-icon">
            <img 
              v-if="getCompanyLogo(sub.company)" 
              :src="getCompanyLogo(sub.company)" 
              :alt="sub.name" 
              class="company-logo"
            />
            <div 
              v-else 
              class="icon-placeholder" 
              :style="{ backgroundColor: getIconColor(sub.name) }"
            >
              {{ getCompanyInitial(sub.name) }}
            </div>
          </div>
          <div class="subsidiary-info">
            <div class="subsidiary-name">{{ sub.name }}</div>
            <div class="subsidiary-details">
              <span>RKAP {{ sub.rkap }}</span>
              <span class="separator">•</span>
              <span>Dividen {{ sub.dividen }}</span>
              <span class="separator">•</span>
              <span>Financial Score <strong :style="{ color: sub.score === 'D' || sub.score === 'D+' ? '#ff4d4f' : '' }">{{ sub.score }}</strong></span>
            </div>
          </div>
        </div>
        <IconifyIcon icon="mdi:chevron-right" width="20" color="#999" />
      </div>
    </div>
    <div v-else class="empty-list">
      <p>No underperforming subsidiaries</p>
    </div>
    <div v-if="subsidiariesList.length > 0" class="card-footer">
      <a-typography-link @click="() => router.push('/subsidiaries')">Learn more</a-typography-link>
    </div>
  </a-card>
</template>
