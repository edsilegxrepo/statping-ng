<template>
  <div class="col-12 mb-3 pb-2 border-bottom" role="alert">
    <span
      class="font-weight-bold text-capitalize"
      :class="{
        'text-success': update.type.toLowerCase() === 'resolved',
        'text-danger': update.type.toLowerCase() === 'investigating',
        'text-warning': update.type.toLowerCase() === 'update',
      }"
      >{{ update.type }}</span
    >
    <span class="text-muted"
      >- {{ update.message }}
      <button v-if="admin" @click="deleteUpdate(update)" type="button" class="close">
        <span aria-hidden="true">&times;</span>
      </button>
    </span>
    <span class="d-block small">{{ timeAgo(update.created_at) }} ago</span>
  </div>
</template>

<script setup>
import Api from "@/API";

const props = defineProps({
	update: {
		type: Object,
		required: true,
	},
	admin: {
		type: Boolean,
		required: true,
	},
	onUpdate: {
		type: Function,
		required: false,
	},
});

function timeAgo(date) {
	const seconds = Math.floor((new Date() - new Date(date)) / 1000);
	if (seconds < 60) return `${seconds} seconds`;
	const minutes = Math.floor(seconds / 60);
	if (minutes < 60) return `${minutes} minutes`;
	const hours = Math.floor(minutes / 60);
	if (hours < 24) return `${hours} hours`;
	const days = Math.floor(hours / 24);
	return `${days} days`;
}

async function deleteUpdate(update) {
	const res = await Api.incident_update_delete(update);
	if (res.status === "success" && props.onUpdate) {
		props.onUpdate();
	}
}
</script>

<style scoped></style>
