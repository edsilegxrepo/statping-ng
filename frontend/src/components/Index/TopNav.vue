<template>
  <nav class="top-nav">
    <div class="top-nav-content">
      <div class="top-nav-left">
        <router-link to="/" class="top-nav-brand">
          <img v-if="core.footer" :src="core.footer" alt="Logo" class="top-nav-logo" />
          <span class="top-nav-title">{{ core.name }}</span>
        </router-link>
      </div>

      <div class="top-nav-center">
        <span class="status-summary" :class="statusClass">
          <span class="status-dot"></span>
          {{ onlineCount }}/{{ totalCount }} online
        </span>

        <span class="status-divider"></span>

        <span class="status-uptime" :class="uptimeClass">
          <font-awesome-icon icon="chart-line" class="me-1" />
          {{ uptimePercent }}% (24h)
        </span>

        <span class="status-divider"></span>

        <span class="status-lastcheck">
          <font-awesome-icon icon="sync-alt" class="me-1" />
          {{ lastCheckText }}
        </span>

        <span v-if="hasActiveIncidents" class="status-divider"></span>

        <span v-if="hasActiveIncidents" class="status-incident">
          <font-awesome-icon icon="exclamation-triangle" class="me-1" />
          {{ activeIncidentCount }} active
        </span>
      </div>

      <div class="top-nav-right">
        <router-link to="/dashboard" class="top-nav-btn">
          <font-awesome-icon icon="tachometer-alt" class="me-1" />
          Dashboard
        </router-link>
        <router-link to="/login" class="top-nav-btn top-nav-btn-login">
          <font-awesome-icon icon="sign-in-alt" class="me-1" />
          Login
        </router-link>
      </div>
    </div>
  </nav>
</template>

<script setup>
import { computed } from 'vue'
import { useMainStore } from '@/stores/main'

const store = useMainStore()

const core = computed(() => store.core)
const messages = computed(() => store.messages || [])

const services = computed(() => store.services || [])
const totalCount = computed(() => services.value.length)
const onlineCount = computed(() => services.value.filter(s => s.online).length)

const statusClass = computed(() => {
  if (totalCount.value === 0) return 'status-unknown'
  if (onlineCount.value === totalCount.value) return 'status-all-online'
  if (onlineCount.value === 0) return 'status-all-offline'
  return 'status-partial'
})

// 24h uptime percentage
const uptimePercent = computed(() => {
  if (services.value.length === 0) return '—'
  const total = services.value.reduce((sum, s) => sum + (s.online_24_hours || 0), 0)
  const avg = total / services.value.length
  return avg.toFixed(1)
})

const uptimeClass = computed(() => {
  const pct = parseFloat(uptimePercent.value)
  if (isNaN(pct)) return ''
  if (pct >= 99) return 'uptime-good'
  if (pct >= 95) return 'uptime-warn'
  return 'uptime-bad'
})

// Last check timestamp
const lastCheckText = computed(() => {
  if (services.value.length === 0) return '—'

  // Find the most recent check across all services
  let mostRecent = null
  for (const s of services.value) {
    if (s.last_success) {
      const d = new Date(s.last_success)
      if (!mostRecent || d > mostRecent) mostRecent = d
    }
  }

  if (!mostRecent) return '—'

  const now = new Date()
  const diffMs = now - mostRecent
  const diffSec = Math.floor(diffMs / 1000)
  const diffMin = Math.floor(diffSec / 60)
  const diffHour = Math.floor(diffMin / 60)

  if (diffSec < 60) return 'just now'
  if (diffMin < 60) return `${diffMin}m ago`
  if (diffHour < 24) return `${diffHour}h ago`
  return '>24h ago'
})

// Active incidents/messages
const activeIncidents = computed(() => {
  const now = new Date()
  return messages.value.filter(m => {
    const start = new Date(m.start_on)
    const end = m.start_on === m.end_on ? new Date(8640000000000000) : new Date(m.end_on)
    return now >= start && now <= end
  })
})

const hasActiveIncidents = computed(() => activeIncidents.value.length > 0)
const activeIncidentCount = computed(() => activeIncidents.value.length)
</script>

<style scoped>
.top-nav {
  position: sticky;
  top: 0;
  z-index: 1000;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  padding: 0 1rem;
  height: 48px;
}

.top-nav-content {
  max-width: 1400px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 100%;
}

.top-nav-left {
  display: flex;
  align-items: center;
}

.top-nav-brand {
  display: flex;
  align-items: center;
  text-decoration: none;
  color: #fff;
  font-weight: 600;
  font-size: 1rem;
}

.top-nav-brand:hover {
  color: #fff;
  opacity: 0.9;
}

.top-nav-logo {
  height: 28px;
  width: auto;
  margin-right: 0.5rem;
}

.top-nav-title {
  white-space: nowrap;
}

.top-nav-center {
  display: flex;
  align-items: center;
}

.status-summary {
  display: flex;
  align-items: center;
  font-size: 0.85rem;
  font-weight: 500;
  padding: 0.25rem 0.75rem;
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.1);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 0.5rem;
}

.status-all-online {
  color: #4ade80;
}
.status-all-online .status-dot {
  background: #4ade80;
  box-shadow: 0 0 6px #4ade80;
}

.status-partial {
  color: #fbbf24;
}
.status-partial .status-dot {
  background: #fbbf24;
  box-shadow: 0 0 6px #fbbf24;
}

.status-all-offline {
  color: #f87171;
}
.status-all-offline .status-dot {
  background: #f87171;
  box-shadow: 0 0 6px #f87171;
}

.status-unknown {
  color: #9ca3af;
}
.status-unknown .status-dot {
  background: #9ca3af;
}

.status-divider {
  width: 1px;
  height: 16px;
  background: rgba(255, 255, 255, 0.2);
  margin: 0 0.75rem;
}

.status-uptime,
.status-lastcheck,
.status-incident {
  display: flex;
  align-items: center;
  font-size: 0.8rem;
  color: rgba(255, 255, 255, 0.7);
}

.uptime-good {
  color: #4ade80;
}

.uptime-warn {
  color: #fbbf24;
}

.uptime-bad {
  color: #f87171;
}

.status-incident {
  color: #fbbf24;
  font-weight: 500;
}

.top-nav-right {
  display: flex;
  align-items: center;
}

.top-nav-btn {
  display: flex;
  align-items: center;
  padding: 0.35rem 0.85rem;
  font-size: 0.85rem;
  font-weight: 500;
  color: #fff;
  background: rgba(255, 255, 255, 0.15);
  border: none;
  border-radius: 6px;
  text-decoration: none;
  transition: all 0.2s ease;
}

.top-nav-btn:hover {
  background: rgba(255, 255, 255, 0.25);
  color: #fff;
}

.top-nav-btn-login {
  margin-left: 0.5rem;
  background: rgba(74, 222, 128, 0.2);
  border: 1px solid rgba(74, 222, 128, 0.4);
}

.top-nav-btn-login:hover {
  background: rgba(74, 222, 128, 0.3);
}

/* Mobile responsive */
@media (max-width: 992px) {
  .status-uptime,
  .status-lastcheck,
  .status-incident {
    display: none;
  }

  .status-divider {
    display: none;
  }
}

@media (max-width: 576px) {
  .top-nav-title {
    display: none;
  }

  .status-summary {
    font-size: 0.75rem;
    padding: 0.2rem 0.5rem;
  }

  .top-nav-btn {
    padding: 0.3rem 0.6rem;
    font-size: 0.8rem;
  }
}
</style>
