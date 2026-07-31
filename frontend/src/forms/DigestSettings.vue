<template>
  <div class="card">
    <div class="card-header">
      Daily Digest Email
    </div>
    <div class="card-body">
      <div v-if="loading" class="text-center py-4">
        <font-awesome-icon icon="circle-notch" spin size="2x" />
      </div>

      <form v-else @submit.prevent="save">
        <div class="form-group row">
          <label class="col-sm-4 col-form-label">Enable Daily Digest</label>
          <div class="col-sm-8">
            <div class="switch">
              <input v-model="digest.digest_enabled" type="checkbox" id="digest_enabled" />
              <label for="digest_enabled">Send daily summary of service status and errors</label>
            </div>
          </div>
        </div>

        <div class="form-group row">
          <label class="col-sm-4 col-form-label">Email Recipients</label>
          <div class="col-sm-8">
            <input
              v-model="digest.digest_emails"
              type="text"
              class="form-control"
              placeholder="admin@example.com, ops@example.com"
              :disabled="!digest.digest_enabled"
            />
            <small class="form-text text-muted">
              Comma-separated list of email addresses to receive the daily digest
            </small>
          </div>
        </div>

        <div class="form-group row">
          <label class="col-sm-4 col-form-label">Send Time (Hour)</label>
          <div class="col-sm-8">
            <select
              v-model.number="digest.digest_hour"
              class="form-control"
              :disabled="!digest.digest_enabled"
            >
              <option v-for="h in 24" :key="h-1" :value="h-1">
                {{ formatHour(h-1) }}
              </option>
            </select>
            <small class="form-text text-muted">
              Time of day to send the digest (server timezone)
            </small>
          </div>
        </div>

        <div class="alert alert-info">
          <strong>Note:</strong> The daily digest uses the SMTP settings configured in the Email notifier.
          Make sure SMTP is configured before enabling the digest.
        </div>

        <hr />

        <div class="form-group row">
          <div class="col-sm-4"></div>
          <div class="col-sm-8">
            <button
              type="button"
              @click="testSmtp"
              :disabled="testingSmtp"
              class="btn btn-outline-secondary mr-2"
            >
              <font-awesome-icon v-if="testingSmtp" icon="circle-notch" spin />
              <font-awesome-icon v-else icon="plug" />
              Test SMTP
            </button>
            <button
              type="button"
              @click="testDigest"
              :disabled="testing || !digest.digest_enabled || !digest.digest_emails"
              class="btn btn-outline-primary mr-2"
            >
              <font-awesome-icon v-if="testing" icon="circle-notch" spin />
              <font-awesome-icon v-else icon="paper-plane" />
              Send Test Digest
            </button>
            <button type="submit" :disabled="saving" class="btn btn-success">
              <font-awesome-icon v-if="saving" icon="circle-notch" spin />
              Save Settings
            </button>
          </div>
        </div>

        <div v-if="smtpResult" class="card mt-3">
          <div class="card-header" :class="smtpResult.connected ? 'bg-success text-white' : 'bg-danger text-white'">
            SMTP Diagnostics: {{ smtpResult.host }}:{{ smtpResult.port }}
          </div>
          <div class="card-body">
            <table class="table table-sm mb-0">
              <tr>
                <td><strong>Connected</strong></td>
                <td>
                  <span :class="smtpResult.connected ? 'text-success' : 'text-danger'">
                    {{ smtpResult.connected ? 'Yes' : 'No' }}
                  </span>
                </td>
              </tr>
              <tr v-if="smtpResult.banner">
                <td><strong>Server Banner</strong></td>
                <td><code>{{ smtpResult.banner }}</code></td>
              </tr>
              <tr v-if="smtpResult.connected">
                <td><strong>TLS/STARTTLS</strong></td>
                <td>
                  <span :class="smtpResult.tls_supported ? 'text-success' : 'text-warning'">
                    {{ smtpResult.tls_supported ? 'Supported' : 'Not supported' }}
                  </span>
                </td>
              </tr>
              <tr v-if="smtpResult.connected">
                <td><strong>Authentication</strong></td>
                <td>
                  {{ smtpResult.auth_supported ? 'Supported' : 'Not required' }}
                  <span v-if="smtpResult.auth_methods" class="text-muted ml-2">({{ smtpResult.auth_methods }})</span>
                </td>
              </tr>
              <tr v-if="smtpResult.error">
                <td><strong>Error</strong></td>
                <td class="text-danger">{{ smtpResult.error }}</td>
              </tr>
            </table>
            <div v-if="smtpResult.recommendations && smtpResult.recommendations.length" class="mt-3">
              <strong>Recommendations:</strong>
              <ul class="mb-0 mt-1">
                <li v-for="(rec, i) in smtpResult.recommendations" :key="i">{{ rec }}</li>
              </ul>
            </div>
          </div>
        </div>

        <div v-if="testResult" class="alert mt-3" :class="testResult.success ? 'alert-success' : 'alert-danger'">
          {{ testResult.message }}
        </div>

        <div v-if="saveMessage" class="alert mt-3" :class="saveSuccess ? 'alert-success' : 'alert-danger'">
          {{ saveMessage }}
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import Api from '@/API'

const loading = ref(true)
const saving = ref(false)
const testing = ref(false)
const testingSmtp = ref(false)
const testResult = ref(null)
const smtpResult = ref(null)
const saveMessage = ref('')
const saveSuccess = ref(false)

const digest = reactive({
  digest_enabled: false,
  digest_emails: '',
  digest_hour: 8
})

onMounted(async () => {
  try {
    const settings = await Api.digest()
    Object.assign(digest, settings)
  } catch (e) {
    console.error('Failed to load digest settings:', e)
  }
  loading.value = false
})

function formatHour(h) {
  if (h === 0) return '12:00 AM (Midnight)'
  if (h === 12) return '12:00 PM (Noon)'
  if (h < 12) return `${h}:00 AM`
  return `${h - 12}:00 PM`
}

async function testDigest() {
  testing.value = true
  testResult.value = null
  saveMessage.value = ''

  try {
    // Save first to ensure settings are current
    await Api.digest_save(digest)
    const result = await Api.digest_test()
    testResult.value = result
  } catch (e) {
    testResult.value = { success: false, message: e.message || 'Failed to send test digest' }
  }

  testing.value = false
}

async function save() {
  saving.value = true
  saveMessage.value = ''
  testResult.value = null

  try {
    await Api.digest_save(digest)
    saveMessage.value = 'Digest settings saved successfully'
    saveSuccess.value = true
  } catch (e) {
    saveMessage.value = e.message || 'Failed to save settings'
    saveSuccess.value = false
  }

  saving.value = false
}

async function testSmtp() {
  testingSmtp.value = true
  smtpResult.value = null
  testResult.value = null
  saveMessage.value = ''

  try {
    smtpResult.value = await Api.digest_smtp_test()
  } catch (e) {
    smtpResult.value = { connected: false, error: e.message || 'Failed to test SMTP' }
  }

  testingSmtp.value = false
}
</script>
