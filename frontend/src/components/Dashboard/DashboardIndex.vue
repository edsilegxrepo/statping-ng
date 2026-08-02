<template>
  <div class="dashboard-content">
    <!-- Stats Cards -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon stat-icon-primary">
          <font-awesome-icon icon="server" />
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ services.length }}</span>
          <span class="stat-label">{{ $t('total_services') }}</span>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon stat-icon-success">
          <font-awesome-icon icon="check-circle" />
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ onlineServicesCount }}</span>
          <span class="stat-label">{{ $t('online_services') }}</span>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon stat-icon-danger">
          <font-awesome-icon icon="exclamation-triangle" />
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ failuresLast24Hours }}</span>
          <span class="stat-label">{{ $t('failures_24_hours') }}</span>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon stat-icon-info">
          <font-awesome-icon icon="percentage" />
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ uptimePercentage }}%</span>
          <span class="stat-label">Avg Uptime (24h)</span>
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div v-if="services.length === 0" class="empty-state">
      <div class="empty-state-icon">
        <font-awesome-icon icon="server" />
      </div>
      <h3>{{ $t('no_services') }}</h3>
      <p>Get started by creating your first monitor</p>
      <router-link v-if="store.admin" to="/dashboard/create_service" class="btn-create">
        <font-awesome-icon icon="plus" class="me-2" />
        {{ $t('create') }} Service
      </router-link>
    </div>

    <!-- Active Announcements -->
    <div v-if="messagesInRange.length > 0" class="announcements-section">
      <h3 class="section-title">
        <font-awesome-icon icon="bullhorn" class="me-2" />
        Active Announcements
      </h3>
      <div v-for="message in messagesInRange" :key="message.id" class="announcement-card">
        <div class="announcement-icon">
          <font-awesome-icon icon="calendar" />
        </div>
        <div class="announcement-content">
          <p class="announcement-text">{{ message.description }}</p>
          <span class="announcement-time">
            <strong>{{ niceDate(message.start_on) }}</strong> — <strong>{{ niceDate(message.end_on) }}</strong>
            <span class="announcement-duration">({{ duration(message.start_on, message.end_on) }})</span>
          </span>
        </div>
      </div>
    </div>

    <!-- Services Grid -->
    <div v-if="services.length > 0" class="services-section">
      <div v-if="servicesNoGroup.length > 0" class="services-grid">
        <div v-for="service in servicesNoGroup" :key="service.id" class="service-col">
          <ServiceInfo :service="service" />
        </div>
      </div>

      <div v-for="group in groups" :key="group.id">
        <GroupedServices :group="group" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useMainStore } from '@/stores/main'
import GroupedServices from '@/components/Dashboard/GroupedServices.vue'
import ServiceInfo from '@/components/Dashboard/ServiceInfo.vue'

const store = useMainStore()

const services = computed(() => store.services)
const servicesNoGroup = computed(() => store.servicesNoGroup)
const groups = computed(() => store.groupsInOrder)
const onlineServicesCount = computed(() => store.onlineServices(true).length)

const failuresLast24Hours = computed(() => {
  let total = 0
  services.value.forEach((s) => {
    total += s.failures_24_hours || 0
  })
  return total
})

const uptimePercentage = computed(() => {
  if (services.value.length === 0) return '—'
  const total = services.value.reduce((sum, s) => sum + (s.online_24_hours || 0), 0)
  const avg = total / services.value.length
  return avg.toFixed(1)
})

const messagesInRange = computed(() => {
  const now = new Date()
  return store.globalMessages.filter((m) => {
    const start = new Date(m.start_on)
    const end = new Date(m.end_on)
    return now >= start && now <= end
  })
})

function niceDate(date) {
  return new Date(date).toLocaleDateString()
}

function duration(start, end) {
  const ms = new Date(end) - new Date(start)
  const hours = Math.floor(ms / (1000 * 60 * 60))
  if (hours < 24) return `${hours} hours`
  const days = Math.floor(hours / 24)
  return `${days} days`
}
</script>

<style scoped>
.dashboard-content {
  max-width: 100%;
}

/* Stats Grid */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--space-4);
  margin-bottom: var(--space-6);
}

.stat-card {
  display: flex;
  align-items: center;
  padding: var(--space-5);
  background: #fff;
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-md);
  transition: all var(--transition-normal);
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-lg);
}

.stat-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  border-radius: var(--radius-lg);
  font-size: 1.5rem;
  margin-right: var(--space-4);
}

.stat-icon-primary {
  background: var(--color-primary-bg);
  color: var(--color-primary);
}

.stat-icon-success {
  background: var(--color-success-bg);
  color: var(--color-success);
}

.stat-icon-danger {
  background: var(--color-danger-bg);
  color: var(--color-danger);
}

.stat-icon-info {
  background: var(--color-info-bg);
  color: var(--color-info);
}

.stat-info {
  display: flex;
  flex-direction: column;
}

.stat-value {
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--color-gray-900);
  line-height: 1.2;
}

.stat-label {
  font-size: 0.875rem;
  color: var(--color-gray-500);
  margin-top: var(--space-1);
}

/* Empty State */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-8) var(--space-4);
  background: #fff;
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-md);
  text-align: center;
}

.empty-state-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 80px;
  height: 80px;
  background: var(--color-gray-100);
  border-radius: var(--radius-full);
  font-size: 2rem;
  color: var(--color-gray-400);
  margin-bottom: var(--space-4);
}

.empty-state h3 {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--color-gray-900);
  margin: 0 0 var(--space-2);
}

.empty-state p {
  color: var(--color-gray-500);
  margin: 0 0 var(--space-4);
}

.btn-create {
  display: inline-flex;
  align-items: center;
  padding: var(--space-3) var(--space-5);
  background: var(--color-success);
  color: #fff;
  font-weight: 500;
  border-radius: var(--radius-md);
  text-decoration: none;
  transition: all var(--transition-fast);
}

.btn-create:hover {
  background: var(--color-success-dark);
  color: #fff;
}

/* Announcements */
.announcements-section {
  margin-bottom: var(--space-6);
}

.section-title {
  font-size: 1rem;
  font-weight: 600;
  color: var(--color-gray-700);
  margin: 0 0 var(--space-3);
}

.announcement-card {
  display: flex;
  align-items: flex-start;
  padding: var(--space-4);
  background: var(--color-warning-bg);
  border: 1px solid rgba(245, 158, 11, 0.2);
  border-radius: var(--radius-lg);
  margin-bottom: var(--space-3);
}

.announcement-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background: var(--color-warning);
  color: #fff;
  border-radius: var(--radius-md);
  margin-right: var(--space-4);
  flex-shrink: 0;
}

.announcement-content {
  flex: 1;
}

.announcement-text {
  font-weight: 500;
  color: var(--color-gray-800);
  margin: 0 0 var(--space-2);
}

.announcement-time {
  font-size: 0.875rem;
  color: var(--color-gray-600);
}

.announcement-duration {
  color: var(--color-gray-500);
  margin-left: var(--space-2);
}

/* Services Grid */
.services-section {
  margin-top: var(--space-4);
}

.services-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-4);
}

.service-col {
  min-width: 0;
}

/* Responsive */
@media (max-width: 1200px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .services-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .dashboard-content {
    padding: var(--space-3);
  }

  .stats-grid {
    grid-template-columns: 1fr;
  }

  .services-grid {
    grid-template-columns: 1fr;
  }

  .stat-card {
    padding: var(--space-4);
  }

  .stat-icon {
    width: 48px;
    height: 48px;
    font-size: 1.25rem;
  }

  .stat-value {
    font-size: 1.5rem;
  }
}
</style>
