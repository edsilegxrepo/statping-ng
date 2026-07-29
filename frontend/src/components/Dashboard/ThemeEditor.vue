<template>
  <div class="card mb-5">
    <div class="card-header">{{ $t('theme_editor') }}</div>
    <div class="card-body">
      <div v-if="error" class="alert alert-danger mt-3" style="white-space: pre-line">{{ error }}</div>

      <h6 v-if="directory" id="assets_dir" class="text-muted text-monospace text-sm-center font-1 mb-4">
        {{ $t('assets_dir') }}: {{ directory }}
      </h6>

      <div v-if="loaded && !directory" class="jumbotron jumbotron-fluid">
        <div class="text-center col-12">
          <h1 class="display-5">{{ $t('enable_assets') }}</h1>
          <span class="lead"
            >{{ $t('assets_desc') }}
            <p>
              <button
                id="enable_assets"
                @click.prevent="createAssets"
                :disabled="pending"
                href="#"
                class="btn btn-primary mt-3"
              >
                <font-awesome-icon v-if="pending" icon="circle-notch" class="mr-2" spin />{{
                  pending ? $t('assets_loading') : $t('assets_btn')
                }}
              </button>
            </p></span
          >
        </div>
      </div>

      <form v-if="loaded && directory" @submit.prevent="saveAssets" :disabled="pending">
        <h3>Variables</h3>
        <textarea v-model="vars" class="form-control code-editor" rows="10"></textarea>

        <h3 class="mt-3">Base {{ $t('theme') }}</h3>
        <textarea v-model="base" class="form-control code-editor" rows="10"></textarea>

        <h3 class="mt-3">Layout {{ $t('theme') }}</h3>
        <textarea v-model="layout" class="form-control code-editor" rows="10"></textarea>

        <h3 class="mt-3">Forms {{ $t('theme') }}</h3>
        <textarea v-model="forms" class="form-control code-editor" rows="10"></textarea>

        <h3 class="mt-3">Mixins</h3>
        <textarea v-model="mixins" class="form-control code-editor" rows="10"></textarea>

        <h3 class="mt-3">Mobile Overwrites</h3>
        <textarea v-model="mobile" class="form-control code-editor" rows="10"></textarea>
      </form>
    </div>

    <div v-if="directory" class="card-footer">
      <div class="row">
        <div class="col-6">
          <button
            id="save_assets"
            @click.prevent="saveAssets"
            type="submit"
            class="btn btn-primary btn-block"
            :disabled="pending"
          >
            {{ pending ? 'Saving...' : 'Save Styles' }}
          </button>
        </div>
        <div class="col-6">
          <button
            id="delete_assets"
            @click.prevent="deleteAssets"
            class="btn btn-danger btn-block confirm-btn"
            :disabled="pending"
          >
            Delete Local Assets
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useMainStore } from '@/stores/main'
import Api from '@/API'

const store = useMainStore()

const base = ref(null)
const layout = ref(null)
const forms = ref(null)
const mixins = ref(null)
const vars = ref(null)
const mobile = ref(null)
const error = ref(null)
const directory = ref(null)
const loaded = ref(false)
const pending = ref(false)

const coreData = computed(() => store.core)

onMounted(async () => {
  await fetchTheme()
})

async function fetchTheme() {
  loaded.value = true
  pending.value = true
  const theme = await Api.theme()
  directory.value = theme.directory
  if (directory.value) {
    base.value = theme.base
    vars.value = theme.variables
    mobile.value = theme.mobile
    layout.value = theme.layout
    forms.value = theme.forms
    mixins.value = theme.mixins
  }
  pending.value = false
  loaded.value = true
}

async function createAssets() {
  pending.value = true
  try {
    await Api.theme_generate(true)
  } catch (e) {
    error.value = e.response?.data?.error || e.message
  }
  pending.value = false
  await fetchTheme()
}

async function deleteConfirm() {
  pending.value = true
  await Api.theme_generate(false)
  await fetchTheme()
  pending.value = false
}

function deleteAssets() {
  store.setModal({
    visible: true,
    title: 'Delete Local Assets',
    body: `Are you sure you want to delete all local assets?`,
    btnColor: 'btn-danger',
    btnText: 'Delete',
    func: () => deleteConfirm(),
  })
}

async function saveAssets() {
  pending.value = true
  const data = {
    base: base.value,
    layout: layout.value,
    forms: forms.value,
    mixins: mixins.value,
    variables: vars.value,
    mobile: mobile.value,
  }
  let resp
  try {
    resp = await Api.theme_save(data)
  } catch (e) {
    resp = { status: 'error', error: e.response?.data?.error || e.message }
  }
  if (resp?.error) {
    error.value = resp.error
    pending.value = false
    return
  } else {
    error.value = null
  }
  pending.value = false
  await fetchTheme()
}
</script>

<style scoped>
.code-editor {
  font-family: monospace;
  font-size: 12px;
}
</style>
