<template>
  <div class="dashboard-page">
    <section class="page-section">
      <div class="section-header">
        <div class="section-title">
          <font-awesome-icon icon="bullhorn" class="section-icon" />
          <h2>{{ $t('announcements') }}</h2>
        </div>
      </div>

      <div class="section-card">
        <div v-if="messages.length === 0" class="empty-state-inline">
          <font-awesome-icon icon="comment-slash" class="empty-icon" />
          <p>You currently don't have any announcements. Create one using the form below.</p>
        </div>

        <table v-else class="modern-table">
          <thead>
            <tr>
              <th>{{ $t('title') }}</th>
              <th class="d-none d-md-table-cell">{{ $t('service', 1) }}</th>
              <th class="d-none d-md-table-cell">{{ $t('begins') }}</th>
              <th>Status</th>
              <th class="text-right">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="msg in messages" :key="msg.id">
              <td>
                <span class="message-title">{{ msg.title }}</span>
              </td>
              <td class="d-none d-md-table-cell">
                <router-link :to="serviceLink(getService(msg.service))" class="service-link">
                  {{ serviceName(getService(msg.service)) }}
                </router-link>
              </td>
              <td class="d-none d-md-table-cell">
                <span class="date-text">{{ niceDate(msg.start_on) }}</span>
              </td>
              <td>
                <span class="status-badge" :class="getStatusClass(msg)">
                  {{ getStatusText(msg) }}
                </span>
              </td>
              <td class="text-right">
                <div v-if="store.admin" class="action-buttons">
                  <button @click.prevent="editMessage(msg)" class="action-btn" title="Edit">
                    <font-awesome-icon icon="edit" />
                  </button>
                  <button @click.prevent="deleteMessage(msg)" class="action-btn action-btn-danger" title="Delete">
                    <font-awesome-icon icon="trash-alt" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <section v-if="store.admin" class="page-section">
      <FormMessage :edit="editChange" :in_message="message" />
    </section>
  </div>
</template>

<script setup>
import { computed, ref } from "vue";
import Api from "@/API";
import FormMessage from "@/forms/Message.vue";
import { useMainStore } from "@/stores/main";

const store = useMainStore();

const edit = ref(false);
const message = ref({});

const messages = computed(() => store.messages);

function niceDate(date) {
	if (!date) return "";
	return new Date(date).toLocaleDateString();
}

function getStatusClass(msg) {
	const now = new Date();
	const start = new Date(msg.start_on);
	const end = new Date(msg.end_on);

	if (now < start) return "status-scheduled";
	if (now >= start && now <= end) return "status-active";
	return "status-expired";
}

function getStatusText(msg) {
	const now = new Date();
	const start = new Date(msg.start_on);
	const end = new Date(msg.end_on);

	if (now < start) return "Scheduled";
	if (now >= start && now <= end) return "Active";
	return "Expired";
}

function editChange(v) {
	message.value = {};
	edit.value = v;
}

function editMessage(m) {
	message.value = m;
	edit.value = !edit.value;
}

function getService(id) {
	return store.serviceById(id) || {};
}

function serviceName(service) {
	return service.name || "Global";
}

function serviceLink(service) {
	if (!service.id) return "/";
	return `/service/${service.permalink || service.id}`;
}

async function deleteMessageConfirm(m) {
	await Api.message_delete(m.id);
	const messagesData = await Api.messages();
	store.setMessages(messagesData);
}

function deleteMessage(m) {
	store.setModal({
		visible: true,
		title: "Delete Announcement",
		body: `Are you sure you want to delete Announcement ${m.title}?`,
		btnColor: "btn-danger",
		btnText: "Delete Announcement",
		func: () => deleteMessageConfirm(m),
	});
}
</script>

<style scoped>
.dashboard-page {
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

.section-card {
  background: #fff;
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-md);
  overflow: hidden;
}

/* Empty State */
.empty-state-inline {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: var(--space-8);
  text-align: center;
}

.empty-icon {
  font-size: 2.5rem;
  color: var(--color-gray-300);
  margin-bottom: var(--space-3);
}

.empty-state-inline p {
  color: var(--color-gray-500);
  margin: 0;
}

/* Modern Table */
.modern-table {
  width: 100%;
  border-collapse: collapse;
}

.modern-table th {
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--color-gray-500);
  padding: var(--space-3) var(--space-4);
  background: var(--color-gray-50);
  border-bottom: 1px solid var(--color-gray-200);
}

.modern-table td {
  padding: var(--space-3) var(--space-4);
  border-bottom: 1px solid var(--color-gray-100);
  vertical-align: middle;
}

.modern-table tbody tr {
  transition: background var(--transition-fast);
}

.modern-table tbody tr:hover {
  background: var(--color-gray-50);
}

.modern-table tbody tr:last-child td {
  border-bottom: none;
}

/* Message Title */
.message-title {
  font-weight: 500;
  color: var(--color-gray-900);
}

/* Service Link */
.service-link {
  color: var(--color-primary);
  text-decoration: none;
  font-weight: 500;
}

.service-link:hover {
  text-decoration: underline;
}

/* Date Text */
.date-text {
  font-size: 0.875rem;
  color: var(--color-gray-500);
}

/* Status Badge */
.status-badge {
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  padding: var(--space-1) var(--space-2);
  border-radius: var(--radius-full);
}

.status-active {
  background: var(--color-success-bg);
  color: var(--color-success-dark);
}

.status-scheduled {
  background: var(--color-primary-bg);
  color: var(--color-primary);
}

.status-expired {
  background: var(--color-gray-100);
  color: var(--color-gray-500);
}

/* Action Buttons */
.action-buttons {
  display: flex;
  gap: var(--space-1);
  justify-content: flex-end;
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
  color: var(--color-danger);
  border-color: rgba(239, 68, 68, 0.3);
}

.action-btn-danger:hover {
  background: var(--color-danger-bg);
  color: var(--color-danger-dark);
}

/* Responsive */
@media (max-width: 768px) {
  .dashboard-page {
    padding: var(--space-3);
  }
}
</style>
