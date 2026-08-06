<template>
  <div v-if="service" class="col-12">
    <h3>
      {{ service.name }} Failures
      <button v-if="failures.length > 0" @click="deleteFailures" class="btn btn-danger float-right">Delete All</button>
    </h3>

    <div class="card mt-4 mb-4">
      <div class="card-header">
        Search and Filter
        <span class="float-right">
          <span class="switch mr-3">
            <input
              v-model="allRecords"
              @change="onAllRecordsChange"
              type="checkbox"
              class="switch"
              id="allrecords"
            />
            <label for="allrecords">All Records</label>
          </span>
          <font-awesome-icon v-if="loading" icon="circle-notch" spin />
        </span>
      </div>
      <div class="card-body">
        <form>
          <div class="form-row">
            <div class="col">
              <label for="fromdate">From Date</label>
              <flat-pickr
                id="fromdate"
                :disabled="loading || allRecords"
                @on-change="onDateChange"
                v-model="start_time"
                :config="dateConfig"
                type="text"
                class="form-control text-left d-block"
                required
              />
            </div>
            <div class="col">
              <label for="todate">To Date</label>
              <flat-pickr
                id="todate"
                :disabled="loading || allRecords"
                @on-change="onDateChange"
                v-model="end_time"
                :config="dateConfig"
                type="text"
                class="form-control text-left d-block"
                required
              />
            </div>
            <div class="col">
              <label for="search">Search Terms</label>
              <input id="search" type="text" v-model="search" :disabled="allRecords" class="form-control" />
            </div>
          </div>
          <div class="form-row mt-3">
            <div class="col">
              <span class="switch float-left">
                <input
                  v-model="show_checkins"
                  type="checkbox"
                  class="switch"
                  id="showcheckins"
                  :disabled="allRecords"
                />
                <label v-if="show_checkins" for="showcheckins">Showing Checkin Failures</label>
                <label v-else for="showcheckins">View Checkin Failures</label>
              </span>
              <span class="switch float-right ml-3">
                <input
                  v-model="autoRefresh"
                  type="checkbox"
                  class="switch"
                  id="autorefresh"
                  :disabled="allRecords"
                />
                <label for="autorefresh">Auto-refresh</label>
              </span>
              <button type="button" @click="searchFailures" :disabled="autoRefresh || allRecords" class="btn btn-outline-danger float-right">Search</button>
            </div>
          </div>
        </form>
      </div>
    </div>

    <div v-if="failures.length === 0" class="alert alert-info">
      <span v-if="search"> Could not find any failures with issue: "{{ search }}" </span>
      <span v-else> You don't have any failures for {{ service.name }}. Way to go! </span>
    </div>

    <table v-else class="table">
      <thead>
        <tr>
          <th scope="col">#</th>
          <th scope="col">Issue</th>
          <th scope="col">Status Code</th>
          <th scope="col">Ping</th>
          <th scope="col">Created</th>
          <th scope="col">Date/Time</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(failure, index) in failures" :key="index">
          <th class="font-1" scope="row">{{ failure.id }}</th>
          <td class="font-1">{{ failure.issue }}</td>
          <td class="font-1">{{ failure.error_code }}</td>
          <td class="font-1">{{ humanTime(failure.ping) }}</td>
          <td class="font-1">{{ ago(failure.created_at) }}</td>
          <td class="font-1">{{ formatDateTime(failure.created_at) }}</td>
        </tr>
      </tbody>
    </table>

    <nav v-if="total > 4 && failures.length !== 0" class="mt-3">
      <ul class="pagination justify-content-center">
        <li class="page-item" :class="{ disabled: page === 1 }">
          <span @click="page > 1 && gotoPage(1)" class="page-link" style="cursor: pointer" aria-label="First">
            <span aria-hidden="true">&laquo;&laquo;</span>
          </span>
        </li>
        <li class="page-item" :class="{ disabled: page === 1 }">
          <span @click="page > 1 && gotoPage(page - 1)" class="page-link" style="cursor: pointer" aria-label="Previous">
            <span aria-hidden="true">&laquo;</span>
          </span>
        </li>
        <li v-if="page > 3" class="page-item">
          <span @click="gotoPage(1)" class="page-link" style="cursor: pointer">1</span>
        </li>
        <li v-if="page > 4" class="page-item disabled">
          <span class="page-link">...</span>
        </li>
        <li v-for="n in visiblePages" :key="n" class="page-item" :class="{ active: page === n }">
          <span @click="gotoPage(n)" class="page-link" style="cursor: pointer">{{ n }}</span>
        </li>
        <li v-if="page < maxPages - 3" class="page-item disabled">
          <span class="page-link">...</span>
        </li>
        <li v-if="page < maxPages - 2" class="page-item">
          <span @click="gotoPage(maxPages)" class="page-link" style="cursor: pointer">{{ maxPages }}</span>
        </li>
        <li class="page-item" :class="{ disabled: page === maxPages }">
          <span
            @click="page < maxPages && gotoPage(page + 1)"
            class="page-link"
            style="cursor: pointer"
            aria-label="Next"
          >
            <span aria-hidden="true">&raquo;</span>
          </span>
        </li>
        <li class="page-item" :class="{ disabled: page === maxPages }">
          <span @click="page < maxPages && gotoPage(maxPages)" class="page-link" style="cursor: pointer" aria-label="Last">
            <span aria-hidden="true">&raquo;&raquo;</span>
          </span>
        </li>
      </ul>
      <div class="text-center">
        <span>{{ total }} Failures</span>
        <span class="ml-4">
          Per page:
          <select v-model.number="limit" @change="onLimitChange" class="form-control form-control-sm d-inline-block" style="width: auto">
            <option :value="25">25</option>
            <option :value="50">50</option>
            <option :value="100">100</option>
            <option :value="200">200</option>
          </select>
        </span>
      </div>
    </nav>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from "vue";
