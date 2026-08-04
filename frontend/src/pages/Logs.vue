<template>
  <div class="logs-page">
    <section class="page-section">
      <div class="section-header">
        <div class="section-title">
          <font-awesome-icon icon="file-alt" class="section-icon" />
          <h2>Application Logs</h2>
        </div>
        <div class="toolbar">
          <button @click="exportLogs('json')" class="toolbar-btn" title="Export JSON">
            <font-awesome-icon icon="file-code" />
          </button>
          <button @click="exportLogs('tsv')" class="toolbar-btn" title="Export TSV">
            <font-awesome-icon icon="file-csv" />
          </button>
          <div class="date-filter">
            <label class="date-label">From:</label>
            <input v-model="dateFrom" type="date" class="date-input" />
            <label class="date-label">To:</label>
            <input v-model="dateTo" type="date" class="date-input" />
            <button v-if="dateFrom || dateTo" @click="clearDates" class="clear-dates-btn" title="Clear dates">
              <font-awesome-icon icon="times" />
            </button>
          </div>
          <div class="search-box">
            <font-awesome-icon icon="search" class="search-icon" />
            <input v-model="search" type="text" placeholder="Search logs..." class="search-input" />
          </div>
        </div>
      </div>

      <div class="section-card">
        <div v-if="loading" class="loading-state">
          <font-awesome-icon icon="circle-notch" spin class="loading-icon" />
          <span>Loading logs...</span>
        </div>

        <div v-else-if="filteredLogs.length === 0" class="loading-state">
          <span>{{ logs.length === 0 ? 'No logs available' : 'No logs match the current filters' }}</span>
        </div>

        <template v-else>
          <div class="logs-container">
            <div
              v-for="log in paginatedLogs"
              :key="log.id"
              class="log-entry"
              :class="log.level"
            >
              <span class="log-time">{{ log.time }}</span>
              <span class="log-message">{{ log.message }}</span>
            </div>
          </div>

          <div class="pagination-bar">
            <span class="pagination-info">
              Showing {{ paginationStart + 1 }}-{{ paginationEnd }} of {{ filteredLogs.length }} logs
            </span>
            <div class="pagination-controls">
              <button @click="currentPage = 1" :disabled="currentPage === 1" class="page-btn" title="First">
                <font-awesome-icon icon="angle-double-left" />
              </button>
              <button @click="currentPage--" :disabled="currentPage === 1" class="page-btn" title="Previous">
                <font-awesome-icon icon="angle-left" />
              </button>
              <span class="page-number">Page {{ currentPage }} / {{ totalPages }}</span>
              <button @click="currentPage++" :disabled="currentPage >= totalPages" class="page-btn" title="Next">
                <font-awesome-icon icon="angle-right" />
              </button>
              <button @click="currentPage = totalPages" :disabled="currentPage >= totalPages" class="page-btn" title="Last">
                <font-awesome-icon icon="angle-double-right" />
              </button>
              <select v-model.number="pageSize" class="page-size-select">
                <option :value="50">50</option>
                <option :value="100">100</option>
                <option :value="250">250</option>
              </select>
            </div>
          </div>
        </template>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import Api from '../API'

const logs = ref([])
const loading = ref(true)
const search = ref('')
const dateFrom = ref('')
const dateTo = ref('')
const currentPage = ref(1)
const pageSize = ref(50)
let lastLogId = ''
let pollInterval = null
let logIdCounter = 0

function parseLog(line) {
  if (!line || typeof line !== 'string') return null
  const trimmed = line.trim()
  if (!trimmed) return null

  // Format: "2026-08-03 23:31:58: message here"
  const colonIdx = trimmed.indexOf(': ')
  if (colonIdx < 19) return null // timestamp is 19 chars

  const time = trimmed.substring(0, 19)
  const message = trimmed.substring(21)
  if (!message) return null

  let level = ''
  const lower = message.toLowerCase()
  if (lower.includes('error') || lower.includes('fatal')) level = 'log-error'
  else if (lower.includes('warn')) level = 'log-warning'
  else if (lower.includes('debug')) level = 'log-debug'

  return { id: ++logIdCounter, time, message, level }
}

