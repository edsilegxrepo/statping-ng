<template>
  <div class="service-chart-container">
    <div class="chart-actions">
      <button class="data-btn" @click="showDataModal = true" title="View raw data">
        <font-awesome-icon icon="table" />
      </button>
    </div>
    <apexchart v-if="ready" width="100%" height="420" type="line" :options="main_chart_options" :series="main_chart"></apexchart>

    <ChartDataModal
      :show="showDataModal"
      :latencyData="main_data"
      :pingData="ping_data"
      :failureData="failure_data"
      :serviceName="service.name"
      @close="showDataModal = false"
    />
  </div>
</template>

<script setup>
import { ref, computed, reactive, watch, onMounted } from 'vue'
import Api from '@/API'
import ChartDataModal from './ChartDataModal.vue'

const props = defineProps({
  service: {
    type: Object,
    required: true,
  },
  start: {
    type: String,
    required: true,
  },
  end: {
    type: String,
    required: true,
  },
  group: {
    type: String,
    required: true,
  },
  updated: {
    type: Function,
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
const loading = ref(true)
const main_data = ref(null)
const ping_data = ref(null)
const failure_data = ref(null)
const showDataModal = ref(false)

function toUnix(date) {
  return Math.floor(date.getTime() / 1000)
}

function fromUnix(ts) {
  return new Date(parseInt(ts, 10) * 1000).toISOString()
}

function humanTime(ms) {
  if (!ms) return '0ms'
  if (ms >= 10000) return `${Math.round(ms / 1000)} ms`
  return `${ms} μs`
}

function convertToChartData(data) {
  if (!data || !Array.isArray(data)) return { data: [] }
  return {
    data: data.map((d) => ({
      x: new Date(d.timeframe).getTime(),
      y: d.amount,
    })),
  }
}

const main_chart_options = reactive({
  noData: {
    text: 'Loading...',
    align: 'center',
    verticalAlign: 'middle',
    offsetX: 0,
    offsetY: -20,
    style: { color: '#bababa', fontSize: '27px' },
  },
  chart: {
    id: 'mainchart',
    stacked: true,
    events: {
      beforeZoom: (_chartContext, { xaxis }) => {
        const start = (xaxis.min / 1000).toFixed(0)
        const end = (xaxis.max / 1000).toFixed(0)
        props.updated(fromUnix(start), fromUnix(end))
        return {
          xaxis: {
            min: fromUnix(start),
            max: fromUnix(end),
          },
        }
      },
    },
    height: 500,
    width: '100%',
    type: 'area',
    animations: { enabled: false, initialAnimation: { enabled: true } },
    selection: { enabled: true },
    zoom: { enabled: true },
    toolbar: { show: true },
    stroke: { show: false, curve: 'stepline', lineCap: 'butt' },
  },
  grid: {
    show: true,
    borderColor: '#f8f9fa',
    padding: { top: 25, bottom: 25 },
  },
  xaxis: {
    type: 'datetime',
    labels: { show: true },
    tooltip: { enabled: false },
  },
  yaxis: [
    {
      title: { text: 'Latency & Ping' },
      labels: { formatter: (value) => humanTime(value) },
    },
    {
      show: false,
      labels: { formatter: (value) => humanTime(value) },
    },
    {
      opposite: true,
      title: { text: 'Failure Spikes' },
      labels: { formatter: (value) => (value ? value.toFixed(0) : '0') },
      min: 0,
      forceNiceScale: true,
    },
  ],
  markers: { size: 0, strokeWidth: 0, hover: { size: undefined, sizeOffset: 0 } },
  tooltip: {
    theme: false,
    enabled: true,
    custom: ({ series, seriesIndex, dataPointIndex, w }) => {
      const ts = w.globals.seriesX[seriesIndex][dataPointIndex]
      const dt = new Date(ts).toLocaleDateString('en-us', timeoptions)
      const latencyVal = series[0][dataPointIndex]
      const pingVal = series[1][dataPointIndex]
      const failuresVal = series[2] ? series[2][dataPointIndex] : 0
      const latText = humanTime(latencyVal)
      const pingText = humanTime(pingVal)
      const failText = failuresVal ? `${failuresVal} Failures` : '0 Failures'
      return `<div class="p-3" style="background: rgba(255, 255, 255, 0.98); border: 1px solid #dee2e6; border-radius: 6px; box-shadow: 0 4px 15px rgba(0, 0, 0, 0.12); min-width: 190px;">
        <div style="margin-bottom: 6px; font-size: 12px; color: #495057; display: flex; justify-content: space-between;">
          <span><strong>Latency:</strong></span>
          <span style="color: #f1771f; font-weight: bold; margin-left: 10px;">${latText}</span>
        </div>
        <div style="margin-bottom: 6px; font-size: 12px; color: #495057; display: flex; justify-content: space-between;">
          <span><strong>Ping:</strong></span>
          <span style="color: #48d338; font-weight: bold; margin-left: 10px;">${pingText}</span>
        </div>
        <div style="margin-bottom: 6px; font-size: 12px; color: #495057; display: flex; justify-content: space-between;">
          <span><strong>Failures:</strong></span>
          <span style="color: #e01a1a; font-weight: bold; margin-left: 10px;">${failText}</span>
        </div>
        <hr style="margin: 8px 0; border: 0; border-top: 1px solid #e9ecef;">
        <div style="font-size: 10px; color: #6c757d; text-align: right;">${dt}</div>
      </div>`
    },
    fixed: { enabled: false },
    x: { show: true },
  },
  legend: {
    show: true,
    position: 'bottom',
    horizontalAlign: 'center',
    offsetY: 10,
    itemMargin: { horizontal: 25, vertical: 5 },
  },
  dataLabels: { enabled: false },
  floating: true,
  axisTicks: { show: true },
  axisBorder: { show: false },
  colors: ['#f1771f', '#48d338', '#e01a1a'],
  fill: {
    colors: ['#f1771f', '#48d338', '#e01a1a'],
    opacity: [0.5, 0.4, 0.8],
    type: 'solid',
  },
  stroke: {
    show: true,
    curve: 'smooth',
    lineCap: 'butt',
    colors: ['#f1771f', '#48d338', 'transparent'],
    width: [2, 2, 0],
  },
})

const params = computed(() => ({
  start: toUnix(new Date(props.start)),
  end: toUnix(new Date(props.end)),
}))

const main_chart = computed(() => {
  const list = [
    { name: 'Latency', type: 'area', ...convertToChartData(main_data.value) },
    { name: 'Ping', type: 'area', ...convertToChartData(ping_data.value) },
  ]
  if (failure_data.value) {
    list.push({ name: 'Failures', type: 'column', ...convertToChartData(failure_data.value) })
  }
  return list
})

watch(() => props.start, update_data)
watch(() => props.end, update_data)
watch(() => props.group, update_data)

onMounted(async () => {
  await update_data()
})

async function update_data() {
  ready.value = false
  loading.value = true
  await chartHits()
  loading.value = false
  ready.value = true
}

async function chartHits() {
  main_data.value = await load_hits()
  ping_data.value = await load_ping()
  failure_data.value = await load_failures()
}

async function load_hits(start = params.value.start, end = params.value.end, group = props.group) {
  return await Api.service_hits(props.service.id, start, end, group, false)
}

async function load_ping(start = params.value.start, end = params.value.end, group = props.group) {
  return await Api.service_ping(props.service.id, start, end, group, false)
}

async function load_failures(start = params.value.start, end = params.value.end, group = props.group) {
  return await Api.service_failures_data(props.service.id, start, end, group, true)
}
</script>

<style scoped>
.service-chart-container {
  position: relative;
  width: 100%;
}

.chart-actions {
  position: absolute;
  top: 5px;
  right: 195px;
  z-index: 20;
  display: flex;
  gap: 0.25rem;
}

.data-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 17px;
  height: 17px;
  background: transparent;
  border: none;
  border-radius: 2px;
  color: #6e8192;
  cursor: pointer;
  transition: all 0.15s;
  font-size: 15px;
  padding: 0;
}

.data-btn:hover {
  color: #3b82f6;
}
</style>
