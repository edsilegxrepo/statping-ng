<template>
  <div class="card contain-card mb-3">
    <div class="card-header">
      {{ user.id ? `${$t('update')} ${user.username}` : $t('user_create') }}
      <transition name="slide-fade">
        <button @click.prevent="removeEdit" v-if="user.id" class="btn btn-sm float-right btn-danger btn-sm">Close</button>
      </transition>
    </div>
    <div class="card-body">
      <form @submit.prevent="saveUser">
        <div class="form-group row">
          <label class="col-sm-4 col-form-label">{{ $t('username') }}</label>
          <div class="col-6 col-md-4">
            <input
              v-model="user.username"
              type="text"
              class="form-control"
              id="username"
              placeholder="Username"
              required
              autocorrect="off"
              autocapitalize="none"
              :readonly="!!user.id"
            />
          </div>
          <div class="col-6 col-md-4">
            <span id="admin_switch" @click="user.admin = !!user.admin" class="switch">
              <input v-model="user.admin" type="checkbox" class="switch" id="user_admin_switch" :checked="user.admin" />
              <label for="user_admin_switch">{{ $t('administrator') }}</label>
            </span>
          </div>
        </div>
        <div class="form-group row">
          <label for="email" class="col-sm-4 col-form-label">{{ $t('email') }}</label>
          <div class="col-sm-8">
            <input
              v-model="user.email"
              type="email"
              class="form-control"
              id="email"
              placeholder="user@domain.com"
              required
              autocapitalize="none"
              spellcheck="false"
            />
          </div>
        </div>
        <div class="form-group row">
          <label class="col-sm-4 col-form-label">{{ $t('password') }}</label>
          <div class="col-sm-8">
            <input v-model="user.password" type="password" id="password" class="form-control" placeholder="Password" required />
          </div>
        </div>
        <div class="form-group row">
          <label class="col-sm-4 col-form-label">{{ $t('confirm_password') }}</label>
          <div class="col-sm-8">
            <input
              v-model="user.confirm_password"
              type="password"
              id="password_confirm"
              class="form-control"
              placeholder="Confirm Password"
              required
            />
            <span v-if="passTooWeak" class="small text-danger d-block"
              >Password must be at least 30 characters and include uppercase, lowercase, and digits</span
            >
          </div>
        </div>
        <div v-if="user.api_key" class="form-group row">
          <label for="user_key_key" class="col-sm-4 col-form-label">API Key</label>
          <div class="col-sm-8">
            <div class="input-group">
              <input :value="user.api_key" type="text" class="form-control" id="user_key_key" readonly />
              <div class="input-group-append copy-btn">
                <button @click.prevent="copyApiKey" class="btn btn-outline-secondary" type="button">Copy</button>
              </div>
            </div>
          </div>
        </div>
        <div class="form-group row">
          <div class="col-sm-12">
            <LoadButton class="btn-primary" :disabled="loading || !canSubmit" :action="saveUser" :label="user.id ? $t('user_update') : $t('user_create')" />
          </div>
        </div>
        <div class="alert alert-danger d-none" id="alerter" role="alert"></div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useMainStore } from '@/stores/main'
import Api from '@/API'
import LoadButton from '@/components/Elements/LoadButton.vue'

const props = defineProps({
  in_user: {
    type: Object,
    default: () => ({}),
  },
  edit: {
    type: Function,
    default: () => {},
  },
})

const store = useMainStore()
const loading = ref(false)
const passTooWeak = ref(false)
const user = ref({
  username: '',
  admin: false,
  email: '',
  password: '',
  confirm_password: '',
  api_key: '',
})

const canSubmit = computed(() => {
  const u = user.value
  const hasUpper = /[A-Z]/.test(u.password)
  const hasLower = /[a-z]/.test(u.password)
  const hasDigit = /[0-9]/.test(u.password)
  const isStrong = u.password && u.password.length >= 30 && hasUpper && hasLower && hasDigit
  const match = u.password === u.confirm_password

  if (u.id) {
    if (!u.password) return u.username && u.email
    return u.username && u.email && isStrong && match
  }
  return u.username && u.email && u.password && isStrong && match
})

watch(
  () => props.in_user,
  (val) => {
    if (val && Object.keys(val).length > 0) {
      user.value = { ...val }
    }
  },
  { immediate: true }
)

function removeEdit() {
  user.value = {
    username: '',
    admin: false,
    email: '',
    password: '',
    confirm_password: '',
    api_key: '',
  }
  props.edit(false)
}

async function copyApiKey() {
  if (user.value.api_key) {
    await navigator.clipboard.writeText(user.value.api_key)
  }
}

async function saveUser() {
  loading.value = true
  const hasUpper = /[A-Z]/.test(user.value.password)
  const hasLower = /[a-z]/.test(user.value.password)
  const hasDigit = /[0-9]/.test(user.value.password)
  if (user.value.password && (user.value.password.length < 30 || !hasUpper || !hasLower || !hasDigit)) {
    passTooWeak.value = true
    loading.value = false
    return
  }
  passTooWeak.value = false
  if (user.value.id) {
    await updateUser()
  } else {
    await createUser()
  }
  loading.value = false
}

async function createUser() {
  const userData = { ...user.value }
  delete userData.confirm_password
  await Api.user_create(userData)
  await update()
  user.value = {
    username: '',
    admin: false,
    email: '',
    password: '',
    confirm_password: '',
    api_key: '',
  }
}

async function updateUser() {
  const userData = { ...user.value }
  if (!userData.password) {
    delete userData.password
  }
  delete userData.confirm_password
  await Api.user_update(userData)
  await update()
  props.edit(false)
}

async function update() {
  const users = await Api.users()
  store.setUsers(users)
}
</script>

<style scoped></style>
