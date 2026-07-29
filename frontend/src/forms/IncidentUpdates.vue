<template>
  <div class="card-body pt-3">
    <div v-if="updates.length === 0" class="alert alert-link text-danger">
      No updates found, create a new Incident Update below.
    </div>

    <div v-for="update in updates" :key="update.id">
      <IncidentUpdate :update="update" :onUpdate="loadUpdates" :admin="true" />
    </div>

    <form class="row" @submit.prevent="createIncidentUpdate">
      <div class="col-12 col-md-3 mb-3 mb-md-0">
        <select v-model="incident_update.type" class="form-control">
          <option value="Investigating">Investigating</option>
          <option value="Update">Update</option>
          <option value="Unknown">Unknown</option>
          <option value="Resolved">Resolved</option>
        </select>
      </div>
      <div class="col-12 col-md-7 mb-3 mb-md-0">
        <input v-model="incident_update.message" name="description" class="form-control" id="message" required />
      </div>

      <div class="col-12 col-md-2">
        <button :disabled="submitting || !canSubmit" type="submit" class="btn btn-block btn-primary">
          {{ submitting ? 'Adding...' : 'Add' }}
        </button>
      </div>
    </form>

    <div v-if="errorMessage" class="alert alert-danger mt-3 mb-0" role="alert">{{ errorMessage }}</div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import Api from '@/API'
import IncidentUpdate from '@/components/Elements/IncidentUpdate.vue'

const props = defineProps({
  incident: {
    type: Object,
    required: true,
  },
})

const updates = ref([])
const submitting = ref(false)
const errorMessage = ref('')
const incident_update = reactive({
  incident: props.incident.id,
  message: '',
  type: 'Investigating',
})

const canSubmit = computed(() => {
  return incident_update.message.trim().length > 0
})

onMounted(async () => {
  await loadUpdates()
})

function extractErrorMessage(error, fallback) {
  const responseData = error?.response?.data || error
  if (typeof responseData === 'string' && responseData.trim()) {
    return responseData.trim()
  }
  if (typeof responseData?.error === 'string' && responseData.error.trim()) {
    return responseData.error
  }
  if (responseData?.error?.message) {
    return responseData.error.message
  }
  if (responseData?.message) {
    return responseData.message
  }
  if (error?.message) {
    return error.message
  }
  return fallback
}

async function createIncidentUpdate() {
  if (submitting.value) return

  const message = incident_update.message.trim()
  if (!message) {
    errorMessage.value = 'Incident update message is required.'
    return
  }

  submitting.value = true
  errorMessage.value = ''

  try {
    const response = await Api.incident_update_create({
      ...incident_update,
      message,
    })

    if (response?.status === 'success' && response.output) {
      updates.value.push(response.output)
      incident_update.message = ''
      incident_update.type = 'Investigating'
      return
    }

    errorMessage.value = extractErrorMessage(response, 'Unable to add the incident update right now.')
  } catch (error) {
    errorMessage.value = extractErrorMessage(error, 'Unable to add the incident update right now.')
  } finally {
    submitting.value = false
  }
}

async function loadUpdates() {
  updates.value = await Api.incident_updates(props.incident)
}
</script>

<style scoped></style>
