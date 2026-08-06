<template>
  <div>
    <form @submit.prevent="login" autocomplete="on">
      <div class="form-group row">
        <label for="username" class="col-4 col-form-label">{{ $t('username') }}</label>
        <div class="col-8">
          <input
            @keyup="checkForm"
            @change="checkForm"
            type="text"
            v-model="username"
            autocomplete="username"
            name="username"
            class="form-control"
            id="username"
            placeholder="admin"
            autocorrect="off"
            autocapitalize="none"
          />
        </div>
      </div>
      <div class="form-group row">
        <label for="password" class="col-4 col-form-label">{{ $t('password') }}</label>
        <div class="col-8">
          <input
            @keyup="checkForm"
            @change="checkForm"
            type="password"
            v-model="password"
            autocomplete="current-password"
            name="password"
            class="form-control"
            id="password"
            placeholder="************"
          />
        </div>
      </div>
      <div class="form-group row">
        <div class="col-sm-12">
          <div v-if="error" class="alert alert-danger" role="alert">
            {{ errorMessage || $t('wrong_login') }}
          </div>
          <button
            @click.prevent="login"
            type="submit"
            class="btn btn-block btn-primary"
            :disabled="disabled || loading"
          >
            <font-awesome-icon v-if="loading" icon="circle-notch" class="mr-2" spin />{{
              loading ? $t('loading') : $t('sign_in')
            }}
          </button>
        </div>
      </div>
    </form>

    <a
      v-if="oauth && oauth.gh_client_id"
      @click.prevent="GHlogin"
      href="#"
      class="mt-4 btn btn-block btn-outline-dark"
    >
      <font-awesome-icon :icon="['fab', 'github']" /> Login with Github
    </a>

    <a
      v-if="oauth && oauth.slack_client_id"
      @click.prevent="Slacklogin"
      href="#"
      class="btn btn-block btn-outline-dark"
    >
      <font-awesome-icon :icon="['fab', 'slack']" /> Login with Slack
    </a>

    <a
      v-if="oauth && oauth.google_client_id"
      @click.prevent="Googlelogin"
      href="#"
      class="btn btn-block btn-outline-dark"
    >
      <font-awesome-icon :icon="['fab', 'google']" /> Login with Google
    </a>

    <a
      v-if="oauth && oauth.oidc_enabled && oauth.oidc_client_id"
      @click.prevent="OIDClogin"
      href="#"
      class="btn btn-block btn-outline-dark"
    >
      <font-awesome-icon :icon="['fas', 'key']" /> Login with {{ oauth.oidc_name || 'OIDC' }}
    </a>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import Api from "@/API";
import { useMainStore } from "@/stores/main";

const router = useRouter();
const route = useRoute();
const store = useMainStore();

const username = ref("");
const password = ref("");
const loading = ref(false);
const error = ref(false);
const errorMessage = ref("");
const disabled = ref(true);

onMounted(() => {
	// Check for OAuth redirect errors
	if (route.query.error === "pending_approval") {
		error.value = true;
		errorMessage.value =
			"Account pending approval - please contact an administrator";
	}
});

const core = computed(() => store.core);
const oauth = computed(() => store.oauth);

function checkForm() {
	disabled.value = !username.value || !password.value;
}

async function login() {
	loading.value = true;
	error.value = false;
	errorMessage.value = "";
	try {
		const auth = await Api.login(username.value, password.value);
		if (auth.error) {
			error.value = true;
			errorMessage.value = auth.error;
		} else if (auth.token) {
			await store.loadAdmin();
			store.setAdmin(auth.admin);
			store.setLoggedIn(true);
			router.push("/dashboard");
		}
	} catch (e) {
		error.value = true;
		errorMessage.value = "";
	}
	loading.value = false;
}

function encode(val) {
	return encodeURI(val);
}

async function getOAuthState() {
	try {
		const response = await fetch("api/oauth/state");
		const data = await response.json();
		return data.state;
	} catch (e) {
		console.error("Failed to get OAuth state:", e);
		return null;
	}
}

async function GHlogin() {
	const state = await getOAuthState();
	if (!state) return;
	window.location = `https://github.com/login/oauth/authorize?client_id=${oauth.value.gh_client_id}&redirect_uri=${encode(`${core.value.domain}/oauth/github`)}&scope=read:user,read:org&state=${state}`;
}

async function Slacklogin() {
	const state = await getOAuthState();
	if (!state) return;
	window.location = `https://slack.com/oauth/authorize?client_id=${oauth.value.slack_client_id}&redirect_uri=${encode(`${core.value.domain}/oauth/slack`)}&scope=identity.basic&state=${state}`;
}

async function Googlelogin() {
	const state = await getOAuthState();
	if (!state) return;
	window.location = `https://accounts.google.com/signin/oauth?client_id=${oauth.value.google_client_id}&redirect_uri=${encode(`${core.value.domain}/oauth/google`)}&response_type=code&scope=https://www.googleapis.com/auth/userinfo.profile+https://www.googleapis.com/auth/userinfo.email&state=${state}`;
}

async function OIDClogin() {
	try {
		const response = await fetch("api/oauth/oidc/auth-url");
		const data = await response.json();
		if (data.url) {
			window.location = data.url;
		} else if (data.error) {
			error.value = true;
			errorMessage.value = data.error;
		}
	} catch (e) {
		console.error("Failed to get OIDC auth URL:", e);
		error.value = true;
		errorMessage.value = "Failed to initiate OIDC login";
	}
}
</script>

<style scoped></style>
