<script setup lang="ts">
import { computed, ref } from 'vue'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler,
} from 'chart.js'
import { Line } from 'vue-chartjs'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
)

interface ChartDataProps {
  labels: string[]
  revenueData: number[]
  npatData: number[]
  rkapData: number[]
}

const props = defineProps<{
  chartData?: ChartDataProps
  loading?: boolean
}>()

const chartDataComputed = computed(() => {
  if (!props.chartData || props.chartData.labels.length === 0) {
    return {
      labels: [],
      datasets: [],
    }
  }

  return {
    labels: props.chartData.labels,
  datasets: [
    {
      label: 'RKAP',
      backgroundColor: 'rgba(24, 144, 255, 0.25)',
      borderColor: '#1890ff',
      borderWidth: 2,
      pointRadius: 0,
      pointHoverRadius: 4,
      pointBackgroundColor: '#1890ff',
      pointBorderColor: '#fff',
      pointBorderWidth: 2,
        data: props.chartData.rkapData,
      fill: '+1',
      tension: 0.4,
    },
    {
      label: 'NPAT Trends',
      backgroundColor: 'rgba(82, 196, 26, 0.35)',
      borderColor: '#52c41a',
      borderWidth: 2,
      pointRadius: 0,
      pointHoverRadius: 4,
      pointBackgroundColor: '#52c41a',
      pointBorderColor: '#fff',
      pointBorderWidth: 2,
        data: props.chartData.npatData,
      fill: true,
      tension: 0.4,
    },
  ],
  }
})

// Calculate summary info for chart extra
const chartInfo = computed(() => {
  if (!props.chartData || props.chartData.revenueData.length === 0 || props.chartData.labels.length === 0) {
    return 'No data'
  }
  
  const latestRevenue = props.chartData.revenueData[props.chartData.revenueData.length - 1] || 0
  const prevRevenue = props.chartData.revenueData.length > 1 
    ? (props.chartData.revenueData[props.chartData.revenueData.length - 2] || 0)
    : latestRevenue
  
  const change = prevRevenue > 0 ? ((latestRevenue - prevRevenue) / prevRevenue) * 100 : 0
  const sign = change >= 0 ? '+' : ''
  
  // Ambil label terakhir untuk mendapatkan bulan dan tahun
  const lastLabel = props.chartData.labels[props.chartData.labels.length - 1]
  
  // Parse label untuk mendapatkan bulan dan tahun (format: "Januari 2025")
  const labelParts = lastLabel.split(' ')
  if (labelParts.length < 2) {
    // Fallback jika format tidak sesuai
    return `Q4 ${new Date().getFullYear()} ${sign}${change.toFixed(0)}% $${latestRevenue.toFixed(0)}M`
  }
  
  const monthName = labelParts[0]
  const year = parseInt(labelParts[1]) || new Date().getFullYear()
  
  // Tentukan quarter berdasarkan nama bulan
  const monthNames = [
    'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
    'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'
  ]
  const monthIndex = monthNames.indexOf(monthName)
  
  // Hitung quarter (Q1: Jan-Mar, Q2: Apr-Jun, Q3: Jul-Sep, Q4: Oct-Dec)
  let quarter = 1
  if (monthIndex >= 0) {
    quarter = Math.floor(monthIndex / 3) + 1
  }
  
  return `Q${quarter} ${year} ${sign}${change.toFixed(0)}% $${latestRevenue.toFixed(0)}M`
})

const chartOptions = ref({
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    mode: 'index' as const,
    intersect: false,
  },
  plugins: {
    legend: {
      position: 'top' as const,
      labels: {
        usePointStyle: true,
        padding: 15,
        font: {
          size: 12,
          weight: '500',
        },
      },
    },
    tooltip: {
      mode: 'index' as const,
      intersect: false,
      backgroundColor: 'rgba(0, 0, 0, 0.8)',
      padding: 12,
      titleFont: {
        size: 14,
      },
      bodyFont: {
        size: 13,
      },
      displayColors: true,
    },
  },
  scales: {
    x: {
      grid: {
        display: false,
      },
      ticks: {
        font: {
          size: 11,
        },
        maxRotation: 45,
        minRotation: 0,
      },
    },
    y: {
      beginAtZero: true,
      grid: {
        color: 'rgba(0, 0, 0, 0.05)',
        drawBorder: false,
      },
      ticks: {
        font: {
          size: 11,
        },
      },
    },
  },
})
</script>

<template>
  <a-card class="chart-card" title="Revenue VS NPAT Trends" :bordered="false" :loading="loading">
    <template #extra>
      <div class="chart-extra">
        <span class="chart-info">{{ chartInfo }}</span>
      </div>
    </template>
    <div class="chart-revenue-dashboard">
      <Line v-if="chartDataComputed.labels.length > 0" :data="chartDataComputed" :options="chartOptions as any" />
      <div v-else class="empty-chart">
        <p>No data available</p>
      </div>
    </div>
  </a-card>
</template>