const filteredLogs = computed(() => {
  let result = logs.value

  if (search.value) {
    const term = search.value.toLowerCase()
    result = result.filter(log => log.message.toLowerCase().includes(term))
  }

  if (dateFrom.value) {
    result = result.filter(log => log.time >= dateFrom.value)
  }

  if (dateTo.value) {
    const endDate = dateTo.value + ' 23:59:59'
    result = result.filter(log => log.time <= endDate)
  }

  return result
})

const totalPages = computed(() => Math.ceil(filteredLogs.value.length / pageSize.value) || 1)
const paginationStart = computed(() => (currentPage.value - 1) * pageSize.value)
const paginationEnd = computed(() => Math.min(currentPage.value * pageSize.value, filteredLogs.value.length))
const paginatedLogs = computed(() => filteredLogs.value.slice(paginationStart.value, paginationEnd.value))

function resetPage() {
  currentPage.value = 1
}

async function loadLogs() {
  try {
    const data = await Api.logs()
    if (!Array.isArray(data)) return

    // Backend already returns newest-first, just parse in order
    const parsed = []
    for (const line of data) {
      const log = parseLog(line)
      if (log) parsed.push(log)
    }
    logs.value = parsed
    if (parsed.length > 0) {
      lastLogId = parsed[0].time + parsed[0].message
    }
  } catch (e) {
    console.error('Failed to load logs:', e)
  } finally {
    loading.value = false
  }
}

async function pollLastLog() {
  try {
    const data = await Api.logs_last()
    const log = parseLog(data)
    if (!log) return

    const logId = log.time + log.message
    if (logId !== lastLogId) {
      lastLogId = logId
      logs.value.unshift(log)
    }
  } catch (e) {
    // Silently ignore polling errors
  }
}


function clearDates() {
  dateFrom.value = ''
  dateTo.value = ''
}

function exportLogs(format) {
  const data = filteredLogs.value
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19)

  let content, filename, mimeType
  if (format === 'json') {
    content = JSON.stringify(data.map(log => ({ time: log.time, message: log.message })), null, 2)
    filename = `logs-${timestamp}.json`
    mimeType = 'application/json'
  } else {
    content = 'Time\tMessage\n' + data.map(log => `${log.time}\t${log.message.replace(/\t/g, ' ')}`).join('\n')
    filename = `logs-${timestamp}.tsv`
    mimeType = 'text/tab-separated-values'
  }

  const blob = new Blob([content], { type: mimeType })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

// Watch for filter changes to reset page
import { watch } from 'vue'
watch([search, dateFrom, dateTo, pageSize], resetPage)

onMounted(async () => {
  await loadLogs()
  pollInterval = setInterval(pollLastLog, 2000)
})

onBeforeUnmount(() => {
  if (pollInterval) clearInterval(pollInterval)
})
</script>

<style scoped>
.logs-page {
  max-width: 100%;
}

.page-section {
  margin-bottom: var(--space-6);
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-4);
  flex-wrap: wrap;
  gap: var(--space-3);
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

.toolbar {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  flex-wrap: wrap;
}

.export-dropdown {
  position: relative;
}

.toolbar-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  background: #fff;
  border: 1px solid var(--color-gray-300);
  border-radius: var(--radius-md);
  color: var(--color-gray-600);
  cursor: pointer;
  transition: all 0.15s;
}

.toolbar-btn:hover {
  background: var(--color-gray-100);
  color: var(--color-gray-800);
}

.dropdown-menu {
  position: absolute;
  top: 100%;
  left: 0;
  margin-top: 4px;
  background: #fff;
  border: 1px solid var(--color-gray-200);
  border-radius: var(--radius-md);
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  z-index: 1000;
  min-width: 150px;
}

