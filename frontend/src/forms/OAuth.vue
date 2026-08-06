<template>
  <form @submit.prevent="saveOAuth">
    <div class="card mb-3">
      <div class="card-header">
        Internal Login
        <span @click="local_enabled = !!local_enabled" class="switch switch-sm switch-rd-gr float-right">
          <input v-model="local_enabled" type="checkbox" id="switch-internal-oauth" :checked="local_enabled" />
          <label for="switch-internal-oauth" class="mb-0"> </label>
        </span>
      </div>
      <div class="card-body">Use Statping's default authentication to allow users you've created to login.</div>
    </div>
    <div class="card mb-3">
      <div class="card-header text-capitalize">
        <font-awesome-icon
          @click="expanded.github = !expanded.github"
          :icon="expanded.github ? 'minus' : 'plus'"
          class="mr-2 pointer"
        />
        Github Settings
        <span @click="github_enabled = !!github_enabled" class="switch switch-sm switch-rd-gr float-right">
          <input v-model="github_enabled" type="checkbox" id="switch-gh-oauth" :checked="github_enabled" />
          <label class="mb-0" for="switch-gh-oauth"> </label>
        </span>
      </div>
      <div class="card-body" :class="{ 'd-none': !expanded.github }">
        <span
          >You will need to create a new <a href="https://github.com/settings/developers">OAuth App</a> within Github.</span
        >

        <div class="form-group row mt-3">
          <label for="github_client" class="col-sm-4 col-form-label">Github Client ID</label>
          <div class="col-sm-8">
            <input v-model="oauth.gh_client_id" type="text" class="form-control" id="github_client" required />
          </div>
        </div>
        <div class="form-group row">
          <label for="github_secret" class="col-sm-4 col-form-label">Github Client Secret</label>
          <div class="col-sm-8">
            <input v-model="oauth.gh_client_secret" type="text" class="form-control" id="github_secret" required />
          </div>
        </div>
        <div class="form-group row">
          <label for="github_users" class="col-sm-4 col-form-label">Restrict Users</label>
          <div class="col-sm-8">
            <input
              v-model="oauth.gh_users"
              type="text"
              class="form-control"
              id="github_users"
              placeholder="user1,user2,user3"
            />
            <small>Optional comma delimited list of usernames</small>
          </div>
        </div>
        <div class="form-group row">
          <label for="github_orgs" class="col-sm-4 col-form-label">Restrict Organizations</label>
          <div class="col-sm-8">
            <input
              v-model="oauth.gh_orgs"
              type="text"
              class="form-control"
              id="github_orgs"
              placeholder="statping,github"
            />
            <small>Optional comma delimited list of Github Organizations</small>
          </div>
        </div>
        <div class="form-group row">
          <label for="gh_callback" class="col-sm-4 col-form-label">Callback URL</label>
          <div class="col-sm-8">
            <div class="input-group">
              <input :value="`${coreData.domain}/oauth/github`" type="text" class="form-control" id="gh_callback" readonly />
              <div class="input-group-append copy-btn">
                <button @click.prevent="copyCallback('github')" class="btn btn-outline-secondary" type="button">Copy</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div class="card mb-3">
      <div class="card-header">
        <font-awesome-icon
          @click="expanded.google = !expanded.google"
          :icon="expanded.google ? 'minus' : 'plus'"
          class="mr-2 pointer"
        />
        Google Settings
        <span @click="google_enabled = !!google_enabled" class="switch switch-sm switch-rd-gr float-right">
          <input v-model="google_enabled" type="checkbox" id="switch-google-oauth" :checked="google_enabled" />
          <label for="switch-google-oauth" class="mb-0"> </label>
        </span>
      </div>
      <div class="card-body" :class="{ 'd-none': !expanded.google }">
        <span
          >Go to <a href="https://console.cloud.google.com/apis/credentials">OAuth Consent Screen</a> on Google Console to
          create a new "Web Application" OAuth application.
        </span>

        <div class="form-group row mt-3">
          <label for="google_client" class="col-sm-4 col-form-label">Google Client ID</label>
          <div class="col-sm-8">
            <input v-model="oauth.google_client_id" type="text" class="form-control" id="google_client" required />
          </div>
        </div>
        <div class="form-group row">
          <label for="google_secret" class="col-sm-4 col-form-label">Google Client Secret</label>
          <div class="col-sm-8">
            <input v-model="oauth.google_client_secret" type="text" class="form-control" id="google_secret" required />
          </div>
        </div>
        <div class="form-group row">
          <label for="google_users" class="col-sm-4 col-form-label">Restrict Users</label>
          <div class="col-sm-8">
            <input
              v-model="oauth.google_users"
              type="text"
              class="form-control"
              id="google_users"
              placeholder="info@gmail.com,example.com"
            />
            <small>Optional comma delimited list of emails and/or domains</small>
          </div>
        </div>
        <div class="form-group row">
          <label for="google_callback" class="col-sm-4 col-form-label">Callback URL</label>
          <div class="col-sm-8">
            <div class="input-group">
              <input
                :value="`${coreData.domain}/oauth/google`"
                type="text"
                class="form-control"
                id="google_callback"
                readonly
              />
              <div class="input-group-append copy-btn">
                <button @click.prevent="copyCallback('google')" class="btn btn-outline-secondary" type="button">Copy</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div class="card mb-3">
      <div class="card-header">
        <font-awesome-icon
          @click="expanded.slack = !expanded.slack"
          :icon="expanded.slack ? 'minus' : 'plus'"
          class="mr-2 pointer"
        />
        Slack Settings
        <span @click="slack_enabled = !!slack_enabled" class="switch switch-sm switch-rd-gr float-right">
          <input v-model="slack_enabled" type="checkbox" id="switch-slack-oauth" :checked="slack_enabled" />
          <label for="switch-slack-oauth" class="mb-0"> </label>
        </span>
      </div>
      <div class="card-body" :class="{ 'd-none': !expanded.slack }">
        <span>Go to <a href="https://api.slack.com/apps">Slack Apps</a> and create a new Application.</span>

        <div class="form-group row mt-3">
          <label for="slack_client" class="col-sm-4 col-form-label">Slack Client ID</label>
          <div class="col-sm-8">
            <input v-model="oauth.slack_client_id" type="text" class="form-control" id="slack_client" required />
          </div>
        </div>
        <div class="form-group row">
          <label for="slack_secret" class="col-sm-4 col-form-label">Slack Client Secret</label>
          <div class="col-sm-8">
            <input v-model="oauth.slack_client_secret" type="text" class="form-control" id="slack_secret" required />
          </div>
        </div>
        <div class="form-group row">
          <label for="slack_team" class="col-sm-4 col-form-label">Team ID</label>
          <div class="col-sm-8">
            <input v-model="oauth.slack_team" type="text" class="form-control" id="slack_team" />
            <small>Optional Slack Team ID</small>
          </div>
        </div>
        <div class="form-group row">
          <label for="slack_users" class="col-sm-4 col-form-label">Restrict Users</label>
          <div class="col-sm-8">
            <input
              v-model="oauth.slack_users"
              type="text"
              class="form-control"
              id="slack_users"
              placeholder="info@example.com,info@domain.net"
            />
            <small>Optional comma delimited list of email addresses</small>
          </div>
        </div>
        <div class="form-group row">
          <label for="slack_callback" class="col-sm-4 col-form-label">Callback URL</label>
          <div class="col-sm-8">
            <div class="input-group">
              <input :value="`${coreData.domain}/oauth/slack`" type="text" class="form-control" id="slack_callback" readonly />
              <div class="input-group-append copy-btn">
                <button @click.prevent="copyCallback('slack')" class="btn btn-outline-secondary" type="button">Copy</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="card mb-3">
      <div class="card-header">
        <font-awesome-icon
          @click="expanded.oidc = !expanded.oidc"
          :icon="expanded.oidc ? 'minus' : 'plus'"
          class="mr-2 pointer"
        />
        OpenID Connect (OIDC)
        <span @click="oidc_enabled = !oidc_enabled" class="switch switch-sm switch-rd-gr float-right">
          <input v-model="oidc_enabled" type="checkbox" id="switch-oidc" :checked="oidc_enabled" />
          <label for="switch-oidc" class="mb-0"> </label>
        </span>
      </div>
      <div class="card-body" :class="{ 'd-none': !expanded.oidc }">
        <div class="alert alert-info mb-3">
          <strong>OpenID Connect</strong> provides standardized authentication with any OIDC-compliant
          identity provider: Keycloak, Authentik, Azure AD, Okta, Auth0, Google Workspace, and others.
          Endpoints are auto-discovered from the issuer URL.
        </div>

        <div class="form-group row">
          <label for="oidc_name" class="col-sm-4 col-form-label">Provider Name</label>
          <div class="col-sm-8">
            <input v-model="oauth.oidc_name" type="text" class="form-control" id="oidc_name" placeholder="Keycloak" />
            <small class="text-muted">Display name shown on login button</small>
          </div>
        </div>
        <div class="form-group row">
          <label for="oidc_issuer" class="col-sm-4 col-form-label">Issuer URL</label>
          <div class="col-sm-8">
            <input v-model="oauth.oidc_issuer_url" type="url" class="form-control" id="oidc_issuer" placeholder="https://keycloak.example.com/realms/main" />
            <small class="text-muted">Base URL for OIDC discovery (/.well-known/openid-configuration)</small>
          </div>
        </div>
        <div class="form-group row">
          <label for="oidc_client_id" class="col-sm-4 col-form-label">Client ID</label>
          <div class="col-sm-8">
            <input v-model="oauth.oidc_client_id" type="text" class="form-control" id="oidc_client_id" />
          </div>
        </div>
        <div class="form-group row">
          <label for="oidc_client_secret" class="col-sm-4 col-form-label">Client Secret</label>
          <div class="col-sm-8">
            <input v-model="oauth.oidc_client_secret" type="password" class="form-control" id="oidc_client_secret" />
          </div>
        </div>
        <div class="form-group row">
          <label for="oidc_scopes" class="col-sm-4 col-form-label">Scopes</label>
          <div class="col-sm-8">
            <input v-model="oauth.oidc_scopes" type="text" class="form-control" id="oidc_scopes" placeholder="openid,profile,email,groups" />
            <small class="text-muted">Comma-separated scopes (openid is required and added automatically)</small>
          </div>
        </div>

        <hr />
        <h6>Authorization</h6>

        <div class="form-group row">
          <label for="oidc_allowed_users" class="col-sm-4 col-form-label">Allowed Users/Domains</label>
          <div class="col-sm-8">
            <input v-model="oauth.oidc_allowed_users" type="text" class="form-control" id="oidc_allowed_users" placeholder="admin@example.com,@company.com" />
            <small class="text-muted">Comma-separated emails or @domain.com patterns. Leave empty to allow all.</small>
          </div>
        </div>
        <div class="form-group row">
          <label for="oidc_admin_groups" class="col-sm-4 col-form-label">Admin Groups</label>
          <div class="col-sm-8">
            <input v-model="oauth.oidc_admin_groups" type="text" class="form-control" id="oidc_admin_groups" placeholder="admins;statping-admins" />
            <small class="text-muted">Semicolon-separated group names that grant admin access</small>
          </div>
        </div>

        <hr />
        <h6>Claim Mapping (Optional)</h6>

        <div class="form-group row">
          <label for="oidc_claim_username" class="col-sm-4 col-form-label">Username Claim</label>
          <div class="col-sm-8">
            <input v-model="oauth.oidc_claim_username" type="text" class="form-control" id="oidc_claim_username" placeholder="preferred_username" />
          </div>
        </div>
        <div class="form-group row">
          <label for="oidc_claim_email" class="col-sm-4 col-form-label">Email Claim</label>
          <div class="col-sm-8">
            <input v-model="oauth.oidc_claim_email" type="text" class="form-control" id="oidc_claim_email" placeholder="email" />
          </div>
        </div>
        <div class="form-group row">
          <label for="oidc_claim_groups" class="col-sm-4 col-form-label">Groups Claim</label>
          <div class="col-sm-8">
            <input v-model="oauth.oidc_claim_groups" type="text" class="form-control" id="oidc_claim_groups" placeholder="groups" />
          </div>
        </div>

        <hr />
        <h6>Security Options</h6>

        <div class="form-group row">
          <label for="switch-oidc-pkce" class="col-sm-4 col-form-label">Use PKCE</label>
          <div class="col-sm-8">
            <span @click="oauth.oidc_use_pkce = !oauth.oidc_use_pkce" class="switch switch-rd-gr">
              <input v-model="oauth.oidc_use_pkce" type="checkbox" id="switch-oidc-pkce" :checked="oauth.oidc_use_pkce" />
              <label for="switch-oidc-pkce" class="mb-0"> </label>
            </span>
            <small class="text-muted d-block">Proof Key for Code Exchange - recommended, required by some IdPs (Azure AD)</small>
          </div>
        </div>

        <div class="form-group row">
          <label for="oidc_callback" class="col-sm-4 col-form-label">Callback URL</label>
          <div class="col-sm-8">
            <div class="input-group">
              <input :value="`${coreData.domain}/oauth/oidc`" type="text" class="form-control" id="oidc_callback" readonly />
              <div class="input-group-append copy-btn">
                <button @click.prevent="copyCallback('oidc')" class="btn btn-outline-secondary" type="button">Copy</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="card mb-3">
      <div class="card-header">
        <font-awesome-icon
          @click="expanded.forwardauth = !expanded.forwardauth"
          :icon="expanded.forwardauth ? 'minus' : 'plus'"
          class="mr-2 pointer"
        />
        Forward Auth (Authelia, Authentik, etc.)
        <span @click="forwardauth.enabled = !forwardauth.enabled" class="switch switch-sm switch-rd-gr float-right">
          <input v-model="forwardauth.enabled" type="checkbox" id="switch-forwardauth" :checked="forwardauth.enabled" />
          <label for="switch-forwardauth" class="mb-0"> </label>
        </span>
      </div>
      <div class="card-body" :class="{ 'd-none': !expanded.forwardauth }">
        <div class="alert alert-info mb-3">
          <strong>Forward Auth</strong> enables authentication via reverse proxy headers. Your proxy (Traefik, NGINX, Caddy)
          authenticates users through Authelia, Authentik, Keycloak, or similar, then passes identity headers to Statping.
          <br /><br />
          <strong>Supported providers:</strong> Authelia, Authentik, Keycloak Gatekeeper, OAuth2-proxy, Pomerium, or any
          proxy that sets <code>Remote-User</code> / <code>X-Auth-*</code> headers.
        </div>

        <div class="form-group row">
          <label for="fa_header_user" class="col-sm-4 col-form-label">User Header</label>
          <div class="col-sm-8">
            <input v-model="forwardauth.header_user" type="text" class="form-control" id="fa_header_user" placeholder="Remote-User" />
            <small class="text-muted">Header containing the username</small>
          </div>
        </div>
        <div class="form-group row">
          <label for="fa_header_email" class="col-sm-4 col-form-label">Email Header</label>
          <div class="col-sm-8">
            <input v-model="forwardauth.header_email" type="text" class="form-control" id="fa_header_email" placeholder="Remote-Email" />
            <small class="text-muted">Header containing the user's email</small>
          </div>
        </div>
        <div class="form-group row">
          <label for="fa_header_groups" class="col-sm-4 col-form-label">Groups Header</label>
          <div class="col-sm-8">
            <input v-model="forwardauth.header_groups" type="text" class="form-control" id="fa_header_groups" placeholder="Remote-Groups" />
            <small class="text-muted">Header containing comma-separated group names</small>
          </div>
        </div>
        <div class="form-group row">
          <label for="fa_header_name" class="col-sm-4 col-form-label">Display Name Header</label>
          <div class="col-sm-8">
            <input v-model="forwardauth.header_name" type="text" class="form-control" id="fa_header_name" placeholder="Remote-Name" />
            <small class="text-muted">Header containing the user's display name</small>
          </div>
        </div>

        <hr />

        <div class="form-group row">
          <label for="fa_admin_groups" class="col-sm-4 col-form-label">Admin Groups</label>
          <div class="col-sm-8">
            <input v-model="forwardauth.admin_groups" type="text" class="form-control" id="fa_admin_groups" placeholder="admins;statping-admins" />
            <small class="text-muted">Semicolon-separated group names that grant admin access</small>
          </div>
        </div>
        <div class="form-group row">
          <label for="fa_trusted_proxies" class="col-sm-4 col-form-label">Trusted Proxies</label>
          <div class="col-sm-8">
            <input v-model="forwardauth.trusted_proxies" type="text" class="form-control" id="fa_trusted_proxies" placeholder="10.0.0.0/8;172.16.0.0/12;192.168.0.0/16" required />
            <small class="text-muted"><strong>Required.</strong> CIDR ranges (semicolon-separated) from which to accept auth headers. Headers from other IPs are ignored.</small>
          </div>
        </div>
        <div class="form-group row">
          <label for="fa_logout_url" class="col-sm-4 col-form-label">Logout URL</label>
          <div class="col-sm-8">
            <input v-model="forwardauth.logout_url" type="text" class="form-control" id="fa_logout_url" placeholder="https://auth.example.com/logout" />
            <small class="text-muted">Optional URL to redirect users on logout</small>
          </div>
        </div>

        <div v-if="forwardAuthError" class="alert alert-danger mt-3">
          {{ forwardAuthError }}
        </div>

        <button class="btn btn-secondary" @click.prevent="saveForwardAuth" :disabled="forwardAuthLoading">
          <font-awesome-icon v-if="forwardAuthLoading" icon="circle-notch" class="mr-2" spin /> Save Forward Auth Settings
        </button>
      </div>
    </div>

    <button class="btn btn-primary btn-block" @click.prevent="saveOAuth" type="submit" :disabled="loading">
      <font-awesome-icon v-if="loading" icon="circle-notch" class="mr-2" spin /> Save OAuth Settings
    </button>
  </form>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from "vue";
