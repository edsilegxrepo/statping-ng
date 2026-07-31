<template>
  <div>
    <TopNav v-if="fromDashboard" :admin="admin" />
    <PublicTopNav v-else />
    <div class="container-fluid px-md-5 mt-md-5">
    <div v-if="!ready" class="row mt-5">
      <div class="col-12 text-center">
        <font-awesome-icon icon="circle-notch" size="3x" spin />
      </div>
      <div class="col-12 text-center mt-3 mb-3">
        <span class="text-muted">Loading Service</span>
      </div>
    </div>

    <div v-if="servicesLoaded && service && !hasData" class="col-12 mb-4">
      <div class="card text-center py-5">
        <div class="card-body">
          <h3 class="text-muted mb-3">{{ service.name }}</h3>
          <span class="badge mb-4" :class="{ 'bg-success': service.online, 'bg-danger': !service.online }">
            {{ service.online ? 'ONLINE' : 'OFFLINE' }}
          </span>
          <p class="text-muted mb-2">No monitoring data available for this service.</p>
          <p class="text-muted small">This may be a static service or monitoring has not started yet.</p>
          <router-link to="/" class="btn btn-outline-primary mt-3">Back to Monitors</router-link>
        </div>
      </div>
    </div>

    <div v-if="servicesLoaded && service && hasData" class="col-12 mb-4">
      <span
        class="mt-3 mb-3 text-white d-md-none btn d-block d-md-none text-uppercase"
        :class="{ 'bg-success': service.online, 'bg-danger': !service.online }"
      >
        {{ service.online ? $t('online') : $t('offline') }}
      </span>

      <span class="mt-2 font-3">
        <router-link to="/" class="text-black-50 text-decoration-none">{{ coreData.name }}</router-link> -
        <span class="text-muted">{{ service.name }}</span>
        <span
          class="badge float-right d-none d-md-block text-uppercase"
          :class="{ 'bg-success': service.online, 'bg-danger': !service.online }"
        >
          {{ service.online ? $t('online') : $t('offline') }}
        </span>
      </span>

      <ServiceTopStats v-if="loaded" :service="service" />

      <MessageBlock
        v-if="loaded"
        v-for="message in messagesInRange"
        :key="message.id"
        :message="message"
      />

      <div class="card text-black-50 bg-white mt-3">
        <div class="card-header text-capitalize">Timeframe</div>
        <div class="card-body pb-4">
          <div class="row">
            <div class="col">
              <flat-pickr
                :disabled="!loaded"
                @on-change="reload"
                v-model="start_time"
                :config="dateConfig"
                type="text"
                class="form-control text-left"
                required
              />
              <small class="d-block">From {{ formatDate(new Date(start_time)) }}</small>
            </div>
            <div class="col">
              <flat-pickr
                :disabled="!loaded"
                @on-change="reload"
                v-model="end_time"
                :config="dateConfig"
                type="text"
                class="form-control text-left"
                required
              />
              <small class="d-block">To {{ formatDate(new Date(end_time)) }}</small>
            </div>
            <div class="col">
              <select :disabled="!loaded" @change="chartHits()" v-model="group" class="form-control">
                <option value="1m">1 Minute</option>
                <option value="5m">5 Minutes</option>
                <option value="15m">15 Minute</option>
                <option value="30m">30 Minutes</option>
                <option value="1h">1 Hour</option>
                <option value="3h">3 Hours</option>
                <option value="6h">6 Hours</option>
                <option value="12h">12 Hours</option>
                <option value="24h">1 Day</option>
                <option value="168h">7 Days</option>
                <option value="360h">15 Days</option>
              </select>
              <small class="d-block d-md-none d-block">Increment Timeframe</small>
            </div>
          </div>
        </div>
      </div>

      <div class="card text-black-50 bg-white mt-3 mb-3">
        <div class="card-header text-capitalize">Service Latency</div>
        <div v-if="loaded" class="card-body">
          <div class="row">
            <AdvancedChart
              :group="group"
              :updated="updated_chart"
              :start="start_time.toString()"
              :end="end_time.toString()"
              :service="service"
            />
          </div>
        </div>
        <div v-else class="row mt-3 mb-3">
          <div class="col-12 text-center">
            <font-awesome-icon icon="circle-notch" size="3x" spin />
          </div>
        </div>
      </div>

      <div class="card text-black-50 bg-white mb-4">
        <div class="card-header text-capitalize d-flex justify-content-between">
          <span>{{ $t('service_failures') }}</span>
          <small class="text-muted">Last 6 Months</small>
        </div>
        <div class="card-body">
          <div class="service-chart-heatmap mt-2">
            <ServiceHeatmap :service="service" @selected-day="showDailyBreakdown" />
          </div>
        </div>
      </div>

      <DailyFailuresChart
        v-if="selectedDay"
        :service="service"
        :selectedDate="selectedDay"
        @close="selectedDay = null"
      />
    </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useMainStore } from '@/stores/main'
