<template>
  <div v-if="show" class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-container">
      <div class="modal-header">
        <h3 class="modal-title">
          <font-awesome-icon icon="table" class="me-2" />
          Failures Data (Last 6 Months)
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
              <th>Date</th>
              <th class="text-right">Failures</th>
              <th>Severity</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(row, idx) in tableData" :key="idx" :class="{ 'has-failures': row.failures > 0 }">
              <td class="date">{{ row.date }}</td>
              <td class="text-right failures">{{ row.failures }}</td>
              <td>
                <span class="severity-badge" :class="row.severityClass">{{ row.severity }}</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { exportTSV, exportJSON } from '@/composables/useExport'

const props = defineProps({
  show: Boolean,
  series: Array,
  serviceName: String,
})

const emit = defineEmits(['close'])

const tableData = computed(() => {
  if (!props.series) return []

  const rows = []
  const year = new Date().getFullYear()

  props.series.forEach(monthData => {
    const monthName = monthData.name
    monthData.data.forEach(dayData => {
      const day = parseInt(dayData.x)
      const failures = dayData.y
      if (failures > 0) {
        rows.push({
          date: `${monthName} ${day}, ${year}`,
          failures,
          severity: getSeverity(failures),
          severityClass: getSeverityClass(failures),
          sortKey: new Date(`${monthName} ${day}, ${year}`).getTime(),
        })
      }
    })
  })

  return rows.sort((a, b) => b.sortKey - a.sortKey)
})

const totalFailures = computed(() => {
  return tableData.value.reduce((sum, row) => sum + row.failures, 0)
})

function getSeverity(failures) {
  if (failures === 0) return 'Healthy'
  if (failures <= 30) return 'Minor'
  if (failures <= 120) return 'Moderate'
  if (failures <= 240) return 'Major'
  return 'Critical'
}

function getSeverityClass(failures) {
  if (failures === 0) return 'severity-healthy'
  if (failures <= 30) return 'severity-minor'
  if (failures <= 120) return 'severity-moderate'
  if (failures <= 240) return 'severity-major'
  return 'severity-critical'
}

function doExportTSV() {
  const headers = ['Date', 'Failures', 'Severity']
  const rows = tableData.value.map(row => [row.date, row.failures, row.severity])
  exportTSV(headers, rows, `${props.serviceName || 'service'}-failures`)
}

function doExportJSON() {
  const data = tableData.value.map(row => ({
    date: row.date,
    failures: row.failures,
    severity: row.severity,
  }))
  exportJSON(data, `${props.serviceName || 'service'}-failures`)
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
  max-width: 600px;
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

.data-table tbody tr.has-failures {
  background: #fef2f2;
}

.data-table tbody tr.has-failures:hover {
  background: #fee2e2;
}

.text-right {
  text-align: right;
}

.date {
  font-family: monospace;
  font-size: 0.8rem;
}

.failures {
  font-weight: 600;
  color: #ef4444;
}

.severity-badge {
  display: inline-block;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
}

.severity-healthy {
  background: #f8f9fa;
  color: #6b7280;
}

.severity-minor {
  background: #dcfce7;
  color: #166534;
}

.severity-moderate {
  background: #fef9c3;
  color: #854d0e;
}

.severity-major {
  background: #ffedd5;
  color: #c2410c;
}

.severity-critical {
  background: #fee2e2;
  color: #b91c1c;
}
</style>
