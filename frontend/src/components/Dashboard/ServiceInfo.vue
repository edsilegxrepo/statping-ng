<template>
  <router-link :to="serviceLink" custom v-slot="{ navigate }">
    <div
      class="service-card"
      :class="service.online ? 'service-card-online' : 'service-card-offline'"
      @click="navigate"
    >
      <!-- Header -->
      <div class="service-header">
        <div class="service-title-row">
          <span class="service-name">{{ service.name }}</span>
          <span class="service-badge" :class="service.online ? 'badge-online' : 'badge-offline'">
            {{ service.online ? $t('online') : $t('offline') }}
          </span>
        </div>
        <div class="service-uptime">
          <font-awesome-icon icon="clock" class="me-1" />
          {{ service.online_7_days }}% uptime (7d)
        </div>
      </div>

      <!-- Chart -->
      <div class="service-chart-area">
        <div v-if="loaded">
          <ServiceSparkLine :title="set2_name" subtitle="Latency Last 24 Hours" :series="set2" />
        </div>
        <div v-else class="service-loading">
          <font-awesome-icon icon="circle-notch" class="text-dim" size="2x" spin />
        </div>
      </div>

      <!-- Events -->
      <div class="service-events">
        <ServiceEvents :service="service" />
      </div>

      <!-- Footer -->
      <div class="service-footer">
        <span class="footer-info">{{ hoverbtn }}</span>
        <div class="footer-actions">
          <button
            @click.stop="goToIncidents"
            @mouseleave="unsetHover"
            @mouseover="setHover($t('incidents'))"
            class="action-btn"
            title="Incidents"
          >
            <font-awesome-icon icon="bullhorn" />
          </button>
          <button
            @click.stop="goToCheckins"
            @mouseleave="unsetHover"
            @mouseover="setHover($t('checkins'))"
            class="action-btn"
            title="Checkins"
          >
            <font-awesome-icon icon="calendar-check" />
          </button>
          <button
            @click.stop="goToFailures"
            @mouseleave="unsetHover"
            @mouseover="setHover($t('failures'))"
            class="action-btn"
            :class="{ 'action-btn-danger': service.stats?.failures }"
            title="Failures"
          >
            <font-awesome-icon icon="exclamation-triangle" />
            <span v-if="service.stats?.failures" class="failure-count">{{ service.stats.failures }}</span>
          </button>
        </div>
      </div>
    </div>
  </router-link>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { useRouter } from "vue-router";
import Api from "@/API";
import ServiceEvents from "@/components/Dashboard/ServiceEvents.vue";
import ServiceSparkLine from "./ServiceSparkLine.vue";

const props = defineProps({
	service: {
		type: Object,
		required: true,
	},
});

const router = useRouter();
const { t } = useI18n();

const uptime = ref(null);
const hoverbtn = ref("");
const set2 = ref([]);
const loaded = ref(false);
const set2_name = ref("");
const failures = ref(null);
const visible = ref(false);

const serviceLink = computed(
	() =>
		`/service/${props.service.permalink || props.service.id}?from=dashboard`,
);

onMounted(async () => {
	unsetHover();
	await loadInfo();
});

function setHover(name) {
	hoverbtn.value = name;
}

function unsetHover() {
	hoverbtn.value = `Avg: ${set2_name.value || "—"}`;
}

function goToIncidents() {
	router.push({ path: `/dashboard/service/${props.service.id}/incidents` });
}

function goToCheckins() {
	router.push({ path: `/dashboard/service/${props.service.id}/checkins` });
}

function goToFailures() {
	router.push({ path: `/dashboard/service/${props.service.id}/failures` });
}

async function loadInfo() {
	set2.value = await getHits(86400 * 3, "60m");
	set2_name.value = calc(set2.value);
	loaded.value = true;
	unsetHover();
}

function nowSubtract(seconds) {
	return new Date(Date.now() - seconds * 1000);
}

function endOfToday() {
	const d = new Date();
	d.setHours(23, 59, 59, 999);
	return d;
}

function toUnix(date) {
	return Math.floor(date.getTime() / 1000);
}

async function getHits(seconds, group) {
	const start = nowSubtract(seconds);
	const end = endOfToday();

	try {
		const fetched = await Api.service_hits(
			props.service.id,
			toUnix(start),
			toUnix(end),
			group,
			true,
		);
		const data = convertToChartData(fetched);
		return [{ name: "Latency", ...data }];
	} catch (e) {
		return [{ name: "Latency", data: [] }];
	}
}

function convertToChartData(arr) {
	if (!arr || !arr.length) return { data: [] };
	return {
		data: arr.map((d) => ({
			x: new Date(d.timeframe).getTime(),
			y: Math.round(d.amount * 0.001),
		})),
	};
}

function calc(s) {
	const data = s[0]?.data;
	if (data && data.length) {
		let total = 0;
		data.forEach((f) => {
			total += parseInt(f.y, 10);
		});
		total = total / data.length;
		return `${Math.round(total)} ms`;
	}
	return "Offline";
}
</script>

<style scoped>
.service-card {
  background: #fff;
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-md);
  cursor: pointer;
  transition: all var(--transition-normal);
  overflow: hidden;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.service-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-xl);
}

.service-card-online {
  border-left: 4px solid var(--color-success);
}

.service-card-offline {
  border-left: 4px solid var(--color-danger);
}

/* Header */
.service-header {
  padding: var(--space-4);
  border-bottom: 1px solid var(--color-gray-100);
}

.service-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-2);
}

.service-name {
  font-size: 1rem;
  font-weight: 600;
  color: var(--color-gray-900);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 70%;
}

.service-badge {
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  padding: var(--space-1) var(--space-2);
  border-radius: var(--radius-full);
}

.badge-online {
  background: var(--color-success-bg);
  color: var(--color-success-dark);
}

.badge-offline {
  background: var(--color-danger-bg);
  color: var(--color-danger-dark);
}

.service-uptime {
  font-size: 0.8rem;
  color: var(--color-gray-500);
}

/* Chart Area */
.service-chart-area {
  padding: var(--space-3) var(--space-4);
  flex: 1;
  min-height: 120px;
}

.service-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100px;
}

/* Events */
.service-events {
  padding: 0 var(--space-4);
}

/* Footer */
.service-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-3) var(--space-4);
  background: var(--color-gray-50);
  border-top: 1px solid var(--color-gray-100);
}

.footer-info {
  font-size: 0.8rem;
  color: var(--color-gray-500);
}

.footer-actions {
  display: flex;
  gap: var(--space-1);
}

.action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: #fff;
  border: 1px solid var(--color-gray-200);
  border-radius: var(--radius-md);
  color: var(--color-gray-500);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.action-btn:hover {
  background: var(--color-gray-100);
  color: var(--color-gray-700);
  border-color: var(--color-gray-300);
}

.action-btn-danger {
  background: var(--color-danger-bg);
  border-color: rgba(239, 68, 68, 0.3);
  color: var(--color-danger);
}

.action-btn-danger:hover {
  background: rgba(239, 68, 68, 0.2);
}

.failure-count {
  font-size: 0.65rem;
  font-weight: 700;
  background: var(--color-danger);
  color: #fff;
  padding: 1px 5px;
  border-radius: var(--radius-full);
  margin-left: 2px;
}
</style>
