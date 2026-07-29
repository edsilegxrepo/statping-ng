<template>
  <apexchart v-if="ready" width="100%" height="400" type="heatmap" :options="plotOptions" :series="series"></apexchart>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import Api from '@/API'

const props = defineProps({
  service: {
    type: Object,
    required: true,
  },
})

const emit = defineEmits(['selected-day'])

const ready = ref(false)
const series = ref([{ data: [] }])

const outageSeverity = {
  minor: { start: 1, end: 30 },
  moderate: { start: 30, end: 120 },
  major: { start: 120, end: 240 },
  critical: { start: 240 },
}

const plotOptions = reactive({
  tooltip: {
    enabled: true,
    custom: ({ series: s, seriesIndex, dataPointIndex }) => {
      const failures = s[seriesIndex][dataPointIndex]
      if (failures > 0) {
        return `<div class="p-2"><strong>${failures} Failures</strong><br><small>Click to view hourly breakdown</small></div>`
      }
      return ''
    },
  },
  chart: {
    selection: { enabled: true },
    zoom: { enabled: true },
    toolbar: { show: true },
    events: {
      dataPointSelection: (_event, _chartContext, config) => {
        const monthName = config.w.globals.seriesNames[config.seriesIndex]
        const day = config.dataPointIndex + 1
        const year = new Date().getFullYear()
        const selectedDate = new Date(`${monthName} ${day}, ${year}`)
        emit('selected-day', selectedDate)
      },
    },
  },
  grid: {
    show: true,
    borderColor: '#dee2e6',
    position: 'front',
    xaxis: { lines: { show: false } },
    yaxis: { lines: { show: true } },
    padding: { top: 25, right: 20, bottom: 10, left: 20 },
  },
  stroke: {
    show: true,
    width: 1.5,
    colors: ['#ffffff'],
  },
  dataLabels: { enabled: false },
  colors: ['#cb3d36'],
  xaxis: {
    type: 'category',
    labels: { show: true },
    tooltip: {
      enabled: true,
      formatter: (_value, { seriesIndex, dataPointIndex, w }) => {
        const month = w.globals.seriesNames[seriesIndex]
        const year = new Date().getFullYear()
        return `${dataPointIndex + 1} ${month} ${year}`
      },
    },
  },
  yaxis: {
    labels: {
      show: true,
      style: { fontSize: '12px', fontWeight: 'bold' },
    },
  },
  plotOptions: {
    heatmap: {
      enableShades: false,
      useFillColorAsStroke: false,
      colorScale: {
        ranges: [
          { from: -1000000, to: 0, color: '#f8f9fa', name: 'Healthy' },
          { from: 1, to: 30, color: '#98EE99', name: 'Minor' },
          { from: 31, to: 120, color: '#FFEB3B', name: 'Moderate' },
          { from: 121, to: 240, color: '#FF9800', name: 'Major' },
          { from: 241, to: 1000000, color: '#F44336', name: 'Critical' },
        ],
      },
    },
  },
})

onMounted(async () => {
  await chartHeatmap()
})

function toUnix(date) {
  return Math.floor(date.getTime() / 1000)
}

function firstDayOfMonth(date) {
  return new Date(date.getFullYear(), date.getMonth(), 1)
}

function lastDayOfMonth(date) {
  return new Date(date.getFullYear(), date.getMonth() + 1, 0, 23, 59, 59, 999)
}

function addMonths(date, months) {
  const d = new Date(date)
  d.setMonth(d.getMonth() + months)
  return d
}

async function chartHeatmap() {
  ready.value = false
  const months = []
  const current = firstDayOfMonth(new Date())

  for (let i = 5; i >= 0; i--) {
    const start = addMonths(current, -i)
    const end = lastDayOfMonth(start)
    months.push({ start, end })
  }

  const results = await Promise.all(months.map((m) => heatmapData(m.start, m.end)))
  series.value = results
  ready.value = true
}

async function heatmapData(start, end) {
  const failuresData = await Api.service_failures_data(props.service.id, toUnix(start), toUnix(end), '24h', true)
  const dataArr = mergeData(failuresData)
  return {
    name: start.toLocaleString('en-us', { month: 'long' }),
    data: dataArr,
  }
}

function mergeData(failuresData) {
  const dataArr = []
  for (let i = 0; i < 31; i++) {
    dataArr.push({ x: (i + 1).toString(), y: 0 })
  }

  failuresData.forEach((d) => {
    const day = new Date(d.timeframe).getUTCDate()
    if (day >= 1 && day <= 31) {
      dataArr[day - 1].y = d.amount
    }
  })

  return dataArr
}
</script>

<style scoped></style>
