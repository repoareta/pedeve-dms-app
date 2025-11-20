<script setup lang="ts">
import { computed } from 'vue'
import { Icon as IconifyIcon } from '@iconify/vue'

const props = defineProps<{
  title: string
  value: string
  change: string
  changeType: 'increase' | 'decrease'
  icon?: string
}>()

// Generate random mini chart data for demonstration
const chartData = computed(() => {
  const points = 10
  const data: number[] = []
  for (let i = 0; i < points; i++) {
    data.push(Math.random() * 40 + 30)
  }
  return data
})

const chartPath = computed(() => {
  const data = chartData.value
  if (!data || data.length === 0) return ''
  
  const width = 60
  const height = 30
  const stepX = width / (data.length - 1)
  const minY = Math.min(...data)
  const maxY = Math.max(...data)
  const rangeY = maxY - minY || 1

  const firstValue = data[0] ?? 0
  let path = `M 0 ${height - ((firstValue - minY) / rangeY) * height}`
  for (let i = 1; i < data.length; i++) {
    const x = i * stepX
    const value = data[i] ?? 0
    const y = height - ((value - minY) / rangeY) * height
    path += ` L ${x} ${y}`
  }

  return path
})

const chartFillPath = computed(() => {
  const data = chartData.value
  if (!data || data.length === 0) return ''
  
  const width = 60
  const height = 30
  const stepX = width / (data.length - 1)
  const minY = Math.min(...data)
  const maxY = Math.max(...data)
  const rangeY = maxY - minY || 1

  const firstValue = data[0] ?? 0
  let path = `M 0 ${height - ((firstValue - minY) / rangeY) * height}`
  for (let i = 1; i < data.length; i++) {
    const x = i * stepX
    const value = data[i] ?? 0
    const y = height - ((value - minY) / rangeY) * height
    path += ` L ${x} ${y}`
  }
  path += ` L ${width} ${height} L 0 ${height} Z`

  return path
})

const chartColor = computed(() => {
  return props.changeType === 'increase' ? '#52c41a' : '#ff4d4f'
})
</script>

<template>
  <a-card class="kpi-card" :bordered="false">
    <div class="kpi-content">
      <div class="kpi-header">
        <IconifyIcon :icon="icon || 'mdi:view-grid'" width="24" class="kpi-icon" />
        <span class="kpi-title">{{ title }}</span>
      </div>
      <div class="kpi-main">
        <div class="kpi-left">
          <div class="kpi-value">{{ value }}</div>
          <div class="kpi-change" :class="changeType">
            <IconifyIcon 
              :icon="changeType === 'increase' ? 'mdi:trending-up' : 'mdi:trending-down'" 
              width="16" 
            />
            <span>{{ change }}</span>
          </div>
        </div>
        <div class="kpi-chart">
          <svg width="60" height="30" viewBox="0 0 60 30" class="mini-chart">
            <defs>
              <linearGradient :id="`gradient-${title}`" x1="0%" y1="0%" x2="0%" y2="100%">
                <stop offset="0%" :style="{ stopColor: chartColor, stopOpacity: 0.3 }" />
                <stop offset="100%" :style="{ stopColor: chartColor, stopOpacity: 0.05 }" />
              </linearGradient>
            </defs>
            <path 
              :d="chartFillPath" 
              :fill="`url(#gradient-${title})`"
              class="chart-fill"
            />
            <path 
              :d="chartPath" 
              :stroke="chartColor"
              stroke-width="2"
              fill="none"
              class="chart-line"
            />
          </svg>
        </div>
      </div>
    </div>
  </a-card>
</template>

