<template>
  <div class="dashboard-page">
    <!-- Users Section -->
    <section class="page-section">
      <div class="section-header">
        <div class="section-title">
          <font-awesome-icon icon="users" class="section-icon" />
          <h2>{{ $t('users') }}</h2>
        </div>
      </div>
      <div class="section-card">
        <table class="modern-table">
          <thead>
            <tr>
              <th>{{ $t('username') }}</th>
              <th>{{ $t('type') }}</th>
              <th class="d-none d-md-table-cell">Provider</th>
              <th class="d-none d-md-table-cell">{{ $t('last_login') }}</th>
              <th class="d-none d-lg-table-cell">Scopes</th>
              <th class="text-right">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(user, index) in users" :key="user.id">
              <td>
                <div class="user-info">
                  <div class="user-avatar" :class="user.admin ? 'avatar-admin' : 'avatar-user'">
                    <font-awesome-icon :icon="user.admin ? 'user-shield' : 'user'" />
                  </div>
                  <span class="user-name">{{ user.username }}</span>
                </div>
              </td>
              <td>
                <span class="role-badge" :class="user.admin ? 'badge-admin' : 'badge-user'">
                  {{ user.admin ? $t('admin') : $t('user') }}
                </span>
              </td>
              <td class="d-none d-md-table-cell">
                <span class="provider-badge" :class="'provider-' + (user.auth_provider || 'local')">
                  {{ formatProvider(user.auth_provider) }}
                </span>
              </td>
              <td class="d-none d-md-table-cell">
                <span class="date-text">{{ niceDate(user.updated_at) }}</span>
              </td>
              <td class="d-none d-lg-table-cell">
                <span class="scopes-text">{{ user.scopes || '—' }}</span>
              </td>
              <td class="text-right">
                <div class="action-buttons">
                  <button @click.prevent="editUser(user)" class="action-btn" title="Edit user">
                    <font-awesome-icon icon="edit" />
                  </button>
                  <button
                    v-if="index !== 0"
                    @click.prevent="deleteUser(user)"
                    class="action-btn action-btn-danger"
                    title="Delete user"
                  >
                    <font-awesome-icon icon="trash-alt" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <!-- Create User Form -->
    <section v-if="store.admin" class="page-section">
      <FormUser :edit="editChange" :in_user="user" />
    </section>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useMainStore } from '@/stores/main'
import Api from '@/API'
import FormUser from '@/forms/User.vue'

const store = useMainStore()

const edit = ref(false)
const user = ref({})

const users = computed(() => store.users)

function niceDate(date) {
  if (!date) return '—'
  return new Date(date).toLocaleDateString()
}

function formatProvider(provider) {
  const providers = {
    local: 'Local',
    ldap: 'LDAP',
    oauth_google: 'Google',
    oauth_github: 'GitHub',
    oauth_slack: 'Slack',
    oauth_custom: 'OAuth',
    forward_auth: 'Forward Auth',
  }
  return providers[provider] || 'Local'
}

function editChange(v) {
  user.value = {}
  edit.value = v
}

function editUser(u) {
  const userCopy = { ...u }
  delete userCopy.password
  delete userCopy.confirm_password
  user.value = userCopy
  edit.value = !edit.value
}

async function deleteUserConfirm(u) {
  await Api.user_delete(u.id)
  const usersData = await Api.users()
  store.setUsers(usersData)
}

function deleteUser(u) {
  store.setModal({
    visible: true,
    title: 'Delete User',
    body: `Are you sure you want to delete user ${u.username}?`,
    btnColor: 'btn-danger',
    btnText: 'Delete User',
    func: () => deleteUserConfirm(u),
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

.section-card {
  background: #fff;
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-md);
  overflow: hidden;
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

/* User Info */
.user-info {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.user-avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: var(--radius-full);
  font-size: 0.875rem;
}

.avatar-admin {
  background: var(--color-warning-bg);
  color: var(--color-warning-dark);
}

.avatar-user {
  background: var(--color-primary-bg);
  color: var(--color-primary);
}

.user-name {
  font-weight: 500;
  color: var(--color-gray-900);
}

/* Badges */
.role-badge {
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  padding: var(--space-1) var(--space-2);
  border-radius: var(--radius-full);
}

.badge-admin {
  background: var(--color-warning-bg);
  color: var(--color-warning-dark);
}

.badge-user {
  background: var(--color-primary-bg);
  color: var(--color-primary);
}

/* Provider Badges */
.provider-badge {
  font-size: 0.7rem;
  font-weight: 500;
  padding: var(--space-1) var(--space-2);
  border-radius: var(--radius-sm);
  background: var(--color-gray-100);
  color: var(--color-gray-600);
}

.provider-local {
  background: #e0f2fe;
  color: #0369a1;
}

.provider-ldap {
  background: #fef3c7;
  color: #b45309;
}

.provider-oauth_google,
.provider-oauth_github,
.provider-oauth_slack,
.provider-oauth_custom {
  background: #f3e8ff;
  color: #7c3aed;
}

.provider-forward_auth {
  background: #dcfce7;
  color: #16a34a;
}


/* Text Styles */
.date-text,
.scopes-text {
  font-size: 0.875rem;
  color: var(--color-gray-500);
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
}
</style>
