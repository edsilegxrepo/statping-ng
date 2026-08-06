<template>
  <div class="col-12">
    <h2>{{ service.name }} Checkins</h2>
    <p class="mb-3">Tell your service to send a routine HTTP request to a Statping Checkin.</p>

    <div v-for="checkin in checkins" :key="checkin.id" class="card text-black-50 bg-white mt-3">
      <div class="card-header text-capitalize">
        {{ checkin.name }}
        <button @click="deleteCheckin(checkin)" class="btn btn-sm small btn-danger float-right text-uppercase">Delete</button>
      </div>
      <div class="card-body">
        <div class="input-group">
          <input type="text" class="form-control" :value="`${coreData.domain}/checkin/${checkin.api_key}`" readonly />
          <div class="input-group-append copy-btn">
            <button @click.prevent="copyUrl(checkin)" class="btn btn-outline-secondary" type="button">Copy</button>
          </div>
        </div>

        <span class="small">Send a GET request to this URL every {{ checkin.interval }} minutes</span>
        <span class="small float-right mt-1 mr-3 d-none d-md-block">Requested {{ ago(checkin.last_hit) }} ago</span>
        <span class="small float-right mt-1 mr-3 d-none d-md-block"
          >Request expected every {{ checkin.interval }} minutes</span
        >

        <div class="card text-black-50 bg-white mt-3">
          <div class="card-header text-capitalize">
            <font-awesome-icon
              @click="expanded = !expanded"
              :icon="expanded ? 'minus' : 'plus'"
              class="mr-2 pointer"
            />
            {{ checkin.name }} Records
          </div>
          <div class="card-body" :class="{ 'd-none': !expanded }">
            <div
              class="alert alert-primary small"
              :class="{ 'alert-success': hit.success, 'alert-danger': !hit.success }"
              v-for="hit in records(checkin)"
              :key="hit.id"
            >
              Checkin {{ hit.success ? 'Request' : 'Failure' }} at {{ hit.created_at }}
            </div>
          </div>
        </div>

        <div class="card text-black-50 bg-white mt-3">
          <div class="card-header text-capitalize">
            <font-awesome-icon
              @click="curl_expanded = !curl_expanded"
              :icon="curl_expanded ? 'minus' : 'plus'"
              class="mr-2 pointer"
            />
            Cronjob Task
          </div>
          <div class="card-body" :class="{ 'd-none': !curl_expanded }">
            This cronjob script will request the checkin endpoint every {{ checkin.interval }} minutes. Add this cronjob
            task to the machine running this service.
            <div class="input-group mt-2">
              <input type="text" class="form-control" :value="getCronJob(checkin)" readonly />
              <div class="input-group-append copy-btn">
                <button @click.prevent="copyCron(checkin)" class="btn btn-outline-secondary" type="button">Copy</button>
              </div>
            </div>
            <span class="small d-block">Using CURL</span>
          </div>
        </div>
      </div>
      <div class="card-footer">
        <span :class="{ 'text-success': last_record(checkin).success, 'text-danger': !last_record(checkin).success }">
          {{
            last_record(checkin).success
              ? 'Checkin is currently working correctly'
              : 'Checkin is currently failing'
          }}
        </span>
      </div>
    </div>

    <div class="card text-black-50 bg-white mt-4">
      <div class="card-header text-capitalize">Create Checkin</div>
      <div class="card-body">
        <form @submit.prevent="saveCheckin">
          <div class="form-group row">
            <div class="col-7 col-md-5">
              <label for="checkin_interval" class="col-form-label">Checkin Name</label>
              <input
                v-model="checkin.name"
                type="text"
                name="name"
                class="form-control"
                id="checkin_name"
                placeholder="New Checkin"
              />
            </div>
            <div class="col-5 col-md-3">
              <label for="checkin_interval" class="col-form-label">Interval (minutes)</label>
              <input
                v-model="checkin.interval"
                type="number"
                name="interval"
                class="form-control"
                id="checkin_interval"
                placeholder="1"
                min="1"
              />
            </div>
            <div class="col-12 col-md-4">
              <label class="col-form-label"></label>
              <button
                :disabled="btn_disabled"
                @click.prevent="saveCheckin"
                type="submit"
                id="submit"
                class="btn btn-primary d-block mt-2"
              >
                Save Checkin
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from "vue";
import { useRoute } from "vue-router";
import Api from "@/API";
import { useMainStore } from "@/stores/main";

const route = useRoute();
const store = useMainStore();

const service = ref({});
const ready = ref(false);
const expanded = ref(false);
const curl_expanded = ref(false);
const checkin = reactive({
	name: "",
	interval: 1,
	service_id: 0,
	hits: [],
	failures: [],
});

const checkins = computed(() => store.serviceCheckins(service.value.id));
const coreData = computed(() => store.core);

const btn_disabled = computed(() => {
	return checkin.name === "" || checkin.interval <= 0;
});

onMounted(async () => {
	if (route.params.id) {
		service.value = await Api.service(route.params.id);
		checkin.service_id = service.value.id;
		ready.value = true;
	}
});

function parseISO(date) {
	return new Date(date);
}

function ago(date) {
	if (!date) return "never";
	const seconds = Math.floor((new Date() - new Date(date)) / 1000);
	if (seconds < 60) return `${seconds} seconds`;
	const minutes = Math.floor(seconds / 60);
	if (minutes < 60) return `${minutes} minutes`;
	const hours = Math.floor(minutes / 60);
	if (hours < 24) return `${hours} hours`;
	const days = Math.floor(hours / 24);
	return `${days} days`;
}

function records(chk) {
	const hits = (chk.hits || []).map((hit) => ({
		success: true,
		created_at: parseISO(hit.created_at),
		id: hit.id,
	}));
	const failures = (chk.failures || []).map((failure) => ({
		success: false,
		created_at: parseISO(failure.created_at),
		id: failure.id,
	}));
	return hits
		.concat(failures)
		.sort((a, b) => a.created_at - b.created_at)
		.reverse()
		.slice(0, 32);
}

function last_record(chk) {
	const r = records(chk);
	if (r.length === 0) return { success: false };
	return r[0];
}

function getCronJob(chk) {
	return `${chk.interval} * * * * /usr/bin/curl ${coreData.value.domain}/checkin/${chk.api_key} >/dev/null 2>&1`;
}

async function copyUrl(chk) {
	await navigator.clipboard.writeText(
		`${coreData.value.domain}/checkin/${chk.api_key}`,
	);
}

async function copyCron(chk) {
	await navigator.clipboard.writeText(getCronJob(chk));
}

async function saveCheckin() {
	checkin.interval = parseInt(checkin.interval, 10);
	await Api.checkin_create(checkin);
	checkin.name = "";
	await load();
}

async function deleteCheckin(chk) {
	await Api.checkin_delete(chk);
	await load();
}

async function load() {
	const chks = await Api.checkins();
	store.setCheckins(chks);
}
</script>
