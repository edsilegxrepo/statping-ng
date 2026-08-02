<template>
  <div class="settings-container">
    <!-- Sidebar -->
    <aside class="settings-sidebar">
      <div class="sidebar-section">
        <h6 class="sidebar-title">{{ $t('main_settings') }}</h6>
        <nav class="sidebar-nav">
          <a
            href="#"
            class="sidebar-link"
            :class="{ active: isActive('v-pills-home-tab') }"
            @click.prevent="changeTab('v-pills-home-tab')"
          >
            <font-awesome-icon icon="cog" class="sidebar-icon" />
            <span>{{ $t('settings') }}</span>
          </a>
          <a
            href="#"
            class="sidebar-link"
            :class="{ active: isActive('v-pills-style-tab') }"
            @click.prevent="changeTab('v-pills-style-tab')"
          >
            <font-awesome-icon icon="palette" class="sidebar-icon" />
            <span>{{ $t('theme') }}</span>
          </a>
          <a
            href="#"
            class="sidebar-link"
            :class="{ active: isActive('v-pills-oauth-tab') }"
            @click.prevent="changeTab('v-pills-oauth-tab')"
          >
            <font-awesome-icon icon="key" class="sidebar-icon" />
            <span>{{ $t('authentication') }}</span>
          </a>
          <a
            href="#"
            class="sidebar-link"
            :class="{ active: isActive('v-pills-ldap-tab') }"
            @click.prevent="changeTab('v-pills-ldap-tab')"
          >
            <font-awesome-icon icon="sitemap" class="sidebar-icon" />
            <span>LDAP</span>
          </a>
          <a
            v-if="coreData.allow_reports !== false"
            href="#"
            class="sidebar-link"
            :class="{ active: isActive('v-pills-digest-tab') }"
            @click.prevent="changeTab('v-pills-digest-tab')"
          >
            <font-awesome-icon icon="envelope" class="sidebar-icon" />
            <span>Daily Digest</span>
          </a>
          <a
            href="#"
            class="sidebar-link"
            :class="{ active: isActive('v-pills-logship-tab') }"
            @click.prevent="changeTab('v-pills-logship-tab')"
          >
            <font-awesome-icon icon="file-export" class="sidebar-icon" />
            <span>Log Shipping</span>
          </a>
          <a
            href="#"
            class="sidebar-link"
            :class="{ active: isActive('v-pills-import-tab') }"
            @click.prevent="changeTab('v-pills-import-tab')"
          >
            <font-awesome-icon icon="cloud-download-alt" class="sidebar-icon" />
            <span>{{ $t('import') }}</span>
          </a>
          <a
            href="#"
            class="sidebar-link"
            :class="{ active: isActive('v-pills-configs-tab') }"
            @click.prevent="changeTab('v-pills-configs-tab')"
          >
            <font-awesome-icon icon="cogs" class="sidebar-icon" />
            <span>{{ $t('configs') }}</span>
          </a>
        </nav>
      </div>

      <div class="sidebar-section">
        <h6 class="sidebar-title">{{ $t('notifiers') }}</h6>
        <nav class="sidebar-nav">
          <a
            v-for="notifier in notifiers"
            :key="notifier.method"
            href="#"
            class="sidebar-link"
            :class="{ active: isActive(`v-pills-${notifier.method.toLowerCase()}-tab`) }"
            @click.prevent="changeTab(`v-pills-${notifier.method.toLowerCase()}-tab`)"
          >
            <font-awesome-icon :icon="iconName(notifier.icon)" class="sidebar-icon" />
            <span>{{ notifier.title }}</span>
            <span v-if="notifier.enabled" class="sidebar-badge badge-success">ON</span>
          </a>
          <a
            href="#"
            class="sidebar-link"
            :class="{ active: isActive('v-pills-notifier-docs-tab') }"
            @click.prevent="changeTab('v-pills-notifier-docs-tab')"
          >
            <font-awesome-icon icon="book" class="sidebar-icon" />
            <span>{{ $t('variables') }}</span>
          </a>
        </nav>
      </div>
    </aside>

    <!-- Content -->
    <main class="settings-content">
      <!-- General Settings -->
      <div v-show="isActive('v-pills-home-tab')" class="settings-panel">
        <CoreSettings />

        <div class="panel-card mt-4">
          <div class="panel-header">
            <font-awesome-icon icon="key" class="me-2" />
            API {{ $t('settings') }}
          </div>
          <div class="panel-body">
            <div class="form-row">
              <label class="form-label">API {{ $t('secret') }}</label>
              <div class="input-with-action">
                <input
                  v-model="coreData.api_secret"
                  @focus="$event.target.select()"
                  type="text"
                  class="form-input"
                  readonly
                />
                <button @click="copySecret" class="input-action-btn" title="Copy to clipboard">
                  <font-awesome-icon icon="copy" />
                </button>
              </div>
              <small class="form-hint">{{ $t('regen_desc') }}</small>
            </div>
          </div>
          <div class="panel-footer">
            <button @click="renewApiKeys" class="btn-danger-outline">
              <font-awesome-icon icon="sync" class="me-2" />
              {{ $t('regen_api') }}
            </button>
          </div>
        </div>
      </div>

      <!-- Theme -->
      <div v-show="isActive('v-pills-style-tab')" class="settings-panel">
        <ThemeEditor />
      </div>

      <!-- OAuth -->
      <div v-show="isActive('v-pills-oauth-tab')" class="settings-panel">
        <OAuth />
      </div>

      <!-- LDAP -->
      <div v-show="isActive('v-pills-ldap-tab')" class="settings-panel">
        <LdapSettings />
      </div>

      <!-- Daily Digest -->
      <div v-if="coreData.allow_reports !== false" v-show="isActive('v-pills-digest-tab')" class="settings-panel">
        <DigestSettings />
      </div>

      <!-- Log Shipping -->
      <div v-show="isActive('v-pills-logship-tab')" class="settings-panel">
        <LogShipSettings />
      </div>

      <!-- Import -->
      <div v-show="isActive('v-pills-import-tab')" class="settings-panel">
        <Importer />
      </div>

      <!-- Configs -->
      <div v-show="isActive('v-pills-configs-tab')" class="settings-panel">
        <Configs />
      </div>

      <!-- Variables Docs -->
      <div v-show="isActive('v-pills-notifier-docs-tab')" class="settings-panel">
        <Variables />
      </div>

      <!-- Notifiers -->
      <div
        v-for="notifier in notifiers"
        :key="`${notifier.method}_panel`"
        v-show="isActive(`v-pills-${notifier.method.toLowerCase()}-tab`)"
        class="settings-panel"
      >
        <Notifier :notifier="notifier" />
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useMainStore } from '@/stores/main'
import { useCookies } from 'vue3-cookies'
import Api from '@/API'
import CoreSettings from '@/forms/CoreSettings.vue'
import Notifier from '@/forms/Notifier.vue'
import OAuth from '@/forms/OAuth.vue'
import LdapSettings from '@/forms/LdapSettings.vue'
import DigestSettings from '@/forms/DigestSettings.vue'
import LogShipSettings from '@/forms/LogShipSettings.vue'
import ThemeEditor from '@/components/Dashboard/ThemeEditor.vue'
import Importer from '@/components/Dashboard/Importer.vue'
import Variables from '@/components/Dashboard/Variables.vue'
import Configs from '@/components/Dashboard/Configs.vue'

