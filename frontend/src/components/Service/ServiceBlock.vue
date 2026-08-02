<template>
  <div class="service-block-wrapper">
    <router-link :to="serviceLink" custom v-slot="{ navigate }">
      <div
        class="service-block"
        :class="service.online ? 'service-block-online' : 'service-block-offline'"
        @click="navigate"
      >
        <!-- Header -->
        <div class="service-block-header">
          <div class="service-block-title-row">
            <span class="service-block-name">{{ service.name }}</span>
            <span class="service-block-badge" :class="service.online ? 'badge-online' : 'badge-offline'">
              {{ service.online ? 'ONLINE' : 'OFFLINE' }}
            </span>
          </div>
          <ServiceTopStats :service="service" />
        </div>

        <!-- Chart -->
        <div v-show="!expanded" class="service-block-chart">
          <ServiceChart :service="service" :visible="visible" :chart_timeframe="chartTimeframe" />
        </div>

        <!-- Footer -->
        <div class="service-block-footer" :class="service.online ? 'footer-online' : 'footer-offline'">
          <div class="footer-controls">
            <div class="dropdown-wrapper">
              <button @click.stop.prevent="openMenu('timeframe')" class="control-btn">
                {{ timeframepick.text }}
                <font-awesome-icon icon="chevron-down" class="ms-1" size="xs" />
              </button>
              <div v-if="dropDownMenu" class="dropdown-menu-custom">
                <a
                  v-for="(tf, i) in timeframes"
                  :key="i"
                  @click.stop.prevent="changeTimeframe(tf)"
                  class="dropdown-item-custom"
                  :class="{ active: timeframepick === tf }"
                >{{ tf.text }}</a>
              </div>
            </div>

            <div class="dropdown-wrapper">
              <button @click.stop.prevent="openMenu('interval')" class="control-btn">
                {{ intervalpick.text }}
                <font-awesome-icon icon="chevron-down" class="ms-1" size="xs" />
              </button>
              <div v-if="intervalMenu" class="dropdown-menu-custom">
                <a
                  v-for="(intv, i) in intervals"
                  :key="i"
                  @click.stop.prevent="changeInterval(intv)"
                  class="dropdown-item-custom"
                  :class="{ active: intervalpick === intv, disabled: disabledInterval(intv) }"
                >{{ intv.text }}</a>
              </div>
            </div>

            <span class="footer-status">{{ smallText }}</span>
          </div>

          <button v-if="!expanded" @click.stop="goToService" class="view-btn">
            {{ $t('view') }}
            <font-awesome-icon icon="arrow-right" class="ms-1" />
          </button>
        </div>
      </div>
    </router-link>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { useMainStore } from '@/stores/main'
import ServiceChart from './ServiceChart.vue'
import ServiceTopStats from '@/components/Service/ServiceTopStats.vue'

const props = defineProps({
  service: {
    type: Object,
    required: true,
  },
})

const router = useRouter()
const store = useMainStore()

const expanded = ref(false)
const visible = ref(true)
const dropDownMenu = ref(false)
const intervalMenu = ref(false)
const interval_val = ref('60m')

function timeset(seconds) {
  return Math.floor((Date.now() - seconds * 1000) / 1000)
}

const timeframe_val = ref(timeset(259200))

const timeframes = [
  { value: timeset(1800), text: '30 Minutes', set: 1 },
  { value: timeset(3600), text: '1 Hour', set: 2 },
  { value: timeset(21600), text: '6 Hours', set: 3 },
  { value: timeset(43200), text: '12 Hours', set: 4 },
  { value: timeset(86400), text: '1 Day', set: 5 },
  { value: timeset(259200), text: '3 Days', set: 6 },
  { value: timeset(604800), text: '7 Days', set: 7 },
  { value: timeset(1209600), text: '14 Days', set: 8 },
  { value: timeset(2592000), text: '1 Month', set: 9 },
  { value: timeset(7776000), text: '3 Months', set: 10 },
  { value: 0, text: 'All Records' },
]

const intervals = [
  { value: '1m', text: '1/min', set: 1 },
  { value: '5m', text: '5/min', set: 2 },
  { value: '15m', text: '15/min', set: 3 },
  { value: '30m', text: '30/min', set: 4 },
  { value: '60m', text: '1/hr', set: 5 },
  { value: '180m', text: '3/hr', set: 6 },
  { value: '360m', text: '6/hr', set: 7 },
  { value: '720m', text: '12/hr', set: 8 },
  { value: '1440m', text: '1/day', set: 9 },
  { value: '4320m', text: '3/day', set: 10 },
  { value: '10080m', text: '7/day', set: 11 },
]

