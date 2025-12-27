<template>
  <div class="balance-sheet-overview-chart">
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

interface OverviewDataItem {
  label: string
  totalAssets: { rkap: number; realisasi: number }
  totalLiabilities: { rkap: number; realisasi: number }
  equity: { rkap: number; realisasi: number }
}

const props = defineProps<{
  data: OverviewDataItem[]
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

  // Group data by month (remove P1/P2 suffix but keep full label for display)
  const groupedByMonth = new Map<string, { p1?: OverviewDataItem; p2?: OverviewDataItem; displayLabel?: string }>()
  
  props.data.forEach(item => {
    // Extract month name without (P1) or (P2) suffix for grouping
    const monthKey = item.label.replace(/\s*\(P[12]\)\s*$/, '').trim()
    // Keep full label for display (includes year if different periods)
    const displayLabel = item.label
    
    if (item.label.includes('(P1)')) {
      const existing = groupedByMonth.get(monthKey) || {}
      groupedByMonth.set(monthKey, { ...existing, p1: item, displayLabel: existing.displayLabel || displayLabel })
    } else if (item.label.includes('(P2)')) {
      const existing = groupedByMonth.get(monthKey) || {}
      groupedByMonth.set(monthKey, { ...existing, p2: item, displayLabel: existing.displayLabel || displayLabel })
    }
  })

  // Create labels from grouped months - use full label if months are different, otherwise use month name
  const labels = Array.from(groupedByMonth.keys()).map(key => {
    const group = groupedByMonth.get(key)
    // If both P1 and P2 exist for same month, use month name only
    // Otherwise, use full label to distinguish different periods
    if (group?.p1 && group?.p2) {
      return key
    }
    return group?.displayLabel || key
  })
  
  // Prepare datasets for grouped bar chart
  const datasets = [
    // Total Assets (RKAP) - P1
    {
      label: 'Total Assets (RKAP) - P1',
      data: labels.map(label => {
        const group = groupedByMonth.get(label)
        return group?.p1?.totalAssets.rkap || 0
      }),
      backgroundColor: 'rgba(24, 144, 255, 0.8)',
      borderColor: '#1890ff',
      borderWidth: 1,
    },
    // Total Assets (RKAP) - P2
    {
      label: 'Total Assets (RKAP) - P2',
      data: labels.map(label => {
        const group = groupedByMonth.get(label)
        return group?.p2?.totalAssets.rkap || 0
      }),
      backgroundColor: 'rgba(24, 144, 255, 0.4)',
      borderColor: '#1890ff',
      borderWidth: 1,
      borderDash: [5, 5],
    },
    // Total Assets (Realisasi) - P1
    {
      label: 'Total Assets (Realisasi) - P1',
      data: labels.map(label => {
        const group = groupedByMonth.get(label)
        return group?.p1?.totalAssets.realisasi || 0
      }),
      backgroundColor: 'rgba(82, 196, 26, 0.8)',
      borderColor: '#52c41a',
      borderWidth: 1,
    },
    // Total Assets (Realisasi) - P2
    {
      label: 'Total Assets (Realisasi) - P2',
      data: labels.map(label => {
        const group = groupedByMonth.get(label)
        return group?.p2?.totalAssets.realisasi || 0
      }),
      backgroundColor: 'rgba(82, 196, 26, 0.4)',
      borderColor: '#52c41a',
      borderWidth: 1,
      borderDash: [5, 5],
    },
    // Total Liabilities (RKAP) - P1
    {
      label: 'Total Liabilities (RKAP) - P1',
      data: labels.map(label => {
        const group = groupedByMonth.get(label)
        return group?.p1?.totalLiabilities.rkap || 0
      }),
      backgroundColor: 'rgba(250, 173, 20, 0.8)',
      borderColor: '#faad14',
      borderWidth: 1,
    },
    // Total Liabilities (RKAP) - P2
    {
      label: 'Total Liabilities (RKAP) - P2',
      data: labels.map(label => {
        const group = groupedByMonth.get(label)
        return group?.p2?.totalLiabilities.rkap || 0
      }),
      backgroundColor: 'rgba(250, 173, 20, 0.4)',
      borderColor: '#faad14',
      borderWidth: 1,
      borderDash: [5, 5],
    },
    // Total Liabilities (Realisasi) - P1
    {
      label: 'Total Liabilities (Realisasi) - P1',
      data: labels.map(label => {
        const group = groupedByMonth.get(label)
        return group?.p1?.totalLiabilities.realisasi || 0
      }),
      backgroundColor: 'rgba(255, 120, 117, 0.8)',
      borderColor: '#ff7875',
      borderWidth: 1,
    },
    // Total Liabilities (Realisasi) - P2
    {
      label: 'Total Liabilities (Realisasi) - P2',
      data: labels.map(label => {
        const group = groupedByMonth.get(label)
        return group?.p2?.totalLiabilities.realisasi || 0
      }),
      backgroundColor: 'rgba(255, 120, 117, 0.4)',
      borderColor: '#ff7875',
      borderWidth: 1,
      borderDash: [5, 5],
    },
    // Equity (RKAP) - P1
    {
      label: 'Equity (RKAP) - P1',
      data: labels.map(label => {
        const group = groupedByMonth.get(label)
        return group?.p1?.equity.rkap || 0
      }),
      backgroundColor: 'rgba(114, 46, 209, 0.8)',
      borderColor: '#722ed1',
      borderWidth: 1,
    },
    // Equity (RKAP) - P2
    {
      label: 'Equity (RKAP) - P2',
      data: labels.map(label => {
        const group = groupedByMonth.get(label)
        return group?.p2?.equity.rkap || 0
      }),
      backgroundColor: 'rgba(114, 46, 209, 0.4)',
      borderColor: '#722ed1',
      borderWidth: 1,
      borderDash: [5, 5],
    },
    // Equity (Realisasi) - P1
    {
      label: 'Equity (Realisasi) - P1',
      data: labels.map(label => {
        const group = groupedByMonth.get(label)
        return group?.p1?.equity.realisasi || 0
      }),
      backgroundColor: 'rgba(235, 47, 150, 0.8)',
      borderColor: '#eb2f96',
      borderWidth: 1,
    },
    // Equity (Realisasi) - P2
    {
      label: 'Equity (Realisasi) - P2',
      data: labels.map(label => {
        const group = groupedByMonth.get(label)
        return group?.p2?.equity.realisasi || 0
      }),
      backgroundColor: 'rgba(235, 47, 150, 0.4)',
      borderColor: '#eb2f96',
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

const formatCurrency = (value: number) => {
  if (value >= 1000000000) {
    return `Rp ${(value / 1000000000).toFixed(2)}M`
  } else if (value >= 1000000) {
    return `Rp ${(value / 1000000).toFixed(2)}Jt`
  } else if (value >= 1000) {
    return `Rp ${(value / 1000).toFixed(2)}Rb`
  }
  return `Rp ${value.toLocaleString('id-ID')}`
}

const formatCurrencyAxis = (value: number | string) => {
  const numValue = typeof value === 'string' ? parseFloat(value) : value
  if (numValue >= 1000000000) {
    return `Rp ${(numValue / 1000000000).toFixed(1)}M`
  } else if (numValue >= 1000000) {
    return `Rp ${(numValue / 1000000).toFixed(1)}Jt`
  } else if (numValue >= 1000) {
    return `Rp ${(numValue / 1000).toFixed(1)}Rb`
  }
  return `Rp ${numValue}`
}

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
          return `${label}: ${formatCurrency(value)}`
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
        callback: formatCurrencyAxis,
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
      text: 'Balance Sheet Overview (Compare Mode)',
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
        size: 12,
      },
      displayColors: true,
      callbacks: {
        label: (context: { parsed: { y: number }; dataset: { label?: string } }) => {
          const value = context.parsed.y
          const label = context.dataset.label || ''
          return `${label}: ${formatCurrency(value)}`
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
        callback: formatCurrencyAxis,
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