import Api from "@/API";
import { useMainStore } from "@/stores/main";

const store = useMainStore();

const coreData = computed(() => store.core);

const google_enabled = ref(false);
const slack_enabled = ref(false);
const github_enabled = ref(false);
const local_enabled = ref(false);
const oidc_enabled = ref(false);
const loading = ref(false);
const forwardAuthLoading = ref(false);

const expanded = reactive({
	github: false,
	google: false,
	slack: false,
	oidc: false,
	forwardauth: false,
});

const forwardauth = reactive({
	enabled: false,
	header_user: "Remote-User",
	header_email: "Remote-Email",
	header_groups: "Remote-Groups",
	header_name: "Remote-Name",
	admin_groups: "",
	trusted_proxies: "",
	logout_url: "",
});

const oauth = reactive({
	gh_client_id: "",
	gh_client_secret: "",
	gh_users: "",
	gh_orgs: "",
	google_client_id: "",
	google_client_secret: "",
	google_users: "",
	oauth_providers: "",
	slack_client_id: "",
	slack_client_secret: "",
	slack_team: "",
	slack_users: "",
	// OIDC fields
	oidc_enabled: false,
	oidc_name: "",
	oidc_issuer_url: "",
	oidc_client_id: "",
	oidc_client_secret: "",
	oidc_scopes: "openid,profile,email",
	oidc_allowed_users: "",
	oidc_admin_groups: "",
	oidc_claim_username: "preferred_username",
	oidc_claim_email: "email",
	oidc_claim_groups: "groups",
	oidc_use_pkce: true,
	oidc_skip_discovery: false,
	oidc_manual_auth_endpoint: "",
	oidc_manual_token_endpoint: "",
	oidc_manual_userinfo_endpoint: "",
	oidc_manual_jwks_url: "",
});

