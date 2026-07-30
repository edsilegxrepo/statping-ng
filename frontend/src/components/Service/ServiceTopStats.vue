<template>
  <div class="row stats_area mt-0 mb-0">
    <div class="col-3" v-if="showResponseTime">
      <span class="font-3 d-block font-weight-bold" :style="{ color: statusColor }">{{ humanTime(service.avg_response) }}</span>
      <span class="font-08 subtitle">Average Response</span>
    </div>
    <div :class="showResponseTime ? 'col-3' : 'col-4'">
      <span class="font-3 d-block font-weight-bold" :style="{ color: statusColor }">{{ service.online_24_hours }}%</span>
      <span class="font-08 subtitle">24h</span>
    </div>
    <div :class="showResponseTime ? 'col-3' : 'col-4'">
      <span class="font-3 d-block font-weight-bold" :style="{ color: statusColor }">{{ service.online_7_days }}%</span>
      <span class="font-08 subtitle">7d</span>
    </div>
    <div :class="showResponseTime ? 'col-3' : 'col-4'">
      <span class="font-3 d-block font-weight-bold" :style="{ color: statusColor }">{{ service.online_1_year }}%</span>
      <span class="font-08 subtitle">12m</span>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  service: {
    type: Object,
    required: true,
  },
})

const showResponseTime = computed(() => props.service.avg_response !== 0 || props.service.type !== 'static')
const statusColor = computed(() => props.service.online ? '#28a745' : '#dc3545')

function humanTime(ms) {
  if (ms < 1000) {
    return `${ms}ms`
  }
  return `${(ms / 1000).toFixed(2)}s`
}
</script>

<style scoped></style>
