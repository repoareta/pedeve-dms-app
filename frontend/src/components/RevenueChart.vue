<script setup lang="ts">
import { ref, onMounted } from 'vue'
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

const chartData = ref({
  labels: [
    'Januari 2025',
    'Februari 2025',
    'Maret 2025',
    'April 2025',
    'Mei 2025',
    'Juni 2025',
    'Juli 2025',
    'Agustus 2025',
    'September 2025',
    'Oktober 2025',
    'November 2025',
    'Desember 2025',
  ],
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
      data: [100, 105, 110, 108, 115, 120, 118, 125, 122, 130, 128, 135],
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
      data: [15, 18, 20, 19, 22, 25, 23, 27, 26, 30, 28, 32],
      fill: true,
      tension: 0.4,
    },
  ],
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
  <a-card class="chart-card" title="Revenue VS NPAT Trends" :bordered="false">
    <template #extra>
      <div class="chart-extra">
        <span class="chart-info">Q1 2024 +10% $120M</span>
      </div>
    </template>
    <div class="chart-container">
      <Line :data="chartData" :options="chartOptions" />
    </div>
  </a-card>
</template>