import flatPickr from 'vue-flatpickr-component'
import 'flatpickr/dist/flatpickr.css'
import Api from '@/API'
import MessageBlock from '@/components/Index/MessageBlock.vue'
import TopNav from '@/components/Dashboard/TopNav.vue'
import PublicTopNav from '@/components/Index/TopNav.vue'
import ServiceTopStats from '@/components/Service/ServiceTopStats.vue'
import ServiceHeatmap from '@/components/Service/ServiceHeatmap.vue'
import AdvancedChart from '@/components/Service/AdvancedChart.vue'
import DailyFailuresChart from '@/components/Service/DailyFailuresChart.vue'

const route = useRoute()
const store = useMainStore()

const ready = ref(false)
const loaded = ref(false)
const group = ref('15m')
const data = ref(null)
const uptime_data = ref(null)
const failures_data = ref(null)
const selectedDay = ref(null)

const start_time = ref(beginningOf('day', nowSubtract(259200 * 3)))
const end_time = ref(endOf('today'))

const dateConfig = {
  wrap: true,
  allowInput: true,
  enableTime: true,
  dateFormat: 'Z',
  altInput: true,
  altFormat: 'Y-m-d h:i K',
  maxDate: endOf('today'),
}

const coreData = computed(() => store.core)
const service = computed(() => store.serviceByAll(route.params.id))
const servicesLoaded = computed(() => store.hasPublicData)
const admin = computed(() => store.admin)
const fromDashboard = computed(() => route.query.from === 'dashboard')
const hasData = computed(() => {
  const s = service.value
  if (!s || !s.stats) return true
  return s.stats.hits > 0 || s.stats.failures > 0
})

const params = computed(() => ({
  start: toUnix(new Date(start_time.value)),
  end: toUnix(new Date(end_time.value)),
}))

const messagesInRange = computed(() => {
  if (!service.value) return []
  return store.serviceMessages(service.value.id).filter((m) => inRange(m))
})

watch(() => route.params.id, fetchData)

onMounted(() => {
  fetchData()
})

function fetchData() {
  if (!route.params.id) {
    ready.value = false
    return
  }
  reload()
  ready.value = true
  loaded.value = true
}

async function reload() {
  if (!ready.value || !service.value) {
    return
  }
  await chartHits()
  await chartFailures()
  await fetchUptime()
}

async function updated_chart(start, end) {
  loaded.value = false
  start_time.value = start
  end_time.value = end
  loaded.value = true
}

async function fetchUptime() {
  const uptime = await Api.service_uptime(service.value.id, params.value.start, params.value.end)
  uptime_data.value = parse_uptime(uptime)
}

function parse_uptime(timedata) {
  if (!timedata.series) return []
  const onData = timedata.series.filter((g) => g.online) || []
  const offData = timedata.series.filter((g) => !g.online) || []
  const arr = []
  onData.forEach((d) => {
    arr.push({
      x: 'Online',
      y: [new Date(d.start).getTime(), new Date(d.end).getTime()],
      fillColor: '#0db407',
    })
  })
  offData.forEach((d) => {
    arr.push({
      x: 'Offline',
      y: [new Date(d.start).getTime(), new Date(d.end).getTime()],
      fillColor: '#b40707',
    })
  })
  return [{ data: arr }]
}

function inRange(message) {
  const now = new Date()
  const start = new Date(message.start_on)
  const end = message.start_on === message.end_on ? new Date(8640000000000000) : new Date(message.end_on)
  return now >= start && now <= end
}

async function chartHits() {
  data.value = await Api.service_hits(service.value.id, params.value.start, params.value.end, group.value, false)
}

async function chartFailures() {
  failures_data.value = await Api.service_failures_data(
    service.value.id,
    params.value.start,
    params.value.end,
    group.value,
    true
  )
}

function showDailyBreakdown(date) {
  selectedDay.value = date
}

function formatDate(date) {
  return date.toLocaleString()
}

function toUnix(date) {
  return Math.floor(date.getTime() / 1000)
}

function nowSubtract(seconds) {
  return new Date(Date.now() - seconds * 1000)
}

function beginningOf(period, date = new Date()) {
  const d = new Date(date)
  d.setHours(0, 0, 0, 0)
  return d.toISOString()
}

function endOf(period) {
  const d = new Date()
  d.setHours(23, 59, 59, 999)
  return d.toISOString()
}
</script>

<style scoped></style>
