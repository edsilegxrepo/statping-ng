<template>
  <div class="col-12">
    <div class="row">
      <div class="col-md-3 col-sm-12 mb-4 mb-md-0">
        <div class="nav flex-column nav-pills" id="v-pills-tab" role="tablist" aria-orientation="vertical">
          <div v-if="version_below" class="col-12 small text-center mt-0 pt-0 pb-0 mb-3">
            Update {{ github.tag_name }} Available
            <div class="row">
              <div class="col-6">
                <a href="https://github.com/statping-ng/statping-ng/releases/latest" class="btn btn-sm text-success mt-2"
                  >Download</a
                >
              </div>
              <div class="col-6">
                <a href="https://github.com/statping-ng/statping-ng/blob/master/CHANGELOG.md" class="btn btn-sm text-dim mt-2"
                  >Changelog</a
                >
              </div>
            </div>
          </div>

          <h6 class="text-muted">{{ $t('main_settings') }}</h6>

          <a
            @click.prevent="changeTab"
            class="nav-link"
            :class="{ active: liClass('v-pills-home-tab') }"
            id="v-pills-home-tab"
            data-toggle="pill"
            href="#v-pills-home"
            role="tab"
            aria-controls="v-pills-home"
            aria-selected="true"
          >
            <font-awesome-icon icon="cog" class="mr-2" /> {{ $t('settings') }}
          </a>
          <a
            @click.prevent="changeTab"
            class="nav-link"
            :class="{ active: liClass('v-pills-style-tab') }"
            id="v-pills-style-tab"
            data-toggle="pill"
            href="#v-pills-style"
            role="tab"
            aria-controls="v-pills-style"
            aria-selected="false"
          >
            <font-awesome-icon icon="image" class="mr-2" /> {{ $t('theme') }}
          </a>
          <a
            @click.prevent="changeTab"
            class="nav-link"
            :class="{ active: liClass('v-pills-oauth-tab') }"
            id="v-pills-oauth-tab"
            data-toggle="pill"
            href="#v-pills-oauth"
            role="tab"
            aria-controls="v-pills-oauth"
            aria-selected="false"
          >
            <font-awesome-icon icon="key" class="mr-2" /> {{ $t('authentication') }}
          </a>
          <a
            @click.prevent="changeTab"
            class="nav-link"
            :class="{ active: liClass('v-pills-import-tab') }"
            id="v-pills-import-tab"
            data-toggle="pill"
            href="#v-pills-import"
            role="tab"
            aria-controls="v-pills-import"
            aria-selected="false"
          >
            <font-awesome-icon icon="cloud-download-alt" class="mr-2" /> {{ $t('import') }}
          </a>
          <a
            @click.prevent="changeTab"
            class="nav-link"
            :class="{ active: liClass('v-pills-configs-tab') }"
            id="v-pills-configs-tab"
            data-toggle="pill"
            href="#v-pills-configs"
            role="tab"
            aria-controls="v-pills-configs"
            aria-selected="false"
          >
            <font-awesome-icon icon="cogs" class="mr-2" /> {{ $t('configs') }}
          </a>

          <h6 class="mt-4 text-muted">{{ $t('notifiers') }}</h6>

          <div id="notifiers_tabs">
            <a
              v-for="notifier in notifiers"
              :key="`${notifier.method}`"
              @click.prevent="changeTab"
              class="nav-link text-capitalize"
              :class="{ active: liClass(`v-pills-${notifier.method.toLowerCase()}-tab`) }"
              :id="`v-pills-${notifier.method.toLowerCase()}-tab`"
              data-toggle="pill"
              :href="`#v-pills-${notifier.method.toLowerCase()}`"
              role="tab"
              :aria-controls="`v-pills-${notifier.method.toLowerCase()}`"
              aria-selected="false"
            >
              <font-awesome-icon :icon="iconName(notifier.icon)" class="mr-2" /> {{ notifier.title }}
              <span
                v-if="notifier.enabled"
                class="badge badge-pill float-right mt-1"
                :class="{
                  'badge-success': !liClass(`v-pills-${notifier.method.toLowerCase()}-tab`),
                  'badge-light': liClass(`v-pills-${notifier.method.toLowerCase()}-tab`),
                  'text-dark': liClass(`v-pills-${notifier.method.toLowerCase()}-tab`),
                }"
                >ON</span
              >
            </a>
            <a
              @click.prevent="changeTab"
              class="nav-link text-capitalize"
              :class="{ active: liClass(`v-pills-notifier-docs-tab`) }"
              :id="`v-pills-notifier-docs-tab`"
              data-toggle="pill"
              :href="`#v-pills-notifier-docs`"
              role="tab"
              :aria-controls="`v-pills-notifier-docs`"
              aria-selected="false"
            >
              <font-awesome-icon icon="question" class="mr-2" /> {{ $t('variables') }}
            </a>
          </div>
        </div>
      </div>
      <div class="col-md-9 col-sm-12">
        <div class="tab-content" id="v-pills-tabContent">
          <div
            class="tab-pane fade"
            :class="{ active: liClass('v-pills-home-tab'), show: liClass('v-pills-home-tab') }"
            id="v-pills-home"
            role="tabpanel"
            aria-labelledby="v-pills-home-tab"
          >
            <CoreSettings />

            <div class="card mt-3">
              <div class="card-header">API {{ $t('settings') }}</div>
              <div class="card-body">
                <div class="form-group row">
                  <label class="col-sm-3 col-form-label">API {{ $t('secret') }}</label>
                  <div class="col-sm-9">
                    <div class="input-group">
                      <input
                        v-model="coreData.api_secret"
                        @focus="$event.target.select()"
                        type="text"
                        class="form-control select-input"
                        id="api_secret"
                        readonly
                      />
                      <div class="input-group-append copy-btn">
                        <button @click="copySecret" class="btn btn-outline-secondary" type="button">{{ $t('copy') }}</button>
                      </div>
                    </div>
                    <small class="form-text text-muted">{{ $t('regen_desc') }}</small>
                  </div>
                </div>
              </div>
              <div class="card-footer">
                <button id="regenkeys" @click="renewApiKeys" class="btn btn-sm btn-danger float-right">
                  {{ $t('regen_api') }}
                </button>
              </div>
            </div>
          </div>

          <div
            class="tab-pane fade"
            :class="{ active: liClass('v-pills-style-tab'), show: liClass('v-pills-style-tab') }"
            id="v-pills-style"
            role="tabpanel"
            aria-labelledby="v-pills-style-tab"
          >
            <ThemeEditor />
          </div>

          <div
            class="tab-pane fade"
            :class="{ active: liClass('v-pills-oauth-tab'), show: liClass('v-pills-oauth-tab') }"
            id="v-pills-oauth"
            role="tabpanel"
            aria-labelledby="v-pills-oauth-tab"
          >
            <OAuth />
          </div>

          <div
            class="tab-pane fade"
            :class="{ active: liClass('v-pills-configs-tab'), show: liClass('v-pills-configs-tab') }"
            id="v-pills-configs"
            role="tabpanel"
            aria-labelledby="v-pills-configs-tab"
          >
            <Configs />
          </div>

          <div
            class="tab-pane fade"
            :class="{ active: liClass('v-pills-import-tab'), show: liClass('v-pills-import-tab') }"
            id="v-pills-import"
            role="tabpanel"
            aria-labelledby="v-pills-import-tab"
          >
            <Importer />
          </div>

          <div
            class="tab-pane fade"
            :class="{ active: liClass(`v-pills-notifier-docs-tab`), show: liClass(`v-pills-notifier-docs-tab`) }"
            :id="`v-pills-notifier-docs-tab`"
            role="tabpanel"
            :aria-labelledby="`v-pills-notifier-docs-tab`"
          >
            <Variables />
          </div>

          <div
            v-for="notifier in notifiers"
            :key="`${notifier.method}_panel`"
            class="tab-pane fade"
            :class="{
              active: liClass(`v-pills-${notifier.method.toLowerCase()}-tab`),
              show: liClass(`v-pills-${notifier.method.toLowerCase()}-tab`),
            }"
            :id="`v-pills-${notifier.method.toLowerCase()}-tab`"
            role="tabpanel"
            :aria-labelledby="`v-pills-${notifier.method.toLowerCase()}-tab`"
          >
            <Notifier :notifier="notifier" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useMainStore } from '@/stores/main'