import flatPickr from "vue-flatpickr-component";
import { useRoute } from "vue-router";
import { useMainStore } from "@/stores/main";
import "flatpickr/dist/flatpickr.css";
import Api from "@/API";

const route = useRoute();
const store = useMainStore();

const loading = ref(true);
const search = ref("");
const show_checkins = ref(false);
const autoRefresh = ref(false);
const allRecords = ref(false);
const service = ref(null);
const fails = ref([]);
const limit = ref(50);
const offset = ref(0);
const total = ref(0);
const page = ref(1);
const start_time = ref(nowSubtract(216000).toISOString());
const end_time = ref(nowSubtract(0).toISOString());

const dateConfig = {
	wrap: true,
	allowInput: false,
	enableTime: true,
	dateFormat: "Z",
	altInput: true,
	altFormat: "Y-m-d h:i K",
	maxDate: new Date(),
};

const failures = computed(() => {
	let sorted = fails.value;
	if (allRecords.value) {
		return sorted;
	}
	if (show_checkins.value) {
		sorted = sorted.filter((f) => f.method === "checkin");
	} else {
		sorted = sorted.filter((f) => f.method !== "checkin");
	}
	if (search.value !== "") {
		sorted = sorted.filter((f) =>
			f.issue.toLowerCase().includes(search.value.toLowerCase()),
		);
	}
	return sorted;
});

const maxPages = computed(() => Math.ceil(total.value / limit.value));

const visiblePages = computed(() => {
	const pages = [];
	const start = Math.max(1, page.value - 2);
	const end = Math.min(maxPages.value, page.value + 2);
	for (let i = start; i <= end; i++) {
		pages.push(i);
	}
	return pages;
});

watch(() => route.params.id, reloadTimes);

onMounted(async () => {
	service.value = await Api.service(route.params.id);
	total.value = service.value.stats?.failures || 0;
	await gotoPage(1);
});

function nowSubtract(seconds) {
	return new Date(Date.now() - seconds * 1000);
}

function toUnix(date) {
	return Math.floor(date.getTime() / 1000);
}

function parseISO(date) {
	return new Date(date);
}

function humanTime(ms) {
	if (!ms) return "0ms";
	if (ms >= 10000) return `${Math.round(ms / 1000)} ms`;
	return `${ms} μs`;
}

function ago(date) {
	if (!date) return "never";
	const seconds = Math.floor((new Date() - new Date(date)) / 1000);
	if (seconds < 60) return `${seconds} seconds ago`;
	const minutes = Math.floor(seconds / 60);
	if (minutes < 60) return `${minutes} minutes ago`;
	const hours = Math.floor(minutes / 60);
	if (hours < 24) return `${hours} hours ago`;
	const days = Math.floor(hours / 24);
	return `${days} days ago`;
}

function formatDateTime(date) {
	if (!date) return "";
	const d = new Date(date);
	return d.toLocaleString("en-US", {
		year: "numeric",
		month: "2-digit",
		day: "2-digit",
		hour: "2-digit",
		minute: "2-digit",
		second: "2-digit",
		hour12: false,
	});
}

async function reloadTimes() {
	if (route.params.id) {
		service.value = await Api.service(route.params.id);
		total.value = service.value.stats?.failures || 0;
		await gotoPage(1);
	}
}

async function deleteConfirm() {
	await Api.service_failures_delete(service.value);
	service.value = await Api.service(service.value.id);
	total.value = 0;
	await load();
}

function deleteFailures() {
	store.setModal({
		visible: true,
		title: "Delete All Failures",
		body: `Are you sure you want to delete all Failures for service ${service.value.name}?`,
		btnColor: "btn-danger",
		btnText: "Delete Failures",
		func: () => deleteConfirm(),
	});
}

async function gotoPage(p) {
	if (p < 1 || p > maxPages.value) return;
	page.value = p;
	offset.value = (p - 1) * limit.value;
	await load();
}

async function load() {
	loading.value = true;
	if (allRecords.value) {
		fails.value = await Api.service_failures(
			service.value.id,
			0,
			toUnix(new Date()),
			limit.value,
			offset.value,
		);
	} else {
		fails.value = await Api.service_failures(
			service.value.id,
			toUnix(parseISO(start_time.value)),
			toUnix(parseISO(end_time.value)),
			limit.value,
			offset.value,
		);
	}
	loading.value = false;
}

async function searchFailures() {
	page.value = 1;
	offset.value = 0;
	await load();
}

async function onAllRecordsChange() {
	page.value = 1;
	offset.value = 0;
	await load();
}

async function onDateChange() {
	if (autoRefresh.value) {
		page.value = 1;
		offset.value = 0;
		await load();
	}
}

async function onLimitChange() {
	page.value = 1;
	offset.value = 0;
	await load();
}
</script>
