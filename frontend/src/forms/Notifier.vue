<template>
  <div>
    <form @submit.prevent="saveNotifier">
      <div class="card contain-card mb-3">
        <div class="card-header text-capitalize">
          {{ notifier.title }}
          <span @click="enableToggle" class="switch switch-sm switch-rd-gr float-right">
            <input
              v-model="notifier.enabled"
              type="checkbox"
              :id="`enable_${notifier.method}`"
              :checked="notifier.enabled"
            />
            <label class="mb-0" :for="`enable_${notifier.method}`"></label>
          </span>
        </div>
        <div class="card-body">
          <p class="small text-muted" v-html="sanitizeHtml(notifier.description)" />

          <div v-if="notifier.method === 'mobile'">
            <div class="form-group row mt-3">
              <label for="statping_domain" class="col-sm-4 col-form-label">Statping Domain</label>
              <div class="col-sm-8">
                <div class="input-group">
                  <input :value="coreData.domain" type="text" class="form-control" id="statping_domain" readonly />
                  <div class="input-group-append copy-btn">
                    <button @click.prevent="copyText(coreData.domain)" class="btn btn-outline-secondary" type="button">
                      Copy
                    </button>
                  </div>
                </div>
              </div>
            </div>
            <div class="form-group row mt-3">
              <label for="apisecret" class="col-sm-4 col-form-label">API Secret</label>
              <div class="col-sm-8">
                <div class="input-group">
                  <input :value="coreData.api_secret" type="text" class="form-control" id="apisecret" readonly />
                  <div class="input-group-append copy-btn">
                    <button @click.prevent="copyText(coreData.api_secret)" class="btn btn-outline-secondary" type="button">
                      Copy
                    </button>
                  </div>
                </div>
              </div>
            </div>
            <div class="col-12 col-md-6 offset-0 offset-md-3">
              <img :src="qrcode" class="img-thumbnail" />
              <span class="text-muted small center">Scan this QR Code on the Statping Mobile App for quick setup</span>
            </div>
          </div>

          <div v-if="notifier.method !== 'mobile'" v-for="(form, index) in notifier.form" :key="index" class="form-group">
            <label class="text-capitalize">{{ form.title }}</label>
            <input
              v-if="formVisible(['text', 'number', 'password', 'email'], form)"
              v-model="notifier[form.field.toLowerCase()]"
              :type="form.type"
              class="form-control"
              :placeholder="form.placeholder"
            />

            <select
              v-if="formVisible(['list'], form)"
              v-model="notifier[form.field.toLowerCase()]"
              class="form-control"
            >
              <option v-for="val in form.list_options" :key="val" :value="val">{{ val }}</option>
            </select>

            <span
              v-if="formVisible(['switch'], form)"
              @click="notifier[form.field.toLowerCase()] = !!notifier[form.field.toLowerCase()]"
              class="switch switch-rd-gr float-right mt-2"
            >
              <input
                v-model="notifier[form.field.toLowerCase()]"
                type="checkbox"
                class="switch-sm"
                :id="`switch_${notifier.name}_${form.field}`"
                :checked="notifier[form.field.toLowerCase()]"
              />
              <label class="mb-0" :for="`switch_${notifier.name}_${form.field}`"></label>
            </span>

            <small class="form-text text-muted" v-html="sanitizeHtml(form.small_text)"></small>
          </div>

          <div class="row mt-4">
            <div class="col-sm-12">
              <span class="slider-info">Limit {{ notifier.limits }} per hour</span>
              <input v-model.number="notifier.limits" type="range" name="limits" class="slider" min="1" max="300" />
              <small class="form-text text-muted"
                >Notifier '{{ notifier.title }}' will send a maximum of {{ notifier.limits }} notifications per hour.</small
              >
            </div>
          </div>
        </div>
      </div>

      <div v-if="notifier.data_type" class="card mb-3">
        <div class="card-header text-capitalize">
          <font-awesome-icon @click="expanded = !expanded" :icon="expanded ? 'minus' : 'plus'" class="mr-2 pointer" />
          {{ notifier.title }} Outgoing Request
          <span class="badge badge-dark float-right text-uppercase mt-1">{{ notifier.data_type }}</span>
        </div>
        <div class="card-body" :class="{ 'd-none': !expanded }">
          <span class="text-muted d-block mb-3" v-if="notifier.request_info" v-html="sanitizeHtml(notifier.request_info)"></span>

          <div class="row">
            <div class="col-12">
              <h5 class="text-capitalize">Success Data</h5>
              <textarea v-model="success_data" class="form-control code-editor" rows="8"></textarea>
            </div>
          </div>

          <div class="row mt-4">
            <div class="col-12">
              <h5 class="text-capitalize">Failure Data</h5>
              <textarea v-model="failure_data" class="form-control code-editor" rows="8"></textarea>
            </div>
          </div>
        </div>
      </div>
    </form>

    <div v-if="error || success" class="card mb-3">
      <div class="card-body">
        <div v-if="error && !success" class="alert alert-danger col-12" role="alert">{{ error }}</div>
        <div v-if="success" class="alert alert-success col-12" role="alert">
          <span class="text-capitalize">{{ notifier.title }}</span> appears to be working!
        </div>

        <h5>Response</h5>
        <pre class="bg-light p-2">{{ response }}</pre>
      </div>
    </div>

    <div class="card mb-3">
      <div class="card-body">
        <div class="notifier-actions">
          <button
            @click.prevent="saveNotifier"
            :disabled="loading"
            type="submit"
            class="btn text-capitalize btn-primary save-notifier"
          >
            <font-awesome-icon v-if="loading" icon="circle-notch" class="mr-2" spin />
            {{ loading ? 'Loading...' : saved ? 'Saved' : 'Save' }}
          </button>
          <button
            @click.prevent="testNotifier('success')"
            :disabled="loadingTest"
            class="btn btn-secondary text-capitalize test-notifier"
          >
            <font-awesome-icon v-if="loadingTest" icon="circle-notch" class="mr-2" spin />{{
              loadingTest ? 'Loading...' : 'Test Success'
            }}
          </button>
          <button
            @click.prevent="testNotifier('failure')"
            :disabled="loadingTest"
            class="btn btn-secondary text-capitalize test-notifier"
          >
            <font-awesome-icon v-if="loadingTest" icon="circle-notch" class="mr-2" spin />{{
              loadingTest ? 'Loading...' : 'Test Failure'
            }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="notifier.logs && notifier.logs.length" class="card mb-3">
      <div class="card-header text-capitalize">
        <font-awesome-icon
          @click="expanded_logs = !expanded_logs"
          :icon="expanded_logs ? 'minus' : 'plus'"
          class="mr-2 pointer"
        />
        {{ notifier.title }} Logs
        <span class="badge badge-info float-right text-uppercase mt-1">{{ notifier.logs.length }}</span>
      </div>
      <div class="card-body" :class="{ 'd-none': !expanded_logs }">
        <div
          v-for="(log, i) in notifier.logs"
          :key="i"
          class="alert"
          :class="{
            'alert-danger': log.error,
            'alert-dark': !log.success && !log.error,
            'alert-success': log.success && !log.error,
          }"
        >
          <span class="d-block">
            Service {{ log.service }}
            {{ log.success ? 'Success Triggered' : 'Failure Triggered' }}
          </span>

          <div v-if="log.message !== ''" class="bg-white p-3 small mt-2">
            <code>{{ log.message }}</code>
          </div>

          <div class="row mt-2">
            <span class="col-6 small">{{ niceDate(log.created_at) }}</span>
          </div>
        </div>
      </div>
    </div>

    <span class="d-block small text-center mb-3">
      <span class="text-capitalize">{{ notifier.title }}</span> Notifier created by
      <a :href="notifier.author_url" target="_blank">{{ notifier.author }}</a>
    </span>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from "vue";
