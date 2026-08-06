<template>
  <form @submit.prevent="updateIntegration">
    <h4 class="text-capitalize">{{ integration.full_name }}</h4>
    <p class="small text-muted" v-html="sanitizeHtml(integration.description)"></p>

    <div v-for="(field, index) in integration.fields" :key="index" class="form-group">
      <label class="text-capitalize">{{ field.name }}</label>

      <textarea v-if="field.type === 'textarea'" v-model="field.value" rows="3" class="form-control"></textarea>

      <input v-else :type="field.type" v-model="field.value" class="form-control" />

      <small class="form-text text-muted" v-html="sanitizeHtml(field.description)"></small>
    </div>

    <div class="col-12">
      <div class="col-3">
        <span @click="integration.enabled = !!integration.enabled" class="switch">
          <input
            type="checkbox"
            name="enabled-option"
            class="switch"
            v-model="integration.enabled"
            :id="`switch-${integration.name}`"
            :checked="integration.enabled"
          />
          <label :for="`switch-${integration.name}`"></label>
        </span>
      </div>

      <div v-if="services.length !== 0" class="col-12">
        <table class="table">
          <thead>
            <tr>
              <th scope="col">Name</th>
              <th scope="col">Domain</th>
              <th scope="col">Port</th>
              <th scope="col">Interval</th>
              <th scope="col">Timeout</th>
              <th scope="col">Type</th>
              <th scope="col"></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(service, index) in services" :key="index">
              <td><input v-model="service.name" type="text" style="width: 80pt" /></td>
              <td>{{ service.domain }}</td>
              <td>{{ service.port }}</td>
              <td><input v-model="service.check_interval" type="number" style="width: 35pt" /></td>
              <td><input v-model="service.timeout" type="number" style="width: 35pt" /></td>
              <td>{{ service.type }}</td>
              <td>
                <button
                  @click.prevent="addService(service)"
                  :disabled="service.added"
                  class="btn btn-sm btn-outline-primary"
                >
                  Add
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="col-12">
        <button @click.prevent="updateIntegration" type="submit" class="btn btn-block btn-info">Fetch Services</button>
      </div>
    </div>

    <div class="alert alert-danger d-none" role="alert"></div>
  </form>
</template>

<script setup>
import { ref } from "vue";
import Api from "@/API";
import { sanitizeHtml } from "@/mixins";
import { useMainStore } from "@/stores/main";

const props = defineProps({
	integration: {
		type: Object,
		required: true,
	},
});

const store = useMainStore();
const services = ref([]);

async function addService(s) {
	const data = {
		name: s.name,
		type: s.type,
		domain: s.domain,
		port: s.port,
		check_interval: s.check_interval,
		timeout: s.timeout,
	};
	await Api.service_create(data);
	const svcList = await Api.services();
	store.setServices(svcList);
	s.added = true;
}

async function updateIntegration() {
	const i = props.integration;
	const data = { name: i.name, enabled: i.enabled, fields: i.fields };
	const out = await Api.integration_save(data);
	if (out != null) {
		services.value = out;
	}
	const integrations = await Api.integrations();
	store.setIntegrations(integrations);
}
</script>

<style scoped></style>
