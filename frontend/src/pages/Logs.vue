<template>
  <div class="logs-page">
    <section class="page-section">
      <div class="section-header">
        <div class="section-title">
          <font-awesome-icon icon="file-alt" class="section-icon" />
          <h2>Application Logs</h2>
        </div>
        <div class="search-box">
          <font-awesome-icon icon="search" class="search-icon" />
          <input v-model="search" type="text" placeholder="Search logs..." class="search-input" />
        </div>
      </div>

      <div class="section-card">
        <div v-if="logs.length === 0" class="loading-state">
          <font-awesome-icon icon="circle-notch" spin class="loading-icon" />
          <span>Loading logs...</span>
        </div>

        <div v-else class="logs-container">
          <div v-for="(log, index) in logs" :key="index" class="log-entry" :class="getLogClass(log.message)">
            <span class="log-time">{{ log.time }}</span>
            <span class="log-message">{{ log.message }}</span>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script>
import Api from '../API'

export default {
  name: 'Logs',
  data() {
    return {
      logs_record: [],
      last: '',
      search: '',
      t: null,
    }
  },
  computed: {
    logs() {
      if (this.search) {
        return this.logs_record.filter((o) => o.message.toLowerCase().includes(this.search.toLowerCase()))
      } else {
        return this.logs_record
      }
    },
  },
  async created() {
    await this.getLogs()
    if (!this.t) {
      this.t = setInterval(async () => {
        await this.lastLog()
      }, 1000)
    }
  },
  beforeUnmount() {
    clearInterval(this.t)
  },
  methods: {
    getLogClass(message) {
      if (!message) return ''
      const lower = message.toLowerCase()
      if (lower.includes('error') || lower.includes('fatal')) return 'log-error'
      if (lower.includes('warn')) return 'log-warning'
      if (lower.includes('debug')) return 'log-debug'
      return ''
    },
    parseLog(data) {
      const ts = data.match(
        /[0-9]{4}-(0[1-9]|1[0-2])-(0[1-9]|[1-2][0-9]|3[0-1]) (2[0-3]|[01][0-9]):[0-5][0-9]:[0-5][0-9]/gm
      )
      return {
        time: ts ? ts[0] : '',
        message: ts ? data.split(`${ts}: `)[1] || data : data,
      }
    },
    cleanLog(l) {
      const splitLog = l.split(': ')
      const last = splitLog.slice(1)
      return last.join(': ')
    },
    async getLogs() {
      const l = await Api.logs()
      l.forEach((d) => {
        this.logs_record.push(this.parseLog(d))
      })
      this.last = this.cleanLog(l[l.length - 1])
    },
    async lastLog() {
      const log = await Api.logs_last()
      const cleanLast = this.cleanLog(log)
      if (this.last !== cleanLast) {
        this.last = cleanLast
        this.logs_record.unshift(this.parseLog(log))
      }
    },
  },
}
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

/* Search Box */
.search-box {
  position: relative;
  width: 300px;
}

.search-icon {
  position: absolute;
  left: var(--space-3);
  top: 50%;
  transform: translateY(-50%);
  color: var(--color-gray-400);
  font-size: 0.875rem;
}

.search-input {
  width: 100%;
  padding: var(--space-2) var(--space-3) var(--space-2) 36px;
  font-size: 0.875rem;
  border: 1px solid var(--color-gray-300);
  border-radius: var(--radius-md);
  background: #fff;
  transition: all var(--transition-fast);
}

.search-input:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px var(--color-primary-bg);
}

/* Section Card */
.section-card {
  background: var(--color-gray-900);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-lg);
  overflow: hidden;
}

/* Loading State */
.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-3);
  padding: var(--space-8);
  color: var(--color-gray-400);
}

.loading-icon {
  font-size: 1.25rem;
}

/* Logs Container */
.logs-container {
  max-height: calc(100vh - 250px);
  overflow-y: auto;
  padding: var(--space-2);
}

.log-entry {
  display: flex;
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-sm);
  transition: background var(--transition-fast);
}

.log-entry:hover {
  background: rgba(255, 255, 255, 0.05);
}

.log-time {
  font-family: var(--font-mono);
  font-size: 0.75rem;
  color: var(--color-gray-500);
  white-space: nowrap;
  margin-right: var(--space-4);
  min-width: 140px;
}

.log-message {
  font-family: var(--font-mono);
  font-size: 0.8rem;
  color: var(--color-gray-300);
  word-break: break-word;
}

/* Log Level Colors */
.log-error .log-message {
  color: var(--color-danger-light);
}

.log-warning .log-message {
  color: var(--color-warning-light);
}

.log-debug .log-message {
  color: var(--color-gray-500);
}

/* Scrollbar */
.logs-container::-webkit-scrollbar {
  width: 8px;
}

.logs-container::-webkit-scrollbar-track {
  background: var(--color-gray-800);
}

.logs-container::-webkit-scrollbar-thumb {
  background: var(--color-gray-600);
  border-radius: 4px;
}

.logs-container::-webkit-scrollbar-thumb:hover {
  background: var(--color-gray-500);
}

/* Responsive */
@media (max-width: 768px) {
  .logs-page {
    padding: var(--space-3);
  }

  .section-header {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-3);
  }

  .search-box {
    width: 100%;
  }

  .log-entry {
    flex-direction: column;
    gap: var(--space-1);
  }

  .log-time {
    min-width: auto;
  }
}
</style>