const serviceLink = computed(() => `/service/${props.service.permalink || props.service.id}`)

const timeframepick = computed(() => timeframes.find((s) => s.value === timeframe_val.value) || timeframes[5])

const intervalpick = computed(() => intervals.find((s) => s.value === interval_val.value) || intervals[4])

const chartTimeframe = computed(() => ({
  start_time: timeframe_val.value,
  interval: interval_val.value,
}))

const smallText = computed(() => (props.service.online ? 'Operational' : 'Degraded'))

function disabledInterval(interval) {
  const min = timeframepick.value.set - interval.set - 1
  return min >= interval.set
}

function openMenu(tm) {
  if (tm === 'interval') {
    intervalMenu.value = !intervalMenu.value
    dropDownMenu.value = false
  } else if (tm === 'timeframe') {
    dropDownMenu.value = !dropDownMenu.value
    intervalMenu.value = false
  }
}

function changeInterval(tm) {
  interval_val.value = tm.value
  intervalMenu.value = false
  dropDownMenu.value = false
}

function changeTimeframe(tm) {
  timeframe_val.value = tm.value
  dropDownMenu.value = false
  intervalMenu.value = false
}

function goToService() {
  store.setService(props.service)
  router.push(`/service/${props.service.id}`)
}
</script>

<style scoped>
.service-block-wrapper {
  margin-bottom: 1.5rem;
}

.service-block {
  background: #fff;
  border-radius: var(--radius-xl, 12px);
  box-shadow: var(--shadow-md, 0 4px 6px -1px rgba(0, 0, 0, 0.1));
  cursor: pointer;
  transition: all 0.2s ease;
  overflow: hidden;
}

.service-block:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-xl, 0 20px 25px -5px rgba(0, 0, 0, 0.1));
}

.service-block-online {
  border-left: 4px solid var(--color-success, #22c55e);
}

.service-block-offline {
  border-left: 4px solid var(--color-danger, #ef4444);
}

/* Header */
.service-block-header {
  padding: 1rem 1.25rem;
}

.service-block-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.75rem;
}

.service-block-name {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--color-gray-900, #111827);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 70%;
}

.service-block-badge {
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  padding: 0.25rem 0.6rem;
  border-radius: 9999px;
}

.badge-online {
  background: var(--color-success-bg, rgba(34, 197, 94, 0.1));
  color: var(--color-success-dark, #15803d);
}

.badge-offline {
  background: var(--color-danger-bg, rgba(239, 68, 68, 0.1));
  color: var(--color-danger-dark, #b91c1c);
}

/* Chart */
.service-block-chart {
  padding: 0 1rem;
  min-height: 180px;
}

/* Footer */
.service-block-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1rem;
  border-top: 1px solid var(--color-gray-100, #f3f4f6);
}

.footer-online {
  background: linear-gradient(135deg, #22c55e 0%, #16a34a 100%);
}

.footer-offline {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
}

.footer-controls {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.dropdown-wrapper {
  position: relative;
}

.control-btn {
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: #fff;
  font-size: 0.75rem;
  font-weight: 500;
  padding: 0.35rem 0.6rem;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s ease;
  display: flex;
  align-items: center;
}

.control-btn:hover {
  background: rgba(255, 255, 255, 0.3);
}

.dropdown-menu-custom {
  position: absolute;
  bottom: 100%;
  left: 0;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.15);
  min-width: 120px;
  padding: 0.5rem 0;
  margin-bottom: 0.5rem;
  z-index: 100;
}

.dropdown-item-custom {
  display: block;
  padding: 0.5rem 1rem;
  font-size: 0.8rem;
  color: #374151;
  text-decoration: none;
  cursor: pointer;
  transition: background 0.15s ease;
}

.dropdown-item-custom:hover {
  background: #f3f4f6;
}

.dropdown-item-custom.active {
  background: #eff6ff;
  color: #2563eb;
  font-weight: 500;
}

.dropdown-item-custom.disabled {
  color: #9ca3af;
  pointer-events: none;
}

.footer-status {
  color: rgba(255, 255, 255, 0.9);
  font-size: 0.8rem;
  font-weight: 500;
  margin-left: 0.5rem;
}

.view-btn {
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: #fff;
  font-size: 0.8rem;
  font-weight: 500;
  padding: 0.4rem 0.8rem;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s ease;
  display: flex;
  align-items: center;
}

.view-btn:hover {
  background: rgba(255, 255, 255, 0.3);
}

@media (max-width: 768px) {
  .footer-status {
    display: none;
  }
}
</style>