const router = useRouter()
const store = useMainStore()
const { cookies } = useCookies()

const tab = ref('v-pills-home-tab')

const coreData = computed(() => store.core)
const notifiers = computed(() => store.notifiers)

onMounted(() => {
  // GitHub version check removed (update command depended on defunct statping.com)
})

function changeTab(tabId) {
  tab.value = tabId
}

function isActive(id) {
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
  let redirectUrl = null
  try {
    const response = await Api.logout()
    if (response.redirect) {
      redirectUrl = response.redirect
    }
  } catch (e) {
    console.error('Backend logout failed', e)
  }
  store.setHasAllData(false)
  store.setToken(null)
  store.setAdmin(false)
  store.setUser(false)
  store.setLoggedIn(false)
  cookies.remove('statping_auth')

  if (redirectUrl) {
    window.location.href = redirectUrl
  } else {
    await router.push('/logout')
  }
}
</script>

<style scoped>
.settings-container {
  display: flex;
  min-height: calc(100vh - 56px);
  background: var(--color-gray-50);
}

/* Sidebar */
.settings-sidebar {
  width: 280px;
  background: #fff;
  border-right: 1px solid var(--color-gray-200);
  padding: var(--space-4);
  flex-shrink: 0;
  overflow-y: auto;
  max-height: calc(100vh - 56px);
  position: sticky;
  top: 56px;
}

