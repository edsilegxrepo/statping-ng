<template>
  <form @submit.prevent="saveOAuth">
    <div class="card mb-3">
      <div class="card-header">
        Internal Login
        <span @click="local_enabled = !!local_enabled" class="switch switch-sm switch-rd-gr float-right">
          <input v-model="local_enabled" type="checkbox" id="switch-internal-oauth" :checked="local_enabled" />
          <label for="switch-internal-oauth" class="mb-0"> </label>
        </span>
      </div>
      <div class="card-body">Use Statping's default authentication to allow users you've created to login.</div>
    </div>
    <div class="card mb-3">
      <div class="card-header text-capitalize">
        <font-awesome-icon
          @click="expanded.github = !expanded.github"
          :icon="expanded.github ? 'minus' : 'plus'"
          class="mr-2 pointer"
        />
        Github Settings
        <span @click="github_enabled = !!github_enabled" class="switch switch-sm switch-rd-gr float-right">
          <input v-model="github_enabled" type="checkbox" id="switch-gh-oauth" :checked="github_enabled" />
          <label class="mb-0" for="switch-gh-oauth"> </label>
        </span>
      </div>
      <div class="card-body" :class="{ 'd-none': !expanded.github }">
        <span
          >You will need to create a new <a href="https://github.com/settings/developers">OAuth App</a> within Github.</span
        >

        <div class="form-group row mt-3">
          <label for="github_client" class="col-sm-4 col-form-label">Github Client ID</label>
          <div class="col-sm-8">
            <input v-model="oauth.gh_client_id" type="text" class="form-control" id="github_client" required />
          </div>
        </div>
        <div class="form-group row">
          <label for="github_secret" class="col-sm-4 col-form-label">Github Client Secret</label>
          <div class="col-sm-8">
            <input v-model="oauth.gh_client_secret" type="text" class="form-control" id="github_secret" required />
          </div>
        </div>
        <div class="form-group row">
          <label for="github_users" class="col-sm-4 col-form-label">Restrict Users</label>
          <div class="col-sm-8">
            <input
              v-model="oauth.gh_users"
              type="text"
              class="form-control"
              id="github_users"
              placeholder="octocat,hunterlong,jimbo123"
            />
            <small>Optional comma delimited list of usernames</small>
          </div>
        </div>
        <div class="form-group row">
          <label for="github_orgs" class="col-sm-4 col-form-label">Restrict Organizations</label>
          <div class="col-sm-8">
            <input
              v-model="oauth.gh_orgs"
              type="text"
              class="form-control"
              id="github_orgs"
              placeholder="statping,github"
            />
            <small>Optional comma delimited list of Github Organizations</small>
          </div>
        </div>
        <div class="form-group row">
          <label for="gh_callback" class="col-sm-4 col-form-label">Callback URL</label>
          <div class="col-sm-8">
            <div class="input-group">
              <input :value="`${coreData.domain}/oauth/github`" type="text" class="form-control" id="gh_callback" readonly />
              <div class="input-group-append copy-btn">
                <button @click.prevent="copyCallback('github')" class="btn btn-outline-secondary" type="button">Copy</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div class="card mb-3">
      <div class="card-header">
        <font-awesome-icon
          @click="expanded.google = !expanded.google"
          :icon="expanded.google ? 'minus' : 'plus'"
          class="mr-2 pointer"
        />
        Google Settings
        <span @click="google_enabled = !!google_enabled" class="switch switch-sm switch-rd-gr float-right">
          <input v-model="google_enabled" type="checkbox" id="switch-google-oauth" :checked="google_enabled" />
          <label for="switch-google-oauth" class="mb-0"> </label>
        </span>
      </div>
      <div class="card-body" :class="{ 'd-none': !expanded.google }">
        <span
          >Go to <a href="https://console.cloud.google.com/apis/credentials">OAuth Consent Screen</a> on Google Console to
          create a new "Web Application" OAuth application.
        </span>

        <div class="form-group row mt-3">
          <label for="google_client" class="col-sm-4 col-form-label">Google Client ID</label>
          <div class="col-sm-8">
            <input v-model="oauth.google_client_id" type="text" class="form-control" id="google_client" required />
          </div>
        </div>
        <div class="form-group row">
          <label for="google_secret" class="col-sm-4 col-form-label">Google Client Secret</label>
          <div class="col-sm-8">
            <input v-model="oauth.google_client_secret" type="text" class="form-control" id="google_secret" required />
          </div>
        </div>
        <div class="form-group row">
          <label for="google_users" class="col-sm-4 col-form-label">Restrict Users</label>
          <div class="col-sm-8">
            <input
              v-model="oauth.google_users"
              type="text"
              class="form-control"
              id="google_users"
              placeholder="info@gmail.com,example.com"
            />
            <small>Optional comma delimited list of emails and/or domains</small>
          </div>
        </div>
        <div class="form-group row">
          <label for="google_callback" class="col-sm-4 col-form-label">Callback URL</label>
          <div class="col-sm-8">
            <div class="input-group">
              <input
                :value="`${coreData.domain}/oauth/google`"
                type="text"
                class="form-control"
                id="google_callback"
                readonly
              />
              <div class="input-group-append copy-btn">
                <button @click.prevent="copyCallback('google')" class="btn btn-outline-secondary" type="button">Copy</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div class="card mb-3">
      <div class="card-header">
        <font-awesome-icon
          @click="expanded.slack = !expanded.slack"
          :icon="expanded.slack ? 'minus' : 'plus'"
          class="mr-2 pointer"
        />
        Slack Settings
        <span @click="slack_enabled = !!slack_enabled" class="switch switch-sm switch-rd-gr float-right">
          <input v-model="slack_enabled" type="checkbox" id="switch-slack-oauth" :checked="slack_enabled" />
          <label for="switch-slack-oauth" class="mb-0"> </label>
        </span>
      </div>
      <div class="card-body" :class="{ 'd-none': !expanded.slack }">
        <span>Go to <a href="https://api.slack.com/apps">Slack Apps</a> and create a new Application.</span>

        <div class="form-group row mt-3">
          <label for="slack_client" class="col-sm-4 col-form-label">Slack Client ID</label>
          <div class="col-sm-8">
            <input v-model="oauth.slack_client_id" type="text" class="form-control" id="slack_client" required />
          </div>
        </div>
        <div class="form-group row">
          <label for="slack_secret" class="col-sm-4 col-form-label">Slack Client Secret</label>
          <div class="col-sm-8">
            <input v-model="oauth.slack_client_secret" type="text" class="form-control" id="slack_secret" required />
          </div>
        </div>
        <div class="form-group row">
          <label for="slack_team" class="col-sm-4 col-form-label">Team ID</label>
          <div class="col-sm-8">
            <input v-model="oauth.slack_team" type="text" class="form-control" id="slack_team" />
            <small>Optional Slack Team ID</small>
          </div>
        </div>
        <div class="form-group row">
          <label for="slack_users" class="col-sm-4 col-form-label">Restrict Users</label>
          <div class="col-sm-8">
            <input
              v-model="oauth.slack_users"
              type="text"
              class="form-control"
              id="slack_users"
              placeholder="info@example.com,info@domain.net"
            />
            <small>Optional comma delimited list of email addresses</small>
          </div>
        </div>
        <div class="form-group row">
          <label for="slack_callback" class="col-sm-4 col-form-label">Callback URL</label>
          <div class="col-sm-8">
            <div class="input-group">
              <input :value="`${coreData.domain}/oauth/slack`" type="text" class="form-control" id="slack_callback" readonly />
              <div class="input-group-append copy-btn">
                <button @click.prevent="copyCallback('slack')" class="btn btn-outline-secondary" type="button">Copy</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="card mb-3">
      <div class="card-header">
        <font-awesome-icon
          @click="expanded.custom = !expanded.custom"
          :icon="expanded.custom ? 'minus' : 'plus'"
          class="mr-2 pointer"
        />
        Custom oAuth Settings
        <span @click="custom_enabled = !!custom_enabled" class="switch switch-sm switch-rd-gr float-right">
          <input v-model="custom_enabled" type="checkbox" id="switch-custom-oauth" :checked="custom_enabled" />
          <label for="switch-custom-oauth" class="mb-0"> </label>
        </span>
      </div>
      <div class="card-body" :class="{ 'd-none': !expanded.custom || !custom_enabled }">
        <div class="form-group row">
          <label for="custom_name" class="col-sm-4 col-form-label">Custom Name</label>
          <div class="col-sm-8">
            <input v-model="oauth.custom_name" type="text" class="form-control" id="custom_name" required />
          </div>
        </div>
        <div class="form-group row mt-3">
          <label for="custom_client" class="col-sm-4 col-form-label">Client ID</label>
          <div class="col-sm-8">
            <input v-model="oauth.custom_client_id" type="text" class="form-control" id="custom_client" required />
          </div>
        </div>
        <div class="form-group row">
          <label for="custom_secret" class="col-sm-4 col-form-label">Client Secret</label>
          <div class="col-sm-8">
            <input v-model="oauth.custom_client_secret" type="text" class="form-control" id="custom_secret" required />
          </div>
        </div>
        <div class="form-group row">
          <label for="custom_endpoint" class="col-sm-4 col-form-label">Auth Endpoint</label>
          <div class="col-sm-8">
            <input v-model="oauth.custom_endpoint_auth" type="text" class="form-control" id="custom_endpoint" required />
          </div>
        </div>
        <div class="form-group row">
          <label for="custom_endpoint_token" class="col-sm-4 col-form-label">Token Endpoint</label>
          <div class="col-sm-8">
            <input
              v-model="oauth.custom_endpoint_token"
              type="text"
              class="form-control"
              id="custom_endpoint_token"
              required
            />
          </div>
        </div>
        <div class="form-group row">
          <label for="custom_scopes" class="col-sm-4 col-form-label">Scopes</label>
          <div class="col-sm-8">
            <input v-model="oauth.custom_scopes" type="text" class="form-control" id="custom_scopes" />
            <small>Optional comma delimited list of oauth scopes</small>
          </div>
        </div>
        <div class="form-group row">
          <label for="switch-custom-openid" class="col-sm-4 col-form-label">Open ID</label>
          <div class="col-sm-8">
            <span @click="oauth.custom_open_id = !!oauth.custom_open_id" class="switch switch-rd-gr float-right">
              <input
                v-model="oauth.custom_open_id"
                type="checkbox"
                id="switch-custom-openid"
                :checked="oauth.custom_open_id"
              />
              <label for="switch-custom-openid" class="mb-0"> </label>
            </span>
            <small>Enable if provider is OpenID</small>
          </div>
        </div>

        <div class="form-group row">
          <label for="custom_callback" class="col-sm-4 col-form-label">Callback URL</label>
          <div class="col-sm-8">
            <div class="input-group">
              <input
                :value="`${coreData.domain}/oauth/custom`"
                type="text"
                class="form-control"
                id="custom_callback"
                readonly
              />
              <div class="input-group-append copy-btn">
                <button @click.prevent="copyCallback('custom')" class="btn btn-outline-secondary" type="button">Copy</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <button class="btn btn-primary btn-block" @click.prevent="saveOAuth" type="submit" :disabled="loading">
      <font-awesome-icon v-if="loading" icon="circle-notch" class="mr-2" spin /> Save OAuth Settings
    </button>
  </form>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useMainStore } from '@/stores/main'