onMounted(async () => {
	const data = await Api.oauth();
	Object.assign(oauth, data);
	local_enabled.value = has("local");
	github_enabled.value = has("github");
	google_enabled.value = has("google");
	slack_enabled.value = has("slack");
	oidc_enabled.value = has("oidc") || oauth.oidc_enabled;

	// Load forward auth settings
	try {
		const faData = await Api.forwardauth();
		forwardauth.enabled = faData.forward_auth_enabled || false;
		forwardauth.header_user = faData.forward_auth_header_user || "Remote-User";
		forwardauth.header_email =
			faData.forward_auth_header_email || "Remote-Email";
		forwardauth.header_groups =
			faData.forward_auth_header_groups || "Remote-Groups";
		forwardauth.header_name = faData.forward_auth_header_name || "Remote-Name";
		forwardauth.admin_groups = faData.forward_auth_admin_groups || "";
		forwardauth.trusted_proxies = faData.forward_auth_trusted_proxies || "";
		forwardauth.logout_url = faData.forward_auth_logout_url || "";
	} catch (e) {
		console.log("Forward auth settings not available");
	}
});

function providers() {
	const list = [];
	if (github_enabled.value) list.push("github");
	if (local_enabled.value) list.push("local");
	if (google_enabled.value) list.push("google");
	if (slack_enabled.value) list.push("slack");
	if (oidc_enabled.value) list.push("oidc");
	return list.join(",");
}

