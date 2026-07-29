<template>
  <div class="col-12">
    <div class="card contain-card mb-4">
      <div class="card-header">{{ $t('announcements') }}</div>
      <div class="card-body pt-0">
        <div v-if="messages.length === 0">
          <div class="alert alert-dark d-block mt-3 mb-0">
            You currently don't have any Announcements! Create one using the form below.
          </div>
        </div>

        <table v-else class="table table-striped">
          <thead>
            <tr>
              <th scope="col">{{ $t('title') }}</th>
              <th scope="col" class="d-none d-md-table-cell">{{ $t('service', 1) }}</th>
              <th scope="col" class="d-none d-md-table-cell">{{ $t('begins') }}</th>
              <th scope="col"></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="msg in messages" :key="msg.id">
              <td>{{ msg.title }}</td>
              <td class="d-none d-md-table-cell">
                <router-link :to="serviceLink(getService(msg.service))">{{
                  serviceName(getService(msg.service))
                }}</router-link>
              </td>
              <td class="d-none d-md-table-cell">{{ niceDate(msg.start_on) }}</td>
              <td class="text-right">
                <div v-if="store.admin" class="btn-group">
                  <button @click.prevent="editMessage(msg)" class="btn btn-sm btn-outline-secondary">
                    <font-awesome-icon icon="edit" />
                  </button>
                  <button @click.prevent="deleteMessage(msg)" class="btn btn-sm btn-danger">
                    <font-awesome-icon icon="times" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
    <FormMessage v-if="store.admin" :edit="editChange" :in_message="message" />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useMainStore } from '@/stores/main'
import Api from '@/API'
import FormMessage from '@/forms/Message.vue'

const store = useMainStore()

const edit = ref(false)
const message = ref({})

const messages = computed(() => store.messages)

function niceDate(date) {
  if (!date) return ''
  return new Date(date).toLocaleDateString()
}

function editChange(v) {
  message.value = {}
  edit.value = v
}

function editMessage(m) {
  message.value = m
  edit.value = !edit.value
}

function getService(id) {
  return store.serviceById(id) || {}
}

function serviceName(service) {
  return service.name || 'Global Message'
}

function serviceLink(service) {
  if (!service.id) return '/'
  return `/service/${service.permalink || service.id}`
}

async function deleteMessageConfirm(m) {
  await Api.message_delete(m.id)
  const messagesData = await Api.messages()
  store.setMessages(messagesData)
}

function deleteMessage(m) {
  store.setModal({
    visible: true,
    title: 'Delete Announcement',
    body: `Are you sure you want to delete Announcement ${m.title}?`,
    btnColor: 'btn-danger',
    btnText: 'Delete Announcement',
    func: () => deleteMessageConfirm(m),
  })
}
</script>

<style scoped></style>
