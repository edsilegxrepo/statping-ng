<template>
  <font-awesome-layers v-if="activeMessagesCount > 0">
    <font-awesome-icon icon="calendar" style="color: dodgerblue" />
    <font-awesome-layers-text class="icon-text" :value="activeMessagesCount" />
  </font-awesome-layers>
</template>

<script setup>
import { computed } from "vue";

const props = defineProps({
	messages: {
		type: Array,
		default: () => [],
	},
});

const activeMessagesCount = computed(() => {
	if (!props.messages) return 0;
	const now = new Date();
	return props.messages.filter((m) => {
		const start = new Date(m.start_on);
		const end = new Date(m.end_on);
		return now >= start && now <= end;
	}).length;
});
</script>

<style scoped></style>
