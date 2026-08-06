<template>
  <div id="app">
    <router-view />
    <Footer v-if="$route.path !== '/setup'" />
  </div>
</template>

<script setup>
import { computed, onBeforeMount, ref } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute, useRouter } from "vue-router";
import Footer from "./components/Index/Footer.vue";
import { useFaviconStatus } from "./composables/useFaviconStatus";
import { useMainStore } from "./stores/main";

const router = useRouter();
const route = useRoute();
const store = useMainStore();
const { locale } = useI18n();

useFaviconStatus();

const loaded = ref(false);

const core = computed(() => store.core);

onBeforeMount(async () => {
	await store.loadCore();

	locale.value = core.value.language || "en";

	if (!core.value.setup) {
		router.push("/setup");
	}

	if (route.path !== "/setup") {
		if (store.admin) {
			await store.loadAdmin();
		} else {
			await store.loadRequired();
		}
		loaded.value = true;
	}
});
</script>

<style lang="scss">
@use './assets/css/bootstrap.min.css';
@use './assets/scss/index';
</style>