function has(val) {
	if (!oauth.oauth_providers) return false;
	return oauth.oauth_providers.split(",").includes(val);
}

async function copyCallback(provider) {
	const url = `${coreData.value.domain}/oauth/${provider}`;
	await navigator.clipboard.writeText(url);
}

async function saveOAuth() {
	loading.value = true;
	oauth.oauth_providers = providers();
	oauth.oidc_enabled = oidc_enabled.value;
	await Api.oauth_save(oauth);
	const data = await Api.oauth();
	store.setOAuth(data);
	loading.value = false;
}

const forwardAuthError = ref("");

function validateForwardAuth() {
	forwardAuthError.value = "";

	if (forwardauth.enabled) {
		// Trusted proxies required when enabled
		if (!forwardauth.trusted_proxies.trim()) {
			forwardAuthError.value =
				"Trusted proxies are required when Forward Auth is enabled";
			return false;
		}

		// Validate CIDR format
		const cidrs = forwardauth.trusted_proxies
			.split(";")
			.map((s) => s.trim())
			.filter((s) => s);
		const cidrRegex =
			/^(\d{1,3}\.){3}\d{1,3}(\/\d{1,2})?$|^[0-9a-fA-F:]+(:\/\d{1,3})?$/;
		for (const cidr of cidrs) {
			if (!cidrRegex.test(cidr)) {
				forwardAuthError.value = `Invalid CIDR format: ${cidr}`;
				return false;
			}
		}
	}

	// Validate logout URL format if provided
	if (forwardauth.logout_url.trim()) {
		if (
			!forwardauth.logout_url.startsWith("http://") &&
			!forwardauth.logout_url.startsWith("https://")
		) {
			forwardAuthError.value = "Logout URL must start with http:// or https://";
			return false;
		}
	}

	// Validate field lengths
	if (forwardauth.trusted_proxies.length > 4096) {
		forwardAuthError.value = "Trusted proxies too long (max 4096 characters)";
		return false;
	}
	if (forwardauth.admin_groups.length > 1024) {
		forwardAuthError.value = "Admin groups too long (max 1024 characters)";
		return false;
	}
	if (forwardauth.logout_url.length > 2048) {
		forwardAuthError.value = "Logout URL too long (max 2048 characters)";
		return false;
	}

	return true;
}

async function saveForwardAuth() {
	if (!validateForwardAuth()) {
		return;
	}

	forwardAuthLoading.value = true;
	forwardAuthError.value = "";
	try {
		await Api.forwardauth_save({
			forward_auth_enabled: forwardauth.enabled,
			forward_auth_header_user: forwardauth.header_user,
			forward_auth_header_email: forwardauth.header_email,
			forward_auth_header_groups: forwardauth.header_groups,
			forward_auth_header_name: forwardauth.header_name,
			forward_auth_admin_groups: forwardauth.admin_groups,
			forward_auth_trusted_proxies: forwardauth.trusted_proxies,
			forward_auth_logout_url: forwardauth.logout_url,
		});
	} catch (e) {
		console.error("Failed to save forward auth settings:", e);
		forwardAuthError.value =
			e.response?.data?.error || "Failed to save settings";
	}
	forwardAuthLoading.value = false;
}
</script>

<style scoped></style>
