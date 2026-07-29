<template>
  <div>
    <div v-for="checkin in checkins" :key="checkin.id" class="col-12 alert alert-light" role="alert">
      <span class="badge badge-pill badge-info text-uppercase">{{ checkin.name }}</span>
      <span class="float-right font-2">Last checkin {{ ago(checkin.last_hit) }}</span>
      <span class="float-right font-2 mr-3">Check Every {{ checkin.interval }} seconds</span>
      <span class="float-right font-2 mr-3">Grace Period {{ checkin.grace }} seconds</span>
      <span class="d-block mt-2">
        <input type="text" class="form-control" :value="`${coreData.domain}/checkin/${checkin.api_key}`" readonly />
        <span class="small"
          >Send a GET request to this URL every {{ checkin.interval }} seconds
          <button @click.prevent="deleteCheckin(checkin)" type="button" class="btn btn-danger btn-xs float-right mt-1">
            Delete
          </button>
        </span>
      </span>
    </div>

    <div class="col-12 alert alert-light">
      <form @submit.prevent="saveCheckin">
        <div class="form-group row">
          <div class="col-12 col-md-5">
            <label for="checkin_interval" class="col-form-label">Checkin Name</label>
            <input
              v-model="checkin.name"
              type="text"
              name="name"
              class="form-control"
              id="checkin_name"
              placeholder="New Checkin"
            />
          </div>
          <div class="col-12 col-md-5">
            <label for="checkin_interval" class="col-form-label">Interval (minutes)</label>
            <input
              v-model.number="checkin.interval"
              type="number"
              name="interval"
              class="form-control"
              id="checkin_interval"
              placeholder="1"
              min="1"
            />
          </div>
          <div class="col-12 col-md-5">
            <label class="col-form-label"></label>
            <button @click.prevent="saveCheckin" type="submit" id="submit" class="btn btn-success d-block mt-2">
              Save Checkin
            </button>
          </div>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { reactive, computed } from 'vue'
import { useMainStore } from '@/stores/main'
import Api from '@/API'

const props = defineProps({
  service: {
    type: Object,
    required: true,
  },
})

const store = useMainStore()

const checkin = reactive({
  name: '',
  interval: 60,
  service_id: props.service.id,
})

const checkins = computed(() => store.serviceCheckins(props.service.id))
const coreData = computed(() => store.core)

function ago(date) {
  if (!date) return 'never'
  const seconds = Math.floor((new Date() - new Date(date)) / 1000)
  if (seconds < 60) return `${seconds} seconds ago`
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes} minutes ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours} hours ago`
  const days = Math.floor(hours / 24)
  return `${days} days ago`
}

async function saveCheckin() {
  checkin.interval = parseInt(checkin.interval, 10)
  checkin.grace = parseInt(checkin.grace || 0, 10)
  await Api.checkin_create(checkin)
  await updateCheckins()
}

async function deleteCheckin(chk) {
  await Api.checkin_delete(chk)
  await updateCheckins()
}

async function updateCheckins() {
  const chks = await Api.checkins()
  store.setCheckins(chks)
}
</script>

<style scoped></style>