.sidebar-section {
  margin-bottom: var(--space-6);
}

.sidebar-title {
  font-size: 0.7rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--color-gray-400);
  margin: 0 0 var(--space-3);
  padding: 0 var(--space-3);
}

.sidebar-nav {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.sidebar-link {
  display: flex;
  align-items: center;
  padding: var(--space-2) var(--space-3);
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-gray-600);
  text-decoration: none;
  border-radius: var(--radius-md);
  transition: all var(--transition-fast);
}

.sidebar-link:hover {
  background: var(--color-gray-100);
  color: var(--color-gray-900);
}

.sidebar-link.active {
  background: var(--color-primary-bg);
  color: var(--color-primary);
}

.sidebar-icon {
  width: 18px;
  margin-right: var(--space-3);
  opacity: 0.7;
}

.sidebar-link.active .sidebar-icon {
  opacity: 1;
}

.sidebar-badge {
  margin-left: auto;
  font-size: 0.6rem;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: var(--radius-full);
}

.badge-success {
  background: var(--color-success-bg);
  color: var(--color-success-dark);
}

/* Content */
.settings-content {
  flex: 1;
  padding: var(--space-6);
  max-width: 800px;
  margin: 0 auto;
}

.settings-panel {
  animation: fadeIn 0.2s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Panel Card */
.panel-card {
  background: #fff;
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-md);
  overflow: hidden;
}

.panel-header {
  padding: var(--space-4) var(--space-5);
  font-weight: 600;
  color: var(--color-gray-800);
  border-bottom: 1px solid var(--color-gray-100);
  background: var(--color-gray-50);
}

.panel-body {
  padding: var(--space-5);
}

.panel-footer {
  padding: var(--space-4) var(--space-5);
  background: var(--color-gray-50);
  border-top: 1px solid var(--color-gray-100);
  display: flex;
  justify-content: flex-end;
}

/* Form Elements */
.form-row {
  margin-bottom: var(--space-4);
}

.form-label {
  display: block;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-gray-700);
  margin-bottom: var(--space-2);
}

.form-input {
  width: 100%;
  padding: var(--space-2) var(--space-3);
  font-size: 0.875rem;
  border: 1px solid var(--color-gray-300);
  border-radius: var(--radius-md);
  background: #fff;
  transition: all var(--transition-fast);
}

.form-input:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px var(--color-primary-bg);
}

.form-hint {
  display: block;
  font-size: 0.8rem;
  color: var(--color-gray-500);
  margin-top: var(--space-1);
}

.input-with-action {
  display: flex;
  gap: var(--space-2);
}

.input-with-action .form-input {
  flex: 1;
}

.input-action-btn {
  padding: var(--space-2) var(--space-3);
  background: var(--color-gray-100);
  border: 1px solid var(--color-gray-300);
  border-radius: var(--radius-md);
  color: var(--color-gray-600);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.input-action-btn:hover {
  background: var(--color-gray-200);
  color: var(--color-gray-800);
}

/* Buttons */
.btn-danger-outline {
  display: inline-flex;
  align-items: center;
  padding: var(--space-2) var(--space-4);
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-danger);
  background: transparent;
  border: 1px solid var(--color-danger);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.btn-danger-outline:hover {
  background: var(--color-danger);
  color: #fff;
}

/* Responsive */
@media (max-width: 1024px) {
  .settings-sidebar {
    width: 240px;
  }
}

@media (max-width: 768px) {
  .settings-container {
    flex-direction: column;
  }

  .settings-sidebar {
    width: 100%;
    max-height: none;
    position: static;
    border-right: none;
    border-bottom: 1px solid var(--color-gray-200);
  }

  .sidebar-nav {
    flex-direction: row;
    flex-wrap: wrap;
    gap: var(--space-2);
  }

  .sidebar-link {
    flex: 0 0 auto;
  }

  .settings-content {
    padding: var(--space-4);
  }
}
</style>