.dropdown-menu button {
  display: block;
  width: 100%;
  padding: 10px 14px;
  text-align: left;
  background: none;
  border: none;
  font-size: 0.875rem;
  color: var(--color-gray-700);
  cursor: pointer;
}

.dropdown-menu button:hover {
  background: var(--color-gray-100);
}

.date-filter {
  display: flex;
  align-items: center;
  gap: 6px;
}

.date-label {
  color: var(--color-gray-600);
  font-size: 0.8rem;
  font-weight: 500;
}

.date-input {
  padding: 6px 8px;
  font-size: 0.8rem;
  border: 1px solid var(--color-gray-300);
  border-radius: var(--radius-md);
  background: #fff;
  width: 130px;
}

.date-input:focus {
  outline: none;
  border-color: var(--color-primary);
}

.clear-dates-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  background: none;
  border: none;
  color: var(--color-gray-400);
  cursor: pointer;
  border-radius: var(--radius-sm);
}

.clear-dates-btn:hover {
  background: var(--color-gray-100);
  color: var(--color-gray-600);
}

.search-box {
  position: relative;
  width: 350px;
}

.search-icon {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--color-gray-400);
  font-size: 0.875rem;
}

.search-input {
  width: 100%;
  padding: 8px 12px 8px 36px;
  font-size: 0.875rem;
  border: 1px solid var(--color-gray-300);
  border-radius: var(--radius-md);
  background: #fff;
}

.search-input:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px var(--color-primary-bg);
}

.section-card {
  background: #1a1a2e;
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-lg);
  overflow: hidden;
}

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 48px;
  color: #888;
}

.loading-icon {
  font-size: 1.25rem;
}

.logs-container {
  max-height: calc(100vh - 230px);
  overflow-y: auto;
  padding: 8px;
  margin-bottom: 8px;
}

.log-entry {
  display: flex;
  padding: 2px 12px;
  line-height: 1.4;
}

.log-entry:hover {
  background: rgba(255, 255, 255, 0.05);
}

.log-time {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 0.75rem;
  color: #666;
  white-space: nowrap;
  margin-right: 16px;
  min-width: 145px;
}

.log-message {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 0.8rem;
  color: #ccc;
  word-break: break-word;
}

.log-error .log-message { color: #ff6b6b; }
.log-warning .log-message { color: #ffd93d; }
.log-debug .log-message { color: #888; }

.logs-container::-webkit-scrollbar { width: 8px; }
.logs-container::-webkit-scrollbar-track { background: #2a2a3e; }
.logs-container::-webkit-scrollbar-thumb { background: #444; border-radius: 4px; }
.logs-container::-webkit-scrollbar-thumb:hover { background: #555; }

.pagination-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  margin-top: 8px;
  background: #16162a;
  border-top: 1px solid #333;
}

.pagination-info {
  font-size: 0.8rem;
  color: #888;
}

.pagination-controls {
  display: flex;
  align-items: center;
  gap: 8px;
}

.page-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: #2a2a3e;
  border: 1px solid #444;
  border-radius: 4px;
  color: #ccc;
  cursor: pointer;
}

.page-btn:hover:not(:disabled) {
  background: #3a3a4e;
  color: #fff;
}

.page-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.page-number {
  font-size: 0.8rem;
  color: #ccc;
  padding: 0 8px;
  min-width: 90px;
  text-align: center;
}

.page-size-select {
  padding: 4px 8px;
  font-size: 0.8rem;
  background: #2a2a3e;
  border: 1px solid #444;
  border-radius: 4px;
  color: #ccc;
  cursor: pointer;
  margin-left: 8px;
}

@media (max-width: 768px) {
  .section-header {
    flex-direction: column;
    align-items: flex-start;
  }
  .toolbar {
    width: 100%;
  }
  .search-box {
    width: 100%;
  }
  .log-entry {
    flex-direction: column;
    gap: 4px;
  }
  .log-time {
    min-width: auto;
  }
}
</style>
