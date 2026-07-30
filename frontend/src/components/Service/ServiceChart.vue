<template>
  <div v-show="showing">
    <apexchart v-if="ready" class="service-chart" width="100%" height="340" type="area" :options="chartOptions" :series="series" />
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import Api from '@/API'

const props = defineProps({
  service: {
    type: Object,
    required: true,
  },
  visible: {
    type: Boolean,
    required: true,
  },
  chart_timeframe: {
    type: Object,
    required: true,
  },
})

const timeoptions = {
  weekday: 'long',
  year: 'numeric',
  month: 'long',
  day: 'numeric',
  hour: 'numeric',
  minute: 'numeric',
}

const ready = ref(false)
const showing = ref(false)
const data = ref(null)
const ping_data = ref(null)
const series = ref(null)

const chartOptions = computed(() => ({
  noData: {
    text: 'No Data Found',
    style: { color: '#999', fontSize: '14px' },
  },
  chart: {
    height: 340,
    width: '100%',
    type: 'area',
    animations: {
      enabled: true,
      easing: 'easeinout',
      speed: 800,
      animateGradually: { enabled: false, delay: 400 },
      dynamicAnimation: { enabled: true, speed: 500 },
      hover: { animationDuration: 0 },
      responsiveAnimationDuration: 0,
    },
    selection: { enabled: false },
    zoom: { enabled: false },
    toolbar: { show: false },
  },
  grid: {
    show: true,
    borderColor: '#f0f0f0',
    padding: { top: 10, right: 10, bottom: 25, left: 10 },
  },
  dropShadow: { enabled: false },
  xaxis: {
    type: 'datetime',
    labels: { show: true, style: { fontSize: '10px', colors: '#999' } },
    tooltip: { enabled: false },
    axisBorder: { show: false },
    axisTicks: { show: false },
  },
  yaxis: {
    labels: {
      show: true,
      style: { fontSize: '10px', colors: '#999' },
      formatter: (val) => val >= 1000 ? `${(val/1000).toFixed(2)}s` : `${Math.round(val)}ms`
    }
  },
  markers: {
    size: 0,
    strokeWidth: 0,
    hover: { size: undefined, sizeOffset: 0 },
  },
  tooltip: {
    theme: false,
    enabled: true,
    custom: ({ series: s, seriesIndex, dataPointIndex, w }) => {
      let ts = w.globals.seriesX[seriesIndex][dataPointIndex]
      const dt = new Date(ts).toLocaleDateString('en-us', timeoptions)
      let val = s[0][dataPointIndex]
      let pingVal = s[1] ? s[1][dataPointIndex] : 0
      return `<div class="chartmarker">
<span>Average Response Time: ${humanTime(val)}/${props.chart_timeframe.interval}</span>
<span>Average Ping: ${humanTime(pingVal)}/${props.chart_timeframe.interval}</span>
<span>${dt}</span>
</div>`
    },
    fixed: { enabled: true, position: 'topRight', offsetX: -30, offsetY: 0 },
    x: { show: false },
    y: { formatter: (value) => `${value} %` },
  },
  dataLabels: { enabled: false },
  floating: true,
  axisTicks: { show: false },
  axisBorder: { show: false },
  fill: {
    type: 'gradient',
    gradient: {
      shadeIntensity: 1,
      opacityFrom: 0.4,
      opacityTo: 0.1,
      stops: [0, 90, 100],
    },
  },
  colors: props.service.online ? ['#48d338', '#17a2b8'] : ['#dd3545', '#fd7e14'],
  stroke: {
    show: true,
    curve: 'smooth',
    width: 2,
    colors: props.service.online ? ['#2fb821', '#138496'] : ['#c60f20', '#e96b0a'],
  },
  legend: {
    show: true,
    position: 'top',
    horizontalAlign: 'right',
    fontSize: '10px',
    markers: { width: 8, height: 8 },
    itemMargin: { horizontal: 10 },
  },
}))

watch(
  () => props.visible,
  (newVal) => {
    if (newVal && !showing.value) {
      showing.value = true
      chartHits(props.chart_timeframe)
    }
  },
  { immediate: true }
)

watch(
  () => props.chart_timeframe,
  (newVal) => {
    if (newVal && showing.value) {
      chartHits(newVal)
    }
  }
)

function humanTime(ms) {
  if (!ms) return '0ms'
  if (ms < 1000) return `${Math.round(ms)}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

function now() {
  return new Date()
}

function fromUnix(unix) {
  return new Date(unix * 1000)
}

function toUnix(date) {
  return Math.floor(date.getTime() / 1000)
}

function beginningOf(unit, date) {
  const d = new Date(date)
  d.setMinutes(0, 0, 0)
  return d
}

function endOf(unit, date) {
  const d = new Date(date)
  d.setMinutes(59, 59, 999)
  return d
}

function convertToChartData(arr) {
  if (!arr || !arr.length) return { data: [] }
  return {
    data: arr.map((d) => ({
      x: new Date(d.timeframe).getTime(),
      y: d.amount,
    })),
  }
}

async function chartHits(val) {
  ready.value = false
  const end = endOf('hour', now())
  const start = beginningOf('hour', fromUnix(val.start_time))

  try {
    const [hitsData, pingData] = await Promise.all([
      Api.service_hits(props.service.id, toUnix(start), toUnix(end), val.interval, false),
      Api.service_ping(props.service.id, toUnix(start), toUnix(end), val.interval, false),
    ])

    data.value = hitsData
    ping_data.value = pingData

    series.value = [
      { name: 'Latency', ...convertToChartData(hitsData) },
      { name: 'Ping', ...convertToChartData(pingData) },
    ]
  } catch (e) {
    console.error('Failed to load chart data:', e)
    series.value = [{ name: 'Latency', data: [] }, { name: 'Ping', data: [] }]
  }

  ready.value = true
}
</script>

<style scoped></style>
