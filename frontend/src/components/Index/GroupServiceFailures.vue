<template>
  <div>
    <div v-if="!loaded" class="row">
      <div class="col-12 text-center mt-3">
        <font-awesome-icon icon="circle-notch" class="text-dim" size="2x" spin />
      </div>
    </div>
    <transition name="fade">
      <div v-if="loaded">
        <div class="d-flex mt-3">
          <div
            v-for="(d, index) in failureData"
            :key="index"
            class="flex-fill service_day"
            @mouseover="mouseover(d)"
            @mouseout="mouseout"
            :class="getDayClass(d)"
          >
            <span v-if="d.amount !== 0" class="d-none d-md-block text-center small"></span>
          </div>
        </div>
        <div class="row mt-2">
          <div class="col-12 no-select">
            <p class="divided">
              <span class="font-2 text-muted">{{ days_to_show }} {{ $t('days_ago') }}</span>
              <span class="divider"></span>
              <span class="text-center font-2" :class="{ 'text-muted': service.online, 'text-danger': !service.online }">{{
                serviceTxt
              }}</span>
              <span class="divider"></span>
              <span class="font-2 text-muted">{{ $t('today') }}</span>
            </p>
          </div>
        </div>
        <div class="daily-failures small text-right text-dim">{{ hover_text }}</div>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useIntersectionObserver } from '@vueuse/core'
import Api from '@/API'

const props = defineProps({
  service: {
    type: Object,
    required: true,
  },
})

const failureData = ref([])
const hover_text = ref('')
const loaded = ref(false)
const visible = ref(false)
const days_to_show = 90

const outageSeverity = {
  minor: { start: 1, end: 30 },
  moderate: { start: 30, end: 120 },
  major: { start: 120, end: 240 },
  critical: { start: 240 },
}

const serviceTxt = computed(() => smallText(props.service))

const target = ref(null)

onMounted(() => {
  visibleChart()
})

function visibleChart() {
  if (!visible.value) {
    visible.value = true
    lastDaysFailures().then(() => (loaded.value = true))
  }
}

function mouseout() {
  hover_text.value = ''
}

function mouseover(e) {
  let txt = `${e.amount} Failures`
  if (e.amount === 0) {
    txt = `No Issues`
  }
  hover_text.value = `${e.date.toUTCString().replace(' 00:00:00 GMT', '')} - ${txt}`
}

function nowSubtract(seconds) {
  return new Date(Date.now() - seconds * 1000)
}

function beginningOf(unit, date) {
  const d = new Date(date)
  d.setUTCHours(0, 0, 0, 0)
  return d
}

function endOf(unit) {
  const d = new Date()
  d.setUTCHours(23, 59, 59, 999)
  return d
}

function toUnix(date) {
  return Math.floor(date.getTime() / 1000)
}

function parseISO(str) {
  return new Date(str)
}

function smallText(service) {
  if (service.online) {
    return 'Operational'
  }
  return 'Degraded'
}

async function lastDaysFailures() {
  const start = beginningOf('day', nowSubtract(86400 * days_to_show))
  const end = endOf('today')

  const failuresPromise = Api.service_failures_data(props.service.id, toUnix(start), toUnix(end), '24h', true)
  const hitsPromise = Api.service_hits(props.service.id, toUnix(start), toUnix(end), '24h', true)

  const [failuresDataResult, hitsDataResult] = await Promise.all([failuresPromise, hitsPromise])

  const mergedData = mergeData(failuresDataResult, hitsDataResult)

  mergedData.forEach((d) => {
    let date = new Date(d.timeframe)
    if (toUnix(date) * 1000 > Date.now()) {
      return
    }

    failureData.value.push({
      month: date.getUTCMonth() + 1,
      day: date.getUTCDate(),
      date: date,
      amount: d.amount,
      hits: d.hits || 0,
    })
  })

  failureData.value = failureData.value.slice(-days_to_show)
}

function mergeData(failuresDataArr, hitsDataArr) {
  const dataMap = new Map()

  hitsDataArr.forEach((d) => {
    dataMap.set(d.timeframe, {
      hits: d.amount,
      amount: 0,
      date: d.timeframe,
    })
  })

  failuresDataArr.forEach((d) => {
    let data = dataMap.get(d.timeframe) || {
      hits: 0,
      amount: 0,
      date: d.timeframe,
    }
    data.amount = d.amount
    dataMap.set(d.timeframe, data)
  })

  return Array.from(dataMap, ([date, data]) => ({
    ...data,
    timeframe: date,
  }))
}

function getDayClass(data) {
  if (data.amount === 0 && data.hits === 0) {
    return 'day-no-data'
  } else if (data.amount === 0 && data.hits > 0) {
    return 'day-success'
  } else {
    const severity = data.amount
    if (severity >= outageSeverity.minor.start && severity < outageSeverity.minor.end) {
      return 'day-minor-outage'
    } else if (severity >= outageSeverity.moderate.start && severity < outageSeverity.moderate.end) {
      return 'day-moderate-outage'
    } else if (severity >= outageSeverity.major.start && severity < outageSeverity.major.end) {
      return 'day-major-outage'
    } else if (severity >= outageSeverity.critical.start) {
      return 'day-critical-outage'
    }
  }
}
</script>
