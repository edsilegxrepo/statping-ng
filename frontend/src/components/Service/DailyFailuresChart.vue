<template>
  <div v-if="selectedDate" class="card text-black-50 bg-white mt-3 mb-5">
    <div class="card-header text-capitalize d-flex justify-content-between align-items-center">
      <span
        >Hourly Failure Breakdown: {{ formatDate(selectedDate) }} (UTC) -
        {{ totalFailures }} total failures</span
      >
      <button class="btn btn-sm btn-outline-secondary" @click="$emit('close')">
        <font-awesome-icon icon="times" />
      </button>
    </div>
    <div class="card-body">
      <div v-if="loading" class="text-center py-5">
        <font-awesome-icon icon="circle-notch" size="3x" spin />
      </div>
      <apexchart v-else width="100%" height="250" type="bar" :options="chartOptions" :series="series"></apexchart>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, reactive } from 'vue'
import Api from '@/API'

const props = defineProps({
  service: {
    type: Object,
    required: true,
  },
  selectedDate: {
    type: Date,
    default: null,
  },
})

defineEmits(['close'])

const loading = ref(false)
const series = ref([])

const chartOptions = reactive({
  chart: {
    id: 'daily-failures',
    toolbar: { show: false },
    zoom: { enabled: false },
  },
  plotOptions: {
    bar: {
      colors: {
        ranges: [
          { from: 0, to: 0, color: '#28a745' },
          { from: 1, to: 10, color: '#98EE99' },
          { from: 11, to: 30, color: '#FFEB3B' },
          { from: 31, to: 60, color: '#FF9800' },
          { from: 61, to: 1000000, color: '#F44336' },
        ],
      },
    },
  },
  xaxis: {
    type: 'category',
    categories: Array.from({ length: 24 }, (_, i) => `${i}:00`),
    title: { text: 'Hour of Day (UTC)' },
  },
  yaxis: {
    title: { text: 'Failures' },
    min: 0,
    forceNiceScale: true,
  },
  tooltip: {
    y: {
      formatter: (val) => `${val} Failures`,
    },
  },
})

const totalFailures = computed(() => {
  if (!series.value || series.value.length === 0 || !series.value[0].data) {
    return 0
  }
  return series.value[0].data.reduce((sum, val) => sum + val, 0)
})

watch(
  () => props.selectedDate,
  (newVal) => {
    if (newVal) {
      fetchDailyData()
    }
  },
  { immediate: true }
)

function formatDate(date) {
  const options = { year: 'numeric', month: 'long', day: 'numeric' }
  return date.toLocaleDateString('en-us', options)
}

async function fetchDailyData() {
  loading.value = true

  const year = props.selectedDate.getFullYear()
  const month = props.selectedDate.getMonth()
  const day = props.selectedDate.getDate()

  const startUnix = Math.floor(Date.UTC(year, month, day, 0, 0, 0) / 1000)
  const endUnix = Math.floor(Date.UTC(year, month, day, 23, 59, 59) / 1000)

  const data = await Api.service_failures_data(props.service.id, startUnix, endUnix, '1h', true)

  const hourlyData = Array(24).fill(0)
  data.forEach((d) => {
    const hour = new Date(d.timeframe).getUTCHours()
    if (hour >= 0 && hour < 24) {
      hourlyData[hour] = d.amount
    }
  })

  series.value = [{ name: 'Failures', data: hourlyData }]
  loading.value = false
}
</script>
