<template>
  <router-link :to="serviceLink" custom v-slot="{ navigate }">
    <div
      class="dashboard_card card mb-4"
      style="cursor: pointer"
      :class="{ 'offline-card': !service.online }"
      @click="navigate"
    >
      <div class="card-header pb-1">
        <h6>
          <span class="no-decoration font-weight-bold text-dark">{{ service.name }}</span>
          <span
            class="badge float-right text-uppercase"
            :class="{ 'badge-success': service.online, 'badge-danger': !service.online }"
          >
            {{ service.online ? $t('online') : $t('offline') }}
          </span>
        </h6>
      </div>

      <div class="card-body pb-1">
        <div v-if="loaded" class="row pl-2">
          <div class="col-md-12 col-sm-12 pl-2 mt-2 mt-md-0 mb-3">
            <ServiceSparkLine :title="set2_name" subtitle="Latency Last 24 Hours" :series="set2" />
          </div>
          <ServiceEvents :service="service" />
        </div>
        <div v-else class="row mb-5">
          <div class="col-12 col-md-12 text-center">
            <font-awesome-icon icon="circle-notch" class="text-dim" size="2x" spin />
          </div>
        </div>
      </div>

      <div class="card-footer">
        <div class="row">
          <div class="col-5 pr-0">
            <span class="small text-dim">{{ hoverbtn }}</span>
          </div>

          <div class="col-7 pr-2 pl-0">
            <div class="btn-group float-right">
              <button
                @click.stop="goToIncidents"
                @mouseleave="unsetHover"
                @mouseover="setHover($t('incidents'))"
                class="btn btn-sm btn-white incident"
              >
                <font-awesome-icon icon="bullhorn" />
              </button>
              <button
                @click.stop="goToCheckins"
                @mouseleave="unsetHover"
                @mouseover="setHover($t('checkins'))"
                class="btn btn-sm btn-white checkins"
              >
                <font-awesome-icon icon="calendar-check" />
              </button>
              <button
                @click.stop="goToFailures"
                @mouseleave="unsetHover"
                @mouseover="setHover($t('failures'))"
                class="btn btn-sm btn-white failures"
              >
                <font-awesome-icon icon="exclamation-triangle" />
                <span v-if="service.stats?.failures" class="badge badge-danger ml-1">{{ service.stats.failures }}</span>
              </button>
            </div>
          </div>
        </div>
      </div>

      <span v-for="(failure, index) in failures" :key="index" class="alert alert-light">
        {{ $t('failed') }} {{ failure.created_at }}<br />
        {{ failure.issue }}
      </span>
    </div>
  </router-link>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Api from '@/API'
import ServiceSparkLine from './ServiceSparkLine.vue'
import ServiceEvents from '@/components/Dashboard/ServiceEvents.vue'

const props = defineProps({
  service: {
    type: Object,
    required: true,
  },
})

const router = useRouter()
const { t } = useI18n()

const uptime = ref(null)
const hoverbtn = ref('')
const set2 = ref([])
const loaded = ref(false)
const set2_name = ref('')
const failures = ref(null)
const visible = ref(false)

const serviceLink = computed(() => `/service/${props.service.permalink || props.service.id}?from=dashboard`)

onMounted(async () => {
  unsetHover()
  await loadInfo()
})

function setHover(name) {
  hoverbtn.value = name
}

function unsetHover() {
  hoverbtn.value = `${t('uptime')} ${props.service.online_7_days}%`
}

function goToIncidents() {
  router.push({ path: `/dashboard/service/${props.service.id}/incidents` })
}

function goToCheckins() {
  router.push({ path: `/dashboard/service/${props.service.id}/checkins` })
}

function goToFailures() {
  router.push({ path: `/dashboard/service/${props.service.id}/failures` })
}

async function loadInfo() {
  set2.value = await getHits(86400 * 3, '60m')
  set2_name.value = calc(set2.value)
  loaded.value = true
}

function nowSubtract(seconds) {
  return new Date(Date.now() - seconds * 1000)
}

function endOfToday() {
  const d = new Date()
  d.setHours(23, 59, 59, 999)
  return d
}

function toUnix(date) {
  return Math.floor(date.getTime() / 1000)
}

async function getHits(seconds, group) {
  const start = nowSubtract(seconds)
  const end = endOfToday()

  try {
    const fetched = await Api.service_hits(props.service.id, toUnix(start), toUnix(end), group, true)
    const data = convertToChartData(fetched)
    return [{ name: 'Latency', ...data }]
  } catch (e) {
    return [{ name: 'Latency', data: [] }]
  }
}

function convertToChartData(arr) {
  if (!arr || !arr.length) return { data: [] }
  return {
    data: arr.map((d) => ({
      x: new Date(d.timeframe).getTime(),
      y: Math.round(d.amount * 0.001),
    })),
  }
}

function calc(s) {
  const data = s[0]?.data
  if (data && data.length) {
    let total = 0
    data.forEach((f) => {
      total += parseInt(f.y, 10)
    })
    total = total / data.length
    return `${Math.round(total)} ms`
  }
  return 'Offline'
}
</script>
