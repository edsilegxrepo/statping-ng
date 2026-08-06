<template>
  <div v-if="show" class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-container">
      <div class="modal-header">
        <h3 class="modal-title">
          <font-awesome-icon icon="table" class="me-2" />
          Hourly Failures - {{ formatDate(selectedDate) }}
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
        <span class="record-count">{{ totalFailures }} total failures</span>
      </div>

      <div class="modal-body">
        <table class="data-table">
          <thead>
            <tr>
              <th>Hour (UTC)</th>
              <th class="text-right">Failures</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(row, idx) in tableData" :key="idx" :class="{ 'has-failures': row.failures > 0 }">
              <td class="hour">{{ row.hour }}</td>
              <td class="text-right failures" :class="{ 'zero': row.failures === 0 }">{{ row.failures }}</td>
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
	series: Array,
	selectedDate: Date,
	serviceName: String,
});

const emit = defineEmits(["close"]);

const tableData = computed(() => {
	if (!props.series || props.series.length === 0 || !props.series[0].data) {
		return [];
	}

	return props.series[0].data
		.map((failures, hour) => ({
			hour: `${hour.toString().padStart(2, "0")}:00 - ${hour.toString().padStart(2, "0")}:59`,
			failures,
		}))
		.filter((row) => row.failures > 0);
});

const totalFailures = computed(() => {
	return tableData.value.reduce((sum, row) => sum + row.failures, 0);
});

function formatDate(date) {
	if (!date) return "";
	const options = { year: "numeric", month: "long", day: "numeric" };
	return date.toLocaleDateString("en-us", options);
}

function doExportTSV() {
	const dateStr = props.selectedDate
		? props.selectedDate.toISOString().split("T")[0]
		: "unknown";
	const headers = ["Hour (UTC)", "Failures"];
	const rows = tableData.value.map((row) => [row.hour, row.failures]);
	exportTSV(
		headers,
		rows,
		`${props.serviceName || "service"}-failures-${dateStr}`,
	);
}

function doExportJSON() {
	const dateStr = props.selectedDate
		? props.selectedDate.toISOString().split("T")[0]
		: "unknown";
	const data = {
		date: dateStr,
		service: props.serviceName,
		totalFailures: totalFailures.value,
		hourly: tableData.value.map((row) => ({
			hour: row.hour,
			failures: row.failures,
		})),
	};
	exportJSON(data, `${props.serviceName || "service"}-failures-${dateStr}`);
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
  max-width: 500px;
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
  padding: 0.5rem 1rem;
  border-bottom: 1px solid #f3f4f6;
  color: #4b5563;
}

.data-table tbody tr:hover {
  background: #f9fafb;
}

.data-table tbody tr.has-failures {
  background: #fef2f2;
}

.data-table tbody tr.has-failures:hover {
  background: #fee2e2;
}

.text-right {
  text-align: right;
}

.hour {
  font-family: monospace;
  font-size: 0.8rem;
}

.failures {
  font-weight: 600;
  color: #ef4444;
}

.failures.zero {
  color: #9ca3af;
  font-weight: 400;
}
</style>
