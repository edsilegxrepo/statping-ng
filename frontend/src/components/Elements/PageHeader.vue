<template>
  <div class="page-header">
    <h1 class="page-header-title">{{ title || core.name }}</h1>
    <p class="page-header-desc">{{ description || core.description }}</p>
    <div class="page-header-accent"></div>
  </div>
</template>

<script setup>
import { storeToRefs } from "pinia";
import { useMainStore } from "@/stores/main";

defineProps({
	title: {
		type: String,
		default: "",
	},
	description: {
		type: String,
		default: "",
	},
});

const store = useMainStore();
const { core } = storeToRefs(store);
</script>

<style scoped>
.page-header {
  text-align: center;
  padding: 2rem 1rem;
  background: var(--color-gray-100, #f3f4f6);
  margin-bottom: 2rem;
}

.page-header-title {
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--color-gray-800, #1f2937);
  margin: 0 0 0.5rem;
}

.page-header-desc {
  font-size: 0.95rem;
  color: var(--color-gray-500, #6b7280);
  margin: 0 0 1rem;
}

.page-header-accent {
  width: 60px;
  height: 3px;
  background: var(--color-primary, #3b82f6);
  margin: 0 auto;
  border-radius: 2px;
}
</style>
