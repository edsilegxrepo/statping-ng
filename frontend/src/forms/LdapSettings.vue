<template>
  <div class="card">
    <div class="card-header">
      LDAP / Active Directory Authentication
    </div>
    <div class="card-body">
      <div v-if="loading" class="text-center py-4">
        <font-awesome-icon icon="circle-notch" spin size="2x" />
      </div>

      <form v-else @submit.prevent="save">
        <div class="form-group row">
          <label class="col-sm-4 col-form-label">Enable LDAP</label>
          <div class="col-sm-8">
            <div class="switch">
              <input v-model="ldap.ldap_enabled" type="checkbox" id="ldap_enabled" />
              <label for="ldap_enabled">Enable LDAP authentication</label>
            </div>
          </div>
        </div>

        <div class="form-group row">
          <label class="col-sm-4 col-form-label">Template</label>
          <div class="col-sm-8">
            <select v-model="ldap.ldap_template" @change="applyTemplate" class="form-control">
              <option value="">Custom</option>
              <option value="openldap">OpenLDAP</option>
              <option value="activedirectory">Microsoft Active Directory</option>
              <option value="freeipa">FreeIPA</option>
            </select>
            <small class="form-text text-muted">Select a template to auto-fill common settings</small>
          </div>
        </div>

        <hr />
        <h6 class="text-muted mb-3">Connection Settings</h6>

        <div class="form-group row">
          <label class="col-sm-4 col-form-label">LDAP Host</label>
          <div class="col-sm-8">
            <input v-model="ldap.ldap_host" type="text" class="form-control" placeholder="ldap.example.com" />
          </div>
        </div>

        <div class="form-group row">
          <label class="col-sm-4 col-form-label">LDAP Port</label>
          <div class="col-sm-8">
            <input v-model.number="ldap.ldap_port" type="number" class="form-control" placeholder="636" />
            <small class="form-text text-muted">
              LDAPS (implicit TLS): 636 or 3269. StartTLS: 389 or 3268
            </small>
          </div>
        </div>

        <div class="form-group row">
          <label class="col-sm-4 col-form-label">Use StartTLS</label>
          <div class="col-sm-8">
            <div class="switch">
              <input v-model="ldap.ldap_start_tls" type="checkbox" id="ldap_start_tls" />
              <label for="ldap_start_tls">Upgrade plain connection to TLS (for ports 389/3268)</label>
            </div>
            <small class="form-text text-muted">
              Ports 636/3269 use implicit TLS automatically - StartTLS is ignored for these ports
            </small>
          </div>
        </div>

        <div class="form-group row">
          <label class="col-sm-4 col-form-label">Skip TLS Verify</label>
          <div class="col-sm-8">
            <div class="switch">
              <input v-model="ldap.ldap_skip_verify" type="checkbox" id="ldap_skip_verify" />
              <label for="ldap_skip_verify">Skip certificate verification (not recommended for production)</label>
            </div>
          </div>
        </div>

        <hr />
        <h6 class="text-muted mb-3">Service Account (Bind DN)</h6>

        <div class="form-group row">
          <label class="col-sm-4 col-form-label">Bind DN</label>
          <div class="col-sm-8">
            <input v-model="ldap.ldap_bind_dn" type="text" class="form-control" placeholder="cn=service,dc=example,dc=com" />
            <small class="form-text text-muted">Service account DN for searching users</small>
          </div>
        </div>

        <div class="form-group row">
          <label class="col-sm-4 col-form-label">Bind Password</label>
          <div class="col-sm-8">
            <input v-model="ldap.ldap_bind_password" type="password" class="form-control" placeholder="Leave blank to keep current" />
          </div>
        </div>

        <hr />
        <h6 class="text-muted mb-3">User Search Settings</h6>

        <div class="form-group row">
          <label class="col-sm-4 col-form-label">Base DN</label>
          <div class="col-sm-8">
            <input v-model="ldap.ldap_base_dn" type="text" class="form-control" placeholder="dc=example,dc=com" />
            <small class="form-text text-muted">Base DN for user searches</small>
          </div>
        </div>

        <div class="form-group row">
          <label class="col-sm-4 col-form-label">User Filter</label>
          <div class="col-sm-8">
            <input v-model="ldap.ldap_user_filter" type="text" class="form-control" placeholder="(&(objectClass=user)(sAMAccountName=%s))" />
            <small class="form-text text-muted">LDAP filter to find users. Use %s as username placeholder</small>
          </div>
        </div>

        <div class="form-group row">
          <label class="col-sm-4 col-form-label">Username Attribute</label>
          <div class="col-sm-8">
            <input v-model="ldap.ldap_username_attr" type="text" class="form-control" placeholder="sAMAccountName or uid" />
          </div>
        </div>

        <div class="form-group row">
          <label class="col-sm-4 col-form-label">Email Attribute</label>
          <div class="col-sm-8">
            <input v-model="ldap.ldap_email_attr" type="text" class="form-control" placeholder="mail" />
          </div>
        </div>

        <hr />
        <h6 class="text-muted mb-3">Authorization</h6>

        <div class="form-group row">
          <label class="col-sm-4 col-form-label">Authorized Group</label>
          <div class="col-sm-8">
            <div class="switch mb-2">
              <input v-model="ldap.ldap_authorized_group_enabled" type="checkbox" id="ldap_authorized_group_enabled" />
              <label for="ldap_authorized_group_enabled">Require group membership to access application</label>
            </div>
            <input
              v-model="ldap.ldap_authorized_group"
              type="text"
              class="form-control"
              :disabled="!ldap.ldap_authorized_group_enabled"
              placeholder="CN=StatpingUsers,OU=Groups,DC=example,DC=com"
            />
            <small class="form-text text-muted">
              Users must be a member of this group to log in. Multiple groups can be comma-separated (user needs to be in ANY).
              Leave disabled to allow any authenticated user.
            </small>
          </div>
        </div>

        <div class="alert alert-info">
          <strong>Note:</strong> LDAP users are auto-provisioned on first login but are <strong>disabled by default</strong>.
          An administrator must enable each user in the Users section before they can access the application.
          Role assignment (admin/user) is managed within the application, not via LDAP groups.
        </div>

        <hr />

        <div class="form-group row">
          <div class="col-sm-4"></div>
          <div class="col-sm-8">
            <button type="button" @click="testConnection" :disabled="testing" class="btn btn-outline-primary mr-2">
              <font-awesome-icon v-if="testing" icon="circle-notch" spin />
              <font-awesome-icon v-else icon="plug" />
              Test Connection
            </button>
            <button type="submit" :disabled="saving" class="btn btn-success">
              <font-awesome-icon v-if="saving" icon="circle-notch" spin />
              Save Settings
            </button>
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
const testResult = ref(null)
const saveMessage = ref('')
const saveSuccess = ref(false)

