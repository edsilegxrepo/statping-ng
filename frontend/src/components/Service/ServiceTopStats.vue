<template>
  <div class="service-stats">
    <div class="stat-item" v-if="showResponseTime">
      <span class="stat-value" :class="statusClass">{{ humanTime(service.avg_response) }}</span>
      <span class="stat-label">Average Response</span>
    </div>
    <div class="stat-item">
      <span class="stat-value" :class="statusClass">{{ service.online_24_hours }}%</span>
      <span class="stat-label">24h</span>
    </div>
    <div class="stat-item">
      <span class="stat-value" :class="statusClass">{{ service.online_7_days }}%</span>
      <span class="stat-label">7d</span>
    </div>
    <div class="stat-item">
      <span class="stat-value" :class="statusClass">{{ service.online_1_year }}%</span>
      <span class="stat-label">12m</span>
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
const statusClass = computed(() => props.service.online ? 'stat-online' : 'stat-offline')

function humanTime(ms) {
  if (ms < 1000) {
    return `${ms}ms`
  }
  return `${(ms / 1000).toFixed(2)}s`
}
</script>

<style scoped>
.service-stats {
  display: flex;
  gap: 1rem;
  margin-top: 1rem;
}

.stat-item {
  flex: 1;
  text-align: center;
  padding: 1rem 1.5rem;
  background: #fff;
  border-radius: var(--radius-lg, 0.5rem);
  box-shadow: var(--shadow-sm, 0 1px 2px rgba(0,0,0,0.05));
  border: 1px solid var(--color-gray-200, #e5e7eb);
}

.stat-value {
  display: block;
  font-size: 1.5rem;
  font-weight: 700;
  line-height: 1.2;
}

.stat-online {
  color: var(--color-success, #22c55e);
}

.stat-offline {
  color: var(--color-danger, #ef4444);
}

.stat-label {
  display: block;
  font-size: 0.75rem;
  color: var(--color-gray-500, #6b7280);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-top: 0.25rem;
}

@media (max-width: 768px) {
  .service-stats {
    flex-wrap: wrap;
  }

  .stat-item {
    flex: 1 1 45%;
    min-width: 120px;
  }
}
</style>
