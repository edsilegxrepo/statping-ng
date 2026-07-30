<template>
  <div>
    <div v-if="servicesList.length === 0">
      <div class="alert alert-dark d-block mt-3 mb-0">You currently don't have any services!</div>
    </div>
    <table v-else class="table">
      <thead>
        <tr>
          <th scope="col">{{ $t('name') }}</th>
          <th scope="col" class="d-none d-md-table-cell">{{ $t('status') }}</th>
          <th scope="col" class="d-none d-md-table-cell">{{ $t('visibility') }}</th>
          <th scope="col" class="d-none d-md-table-cell">{{ $t('group') }}</th>
          <th scope="col" class="d-none d-md-table-cell" style="width: 130px">
            {{ $t('failures') }}
            <div class="btn-group float-right" role="group">
              <a
                @click="list_timeframe = '3h'"
                class="small"
                :class="{ 'text-success': list_timeframe === '3h', 'text-muted': list_timeframe !== '3h' }"
                >3h</a
              >
              <a
                @click="list_timeframe = '12h'"
                class="small"
                :class="{ 'text-success': list_timeframe === '12h', 'text-muted': list_timeframe !== '12h' }"
                >12h</a
              >
              <a
                @click="list_timeframe = '24h'"
                class="small"
                :class="{ 'text-success': list_timeframe === '24h', 'text-muted': list_timeframe !== '24h' }"
                >24h</a
              >
              <a
                @click="list_timeframe = '7d'"
                class="small"
                :class="{ 'text-success': list_timeframe === '7d', 'text-muted': list_timeframe !== '7d' }"
                >7d</a
              >
            </div>
          </th>
          <th scope="col"></th>
        </tr>
      </thead>
      <draggable id="services_list" tag="tbody" v-model="servicesList" item-key="id" handle=".drag_icon">
        <template #item="{ element: service }">
          <tr>
            <td>
              <span v-if="store.admin" class="drag_icon d-none d-md-inline">
                <font-awesome-icon icon="bars" class="mr-3" />
              </span>
              {{ service.name }}
            </td>
            <td class="d-none d-md-table-cell">
              <span
                class="badge text-uppercase"
                :class="{ 'badge-success': service.online, 'badge-danger': !service.online }"
              >
                {{ service.online ? $t('online') : $t('offline') }}
              </span>
            </td>
            <td class="d-none d-md-table-cell">
              <span
                class="badge text-uppercase"
                :class="{ 'badge-primary': service.public, 'badge-secondary': !service.public }"
              >
                {{ service.public ? $t('public') : $t('private') }}
              </span>
            </td>
            <td class="d-none d-md-table-cell">
              <div v-if="service.group_id !== 0">
                <span class="badge badge-secondary">{{ serviceGroup(service) }}</span>
              </div>
            </td>
            <td class="d-none d-md-table-cell">
              <ServiceSparkList :service="service" :timeframe="list_timeframe" />
            </td>
            <td class="text-right">
              <div class="btn-group">
                <button
                  :disabled="loading"
                  v-if="store.admin"
                  @click.prevent="gotoEdit(service)"
                  class="btn btn-sm btn-outline-secondary"
                >
                  <font-awesome-icon icon="edit" />
                </button>
                <button :disabled="loading" @click.prevent="gotoService(service)" class="btn btn-sm btn-outline-secondary">
                  <font-awesome-icon icon="chart-area" />
                </button>
                <button
                  :disabled="loading"
                  v-if="store.admin"
                  @click.prevent="deleteService(service)"
                  class="btn btn-sm btn-danger"
                >
                  <font-awesome-icon v-if="!loading" icon="times" />
                  <font-awesome-icon v-if="loading" icon="circle-notch" spin />
                </button>
              </div>
            </td>
          </tr>
        </template>
      </draggable>
    </table>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useMainStore } from '@/stores/main'
import draggable from 'vuedraggable'
import Api from '@/API'
import ServiceSparkList from '@/components/Service/ServiceSparkList.vue'

const router = useRouter()
const store = useMainStore()

const loading = ref(false)
const list_timeframe = ref('12h')

const servicesList = computed({
  get() {
    return store.servicesInOrder
  },
  set(value) {
    updateOrder(value)
  },
})

function gotoEdit(service) {
  router.push({ path: `/dashboard/edit_service/${service.id}` })
}

function gotoService(service) {
  router.push({ path: `/service/${service.permalink || service.id}`, query: { from: 'dashboard' } })
}

async function updateOrder(value) {
  const data = value.map((s, k) => ({ service: s.id, order: k + 1 }))
  await Api.services_reorder(data)
  await update()
}

async function deleteServiceConfirm(s) {
  loading.value = true
  await Api.service_delete(s.id)
  await update()
  loading.value = false
}

function deleteService(s) {
  store.setModal({
    visible: true,
    title: 'Delete Service',
    body: `Are you sure you want to delete service ${s.name}? This will also delete all failures, checkins, and incidents for this service.`,
    btnColor: 'btn-danger',
    btnText: 'Delete Service',
    func: () => deleteServiceConfirm(s),
  })
}

function serviceGroup(s) {
  const group = store.groupById(s.group_id)
  return group ? group.name : ''
}

async function update() {
  const services = await Api.services()
  store.setServices(services)
}
</script>

<style scoped></style>
