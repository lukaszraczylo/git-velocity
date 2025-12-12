<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { Chart, registerables } from 'chart.js'

Chart.register(...registerables)

const props = defineProps({
  timeline: {
    type: Object,
    required: true
    // Expected shape: { labels: string[], series: [{ name, color, data }] }
  },
  height: {
    type: String,
    default: '300px'
  },
  showScore: {
    type: Boolean,
    default: false
  }
})

const chartRef = ref(null)
let chartInstance = null

const visibleSeries = computed(() => {
  if (!props.timeline?.series) return []
  // Filter out Score series unless showScore is true
  return props.timeline.series.filter(s => props.showScore || s.name !== 'Score')
})

const chartData = computed(() => {
  if (!props.timeline?.labels || !visibleSeries.value.length) {
    return { labels: [], datasets: [] }
  }

  return {
    labels: props.timeline.labels,
    datasets: visibleSeries.value.map(series => ({
      label: series.name,
      data: series.data,
      borderColor: series.color,
      backgroundColor: series.color + '20', // Add transparency
      fill: true,
      tension: 0.4,
      pointRadius: 3,
      pointHoverRadius: 5
    }))
  }
})

const isMobile = ref(window.innerWidth < 640)

// Dark mode colors
const themeColors = {
  gridColor: 'rgba(255, 255, 255, 0.1)',
  textColor: 'rgba(255, 255, 255, 0.7)',
  tooltipBg: 'rgba(30, 30, 30, 0.95)',
  tooltipText: '#fff',
  tooltipBorder: 'rgba(255, 255, 255, 0.1)'
}

const chartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    mode: 'index',
    intersect: false
  },
  plugins: {
    legend: {
      position: 'top',
      labels: {
        usePointStyle: true,
        padding: isMobile.value ? 10 : 20,
        boxWidth: isMobile.value ? 8 : 12,
        color: themeColors.textColor,
        font: {
          size: isMobile.value ? 10 : 12
        }
      }
    },
    tooltip: {
      backgroundColor: themeColors.tooltipBg,
      titleColor: themeColors.tooltipText,
      bodyColor: themeColors.tooltipText,
      borderColor: themeColors.tooltipBorder,
      borderWidth: 1,
      padding: isMobile.value ? 8 : 12,
      titleFont: {
        size: isMobile.value ? 12 : 14
      },
      bodyFont: {
        size: isMobile.value ? 11 : 13
      },
      callbacks: {
        label: (context) => {
          return `${context.dataset.label}: ${context.parsed.y.toLocaleString()}`
        }
      }
    }
  },
  scales: {
    x: {
      grid: {
        display: false
      },
      ticks: {
        color: themeColors.textColor,
        font: {
          size: isMobile.value ? 9 : 11
        },
        maxRotation: isMobile.value ? 45 : 0,
        autoSkip: true,
        maxTicksLimit: isMobile.value ? 6 : 12
      }
    },
    y: {
      beginAtZero: true,
      grid: {
        color: themeColors.gridColor
      },
      ticks: {
        color: themeColors.textColor,
        font: {
          size: isMobile.value ? 9 : 11
        },
        callback: (value) => {
          if (value >= 1000) {
            return (value / 1000).toFixed(1) + 'k'
          }
          return value
        }
      }
    }
  }
}))

function createChart() {
  if (!chartRef.value || !chartData.value.labels.length) return

  if (chartInstance) {
    chartInstance.destroy()
  }

  const ctx = chartRef.value.getContext('2d')
  chartInstance = new Chart(ctx, {
    type: 'line',
    data: chartData.value,
    options: chartOptions.value
  })
}

function updateChart() {
  if (chartInstance) {
    chartInstance.data = chartData.value
    chartInstance.update()
  } else {
    createChart()
  }
}

function handleResize() {
  const newIsMobile = window.innerWidth < 640
  if (newIsMobile !== isMobile.value) {
    isMobile.value = newIsMobile
    createChart() // Recreate chart with new options
  }
}

onMounted(() => {
  createChart()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  if (chartInstance) {
    chartInstance.destroy()
  }
})

watch(() => props.timeline, () => {
  updateChart()
}, { deep: true })

watch(() => props.showScore, () => {
  updateChart()
})
</script>

<template>
  <div class="velocity-chart" :style="{ height }">
    <canvas ref="chartRef"></canvas>
    <div v-if="!timeline?.labels?.length" class="flex items-center justify-center h-full">
      <p class="text-gray-400">No velocity data available</p>
    </div>
  </div>
</template>

<style scoped>
.velocity-chart {
  position: relative;
  width: 100%;
}
</style>
