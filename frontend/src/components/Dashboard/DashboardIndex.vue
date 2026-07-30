<template>
  <div class="col-12 mt-4 mt-md-3">
    <div class="row stats_area mb-5">
      <div class="col-4">
        <span class="font-6 font-weight-bold d-block">{{ services.length }}</span>
        <span class="font-2">{{ $t('total_services') }}</span>
      </div>
      <div class="col-4">
        <span class="font-6 font-weight-bold d-block">{{ failuresLast24Hours }}</span>
        <span class="font-2">{{ $t('failures_24_hours') }}</span>
      </div>
      <div class="col-4">
        <span class="font-6 font-weight-bold d-block">{{ onlineServicesCount }}</span>
        <span class="font-2">{{ $t('online_services') }}</span>
      </div>
    </div>

    <div class="col-12" v-if="services.length === 0">
      <div class="alert alert-dark d-block">
        {{ $t('no_services') }}
        <router-link v-if="store.admin" to="/dashboard/create_service" class="btn btn-sm btn-success float-right">
          <font-awesome-icon icon="plus" /> {{ $t('create') }}
        </router-link>
      </div>
    </div>

    <div v-for="message in messagesInRange" :key="message.id" class="bg-light shadow-sm p-3 pr-4 pl-4 col-12 mb-4">
      <font-awesome-icon icon="calendar" class="mr-3" size="1x" /> {{ message.description }}
      <span class="d-block small text-muted mt-3">
        Starts at <strong>{{ niceDate(message.start_on) }}</strong> till <strong>{{ niceDate(message.end_on) }}</strong>
        ({{ duration(message.start_on, message.end_on) }})
      </span>
    </div>

    <div class="row">
      <div v-for="service in servicesNoGroup" :key="service.id" class="col-lg-4 col-md-6 col-12">
        <ServiceInfo :service="service" />
      </div>
    </div>

    <div v-for="group in groups" :key="group.id">
      <GroupedServices :group="group" />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useMainStore } from '@/stores/main'
import GroupedServices from '@/components/Dashboard/GroupedServices.vue'
import ServiceInfo from '@/components/Dashboard/ServiceInfo.vue'

const store = useMainStore()

const services = computed(() => store.services)
const servicesNoGroup = computed(() => store.servicesNoGroup)
const groups = computed(() => store.groupsInOrder)
const onlineServicesCount = computed(() => store.onlineServices(true).length)

const failuresLast24Hours = computed(() => {
  let total = 0
  services.value.forEach((s) => {
    total += s.failures_24_hours || 0
  })
  return total
})

const messagesInRange = computed(() => {
  const now = new Date()
  return store.globalMessages.filter((m) => {
    const start = new Date(m.start_on)
    const end = new Date(m.end_on)
    return now >= start && now <= end
  })
})

function niceDate(date) {
  return new Date(date).toLocaleDateString()
}

function duration(start, end) {
  const ms = new Date(end) - new Date(start)
  const hours = Math.floor(ms / (1000 * 60 * 60))
  if (hours < 24) return `${hours} hours`
  const days = Math.floor(hours / 24)
  return `${days} days`
}
</script>

<style scoped></style>