import Api from "@/API";
import { isNumeric, sanitizeHtml } from "@/mixins";
import { useMainStore } from "@/stores/main";

const props = defineProps({
	notifier: {
		type: Object,
		required: true,
	},
});

const store = useMainStore();

const loading = ref(false);
const loadingTest = ref(false);
const error = ref(null);
const response = ref(null);
const success = ref(false);
const saved = ref(false);
const expanded = ref(false);
const expanded_logs = ref(false);
const success_data = ref(null);
const failure_data = ref(null);
const form = reactive({});

const coreData = computed(() => store.core);

const qrcode = computed(() => {
	const u = `statping://setup?domain=${coreData.value.domain}&api=${coreData.value.api_secret}`;
	return (
		"https://chart.googleapis.com/chart?chs=500x500&cht=qr&chl=" +
		encodeURIComponent(u)
	);
});

onMounted(() => {
	success_data.value = props.notifier.success_data;
	failure_data.value = props.notifier.failure_data;
});

function formVisible(want, formField) {
	return want.includes(formField.type);
}

async function copyText(text) {
	await navigator.clipboard.writeText(text);
}

function niceDate(date) {
	return new Date(date).toLocaleString();
}

async function enableToggle() {
	props.notifier.enabled = !!props.notifier.enabled;
	const formData = {
		enabled: !props.notifier.enabled,
		method: props.notifier.method,
	};
	await Api.notifier_save(formData);
}

async function saveNotifier() {
	loading.value = true;
	form.enabled = props.notifier.enabled;
	form.limits = parseInt(props.notifier.limits, 10);
	form.method = props.notifier.method;
	if (props.notifier.form) {
		props.notifier.form.forEach((f) => {
			const field = f.field.toLowerCase();
			let val = props.notifier[field];
			if (isNumeric(val) && form.method !== "telegram") {
				val = parseInt(val, 10);
			}
			form[field] = val;
		});
	}
	form.success_data = success_data.value;
	form.failure_data = failure_data.value;
	await Api.notifier_save(form);
	const notifiers = await Api.notifiers();
	store.setNotifiers(notifiers);
	saved.value = true;
	loading.value = false;
}

async function testNotifier(method = "success") {
	success.value = false;
	loadingTest.value = true;
	form.method = props.notifier.method;
	if (props.notifier.form) {
		props.notifier.form.forEach((f) => {
			const field = f.field.toLowerCase();
			let val = props.notifier[field];
			if (isNumeric(val) && form.method !== "telegram") {
				val = parseInt(val, 10);
			}
			form[field] = val;
		});
	}
	const req = {
		notifier: form,
		method: method,
	};
	const tested = await Api.notifier_test(props.notifier.method, req);
	if (tested.success) {
		success.value = true;
	} else {
		error.value = tested.error;
	}
	response.value = tested.response;
	loadingTest.value = false;
}
</script>

<style scoped>
.code-editor {
  font-family: monospace;
  font-size: 12px;
}
.pointer {
  cursor: pointer;
}
.notifier-actions {
  display: flex;
  justify-content: center;
  gap: 1rem;
  flex-wrap: wrap;
}
.notifier-actions .btn {
  min-width: 140px;
}
</style>
