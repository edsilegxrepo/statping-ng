<template>
  <div v-if="services.length > 0" class="col-12 full-col-12">
    <h4 v-if="group.name !== 'Empty Group'" class="group_header mb-3 mt-4">{{ group.name }}</h4>
    <div class="list-group online_list mb-4">
      <router-link
        v-for="(service, index) in services"
        :key="index"
        :to="serviceLink(service)"
        custom
        v-slot="{ navigate }"
      >
        <div class="list-group-item list-group-item-action" style="cursor: pointer" @click="navigate">
          <span class="no-decoration font-3 text-dark font-weight-bold">
            {{ service.name }}
            <MessagesIcon :messages="service.messages" />
          </span>
          <span
            class="badge text-uppercase float-right"
            :class="{ 'bg-success': service.online, 'bg-danger': !service.online }"
          >
            {{ service.online ? $t('online') : $t('offline') }}
          </span>

          <GroupServiceFailures :service="service" />
          <IncidentsBlock :service="service" />
        </div>
      </router-link>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useMainStore } from '@/stores/main'
import GroupServiceFailures from './GroupServiceFailures.vue'
import IncidentsBlock from './IncidentsBlock.vue'
import MessagesIcon from '@/components/Index/MessagesIcon.vue'

const props = defineProps({
  group: {
    type: Object,
    required: true,
  },
})

const store = useMainStore()
const services = computed(() => store.servicesInGroup(props.group.id))

function serviceLink(service) {
  return `/service/${service.permalink || service.id}`
}
</script>
