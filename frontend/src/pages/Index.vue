<template>
  <div>
    <TopNav />
    <Header />
    <div class="container-fluid px-md-5 sm-container">

    <div v-if="loaded && groups.length === 0 && services.length === 0" class="row mt-5 mb-5">
      <div class="col-12 mt-5 mb-4 text-center">
        <div class="card shadow-sm border-0 py-5">
          <div class="card-body">
            <div class="mb-4">
              <font-awesome-icon icon="info-circle" class="text-muted" size="4x" />
            </div>
            <h3 class="font-weight-bold mb-3">No Services Monitored</h3>
            <p class="text-muted mb-4">
              There is currently no data available to display. Please access the Dashboard to create and manage your
              services.
            </p>
            <router-link to="/dashboard" class="btn btn-primary px-5 shadow-sm"> Access Dashboard </router-link>
          </div>
        </div>
      </div>
    </div>

    <div v-if="!loaded" class="row mt-5 mb-5">
      <div class="col-12 mt-5 mb-2 text-center">
        <font-awesome-icon icon="circle-notch" class="text-dim" size="2x" spin />
      </div>
      <div class="col-12 text-center mt-3 mb-3">
        <span class="text-dim">{{ loadingText }}</span>
      </div>
    </div>

    <div class="col-12 full-col-12">
      <MessageBlock v-for="message in visibleMessages" :key="message.id" :message="message" />
    </div>

    <div class="col-12 full-col-12">
      <div v-for="service in servicesNoGroup" :key="service.id" class="list-group online_list mb-4">
        <router-link
          :to="serviceLink(service)"
          custom
          v-slot="{ navigate }"
        >
          <div class="list-group-item list-group-item-action" style="cursor: pointer" @click="navigate">
            <span class="no-decoration font-3 text-dark font-weight-bold">
              {{ service.name }}
              <MessagesIcon :messages="service.messages" />
            </span>
            <span class="badge float-right" :class="{ 'bg-success': service.online, 'bg-danger': !service.online }">{{
              service.online ? 'ONLINE' : 'OFFLINE'
            }}</span>
            <GroupServiceFailures :service="service" />
            <IncidentsBlock :service="service" />
          </div>
        </router-link>
      </div>
    </div>

    <Group v-for="group in groups" :key="group.id" :group="group" />

    <div class="row">
      <div v-for="service in services" :ref="(el) => setServiceRef(service.id, el)" :key="service.id" class="col-lg-4 col-md-6 col-12">
        <ServiceBlock :service="service" />
      </div>
    </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useMainStore } from '@/stores/main'
import { useCookies } from 'vue3-cookies'
import Api from '@/API'

import Group from '@/components/Index/Group.vue'
import Header from '@/components/Index/Header.vue'
import TopNav from '@/components/Index/TopNav.vue'
import MessageBlock from '@/components/Index/MessageBlock.vue'
import ServiceBlock from '@/components/Service/ServiceBlock.vue'
import GroupServiceFailures from '@/components/Index/GroupServiceFailures.vue'
import IncidentsBlock from '@/components/Index/IncidentsBlock.vue'
import MessagesIcon from '@/components/Index/MessagesIcon.vue'

const store = useMainStore()
const { cookies } = useCookies()

const serviceRefs = ref({})

const loaded = computed(() => store.hasPublicData)
const core = computed(() => store.core)
const messages = computed(() => store.messages)
const groups = computed(() => store.groupsInOrder)
const services = computed(() => store.servicesInOrder)
const servicesNoGroup = computed(() => store.servicesNoGroup)

const loadingText = computed(() => {
  if (store.groups.length === 0) {
    return 'Loading Groups'
  } else if (store.services.length === 0) {
    return 'Loading Services'
  } else if (store.messages == null) {
    return 'Loading Announcements'
  }
  return 'Loading...'
})

const visibleMessages = computed(() => {
  return messages.value.filter((m) => inRange(m) && m.service === 0)
})

function setServiceRef(id, el) {
  if (el) {
    serviceRefs.value[id] = el
  }
}

function serviceLink(service) {
  return `/service/${service.permalink || service.id}`
}

function now() {
  return new Date()
}

function maxDate() {
  return new Date(8640000000000000)
}

function isBetween(date, start, end) {
  const d = new Date(date)
  const s = new Date(start)
  const e = new Date(end)
  return d >= s && d <= e
}

function inRange(message) {
  return isBetween(
    now(),
    message.start_on,
    message.start_on === message.end_on ? maxDate().toISOString() : message.end_on
  )
}

async function checkLogin() {
  const token = cookies.get('statping_auth')
  if (!token) {
    store.setLoggedIn(false)
    return
  }
  try {
    const jwt = await Api.check_token(token)
    store.setAdmin(jwt.admin)
    if (jwt.username) {
      store.setLoggedIn(true)
    }
  } catch (e) {
    console.error(e)
  }
}
</script>
