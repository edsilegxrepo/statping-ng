<template>
  <div class="col-12">
    <div class="card contain-card mb-4">
      <div class="card-header">{{ $t('users') }}</div>
      <div class="card-body pt-0">
        <table class="table table-striped">
          <thead>
            <tr>
              <th scope="col">{{ $t('username') }}</th>
              <th scope="col">{{ $t('type') }}</th>
              <th scope="col" class="d-none d-md-table-cell">{{ $t('last_login') }}</th>
              <th scope="col" class="d-none d-md-table-cell">Scopes</th>
              <th scope="col"></th>
            </tr>
          </thead>
          <tbody id="users_table">
            <tr v-for="(user, index) in users" :key="user.id">
              <td>{{ user.username }}</td>
              <td>
                <span
                  class="badge text-uppercase"
                  :class="{ 'badge-danger': user.admin, 'badge-primary': !user.admin }"
                >
                  {{ user.admin ? $t('admin') : $t('user') }}
                </span>
              </td>
              <td class="d-none d-md-table-cell">{{ niceDate(user.updated_at) }}</td>
              <td class="d-none d-md-table-cell">{{ user.scopes }}</td>
              <td class="text-right">
                <div class="btn-group">
                  <a @click.prevent="editUser(user)" href="#" class="btn btn-outline-secondary edit-user">
                    <font-awesome-icon icon="user" /> {{ $t('edit') }}
                  </a>
                  <a
                    v-if="index !== 0"
                    @click.prevent="deleteUser(user)"
                    href="#"
                    class="btn btn-danger delete-user"
                  >
                    <font-awesome-icon icon="times" />
                  </a>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <FormUser v-if="store.admin" :edit="editChange" :in_user="user" />
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
  if (!date) return ''
  return new Date(date).toLocaleDateString()
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

<style scoped></style>
