<template>
  <div class="row p-2">
    <div v-if="loaded && lastFailure && failureBefore" class="col-12 text-danger font-2 m-0 mb-2">
      <font-awesome-icon icon="exclamation" class="mr-1 text-danger font-weight-bold" size="1x" /> Recent Failure<br />
      <span class="font-italic font-weight-light text-dim mt-1" style="max-width: 270px">
        Last failure was {{ timeAgo(lastFailure.created_at) }} ago. {{ lastFailure.issue }}
      </span>
    </div>

    <div v-if="loaded" v-for="message in messages" :key="message.id" class="col-12 font-2 m-0 mb-2">
      <font-awesome-icon icon="calendar" class="mr-1" size="1x" /> Upcoming Announcement<br />
      <span class="font-italic font-weight-light text-dim mt-1">{{ message.description }}</span>
      <span class="font-0 text-dim float-right font-weight-light mt-1"
        >@ <strong>{{ niceDate(message.start_on) }}</strong>
      </span>
    </div>

    <div v-if="loaded" v-for="incident in incidents" :key="incident.id" class="col-12 font-2 m-0 mb-2">
      <font-awesome-icon icon="bullhorn" class="mr-1" size="1x" />Recent Incident<br />
      <span class="font-italic font-weight-light text-dim mt-1" style="max-width: 270px"
        >{{ incident.title }} - {{ incident.description }}</span
      >
      <span class="font-0 text-dim float-right font-weight-light mt-1"
        >@ <strong>{{ niceDate(incident.created_at) }}</strong></span
      >
    </div>

    <div v-if="successEvent && !failureBefore" class="col-12 font-2 m-0 mb-2">
      <span class="text-success"><font-awesome-icon icon="check" class="mr-1" size="1x" />No New Events</span>
      <span
        v-if="service.last_error && !isZeroDate(service.last_error)"
        class="font-italic d-inline-block text-truncate text-dim mt-1"
        style="max-width: 270px"
      >
        Last failure was {{ timeAgo(service.last_error) }} ago.
      </span>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useMainStore } from '@/stores/main'
import Api from '@/API'

const props = defineProps({
  service: {
    type: Object,
    required: true,
  },
})

const store = useMainStore()
const incidents = ref(null)
const loaded = ref(false)

const lastFailure = computed(() => {
  if (!props.service.failures || !props.service.failures.length) {
    return null
  }
  return props.service.failures[0]
})

const failureBefore = computed(() => {
  if (!props.service.last_error) return false
  const lastError = new Date(props.service.last_error)
  const threshold = new Date(Date.now() - 43200 * 1000)
  return lastError > threshold
})

const messages = computed(() => store.serviceMessages(props.service.id))

const successEvent = computed(() => {
  if (
    props.service.online &&
    (!props.service.messages || props.service.messages.length === 0) &&
    (!props.service.incidents || props.service.incidents.length === 0)
  ) {
    return true
  }
  return false
})

onMounted(() => {
  load()
})

async function load() {
  loaded.value = false
  await getIncidents()
  loaded.value = true
}

async function getIncidents() {
  try {
    incidents.value = await Api.incidents_service(props.service.id)
  } catch (e) {
    incidents.value = []
  }
}

function niceDate(date) {
  return new Date(date).toLocaleDateString()
}

function timeAgo(date) {
  const seconds = Math.floor((new Date() - new Date(date)) / 1000)
  if (seconds < 60) return `${seconds} seconds`
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes} minutes`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours} hours`
  const days = Math.floor(hours / 24)
  return `${days} days`
}

function isZeroDate(date) {
  if (!date) return true
  const d = new Date(date)
  return d.getFullYear() < 2000
}
</script>