const ldap = reactive({
  ldap_enabled: false,
  ldap_host: '',
  ldap_port: 636,
  ldap_start_tls: false,
  ldap_skip_verify: false,
  ldap_bind_dn: '',
  ldap_bind_password: '',
  ldap_base_dn: '',
  ldap_user_filter: '',
  ldap_username_attr: '',
  ldap_email_attr: '',
  ldap_authorized_group_enabled: false,
  ldap_authorized_group: '',
  ldap_template: ''
})

const templates = {
  openldap: {
    ldap_user_filter: '(&(objectClass=inetOrgPerson)(uid=%s))',
    ldap_username_attr: 'uid',
    ldap_email_attr: 'mail'
  },
  activedirectory: {
    ldap_user_filter: '(&(objectClass=user)(sAMAccountName=%s))',
    ldap_username_attr: 'sAMAccountName',
    ldap_email_attr: 'mail'
  },
  freeipa: {
    ldap_user_filter: '(&(objectClass=person)(uid=%s))',
    ldap_username_attr: 'uid',
    ldap_email_attr: 'mail'
  }
}

onMounted(async () => {
  try {
    const settings = await Api.ldap()
    Object.assign(ldap, settings)
  } catch (e) {
    console.error('Failed to load LDAP settings:', e)
  }
  loading.value = false
})

function applyTemplate() {
  const template = templates[ldap.ldap_template]
  if (template) {
    Object.assign(ldap, template)
  }
}

async function testConnection() {
  testing.value = true
  testResult.value = null
  saveMessage.value = ''

  try {
    const result = await Api.ldap_test({
      ldap_host: ldap.ldap_host,
      ldap_port: ldap.ldap_port,
      ldap_start_tls: ldap.ldap_start_tls,
      ldap_skip_verify: ldap.ldap_skip_verify,
      ldap_bind_dn: ldap.ldap_bind_dn,
      ldap_bind_password: ldap.ldap_bind_password,
      ldap_base_dn: ldap.ldap_base_dn
    })
    testResult.value = result
  } catch (e) {
    testResult.value = { success: false, message: e.message || 'Connection test failed' }
  }

  testing.value = false
}

async function save() {
  saving.value = true
  saveMessage.value = ''
  testResult.value = null

  try {
    await Api.ldap_save(ldap)
    saveMessage.value = 'LDAP settings saved successfully'
    saveSuccess.value = true
  } catch (e) {
    saveMessage.value = e.message || 'Failed to save settings'
    saveSuccess.value = false
  }

  saving.value = false
}
</script>
