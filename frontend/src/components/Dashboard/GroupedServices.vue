<template>
  <div class="row">
    <h5 v-if="group.name && groupServices.length" class="h5 col-12 mb-3 mt-2 text-dim">
      <font-awesome-icon @click="toggle" :icon="expanded ? 'minus' : 'plus'" class="pointer mr-3" /> {{ group.name }}
      <span class="badge badge-success text-uppercase float-right ml-2">{{ servicesOnline.length }} {{ $t('online') }}</span>
      <span v-if="servicesOffline.length > 0" class="badge badge-danger text-uppercase float-right">
        {{ servicesOffline.length }} {{ $t('offline') }}
      </span>
    </h5>
    <div v-if="expanded" v-for="service in groupServices" :key="service.id" class="col-lg-4 col-md-6 col-12">
      <ServiceInfo :service="service" />
    </div>
  </div>
</template>

<script setup>
import { computed, ref } from "vue";
import ServiceInfo from "@/components/Dashboard/ServiceInfo.vue";
import { useMainStore } from "@/stores/main";

const props = defineProps({
	group: {
		required: true,
		type: Object,
	},
});

const store = useMainStore();
const expanded = ref(true);

const groupServices = computed(() => store.servicesInGroup(props.group.id));
const servicesOnline = computed(() =>
	groupServices.value.filter((s) => s.online),
);
const servicesOffline = computed(() =>
	groupServices.value.filter((s) => !s.online),
);

function toggle() {
	expanded.value = !expanded.value;
}
</script>
