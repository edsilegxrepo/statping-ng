<template>
  <div v-if="selectedDate" class="card text-black-50 bg-white mt-3 mb-5">
    <div class="card-header text-capitalize d-flex justify-content-between align-items-center">
      <span
        >Hourly Failure Breakdown: {{ formatDate(selectedDate) }} (UTC) -
        {{ totalFailures }} total failures</span
      >
      <div class="header-actions">
        <button class="action-btn" @click="showDataModal = true" title="View raw data">
          <font-awesome-icon icon="table" />
        </button>
        <div class="dropdown">
          <button class="action-btn" @click="toggleDownloadMenu" title="Download">
            <font-awesome-icon icon="bars" />
          </button>
          <div v-if="showDownloadMenu" class="dropdown-menu-custom">
            <a @click="downloadSVG" class="dropdown-item-custom">Download SVG</a>
            <a @click="downloadPNG" class="dropdown-item-custom">Download PNG</a>
          </div>
        </div>
        <button class="action-btn" @click="$emit('close')" title="Close">
          <font-awesome-icon icon="times" />
        </button>
      </div>
    </div>
    <div class="card-body">
      <div v-if="loading" class="text-center py-5">
        <font-awesome-icon icon="circle-notch" size="3x" spin />
      </div>
      <apexchart v-else ref="chartRef" width="100%" height="250" type="bar" :options="chartOptions" :series="series"></apexchart>

      <DailyFailuresDataModal
        :show="showDataModal"
        :series="series"
        :selectedDate="selectedDate"
        :serviceName="service.name"
        @close="showDataModal = false"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, reactive } from 'vue'
import Api from '@/API'
import DailyFailuresDataModal from './DailyFailuresDataModal.vue'

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
const showDataModal = ref(false)
const showDownloadMenu = ref(false)
const chartRef = ref(null)

function toggleDownloadMenu() {
  showDownloadMenu.value = !showDownloadMenu.value
}

function downloadSVG() {
  if (chartRef.value) {
    chartRef.value.chart.exports.exportToSVG()
  }
  showDownloadMenu.value = false
}

function downloadPNG() {
  if (chartRef.value) {
    chartRef.value.chart.exports.exportToPng()
  }
  showDownloadMenu.value = false
}

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

<style scoped>
.header-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  background: transparent;
  border: 1px solid #dee2e6;
  border-radius: 4px;
  color: #6e8192;
  cursor: pointer;
  transition: all 0.15s;
  font-size: 14px;
}

.action-btn:hover {
  background: #f3f4f6;
  color: #374151;
}

.dropdown {
  position: relative;
}

.dropdown-menu-custom {
  position: absolute;
  top: 100%;
  right: 0;
  background: #fff;
  border: 1px solid #dee2e6;
  border-radius: 6px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  min-width: 140px;
  padding: 0.5rem 0;
  margin-top: 0.25rem;
  z-index: 100;
}

.dropdown-item-custom {
  display: block;
  padding: 0.5rem 1rem;
  font-size: 0.85rem;
  color: #374151;
  text-decoration: none;
  cursor: pointer;
  transition: background 0.15s;
}

.dropdown-item-custom:hover {
  background: #f3f4f6;
}
</style>
