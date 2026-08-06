<template>
  <div class="col-12">
    <div v-if="!ready" class="row mt-5">
      <div class="col-12 text-center">
        <font-awesome-icon icon="circle-notch" size="3x" spin />
      </div>
      <div class="col-12 text-center mt-3 mb-3">
        <span class="text-muted">Loading Service</span>
      </div>
    </div>
    <FormService v-if="ready" :in_service="service" />
  </div>
</template>

<script setup>
import { onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";
import Api from "@/API";
import FormService from "@/forms/Service.vue";

const route = useRoute();

const service = ref(null);
const ready = ref(false);

watch(() => route.params.id, fetchData);

onMounted(() => {
	fetchData();
});

async function fetchData() {
	if (!route.params.id) {
		ready.value = true;
		return;
	}
	service.value = await Api.service(route.params.id);
	ready.value = true;
}
</script>

<style scoped></style>
