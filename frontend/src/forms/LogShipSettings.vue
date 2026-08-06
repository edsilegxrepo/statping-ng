<template>
  <div class="card">
    <div class="card-header">
      Log Shipping
    </div>
    <div class="card-body">
      <div v-if="loading" class="text-center py-4">
        <font-awesome-icon icon="circle-notch" spin size="2x" />
      </div>

      <form v-else @submit.prevent="save">
        <div v-if="settings.env_override" class="alert alert-info">
          <font-awesome-icon icon="info-circle" class="mr-2" />
          Log shipping is configured via environment variables. UI settings are disabled.
        </div>

        <div class="form-group row">
          <label class="col-sm-4 col-form-label">Enable Log Shipping</label>
          <div class="col-sm-8">
            <div class="switch">
              <input v-model="settings.log_ship_enabled" type="checkbox" id="log_ship_enabled" :disabled="settings.env_override" />
              <label for="log_ship_enabled">Ship logs to external system</label>
            </div>
          </div>
        </div>

        <div class="form-group row">
          <label class="col-sm-4 col-form-label">Destination Type</label>
          <div class="col-sm-8">
            <select v-model="settings.log_ship_type" class="form-control" :disabled="settings.env_override">
              <option value="">Select destination...</option>
              <option v-for="t in settings.types" :key="t.value" :value="t.value">{{ t.label }}</option>
            </select>
          </div>
        </div>

        <div class="form-group row">
          <label class="col-sm-4 col-form-label">Endpoint URL</label>
          <div class="col-sm-8">
            <input v-model="settings.log_ship_endpoint" type="url" class="form-control"
              :placeholder="endpointPlaceholder" :disabled="settings.env_override" />
            <small class="form-text text-muted">{{ endpointHelp }}</small>
          </div>
        </div>

        <div class="form-group row">
          <label class="col-sm-4 col-form-label">Authentication Token</label>
          <div class="col-sm-8">
            <input v-model="settings.log_ship_token" type="password" class="form-control"
              placeholder="Leave blank to keep current" :disabled="settings.env_override" />
            <small class="form-text text-muted">{{ tokenHelp }}</small>
          </div>
        </div>

        <div v-if="showSplunkFields" class="form-group row">
          <label class="col-sm-4 col-form-label">Splunk Index</label>
          <div class="col-sm-8">
            <input v-model="settings.log_ship_index" type="text" class="form-control"
              placeholder="main" :disabled="settings.env_override" />
          </div>
        </div>

        <div v-if="showSourcetypeField" class="form-group row">
          <label class="col-sm-4 col-form-label">Sourcetype</label>
          <div class="col-sm-8">
            <input v-model="settings.log_ship_sourcetype" type="text" class="form-control"
              placeholder="statping" :disabled="settings.env_override" />
          </div>
        </div>

        <div class="form-group row">
          <label class="col-sm-4 col-form-label">Labels</label>
          <div class="col-sm-8">
            <input v-model="settings.log_ship_labels" type="text" class="form-control"
              placeholder="env=prod,datacenter=us-east" :disabled="settings.env_override" />
            <small class="form-text text-muted">Additional metadata labels (key=value,key2=value2)</small>
          </div>
        </div>

        <hr />

        <div class="d-flex justify-content-between">
          <button type="button" @click="testConnection" class="btn btn-outline-primary"
            :disabled="settings.env_override || testing || !settings.log_ship_type || !settings.log_ship_endpoint">
            <font-awesome-icon v-if="testing" icon="circle-notch" spin class="mr-1" />
            <font-awesome-icon v-else icon="plug" class="mr-1" />
            Test Connection
          </button>
          <button type="submit" class="btn btn-primary" :disabled="settings.env_override || saving">
            <font-awesome-icon v-if="saving" icon="circle-notch" spin class="mr-1" />
            Save Settings
          </button>
        </div>

        <div v-if="testResult" class="mt-3 alert" :class="testResult.success ? 'alert-success' : 'alert-danger'">
          <font-awesome-icon :icon="testResult.success ? 'check-circle' : 'times-circle'" class="mr-2" />
          {{ testResult.message }}
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import Api from "@/API";

const loading = ref(true);
const saving = ref(false);
const testing = ref(false);
const testResult = ref(null);
const settings = ref({
	log_ship_enabled: false,
	log_ship_type: "",
	log_ship_endpoint: "",
	log_ship_token: "",
	log_ship_index: "",
	log_ship_sourcetype: "",
	log_ship_labels: "",
	env_override: false,
	types: [],
});

const showSplunkFields = computed(
	() => settings.value.log_ship_type === "splunk",
);
const showSourcetypeField = computed(() =>
	["splunk", "cribl"].includes(settings.value.log_ship_type),
);

const endpointPlaceholder = computed(() => {
	switch (settings.value.log_ship_type) {
		case "loki":
			return "http://loki:3100";
		case "elasticsearch":
			return "http://elasticsearch:9200";
		case "splunk":
			return "https://splunk:8088";
		case "cribl":
			return "http://cribl:10080/api/v1/http";
		case "webhook":
			return "https://your-webhook-endpoint.com/logs";
		default:
			return "https://...";
	}
});

const endpointHelp = computed(() => {
	switch (settings.value.log_ship_type) {
		case "loki":
			return "Loki push API (path /loki/api/v1/push is added automatically)";
		case "elasticsearch":
			return "Elasticsearch base URL (bulk API is used)";
		case "splunk":
			return "Splunk HEC endpoint (path /services/collector/event is added automatically)";
		case "cribl":
			return "Cribl HTTP source endpoint";
		case "webhook":
			return "Any endpoint accepting JSON POST requests";
		default:
			return "";
	}
});

const tokenHelp = computed(() => {
	switch (settings.value.log_ship_type) {
		case "splunk":
			return "Splunk HEC token (from Settings > Data Inputs > HTTP Event Collector)";
		default:
			return "Bearer token for authentication (optional)";
	}
});

onMounted(async () => {
	try {
		const data = await Api.logship_get();
		settings.value = { ...settings.value, ...data };
	} catch (e) {
		console.error("Failed to load log shipping settings:", e);
	}
	loading.value = false;
});

async function save() {
	saving.value = true;
	testResult.value = null;
	try {
		await Api.logship_save(settings.value);
		testResult.value = {
			success: true,
			message: "Settings saved successfully",
		};
	} catch (e) {
		testResult.value = { success: false, message: "Failed to save settings" };
	}
	saving.value = false;
}

async function testConnection() {
	testing.value = true;
	testResult.value = null;
	try {
		const result = await Api.logship_test({
			type: settings.value.log_ship_type,
			endpoint: settings.value.log_ship_endpoint,
			token: settings.value.log_ship_token,
			index: settings.value.log_ship_index,
			sourcetype: settings.value.log_ship_sourcetype,
		});
		testResult.value = result;
	} catch (e) {
		testResult.value = { success: false, message: "Connection test failed" };
	}
	testing.value = false;
}
</script>