import Api from '@/API'

const store = useMainStore()

const coreData = computed(() => store.core)

const google_enabled = ref(false)
const slack_enabled = ref(false)
const github_enabled = ref(false)
const local_enabled = ref(false)
const custom_enabled = ref(false)
const loading = ref(false)

const expanded = reactive({
  github: false,
  google: false,
  slack: false,
  custom: false,
  openid: false,
})

const oauth = reactive({
  gh_client_id: '',
  gh_client_secret: '',
  gh_users: '',
  gh_orgs: '',
  google_client_id: '',
  google_client_secret: '',
  google_users: '',
  oauth_providers: '',
  slack_client_id: '',
  slack_client_secret: '',
  slack_team: '',
  slack_users: '',
  custom_name: '',
  custom_client_id: '',
  custom_client_secret: '',
  custom_endpoint_auth: '',
  custom_endpoint_token: '',
  custom_scopes: '',
  custom_open_id: false,
})

onMounted(async () => {
  const data = await Api.oauth()
  Object.assign(oauth, data)
  local_enabled.value = has('local')
  github_enabled.value = has('github')
  google_enabled.value = has('google')
  slack_enabled.value = has('slack')
  custom_enabled.value = has('custom')
})

function providers() {
  const list = []
  if (github_enabled.value) list.push('github')
  if (local_enabled.value) list.push('local')
  if (google_enabled.value) list.push('google')
  if (slack_enabled.value) list.push('slack')
  if (custom_enabled.value) list.push('custom')
  return list.join(',')
}

function has(val) {
  if (!oauth.oauth_providers) return false
  return oauth.oauth_providers.split(',').includes(val)
}

async function copyCallback(provider) {
  const url = `${coreData.value.domain}/oauth/${provider}`
  await navigator.clipboard.writeText(url)
}

async function saveOAuth() {
  loading.value = true
  oauth.oauth_providers = providers()
  await Api.oauth_save(oauth)
  const data = await Api.oauth()
  store.setOAuth(data)
  loading.value = false
}
</script>

<style scoped></style>
