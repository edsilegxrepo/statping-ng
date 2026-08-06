<template>
  <div class="col-12">
    <div v-for="incident in incidents" :key="incident.id" class="card contain-card mb-4">
      <div class="card-header">
        Incident: {{ incident.title }}
        <button @click="deleteIncident(incident)" class="btn btn-sm btn-danger float-right">
          <font-awesome-icon icon="times" />
        </button>
      </div>

      <FormIncidentUpdates :incident="incident" />

      <span class="font-2 p-2 pl-3"
        >Created: {{ niceDate(incident.created_at) }} | Last Update: {{ niceDate(incident.updated_at) }}</span
      >
    </div>

    <div class="card contain-card">
      <div class="card-header">Create Incident</div>
      <div class="card-body">
        <form @submit.prevent="createIncident">
          <div class="form-group row">
            <label class="col-sm-4 col-form-label">Title</label>
            <div class="col-sm-8">
              <input
                v-model="incident.title"
                type="text"
                name="title"
                class="form-control"
                id="title"
                placeholder="Incident Title"
                required
              />
            </div>
          </div>

          <div class="form-group row">
            <label class="col-sm-4 col-form-label">Description</label>
            <div class="col-sm-8">
              <textarea
                v-model="incident.description"
                rows="5"
                name="description"
                class="form-control"
                id="description"
                required
              ></textarea>
            </div>
          </div>

          <div class="form-group row">
            <div class="col-sm-12">
              <button :disabled="submitting || !canCreateIncident" type="submit" class="btn btn-block btn-primary">
                {{ submitting ? 'Creating Incident...' : 'Create Incident' }}
              </button>
            </div>
          </div>
          <div v-if="errorMessage" class="alert alert-danger" role="alert">{{ errorMessage }}</div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from "vue";
import { useRoute } from "vue-router";
import Api from "@/API";
import FormIncidentUpdates from "@/forms/IncidentUpdates.vue";
import { useMainStore } from "@/stores/main";

const route = useRoute();
const store = useMainStore();

const serviceID = ref(0);
const submitting = ref(false);
const errorMessage = ref("");
const incidents = ref([]);
const incident = reactive({
	title: "",
	description: "",
	service: 0,
});

const canCreateIncident = computed(() => {
	return (
		incident.title.trim().length > 0 && incident.description.trim().length > 0
	);
});

onMounted(async () => {
	serviceID.value = Number(route.params.id);
	incident.service = Number(route.params.id);
	await loadIncidents();
});

function niceDate(date) {
	return new Date(date).toLocaleString();
}

function extractErrorMessage(error, fallback) {
	const responseData = error?.response?.data || error;
	if (typeof responseData === "string" && responseData.trim()) {
		return responseData.trim();
	}
	if (typeof responseData?.error === "string" && responseData.error.trim()) {
		return responseData.error;
	}
	if (responseData?.error?.message) {
		return responseData.error.message;
	}
	if (responseData?.message) {
		return responseData.message;
	}
	if (error?.message) {
		return error.message;
	}
	return fallback;
}

async function deleteConfirm(i) {
	const res = await Api.incident_delete(i);
	if (res.status === "success") {
		incidents.value = incidents.value.filter((obj) => obj.id !== i.id);
	}
}

function deleteIncident(inc) {
	store.setModal({
		visible: true,
		title: "Delete Incident",
		body: `Are you sure you want to delete Incident ${inc.title}?`,
		btnColor: "btn-danger",
		btnText: "Delete Incident",
		func: () => deleteConfirm(inc),
	});
}

async function createIncident() {
	if (submitting.value) return;

	const title = incident.title.trim();
	const description = incident.description.trim();

	if (!title || !description) {
		errorMessage.value = "Incident title and description are required.";
		return;
	}

	submitting.value = true;
	errorMessage.value = "";

	try {
		const response = await Api.incident_create(serviceID.value, {
			...incident,
			title,
			description,
			service: serviceID.value,
		});

		if (response?.status === "success" && response.output) {
			incidents.value.push(response.output);
			incident.title = "";
			incident.description = "";
			incident.service = serviceID.value;
			return;
		}

		errorMessage.value = extractErrorMessage(
			response,
			"Unable to create the incident right now.",
		);
	} catch (error) {
		errorMessage.value = extractErrorMessage(
			error,
			"Unable to create the incident right now.",
		);
	} finally {
		submitting.value = false;
	}
}

async function loadIncidents() {
	incidents.value = await Api.incidents_service(serviceID.value);
}
</script>
