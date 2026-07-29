<template>
  <div class="mb-md-4 mb-4">
    <router-link :to="serviceLink" custom v-slot="{ navigate }">
      <div class="card index-chart" :class="{ 'expanded-service': expanded }" style="cursor: pointer" @click="navigate">
        <div class="card-body">
          <div class="col-12">
            <h4 class="mt-2">
              <span class="d-inline-block text-truncate font-4 text-dark font-weight-bold" style="max-width: 65vw">{{
                service.name
              }}</span>
              <span class="badge float-right" :class="{ 'bg-success': service.online, 'bg-danger': !service.online }">{{
                service.online ? 'ONLINE' : 'OFFLINE'
              }}</span>
            </h4>

            <ServiceTopStats :service="service" />
          </div>
        </div>

        <div v-show="!expanded" class="chart-container">
          <ServiceChart :service="service" :visible="visible" :chart_timeframe="chartTimeframe" />
        </div>

        <div
          class="row lower_canvas full-col-12 text-white"
          :class="{ 'bg-success': service.online, 'bg-danger': !service.online }"
        >
          <div class="col-md-10 col-6">
            <div class="dropup" :class="{ show: dropDownMenu }">
              <button
                style="font-size: 10pt"
                @click.stop.prevent="openMenu('timeframe')"
                type="button"
                class="col-4 float-left btn btn-sm float-right btn-block text-white dropdown-toggle service_scale pr-2"
              >
                {{ timeframepick.text }}
              </button>
              <div class="service-tm-menu" :class="{ 'd-none': !dropDownMenu }">
                <a
                  v-for="(tf, i) in timeframes"
                  :key="i"
                  @click.stop.prevent="changeTimeframe(tf)"
                  class="dropdown-item"
                  href="#"
                  :class="{ active: timeframepick === tf }"
                  >{{ tf.text }}</a
                >
              </div>
            </div>

            <div class="dropup" :class="{ show: intervalMenu }">
              <button
                style="font-size: 10pt"
                @click.stop.prevent="openMenu('interval')"
                type="button"
                class="col-4 float-left btn btn-sm float-right btn-block text-white dropdown-toggle service_scale pr-2"
              >
                {{ intervalpick.text }}
              </button>
              <div class="service-tm-menu" :class="{ 'd-none': !intervalMenu }">
                <a
                  v-for="(intv, i) in intervals"
                  :key="i"
                  @click.stop.prevent="changeInterval(intv)"
                  class="dropdown-item"
                  href="#"
                  :class="{ active: intervalpick === intv, disabled: disabledInterval(intv) }"
                >
                  {{ intv.text }}
                </a>
              </div>

              <span class="d-none float-left d-md-inline">
                {{ smallText }}
              </span>
            </div>
          </div>

          <div class="col-md-2 col-6 float-right">
            <button
              v-if="!expanded"
              @click.stop="goToService"
              class="btn btn-sm float-right dyn-dark text-white"
              :class="{ 'bg-success': service.online, 'bg-danger': !service.online }"
            >
              {{ $t('view') }}
            </button>
          </div>
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
