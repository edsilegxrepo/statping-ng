<template>
  <button
    v-html="loading ? loadLabel : sanitizeHtml(label)"
    @click.prevent="runAction"
    type="submit"
    :disabled="loading || disabled"
    class="btn btn-block"
    :class="{ 'btn-outline-light': loading }"
  ></button>
</template>

<script setup>
import { ref } from "vue";
import { sanitizeHtml } from "@/mixins";

const props = defineProps({
	action: {
		type: Function,
		required: true,
	},
	label: {
		type: String,
		required: true,
	},
	disabled: {
		type: Boolean,
		default: false,
	},
});

const loading = ref(false);
const loadLabel =
	'<div class="spinner-border text-dark"><span class="sr-only">Loading</span></div>';

async function runAction() {
	loading.value = true;
	await props.action();
	loading.value = false;
}
</script>

<style scoped></style>
