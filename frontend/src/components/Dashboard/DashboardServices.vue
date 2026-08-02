<template>
  <div class="dashboard-page">
    <!-- Services Section -->
    <section class="page-section">
      <div class="section-header">
        <div class="section-title">
          <font-awesome-icon icon="server" class="section-icon" />
          <h2>{{ $t('services') }}</h2>
        </div>
        <router-link v-if="store.admin" to="/dashboard/create_service" class="btn-create">
          <font-awesome-icon icon="plus" class="me-2" />
          {{ $t('create') }}
        </router-link>
      </div>
      <div class="section-card">
        <ServicesList />
      </div>
    </section>

    <!-- Groups Section -->
    <section class="page-section">
      <div class="section-header">
        <div class="section-title">
          <font-awesome-icon icon="layer-group" class="section-icon" />
          <h2>{{ $t('groups') }}</h2>
        </div>
      </div>
      <div class="section-card">
        <div v-if="groupsList.length === 0" class="empty-state-inline">
          <font-awesome-icon icon="folder-open" class="empty-icon" />
          <p>You currently don't have any groups. Create one using the form below.</p>
        </div>

        <table v-else class="modern-table">
          <thead>
            <tr>
              <th>{{ $t('name') }}</th>
              <th class="d-none d-md-table-cell">{{ $t('service', 2) }}</th>
              <th>{{ $t('visibility') }}</th>
              <th class="text-right">Actions</th>
            </tr>
          </thead>

          <draggable tag="tbody" v-model="groupsList" item-key="id" handle=".drag-handle">
            <template #item="{ element: group }">
              <tr>
                <td>
                  <div class="cell-with-drag">
                    <span class="drag-handle d-none d-md-inline">
                      <font-awesome-icon icon="grip-vertical" />
                    </span>
                    <span class="group-name">{{ group.name }}</span>
                  </div>
                </td>
                <td class="d-none d-md-table-cell">
                  <span class="count-badge">{{ store.servicesInGroup(group.id).length }}</span>
                </td>
                <td>
                  <span class="visibility-badge" :class="group.public ? 'badge-public' : 'badge-private'">
                    {{ group.public ? $t('public') : $t('private') }}
                  </span>
                </td>
                <td class="text-right">
                  <div v-if="store.admin" class="action-buttons">
                    <button @click.prevent="editGroup(group)" class="action-btn" title="Edit">
                      <font-awesome-icon icon="edit" />
                    </button>
                    <button @click.prevent="deleteGroup(group)" class="action-btn action-btn-danger" title="Delete">
                      <font-awesome-icon icon="trash-alt" />
                    </button>
                  </div>
                </td>
              </tr>
            </template>
          </draggable>
        </table>
      </div>
    </section>

    <!-- Create Group Form -->
    <section v-if="store.admin" class="page-section">
      <FormGroup :edit="editChange" :in_group="group" />
    </section>
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

<style scoped>
.dashboard-page {
  max-width: 100%;
}

/* Section */
.page-section {
  margin-bottom: var(--space-6);
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-4);
}

.section-title {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.section-title h2 {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--color-gray-900);
  margin: 0;
}

.section-icon {
  color: var(--color-primary);
  font-size: 1.25rem;
}

.btn-create {
  display: inline-flex;
  align-items: center;
  padding: var(--space-2) var(--space-4);
  background: var(--color-success);
  color: #fff;
  font-size: 0.875rem;
  font-weight: 500;
  border-radius: var(--radius-md);
  text-decoration: none;
  transition: all var(--transition-fast);
}

.btn-create:hover {
  background: var(--color-success-dark);
  color: #fff;
}

.section-card {
  background: #fff;
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-md);
  overflow: hidden;
}

/* Empty State */
.empty-state-inline {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: var(--space-8);
  text-align: center;
}

.empty-icon {
  font-size: 2.5rem;
  color: var(--color-gray-300);
  margin-bottom: var(--space-3);
}

.empty-state-inline p {
  color: var(--color-gray-500);
  margin: 0;
}

/* Modern Table */
.modern-table {
  width: 100%;
  border-collapse: collapse;
}

.modern-table th {
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--color-gray-500);
  padding: var(--space-3) var(--space-4);
  background: var(--color-gray-50);
  border-bottom: 1px solid var(--color-gray-200);
}

.modern-table td {
  padding: var(--space-3) var(--space-4);
  border-bottom: 1px solid var(--color-gray-100);
  vertical-align: middle;
}

.modern-table tbody tr {
  transition: background var(--transition-fast);
}

.modern-table tbody tr:hover {
  background: var(--color-gray-50);
}

.modern-table tbody tr:last-child td {
  border-bottom: none;
}

/* Cell with drag */
.cell-with-drag {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.drag-handle {
  color: var(--color-gray-400);
  cursor: grab;
  padding: var(--space-1);
}

.drag-handle:hover {
  color: var(--color-gray-600);
}

.group-name {
  font-weight: 500;
  color: var(--color-gray-900);
}

/* Badges */
.count-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 24px;
  height: 24px;
  padding: 0 var(--space-2);
  background: var(--color-gray-100);
  color: var(--color-gray-700);
  font-size: 0.8rem;
  font-weight: 600;
  border-radius: var(--radius-full);
}

.visibility-badge {
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  padding: var(--space-1) var(--space-2);
  border-radius: var(--radius-full);
}

.badge-public {
  background: var(--color-primary-bg);
  color: var(--color-primary);
}

.badge-private {
  background: var(--color-gray-100);
  color: var(--color-gray-600);
}

/* Action Buttons */
.action-buttons {
  display: flex;
  gap: var(--space-1);
  justify-content: flex-end;
}

.action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: #fff;
  border: 1px solid var(--color-gray-200);
  border-radius: var(--radius-md);
  color: var(--color-gray-500);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.action-btn:hover {
  background: var(--color-gray-100);
  color: var(--color-gray-700);
  border-color: var(--color-gray-300);
}

.action-btn-danger {
  color: var(--color-danger);
  border-color: rgba(239, 68, 68, 0.3);
}

.action-btn-danger:hover {
  background: var(--color-danger-bg);
  color: var(--color-danger-dark);
}

/* Responsive */
@media (max-width: 768px) {
  .dashboard-page {
    padding: var(--space-3);
  }

  .section-header {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-3);
  }
}
</style>
