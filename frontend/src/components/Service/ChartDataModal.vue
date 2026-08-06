<template>
  <div v-if="show" class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-container">
      <div class="modal-header">
        <h3 class="modal-title">
          <font-awesome-icon icon="table" class="me-2" />
          Chart Data
        </h3>
        <button class="close-btn" @click="$emit('close')">
          <font-awesome-icon icon="times" />
        </button>
      </div>

      <div class="modal-toolbar">
        <div class="export-buttons">
          <button class="export-btn" @click="doExportTSV">
            <font-awesome-icon icon="file-export" class="me-1" />
            Export TSV
          </button>
          <button class="export-btn" @click="doExportJSON">
            <font-awesome-icon icon="file-code" class="me-1" />
            Export JSON
          </button>
        </div>
        <span class="record-count">{{ tableData.length }} records</span>
      </div>

      <div class="modal-body">
        <table class="data-table">
          <thead>
            <tr>
              <th>Timestamp</th>
              <th class="text-right">Latency</th>
              <th class="text-right">Ping</th>
              <th class="text-right">Failures</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(row, idx) in tableData" :key="idx">
              <td class="timestamp">{{ formatTimestamp(row.timestamp) }}</td>
              <td class="text-right latency">{{ formatTime(row.latency) }}</td>
              <td class="text-right ping">{{ formatTime(row.ping) }}</td>
              <td class="text-right failures" :class="{ 'has-failures': row.failures > 0 }">
                {{ row.failures || 0 }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from "vue";
import { exportJSON, exportTSV } from "@/composables/useExport";

const props = defineProps({
	show: Boolean,
	latencyData: Array,
	pingData: Array,
	failureData: Array,
	serviceName: String,
});

const emit = defineEmits(["close"]);

const tableData = computed(() => {
	const latency = props.latencyData || [];
	const ping = props.pingData || [];
	const failures = props.failureData || [];

	const map = new Map();

	latency.forEach((d) => {
		const ts = new Date(d.timeframe).getTime();
		if (!map.has(ts)) map.set(ts, { timestamp: ts });
		map.get(ts).latency = d.amount;
	});

	ping.forEach((d) => {
		const ts = new Date(d.timeframe).getTime();
		if (!map.has(ts)) map.set(ts, { timestamp: ts });
		map.get(ts).ping = d.amount;
	});

	failures.forEach((d) => {
		const ts = new Date(d.timeframe).getTime();
		if (!map.has(ts)) map.set(ts, { timestamp: ts });
		map.get(ts).failures = d.amount;
	});

	return Array.from(map.values()).sort((a, b) => a.timestamp - b.timestamp);
});

function formatTimestamp(ts) {
	return new Date(ts).toLocaleString();
}

function formatTime(microseconds) {
	if (!microseconds) return "—";
	if (microseconds >= 1000000) {
		return `${(microseconds / 1000000).toFixed(1)}s`;
	}
	if (microseconds >= 1000) {
		return `${(microseconds / 1000).toFixed(1)}ms`;
	}
	return `${microseconds}μs`;
}

function doExportTSV() {
	const headers = ["Timestamp", "Latency (μs)", "Ping (μs)", "Failures"];
	const rows = tableData.value.map((row) => [
		new Date(row.timestamp).toISOString(),
		row.latency || "",
		row.ping || "",
		row.failures || 0,
	]);
	exportTSV(headers, rows, `${props.serviceName || "chart"}-data`);
}

function doExportJSON() {
	const data = tableData.value.map((row) => ({
		timestamp: new Date(row.timestamp).toISOString(),
		latency_us: row.latency || null,
		ping_us: row.ping || null,
		failures: row.failures || 0,
	}));
	exportJSON(data, `${props.serviceName || "chart"}-data`);
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}

.modal-container {
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  width: 90%;
  max-width: 900px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid #e5e7eb;
}

.modal-title {
  font-size: 1.1rem;
  font-weight: 600;
  color: #111827;
  margin: 0;
}

.close-btn {
  background: none;
  border: none;
  color: #6b7280;
  cursor: pointer;
  padding: 0.5rem;
  border-radius: 6px;
  transition: all 0.15s;
}

.close-btn:hover {
  background: #f3f4f6;
  color: #111827;
}

.modal-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1.5rem;
  background: #f9fafb;
  border-bottom: 1px solid #e5e7eb;
}

.export-buttons {
  display: flex;
  gap: 0.5rem;
}

.export-btn {
  display: flex;
  align-items: center;
  padding: 0.5rem 1rem;
  font-size: 0.85rem;
  font-weight: 500;
  color: #374151;
  background: #fff;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s;
}

.export-btn:hover {
  background: #f3f4f6;
  border-color: #9ca3af;
}

.record-count {
  font-size: 0.85rem;
  color: #6b7280;
}

.modal-body {
  flex: 1;
  overflow: auto;
  padding: 0;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.85rem;
}

.data-table th {
  position: sticky;
  top: 0;
  background: #f9fafb;
  padding: 0.75rem 1rem;
  text-align: left;
  font-weight: 600;
  color: #374151;
  border-bottom: 2px solid #e5e7eb;
}

.data-table td {
  padding: 0.6rem 1rem;
  border-bottom: 1px solid #f3f4f6;
  color: #4b5563;
}

.data-table tbody tr:hover {
  background: #f9fafb;
}

.text-right {
  text-align: right;
}

.timestamp {
  font-family: monospace;
  font-size: 0.8rem;
  color: #6b7280;
}

.latency {
  color: #f1771f;
  font-weight: 500;
}

.ping {
  color: #48d338;
  font-weight: 500;
}

.failures {
  color: #9ca3af;
}

.failures.has-failures {
  color: #ef4444;
  font-weight: 600;
}
</style>
