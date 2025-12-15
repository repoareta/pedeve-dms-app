<template>
  <div class="balance-sheet-overview-chart">
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

interface OverviewDataItem {
  label: string
  totalAssets: { rkap: number; realisasi: number }
  totalLiabilities: { rkap: number; realisasi: number }
  equity: { rkap: number; realisasi: number }
}

const props = defineProps<{
  data: OverviewDataItem[]
}>()

const chartData = computed(() => {
  if (!props.data || props.data.length === 0) {
    return {
      labels: [],
      datasets: [],
    }
  }

  const labels = props.data.map(item => item.label)
  
  // Total Assets
  const totalAssetsRKAP = props.data.map(item => item.totalAssets.rkap)
  const totalAssetsRealisasi = props.data.map(item => item.totalAssets.realisasi)
  
  // Total Liabilities
  const totalLiabilitiesRKAP = props.data.map(item => item.totalLiabilities.rkap)
  const totalLiabilitiesRealisasi = props.data.map(item => item.totalLiabilities.realisasi)
  
  // Equity
  const equityRKAP = props.data.map(item => item.equity.rkap)
  const equityRealisasi = props.data.map(item => item.equity.realisasi)

  return {
    labels,
    datasets: [
      {
        label: 'Total Assets (RKAP)',
        data: totalAssetsRKAP,
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
        label: 'Total Assets (Realisasi)',
        data: totalAssetsRealisasi,
        borderColor: '#52c41a',
        backgroundColor: 'rgba(82, 196, 26, 0.1)',
        borderWidth: 2,
        pointRadius: 4,
        pointHoverRadius: 6,
        pointBackgroundColor: '#52c41a',
        pointBorderColor: '#fff',
        pointBorderWidth: 2,
        fill: false,
        tension: 0.4,
      },
      {
        label: 'Total Liabilities (RKAP)',
        data: totalLiabilitiesRKAP,
        borderColor: '#faad14',
        backgroundColor: 'rgba(250, 173, 20, 0.1)',
        borderWidth: 2,
        pointRadius: 4,
        pointHoverRadius: 6,
        pointBackgroundColor: '#faad14',
        pointBorderColor: '#fff',
        pointBorderWidth: 2,
        fill: false,
        tension: 0.4,
        borderDash: [5, 5],
      },
      {
        label: 'Total Liabilities (Realisasi)',
        data: totalLiabilitiesRealisasi,
        borderColor: '#ff7875',
        backgroundColor: 'rgba(255, 120, 117, 0.1)',
        borderWidth: 2,
        pointRadius: 4,
        pointHoverRadius: 6,
        pointBackgroundColor: '#ff7875',
        pointBorderColor: '#fff',
        pointBorderWidth: 2,
        fill: false,
        tension: 0.4,
        borderDash: [5, 5],
      },
      {
        label: 'Equity (RKAP)',
        data: equityRKAP,
        borderColor: '#722ed1',
        backgroundColor: 'rgba(114, 46, 209, 0.1)',
        borderWidth: 2,
        pointRadius: 4,
        pointHoverRadius: 6,
        pointBackgroundColor: '#722ed1',
        pointBorderColor: '#fff',
        pointBorderWidth: 2,
        fill: false,
        tension: 0.4,
      },
      {
        label: 'Equity (Realisasi)',
        data: equityRealisasi,
        borderColor: '#eb2f96',
        backgroundColor: 'rgba(235, 47, 150, 0.1)',
        borderWidth: 2,
        pointRadius: 4,
        pointHoverRadius: 6,
        pointBackgroundColor: '#eb2f96',
        pointBorderColor: '#fff',
        pointBorderWidth: 2,
        fill: false,
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
    title: {
      display: true,
      text: 'Balance Sheet Overview',
      font: {
        size: 16,
        weight: '600' as const,
      },
      padding: {
        top: 10,
        bottom: 20,
      },
    },
    legend: {
      position: 'top' as const,
      labels: {
        usePointStyle: true,
        padding: 15,
        font: {
          size: 11,
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
          // Format currency
          if (value >= 1000000000) {
            return `${label}: Rp ${(value / 1000000000).toFixed(2)}M`
          } else if (value >= 1000000) {
            return `${label}: Rp ${(value / 1000000).toFixed(2)}Jt`
          } else if (value >= 1000) {
            return `${label}: Rp ${(value / 1000).toFixed(2)}Rb`
          }
          return `${label}: Rp ${value.toLocaleString('id-ID')}`
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
          // Format currency untuk y-axis
          if (numValue >= 1000000000) {
            return `Rp ${(numValue / 1000000000).toFixed(1)}M`
          } else if (numValue >= 1000000) {
            return `Rp ${(numValue / 1000000).toFixed(1)}Jt`
          } else if (numValue >= 1000) {
            return `Rp ${(numValue / 1000).toFixed(1)}Rb`
          }
          return `Rp ${numValue}`
        },
      },
    },
  },
}))
</script>

<style scoped>
.balance-sheet-overview-chart {
  height: 400px;
  margin-top: 24px;
  padding: 16px;
  background: #fafafa;
  border-radius: 8px;
}

.empty-chart {
  height: 400px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #999;
  font-size: 14px;
}
</style>
