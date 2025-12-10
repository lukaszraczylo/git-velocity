<script setup>
import { ref, computed, onMounted, watch } from 'vue'
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

const chartOptions = {
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
        padding: 20,
        font: {
          size: 12
        }
      }
    },
    tooltip: {
      backgroundColor: 'rgba(0, 0, 0, 0.8)',
      padding: 12,
      titleFont: {
        size: 14
      },
      bodyFont: {
        size: 13
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
        font: {
          size: 11
        }
      }
    },
    y: {
      beginAtZero: true,
      grid: {
        color: 'rgba(0, 0, 0, 0.05)'
      },
      ticks: {
        font: {
          size: 11
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
}

function createChart() {
  if (!chartRef.value || !chartData.value.labels.length) return

  if (chartInstance) {
    chartInstance.destroy()
  }

  const ctx = chartRef.value.getContext('2d')
  chartInstance = new Chart(ctx, {
    type: 'line',
    data: chartData.value,
    options: chartOptions
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

onMounted(() => {
  createChart()
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
