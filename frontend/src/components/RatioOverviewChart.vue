<template>
  <div class="ratio-overview-chart">
    <Line v-if="chartData.labels.length > 0 && !isCompareMode" :data="chartData" :options="lineChartOptions as any" />
    <Bar v-else-if="chartData.labels.length > 0 && isCompareMode" :data="groupedChartData" :options="barChartOptions as any" />
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
import { Line, Bar } from 'vue-chartjs'

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

interface RatioDataItem {
  label: string
  roe: { rkap: number; realisasi: number }
  roi: { rkap: number; realisasi: number }
  currentRatio: { rkap: number; realisasi: number }
  debtToEquity: { rkap: number; realisasi: number }
}

const props = defineProps<{
  data: RatioDataItem[]
}>()

// Detect compare mode by checking if labels contain (P1) or (P2)
const isCompareMode = computed(() => {
  if (!props.data || props.data.length === 0) return false
  return props.data.some(item => item.label.includes('(P1)') || item.label.includes('(P2)'))
})

// Group data by month for compare mode
const groupedChartData = computed(() => {
  if (!isCompareMode.value || !props.data || props.data.length === 0) {
    return { labels: [], datasets: [] }
  }

  // Group data by month (remove P1/P2 suffix)
  const groupedByMonth = new Map<string, { p1?: RatioDataItem; p2?: RatioDataItem }>()
  
  props.data.forEach(item => {
    const monthKey = item.label.replace(/\s*\(P[12]\)\s*$/, '').trim()
    if (item.label.includes('(P1)')) {
      const existing = groupedByMonth.get(monthKey) || {}
      groupedByMonth.set(monthKey, { ...existing, p1: item })
    } else if (item.label.includes('(P2)')) {
      const existing = groupedByMonth.get(monthKey) || {}
      groupedByMonth.set(monthKey, { ...existing, p2: item })
    }
  })

  const labels = Array.from(groupedByMonth.keys())
  
  const datasets = [
    // ROE (RKAP) - P1
    {
      label: 'ROE (RKAP) - P1',
      data: labels.map(label => {
        const group = groupedByMonth.get(label)
        return group?.p1?.roe.rkap || 0
      }),
      backgroundColor: 'rgba(24, 144, 255, 0.8)',
      borderColor: '#1890ff',
      borderWidth: 1,
    },
    // ROE (RKAP) - P2
    {
      label: 'ROE (RKAP) - P2',
      data: labels.map(label => {
        const group = groupedByMonth.get(label)
        return group?.p2?.roe.rkap || 0
      }),
      backgroundColor: 'rgba(24, 144, 255, 0.4)',
      borderColor: '#1890ff',
      borderWidth: 1,
      borderDash: [5, 5],
    },
    // ROE (Realisasi) - P1
    {
      label: 'ROE (Realisasi) - P1',
      data: labels.map(label => {
        const group = groupedByMonth.get(label)
        return group?.p1?.roe.realisasi || 0
      }),
      backgroundColor: 'rgba(82, 196, 26, 0.8)',
      borderColor: '#52c41a',
      borderWidth: 1,
    },
    // ROE (Realisasi) - P2
    {
      label: 'ROE (Realisasi) - P2',
      data: labels.map(label => {
        const group = groupedByMonth.get(label)
        return group?.p2?.roe.realisasi || 0
      }),
      backgroundColor: 'rgba(82, 196, 26, 0.4)',
      borderColor: '#52c41a',
      borderWidth: 1,
      borderDash: [5, 5],
    },
    // ROI (RKAP) - P1
    {
      label: 'ROI (RKAP) - P1',
      data: labels.map(label => {
        const group = groupedByMonth.get(label)
        return group?.p1?.roi.rkap || 0
      }),
      backgroundColor: 'rgba(250, 173, 20, 0.8)',
      borderColor: '#faad14',
      borderWidth: 1,
    },
    // ROI (RKAP) - P2
    {
      label: 'ROI (RKAP) - P2',
      data: labels.map(label => {
        const group = groupedByMonth.get(label)
        return group?.p2?.roi.rkap || 0
      }),
      backgroundColor: 'rgba(250, 173, 20, 0.4)',
      borderColor: '#faad14',
      borderWidth: 1,
      borderDash: [5, 5],
    },
    // ROI (Realisasi) - P1
    {
      label: 'ROI (Realisasi) - P1',
      data: labels.map(label => {
        const group = groupedByMonth.get(label)
        return group?.p1?.roi.realisasi || 0
      }),
      backgroundColor: 'rgba(255, 120, 117, 0.8)',
      borderColor: '#ff7875',
      borderWidth: 1,
    },
    // ROI (Realisasi) - P2
    {
      label: 'ROI (Realisasi) - P2',
      data: labels.map(label => {
        const group = groupedByMonth.get(label)
        return group?.p2?.roi.realisasi || 0
      }),
      backgroundColor: 'rgba(255, 120, 117, 0.4)',
      borderColor: '#ff7875',
      borderWidth: 1,
      borderDash: [5, 5],
    },
    // Current Ratio (RKAP) - P1
    {
      label: 'Current Ratio (RKAP) - P1',
      data: labels.map(label => {
        const group = groupedByMonth.get(label)
        return group?.p1?.currentRatio.rkap || 0
      }),
      backgroundColor: 'rgba(114, 46, 209, 0.8)',
      borderColor: '#722ed1',
      borderWidth: 1,
    },
    // Current Ratio (RKAP) - P2
    {
      label: 'Current Ratio (RKAP) - P2',
      data: labels.map(label => {
        const group = groupedByMonth.get(label)
        return group?.p2?.currentRatio.rkap || 0
      }),
      backgroundColor: 'rgba(114, 46, 209, 0.4)',
      borderColor: '#722ed1',
      borderWidth: 1,
      borderDash: [5, 5],
    },
    // Current Ratio (Realisasi) - P1
    {
      label: 'Current Ratio (Realisasi) - P1',
      data: labels.map(label => {
        const group = groupedByMonth.get(label)
        return group?.p1?.currentRatio.realisasi || 0
      }),
      backgroundColor: 'rgba(235, 47, 150, 0.8)',
      borderColor: '#eb2f96',
      borderWidth: 1,
    },
    // Current Ratio (Realisasi) - P2
    {
      label: 'Current Ratio (Realisasi) - P2',
      data: labels.map(label => {
        const group = groupedByMonth.get(label)
        return group?.p2?.currentRatio.realisasi || 0
      }),
      backgroundColor: 'rgba(235, 47, 150, 0.4)',
      borderColor: '#eb2f96',
      borderWidth: 1,
      borderDash: [5, 5],
    },
    // Debt to Equity (RKAP) - P1
    {
      label: 'Debt to Equity (RKAP) - P1',
      data: labels.map(label => {
        const group = groupedByMonth.get(label)
        return group?.p1?.debtToEquity.rkap || 0
      }),
      backgroundColor: 'rgba(19, 194, 194, 0.8)',
      borderColor: '#13c2c2',
      borderWidth: 1,
    },
    // Debt to Equity (RKAP) - P2
    {
      label: 'Debt to Equity (RKAP) - P2',
      data: labels.map(label => {
        const group = groupedByMonth.get(label)
        return group?.p2?.debtToEquity.rkap || 0
      }),
      backgroundColor: 'rgba(19, 194, 194, 0.4)',
      borderColor: '#13c2c2',
      borderWidth: 1,
      borderDash: [5, 5],
    },
    // Debt to Equity (Realisasi) - P1
    {
      label: 'Debt to Equity (Realisasi) - P1',
      data: labels.map(label => {
        const group = groupedByMonth.get(label)
        return group?.p1?.debtToEquity.realisasi || 0
      }),
      backgroundColor: 'rgba(250, 140, 22, 0.8)',
      borderColor: '#fa8c16',
      borderWidth: 1,
    },
    // Debt to Equity (Realisasi) - P2
    {
      label: 'Debt to Equity (Realisasi) - P2',
      data: labels.map(label => {
        const group = groupedByMonth.get(label)
        return group?.p2?.debtToEquity.realisasi || 0
      }),
      backgroundColor: 'rgba(250, 140, 22, 0.4)',
      borderColor: '#fa8c16',
      borderWidth: 1,
      borderDash: [5, 5],
    },
  ]

  return {
    labels,
    datasets,
  }
})

