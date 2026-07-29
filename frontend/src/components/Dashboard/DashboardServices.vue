<template>
  <div class="col-12">
    <div class="card contain-card mb-4">
      <div class="card-header">
        {{ $t('services') }}
        <router-link v-if="store.admin" to="/dashboard/create_service" class="btn btn-sm btn-success float-right">
          <font-awesome-icon icon="plus" /> {{ $t('create') }}
        </router-link>
      </div>
      <div class="card-body pt-0">
        <ServicesList />
      </div>
    </div>

    <div class="card contain-card mb-4">
      <div class="card-header">{{ $t('groups') }}</div>
      <div class="card-body pt-0">
        <div v-if="groupsList.length === 0">
          <div class="alert alert-dark d-block mt-3 mb-0">
            You currently don't have any groups! Create one using the form below.
          </div>
        </div>

        <table v-else class="table">
          <thead>
            <tr>
              <th scope="col">{{ $t('name') }}</th>
              <th scope="col" class="d-none d-md-table-cell">{{ $t('service', 2) }}</th>
              <th scope="col">{{ $t('visibility') }}</th>
              <th scope="col"></th>
            </tr>
          </thead>

          <draggable tag="tbody" v-model="groupsList" item-key="id" handle=".drag_icon" class="sortable_groups">
            <template #item="{ element: group }">
              <tr>
                <td>
                  <span class="drag_icon d-none d-md-inline">
                    <font-awesome-icon icon="bars" class="mr-3" /> </span
                  >{{ group.name }}
                </td>
                <td class="d-none d-md-table-cell">{{ store.servicesInGroup(group.id).length }}</td>
                <td>
                  <span
                    class="badge text-uppercase"
                    :class="{ 'badge-primary': group.public, 'badge-secondary': !group.public }"
                  >
                    {{ group.public ? $t('public') : $t('private') }}
                  </span>
                </td>
                <td class="text-right">
                  <div v-if="store.admin" class="btn-group">
                    <button @click.prevent="editGroup(group)" class="btn btn-sm btn-outline-secondary">
                      <font-awesome-icon icon="edit" />
                    </button>
                    <button @click.prevent="deleteGroup(group)" class="btn btn-sm btn-danger">
                      <font-awesome-icon icon="times" />
                    </button>
                  </div>
                </td>
              </tr>
            </template>
          </draggable>
        </table>
      </div>
    </div>

    <FormGroup v-if="store.admin" :edit="editChange" :in_group="group" />
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useMainStore } from '@/stores/main'
import draggable from 'vuedraggable'
import Api from '@/API'
import ServicesList from '@/components/Dashboard/ServicesList.vue'
import FormGroup from '@/forms/Group.vue'

const store = useMainStore()

const edit = ref(false)
const group = ref({})

const groupsList = computed({
  get() {
    return store.groupsClean
  },
  set(value) {
    reorderGroups(value)
  },
})

async function reorderGroups(value) {
  const data = value.map((s, k) => ({ group: s.id, order: k + 1 }))
  await Api.groups_reorder(data)
  const groups = await Api.groups()
  store.setGroups(groups)
}

function editChange(v) {
  group.value = {}
  edit.value = v
}

function editGroup(g) {
  group.value = g
  edit.value = !edit.value
}

async function deleteGroupConfirm(g) {
  await Api.group_delete(g.id)
  const groups = await Api.groups()
  store.setGroups(groups)
}

function deleteGroup(g) {
  store.setModal({
    visible: true,
    title: 'Delete Group',
    body: `Are you sure you want to delete group ${g.name}? All services attached will be removed from this group.`,
    btnColor: 'btn-danger',
    btnText: 'Delete Group',
    func: () => deleteGroupConfirm(g),
  })
}
</script>

<style scoped></style>
