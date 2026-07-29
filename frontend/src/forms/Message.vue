<template>
  <div>
    <div class="card contain-card mb-5">
      <div class="card-header">
        {{ message.id ? `${$t('update')} ${message.title}` : $t('message_create') }}
        <transition name="slide-fade">
          <button @click="removeEdit" v-if="message.id" class="btn btn-sm float-right btn-danger btn-sm">
            {{ $t('close') }}
          </button>
        </transition>
      </div>
      <div class="card-body">
        <form @submit.prevent="saveMessage">
          <div class="form-group row">
            <label class="col-sm-4 col-form-label">{{ $t('title') }}</label>
            <div class="col-sm-8">
              <input
                v-model="message.title"
                type="text"
                name="title"
                class="form-control"
                id="title"
                placeholder="Announcement Title"
                required
              />
            </div>
          </div>

          <div class="form-group row">
            <label class="col-sm-4 col-form-label">{{ $t('description') }}</label>
            <div class="col-sm-8">
              <textarea
                v-model="message.description"
                rows="5"
                name="description"
                class="form-control"
                id="description"
                required
              ></textarea>
            </div>
          </div>

          <div class="form-group row">
            <label class="col-sm-4 col-form-label">{{ $t('service') }}</label>
            <div class="col-sm-8">
              <select v-model.number="message.service" name="service_id" class="form-control">
                <option :value="0">{{ $t('global_announcement') }}</option>
                <option v-for="service in services" :value="service.id" :key="service.id">{{ service.name }}</option>
              </select>
            </div>
          </div>

          <div class="form-group row">
            <label class="col-sm-4 col-form-label">{{ $t('announcement_date') }}</label>
            <div class="col-sm-4">
              <flat-pickr
                v-model="message.start_on"
                :config="config"
                type="text"
                name="start_on"
                class="form-control form-control-plaintext"
                id="start_on"
                required
              />
            </div>
            <div class="col-sm-4 mt-3 mt-md-0">
              <flat-pickr
                v-model="message.end_on"
                :config="config"
                type="text"
                name="end_on"
                class="form-control form-control-plaintext"
                id="end_on"
                required
              />
            </div>
          </div>

          <div class="form-group row">
            <label for="notify_method" class="col-sm-4 col-form-label">{{ $t('notify_users') }}</label>
            <div class="col-sm-8">
              <span @click="message.notify = !!message.notify" class="switch">
                <input v-model="message.notify" type="checkbox" class="switch" id="switch-normal" />
                <label for="switch-normal">{{ $t('notify_desc') }}</label>
              </span>
            </div>
          </div>

          <div v-if="message.notify" class="form-group row">
            <label for="notify_method" class="col-sm-4 col-form-label">{{ $t('notify_method') }}</label>
            <div class="col-sm-8">
              <input
                v-model="message.notify_method"
                type="text"
                name="notify_method"
                class="form-control"
                id="notify_method"
                placeholder="email"
              />
            </div>
          </div>

          <div v-if="message.notify" class="form-group row">
            <label for="notify_before" class="col-sm-4 col-form-label">{{ $t('notify_before') }}</label>
            <div class="col-sm-8">
              <div class="form-inline">
                <input
                  v-model.number="message.notify_before"
                  type="number"
                  name="notify_before"
                  class="col-4 form-control"
                  id="notify_before"
                />
                <select v-model="message.notify_before_scale" class="ml-2 col-7 form-control" name="notify_before_scale" id="notify_before_scale">
                  <option value="minute">{{ $t('minutes') }}</option>
                  <option value="hour">{{ $t('hours') }}</option>
                  <option value="day">{{ $t('days') }}</option>
                </select>
              </div>
            </div>
          </div>

          <div class="form-group row">
            <div class="col-sm-12">
              <button
                @click.prevent="saveMessage"
                :disabled="!message.title || !message.description"
                type="submit"
                class="btn btn-block"
                :class="{ 'btn-primary': !message.id, 'btn-secondary': message.id }"
              >
                {{ message.id ? $t('message_edit') : $t('message_create') }}
              </button>
            </div>
          </div>
          <div class="alert alert-danger d-none" id="alerter" role="alert"></div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useMainStore } from '@/stores/main'
import flatPickr from 'vue-flatpickr-component'
import 'flatpickr/dist/flatpickr.css'
import Api from '@/API'

const props = defineProps({
  in_message: {
    type: Object,
    default: () => ({}),
  },
  service: {
    type: Object,
    default: null,
  },
  edit: {
    type: Function,
    default: () => {},
  },
})

const store = useMainStore()

const message = ref({
  title: '',
  description: '',
  start_on: new Date(),
  end_on: new Date(),
  service_id: 0,
  service: 0,
  notify_method: '',
  notify: false,
  notify_before: 0,
  notify_before_scale: 'minute',
})

const config = {
  altFormat: 'l M J, \\at h:iK',
  altInput: true,
  enableTime: true,
  dateFormat: 'Z',
}

const services = computed(() => store.services)

watch(
  () => props.in_message,
  (val) => {
    if (val && Object.keys(val).length > 0) {
      message.value = { ...val }
    }
  },
  { immediate: true }
)

onMounted(() => {
  if (props.service) {
    message.value.service = props.service.id
  }
})

function removeEdit() {
  message.value = {
    title: '',
    description: '',
    start_on: new Date(),
    end_on: new Date(),
    service_id: 0,
    service: 0,
    notify_method: '',
    notify: false,
    notify_before: 0,
    notify_before_scale: 'minute',
  }
  props.edit(false)
}

async function saveMessage() {
  if (message.value.id) {
    await updateMessage()
  } else {
    await createMessage()
  }
}

async function createMessage() {
  await Api.message_create(message.value)
  const messages = await Api.messages()
  store.setMessages(messages)
  message.value = {
    title: '',
    description: '',
    start_on: new Date(),
    end_on: new Date(),
    service_id: 0,
    service: 0,
    notify_method: '',
    notify: false,
    notify_before: 0,
    notify_before_scale: 'minute',
  }
}

async function updateMessage() {
  await Api.message_update(message.value)
  const messages = await Api.messages()
  store.setMessages(messages)
  props.edit(false)
}
</script>

<style scoped></style>
