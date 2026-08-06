<template>
  <form @submit.prevent="saveSettings">
    <div class="card">
      <div class="card-header">
        <div class="d-flex justify-content-between align-items-center">
          <span>Polling Engine Settings</span>
          <button type="button" class="btn btn-sm btn-outline-info" @click="loadStats">
            <font-awesome-icon icon="sync" :spin="loadingStats" /> Refresh Stats
          </button>
        </div>
      </div>
      <div class="card-body">
        <div class="form-group">
          <label>Worker Count</label>
          <input
            v-model.number="settings.polling_workers"
            type="number"
            class="form-control"
            min="5"
            max="500"
            placeholder="50"
          />
          <small class="form-text text-muted">
            Number of concurrent service checks (5-500, default: 50). Changes require restart.
          </small>
        </div>

        <div class="form-group">
          <label>Queue Size</label>
          <input
            v-model.number="settings.polling_queue_size"
            type="number"
            class="form-control"
            min="100"
            max="10000"
            placeholder="1000"
          />
          <small class="form-text text-muted">
            Maximum pending checks before backpressure (100-10000, default: 1000). Changes require restart.
          </small>
        </div>

        <div class="form-group">
          <label>Rate Limit per Domain</label>
          <input
            v-model.number="settings.polling_rate_limit"
            type="number"
            class="form-control"
            min="0"
            max="1000"
            placeholder="60"
          />
          <small class="form-text text-muted">
            Maximum checks per domain per minute (0 to disable, default: 60). Applies immediately.
          </small>
        </div>

        <div v-if="stats" class="mt-4">
          <h6>Worker Pool Statistics</h6>
          <div class="row">
            <div class="col-md-4">
              <div class="stat-card">
                <div class="stat-value">{{ stats.active_workers }} / {{ stats.workers }}</div>
                <div class="stat-label">Active Workers</div>
              </div>
            </div>
            <div class="col-md-4">
              <div class="stat-card">
                <div class="stat-value">{{ stats.pending_jobs }}</div>
                <div class="stat-label">Pending Jobs</div>
              </div>
            </div>
            <div class="col-md-4">
              <div class="stat-card">
                <div class="stat-value">{{ stats.scheduled_jobs }}</div>
                <div class="stat-label">Scheduled Jobs</div>
              </div>
            </div>
          </div>
          <div class="row mt-2">
            <div class="col-md-4">
              <div class="stat-card">
                <div class="stat-value">{{ formatNumber(stats.completed_jobs) }}</div>
                <div class="stat-label">Completed Jobs</div>
              </div>
            </div>
            <div class="col-md-4">
              <div class="stat-card">
                <div class="stat-value">{{ stats.rate_limited_jobs }}</div>
                <div class="stat-label">Rate Limited</div>
              </div>
            </div>
            <div class="col-md-4">
              <div class="stat-card">
                <div class="stat-value" :class="stats.running ? 'text-success' : 'text-danger'">
                  {{ stats.running ? 'Running' : 'Stopped' }}
                </div>
                <div class="stat-label">Status</div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div class="card-footer">
        <button
          @click.prevent="saveSettings"
          type="submit"
          class="btn btn-primary btn-block"
          :disabled="loading"
        >
          <font-awesome-icon v-if="loading" icon="circle-notch" class="mr-2" spin />Save Settings
        </button>
      </div>
    </div>
  </form>
</template>

<script setup>
import { onMounted, reactive, ref } from "vue";
import Api from "@/API";

const loading = ref(false);
const loadingStats = ref(false);

const settings = reactive({
	polling_workers: 50,
	polling_queue_size: 1000,
	polling_rate_limit: 60,
});

const stats = ref(null);

onMounted(async () => {
	await loadSettings();
	await loadStats();
});

async function loadSettings() {
	try {
		const resp = await Api.polling_settings();
		settings.polling_workers = resp.polling_workers;
		settings.polling_queue_size = resp.polling_queue_size;
		settings.polling_rate_limit = resp.polling_rate_limit;
	} catch (e) {
		console.error("Failed to load polling settings:", e);
	}
}

async function loadStats() {
	loadingStats.value = true;
	try {
		stats.value = await Api.polling_stats();
	} catch (e) {
		console.error("Failed to load polling stats:", e);
	}
	loadingStats.value = false;
}

async function saveSettings() {
	loading.value = true;
	try {
		await Api.polling_save(settings);
	} catch (e) {
		console.error("Failed to save polling settings:", e);
	}
	loading.value = false;
}

function formatNumber(n) {
	if (n >= 1000000) return (n / 1000000).toFixed(1) + "M";
	if (n >= 1000) return (n / 1000).toFixed(1) + "K";
	return n;
}
</script>

<style scoped>
.stat-card {
  background: var(--card-bg, #f8f9fa);
  border-radius: 8px;
  padding: 12px;
  text-align: center;
}
.stat-value {
  font-size: 1.5rem;
  font-weight: 600;
}
.stat-label {
  font-size: 0.85rem;
  color: #6c757d;
}
</style>
