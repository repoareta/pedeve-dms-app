<template>
  <div class="financial-comparison-chart">
    <Line v-if="chartData.labels.length > 0" :data="chartData" :options="chartOptions as any" />
    <div v-else class="empty-chart">
      <p>Data belum tersedia</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
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
  BarElement,
  Title,
  Tooltip,
  Legend,
  Filler
)

interface ChartDataItem {
  label: string
  rkap: number | string | undefined
  realisasi: number | string | undefined
}

const props = defineProps<{
  data: ChartDataItem[]
  title?: string
  isRatio?: boolean // true untuk rasio, false untuk currency
  isMini?: boolean // true untuk mini chart (height lebih kecil)
}>()

// Format value untuk chart
const formatValue = (value: number | string | undefined): number => {
  if (value === undefined || value === null) return 0
  const numValue = typeof value === 'string' ? parseFloat(value) : value
  return isNaN(numValue) ? 0 : numValue
}

const chartData = computed(() => {
  if (!props.data || props.data.length === 0) {
    return {
      labels: [],
      datasets: [],
    }
  }

  const labels = props.data.map(item => item.label)
  const rkapData = props.data.map(item => formatValue(item.rkap))
  const realisasiData = props.data.map(item => formatValue(item.realisasi))

  return {
    labels,
    datasets: [
      {
        label: 'RKAP',
        data: rkapData,
        borderColor: '#1890ff',
        backgroundColor: 'rgba(24, 144, 255, 0.1)',
        borderWidth: 2,
        pointRadius: 4,
        pointHoverRadius: 6,
        pointBackgroundColor: '#1890ff',
        pointBorderColor: '#fff',
        pointBorderWidth: 2,
        fill: false,
        tension: 0.4,
      },
      {
        label: 'Realisasi YTD',
        data: realisasiData,
        borderColor: '#52c41a',
        backgroundColor: 'rgba(82, 196, 26, 0.1)',
        borderWidth: 2,
        pointRadius: 4,
        pointHoverRadius: 6,
        pointBackgroundColor: '#52c41a',
        pointBorderColor: '#fff',
        pointBorderWidth: 2,
        fill: true,
        tension: 0.4,
      },
    ],
  }
})

const chartOptions = computed(() => ({
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
        padding: props.isMini ? 10 : 15,
        font: {
          size: props.isMini ? 10 : 12,
          weight: '500' as const,
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
      callbacks: {
        label: (context: { parsed: { y: number }; dataset: { label?: string } }) => {
          const value = context.parsed.y
          const label = context.dataset.label || ''
          if (props.isRatio) {
            return `${label}: ${value.toFixed(2)}`
          } else {
            // Format currency
            if (value >= 1000000000) {
              return `${label}: Rp ${(value / 1000000000).toFixed(2)}M`
            } else if (value >= 1000000) {
              return `${label}: Rp ${(value / 1000000).toFixed(2)}Jt`
            } else if (value >= 1000) {
              return `${label}: Rp ${(value / 1000).toFixed(2)}Rb`
            }
            return `${label}: Rp ${value.toLocaleString('id-ID')}`
          }
        },
      },
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
        callback: (value: number | string) => {
          const numValue = typeof value === 'string' ? parseFloat(value) : value
          if (props.isRatio) {
            return numValue.toFixed(2)
          } else {
            // Format currency untuk y-axis
            if (numValue >= 1000000000) {
              return `Rp ${(numValue / 1000000000).toFixed(1)}M`
            } else if (numValue >= 1000000) {
              return `Rp ${(numValue / 1000000).toFixed(1)}Jt`
            } else if (numValue >= 1000) {
              return `Rp ${(numValue / 1000).toFixed(1)}Rb`
            }
            return `Rp ${numValue}`
          }
        },
      },
    },
  },
}))
</script>

<style scoped>
.financial-comparison-chart {
  height: v-bind(props.isMini ? '200px' : '350px');
  margin-top: 24px;
  padding: 16px;
  background: #fafafa;
  border-radius: 8px;
}

.empty-chart {
  height: v-bind(props.isMini ? '200px' : '350px');
  display: flex;
  align-items: center;
  justify-content: center;
  color: #999;
  font-size: 14px;
}
</style>