const chartData = computed(() => {
  if (!props.data || props.data.length === 0) {
    return {
      labels: [],
      datasets: [],
    }
  }

  const labels = props.data.map(item => item.label)
  
  // ROE
  const roeRKAP = props.data.map(item => item.roe.rkap)
  const roeRealisasi = props.data.map(item => item.roe.realisasi)
  
  // ROI
  const roiRKAP = props.data.map(item => item.roi.rkap)
  const roiRealisasi = props.data.map(item => item.roi.realisasi)
  
  // Current Ratio
  const currentRatioRKAP = props.data.map(item => item.currentRatio.rkap)
  const currentRatioRealisasi = props.data.map(item => item.currentRatio.realisasi)
  
  // Debt to Equity
  const debtToEquityRKAP = props.data.map(item => item.debtToEquity.rkap)
  const debtToEquityRealisasi = props.data.map(item => item.debtToEquity.realisasi)

  return {
    labels,
    datasets: [
      {
        label: 'ROE (RKAP)',
        data: roeRKAP,
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
        label: 'ROE (Realisasi)',
        data: roeRealisasi,
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
        label: 'ROI (RKAP)',
        data: roiRKAP,
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
      },
      {
        label: 'ROI (Realisasi)',
        data: roiRealisasi,
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
      },
      {
        label: 'Current Ratio (RKAP)',
        data: currentRatioRKAP,
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
        label: 'Current Ratio (Realisasi)',
        data: currentRatioRealisasi,
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
      {
        label: 'Debt to Equity (RKAP)',
        data: debtToEquityRKAP,
        borderColor: '#13c2c2',
        backgroundColor: 'rgba(19, 194, 194, 0.1)',
        borderWidth: 2,
        pointRadius: 4,
        pointHoverRadius: 6,
        pointBackgroundColor: '#13c2c2',
        pointBorderColor: '#fff',
        pointBorderWidth: 2,
        fill: false,
        tension: 0.4,
      },
      {
        label: 'Debt to Equity (Realisasi)',
        data: debtToEquityRealisasi,
        borderColor: '#fa8c16',
        backgroundColor: 'rgba(250, 140, 22, 0.1)',
        borderWidth: 2,
        pointRadius: 4,
        pointHoverRadius: 6,
        pointBackgroundColor: '#fa8c16',
        pointBorderColor: '#fff',
        pointBorderWidth: 2,
        fill: false,
        tension: 0.4,
      },
    ],
  }
})

const lineChartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    mode: 'index' as const,
    intersect: false,
  },
  plugins: {
    title: {
      display: true,
      text: 'Rasio Keuangan Overview',
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
          size: 10,
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
          return `${label}: ${value.toFixed(2)}%`
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
          return `${numValue.toFixed(2)}%`
        },
      },
    },
  },
}))

const barChartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    mode: 'index' as const,
    intersect: false,
  },
  plugins: {
    title: {
      display: true,
      text: 'Rasio Keuangan Overview (Compare Mode)',
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
          size: 9,
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
        size: 12,
      },
      displayColors: true,
      callbacks: {
        label: (context: { parsed: { y: number }; dataset: { label?: string } }) => {
          const value = context.parsed.y
          const label = context.dataset.label || ''
          return `${label}: ${value.toFixed(2)}%`
        },
      },
    },
  },
  scales: {
    x: {
      stacked: false,
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
      stacked: false,
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
          return `${numValue.toFixed(2)}%`
        },
      },
    },
  },
}))
</script>

<style scoped>
.ratio-overview-chart {
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
