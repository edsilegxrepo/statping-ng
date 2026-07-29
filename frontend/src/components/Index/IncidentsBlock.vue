<template>
  <div class="row">
    <div v-for="(incident, i) in incidents" :key="i" class="col-12 mt-2">
      <span class="braker mt-1 mb-3"></span>
      <h6>
        {{ incident.title }}
        <span class="font-2 float-right">{{ niceDate(incident.created_at) }}</span>
      </h6>
      <div class="font-2 mb-3" v-html="sanitizedDescription(incident.description)"></div>
      <IncidentUpdate v-for="(update, j) in incident.updates" :key="j" :update="update" :admin="false" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import DOMPurify from 'dompurify'
import IncidentUpdate from '@/components/Elements/IncidentUpdate.vue'
import Api from '@/API'

const props = defineProps({
  service: {
    type: Object,
    required: true,
  },
})

const incidents = ref(null)

onMounted(() => {
  getIncidents()
})

function niceDate(date) {
  return new Date(date).toLocaleDateString()
}

function sanitizedDescription(desc) {
  return DOMPurify.sanitize(desc || '')
}

async function getIncidents() {
  incidents.value = await Api.incidents_service(props.service.id)
}
</script>

<style scoped></style>