import { useCookies } from 'vue3-cookies'
import semver from 'semver'
import Api from '@/API'
import CoreSettings from '@/forms/CoreSettings.vue'
import Notifier from '@/forms/Notifier.vue'
import OAuth from '@/forms/OAuth.vue'
import ThemeEditor from '@/components/Dashboard/ThemeEditor.vue'
import Importer from '@/components/Dashboard/Importer.vue'
import Variables from '@/components/Dashboard/Variables.vue'
import Configs from '@/components/Dashboard/Configs.vue'

const router = useRouter()
const store = useMainStore()
const { cookies } = useCookies()

const tab = ref('v-pills-home-tab')
const github = ref(null)

const coreData = computed(() => store.core)
const notifiers = computed(() => store.notifiers)

const version_below = computed(() => {
  if (!github.value || !coreData.value.version) {
    return false
  }
  return semver.gt(semver.coerce(github.value.tag_name), semver.coerce(coreData.value.version))
})

onMounted(() => {
  update()
})

async function update() {
  await getGithub()
}

async function getGithub() {
  try {
    github.value = await Api.github_release()
  } catch (e) {
    console.error(e)
  }
}

function changeTab(e) {
  tab.value = e.target.id
}

function liClass(id) {
  return tab.value === id
}

function iconName(icon) {
  return icon || 'bell'
}

async function copySecret() {
  if (coreData.value.api_secret) {
    await navigator.clipboard.writeText(coreData.value.api_secret)
  }
}

async function renew() {
  await Api.renewApiKeys()
  const core = await Api.core()
  store.setCore(core)
  await logout()
}

function renewApiKeys() {
  store.setModal({
    visible: true,
    title: 'Reset API Key',
    body: `Are you sure you want to reset the API keys? You will be logged out.`,
    btnColor: 'btn-danger',
    btnText: 'Reset',
    func: () => renew(),
  })
}

async function logout() {
  try {
    await Api.logout()
  } catch (e) {
    console.error('Backend logout failed', e)
  }
  store.setHasAllData(false)
  store.setToken(null)
  store.setAdmin(false)
  store.setUser(false)
  store.setLoggedIn(false)
  cookies.remove('statping_auth')
  await router.push('/logout')
}
</script>
