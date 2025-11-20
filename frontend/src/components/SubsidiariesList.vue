<script setup lang="ts">
import { Icon as IconifyIcon } from '@iconify/vue'

const subsidiaries = [
  {
    id: 1,
    name: 'PT Petro Nusantara',
    icon: 'mdi:office-building',
    iconColor: '#722ed1',
    rkap: '55%',
    dividen: '20%',
    score: 'D',
  },
  {
    id: 2,
    name: 'PT Pertamina Marine Int',
    icon: 'mdi:ship',
    iconColor: '#faad14',
    rkap: '55%',
    dividen: '20%',
    score: 'D',
  },
  {
    id: 3,
    name: 'PT Energi Abadi Raya',
    icon: 'mdi:lightning-bolt',
    iconColor: '#1890ff',
    rkap: '55%',
    dividen: '20%',
    score: 'D',
  },
  {
    id: 4,
    name: 'PT Geo Minyak Bumi',
    icon: 'mdi:barrel',
    iconColor: '#ff4d4f',
    rkap: '55%',
    dividen: '20%',
    score: 'D',
  },
  {
    id: 5,
    name: 'PT Mitra Fuel Logistik',
    icon: 'mdi:truck',
    iconColor: '#52c41a',
    rkap: '55%',
    dividen: '20%',
    score: 'D',
  },
]
</script>

<template>
  <a-card class="subsidiaries-card" title="Underperforming Subsidiaries" :bordered="false">
    <div class="subsidiaries-list">
      <div 
        v-for="sub in subsidiaries" 
        :key="sub.id" 
        class="subsidiary-item"
      >
        <div class="subsidiary-left">
          <div class="subsidiary-icon" :style="{ background: sub.iconColor }">
            <IconifyIcon :icon="sub.icon" width="20" color="white" />
          </div>
          <div class="subsidiary-info">
            <div class="subsidiary-name">{{ sub.name }}</div>
            <div class="subsidiary-details">
              <span>RKAP {{ sub.rkap }}</span>
              <span class="separator">•</span>
              <span>Dividen {{ sub.dividen }}</span>
              <span class="separator">•</span>
              <span>Financial Score <strong>{{ sub.score }}</strong></span>
            </div>
          </div>
        </div>
        <IconifyIcon icon="mdi:chevron-right" width="20" color="#999" />
      </div>
    </div>
    <div class="card-footer">
      <a-typography-link>Learn more</a-typography-link>
    </div>
  </a-card>
</template>

