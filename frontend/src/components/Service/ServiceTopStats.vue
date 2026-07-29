<template>
  <div class="row stats_area mt-5 mb-4">
    <div class="col-3" v-if="showResponseTime">
      <span class="font-5 d-block font-weight-bold">{{ humanTime(service.avg_response) }}</span>
      <span class="font-1 subtitle">{{ $t('average_response') }}</span>
    </div>
    <div :class="showResponseTime ? 'col-3' : 'col-6'">
      <span class="font-5 d-block font-weight-bold">{{ service.online_24_hours }} %</span>
      <span class="font-1 subtitle">{{ $t('last_uptime') }} 24 {{ $t('hour', 24) }}</span>
    </div>
    <div :class="showResponseTime ? 'col-3' : 'col-6'">
      <span class="font-5 d-block font-weight-bold">{{ service.online_7_days }} %</span>
      <span class="font-1 subtitle">{{ $t('last_uptime') }} 7 {{ $t('day', 7) }}</span>
    </div>
    <div :class="showResponseTime ? 'col-3' : 'col-6'">
      <span class="font-5 d-block font-weight-bold">{{ service.online_1_year }} %</span>
      <span class="font-1 subtitle">{{ $t('last_uptime') }} 12 {{ $t('month', 12) }}</span>
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

function humanTime(ms) {
  if (ms < 1000) {
    return `${ms}ms`
  }
  return `${(ms / 1000).toFixed(2)}s`
}
</script>

<style scoped></style>
