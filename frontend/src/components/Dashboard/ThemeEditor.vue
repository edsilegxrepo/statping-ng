<template>
  <div class="card mb-5">
    <div class="card-header">{{ $t('theme_editor') }}</div>
    <div class="card-body">
      <div v-if="error" class="alert alert-danger mt-3" style="white-space: pre-line">{{ error }}</div>
      <div v-if="success" class="alert alert-success mt-3">{{ success }}</div>

      <p class="text-muted mb-4">
        Add custom CSS to override the default theme. Changes are applied immediately after saving.
      </p>

      <form @submit.prevent="saveCSS">
        <h5>Custom CSS</h5>
        <textarea
          v-model="css"
          class="form-control code-editor"
          rows="30"
          :placeholder="placeholderCSS"
        ></textarea>
      </form>
    </div>

    <div class="card-footer">
      <div class="row">
        <div class="col-6">
          <button
            id="save_css"
            @click.prevent="saveCSS"
            type="submit"
            class="btn btn-primary btn-block"
            :disabled="pending"
          >
            {{ pending ? 'Saving...' : 'Save CSS' }}
          </button>
        </div>
        <div class="col-6">
          <button
            id="delete_css"
            @click.prevent="deleteCSS"
            class="btn btn-danger btn-block"
            :disabled="pending || !hasCSS"
          >
            Delete Custom CSS
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import Api from '@/API'

const css = ref('')
const error = ref(null)
const success = ref(null)
const loaded = ref(false)
const pending = ref(false)
const hasCSS = ref(false)

const placeholderCSS = `/* ============================================
   STATPING CUSTOM CSS EXAMPLES
   ============================================
   Copy any examples below and modify as needed.
   Use !important to override default styles.
   ============================================ */


/* --------------------------------------------
   BRAND COLORS & VARIABLES
   -------------------------------------------- */

:root {
  --primary-color: #007bff;
  --success-color: #28a745;
  --danger-color: #dc3545;
  --warning-color: #ffc107;
  --dark-bg: #1a1a2e;
  --card-shadow: 0 4px 6px rgba(0,0,0,0.1);
}


/* --------------------------------------------
   NAVIGATION BAR
   -------------------------------------------- */

/* Dark navigation bar */
.navbar {
  background-color: var(--dark-bg) !important;
  border-bottom: 2px solid #007bff;
}

/* Navigation links */
.navbar .nav-link {
  color: #ffffff !important;
  font-weight: 500;
}

.navbar .nav-link:hover {
  color: #007bff !important;
}


/* --------------------------------------------
   STATUS PAGE HEADER
   -------------------------------------------- */

/* Top status banner */
.status-banner {
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
  padding: 20px;
  border-radius: 8px;
}

/* Page title */
h1.display-4, .page-title {
  color: #333;
  font-weight: 700;
}


/* --------------------------------------------
   SERVICE CARDS
   -------------------------------------------- */

/* All service cards */
.card {
  border-radius: 12px;
  box-shadow: var(--card-shadow);
  border: none;
  transition: transform 0.2s ease;
}

.card:hover {
  transform: translateY(-2px);
}

/* Online service indicator */
.service-online {
  border-left: 4px solid var(--success-color) !important;
}

/* Offline service indicator */
.service-offline {
  border-left: 4px solid var(--danger-color) !important;
}

/* Service name */
.card-title {
  font-weight: 600;
  color: #2c3e50;
}


/* --------------------------------------------
   STATUS BADGES
   -------------------------------------------- */

.badge-success {
  background-color: var(--success-color) !important;
}

.badge-danger {
  background-color: var(--danger-color) !important;
}

.badge-warning {
  background-color: var(--warning-color) !important;
  color: #000 !important;
}


/* --------------------------------------------
   UPTIME CHART
   -------------------------------------------- */

/* Chart container */
.uptime-chart {
  background: #f8f9fa;
  border-radius: 8px;
  padding: 15px;
}

/* Chart bars - online */
.chart-bar.online {
  background-color: var(--success-color);
}

/* Chart bars - offline */
.chart-bar.offline {
  background-color: var(--danger-color);
}


/* --------------------------------------------
   BUTTONS
   -------------------------------------------- */

.btn-primary {
  background: linear-gradient(135deg, #007bff 0%, #0056b3 100%);
  border: none;
  border-radius: 6px;
  font-weight: 500;
}

.btn-primary:hover {
  background: linear-gradient(135deg, #0056b3 0%, #003d80 100%);
}

.btn-danger {
  background: linear-gradient(135deg, #dc3545 0%, #a71d2a 100%);
  border: none;
}


/* --------------------------------------------
   FOOTER
   -------------------------------------------- */

footer {
  background-color: #f8f9fa;
  border-top: 1px solid #e9ecef;
  padding: 20px 0;
}


/* --------------------------------------------
   DARK MODE (if you want full dark theme)
   -------------------------------------------- */

/*
body {
  background-color: #121212 !important;
  color: #e0e0e0 !important;
}

.card {
  background-color: #1e1e1e !important;
  color: #e0e0e0 !important;
}

.navbar {
  background-color: #1e1e1e !important;
}
*/


/* --------------------------------------------
   RESPONSIVE / MOBILE
   -------------------------------------------- */

@media (max-width: 768px) {
  .card {
    margin-bottom: 15px;
  }

  h1.display-4 {
    font-size: 1.8rem;
  }
}


/* --------------------------------------------
   ANIMATIONS
   -------------------------------------------- */

/* Pulse animation for status indicators */
@keyframes pulse {
  0% { opacity: 1; }
  50% { opacity: 0.5; }
  100% { opacity: 1; }
}

.status-indicator.online {
  animation: pulse 2s infinite;
}
`

onMounted(async () => {
  await fetchTheme()
})

async function fetchTheme() {
  pending.value = true
  try {
    const theme = await Api.theme()
    css.value = theme.css || ''
    hasCSS.value = theme.enabled || false
  } catch (e) {
    error.value = e.response?.data?.error || e.message
  }
  pending.value = false
  loaded.value = true
}

async function saveCSS() {
  pending.value = true
  error.value = null
  success.value = null

  try {
    await Api.theme_save({ css: css.value })
    success.value = 'Custom CSS saved successfully'
    hasCSS.value = css.value.trim().length > 0
    // Clear success message after 3 seconds
    setTimeout(() => { success.value = null }, 3000)
  } catch (e) {
    error.value = e.response?.data?.error || e.message
  }
  pending.value = false
}

async function deleteCSS() {
  pending.value = true
  error.value = null
  success.value = null

  try {
    await Api.theme_generate(false) // false = delete
    css.value = ''
    hasCSS.value = false
    success.value = 'Custom CSS deleted'
    setTimeout(() => { success.value = null }, 3000)
  } catch (e) {
    error.value = e.response?.data?.error || e.message
  }
  pending.value = false
}
</script>

<style scoped>
.code-editor {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  background-color: #1e1e1e;
  color: #d4d4d4;
  border: 1px solid #333;
  padding: 12px;
}

.code-editor::placeholder {
  color: #6a6a6a;
}
</style>
