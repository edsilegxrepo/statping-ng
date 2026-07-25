<template>
    <div>
    <form @submit.prevent="login" autocomplete="on">
        <div class="form-group row">
            <label for="username" class="col-4 col-form-label">{{$t('username')}}</label>
            <div class="col-8">
                <input @keyup="checkForm" @change="checkForm" type="text" v-model="username" autocomplete="username" name="username" class="form-control" id="username" placeholder="admin" autocorrect="off" autocapitalize="none">
            </div>
        </div>
        <div class="form-group row">
            <label for="password" class="col-4 col-form-label">{{$t('password')}}</label>
            <div class="col-8">
                <input @keyup="checkForm" @change="checkForm" type="password" v-model="password" autocomplete="current-password" name="password" class="form-control" id="password" placeholder="************">
            </div>
        </div>
        <div class="form-group row">
            <div class="col-sm-12">
                <div v-if="error" class="alert alert-danger" role="alert">
                    {{$t('wrong_login')}}
                </div>
                <button @click.prevent="login" type="submit" class="btn btn-block btn-primary" :disabled="disabled || loading">
                    <font-awesome-icon v-if="loading" icon="circle-notch" class="mr-2" spin/>{{loading ? $t('loading') : $t('sign_in')}}
                </button>
            </div>
        </div>
    </form>

        <a v-if="oauth && oauth.gh_client_id" @click.prevent="GHlogin" href="#" class="mt-4 btn btn-block btn-outline-dark">
            <font-awesome-icon :icon="['fab', 'github']" /> Login with Github
        </a>

        <a v-if="oauth && oauth.slack_client_id" @click.prevent="Slacklogin" href="#" class="btn btn-block btn-outline-dark">
            <font-awesome-icon :icon="['fab', 'slack']" /> Login with Slack
        </a>

        <a v-if="oauth && oauth.google_client_id" @click.prevent="Googlelogin" href="#" class="btn btn-block btn-outline-dark">
            <font-awesome-icon :icon="['fab', 'google']" /> Login with Google
        </a>

        <a v-if="oauth && oauth.custom_client_id" @click.prevent="Customlogin" href="#" class="btn btn-block btn-outline-dark">
            <font-awesome-icon :icon="['fas', 'address-card']" /> Login with {{oauth.custom_name}}
        </a>

    </div>
</template>

<script>
import Api from "../API";

export default {
	name: "FormLogin",
	computed: {
		core() {
			return this.$store.getters.core;
		},
		oauth() {
			return this.$store.getters.oauth;
		},
	},
	data() {
		return {
			username: "",
			password: "",
			auth: {},
			loading: false,
			error: false,
			disabled: true,
			google_scope:
				"https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fuserinfo.profile+https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fuserinfo.email",
			slack_scope: "identity.email,identity.basic",
		};
	},
	mounted() {
		this.$cookies.remove("statping_auth");
	},
	methods: {
		checkForm() {
			if (!this.username || !this.password) {
				this.disabled = true;
			} else {
				this.disabled = false;
			}
		},
		async login() {
			this.loading = true;
			this.error = false;
			const auth = await Api.login(this.username, this.password);
			if (auth.error) {
				this.error = true;
			} else if (auth.token) {
				// Note: JWT token is set as HttpOnly cookie by the server
				// We only use the token info for UI state
				await this.$store.dispatch("loadAdmin");
				this.$store.commit("setAdmin", auth.admin);
				this.$store.commit("setLoggedIn", true);
				this.$router.push("/dashboard");
			}
			this.loading = false;
		},
		encode(val) {
			return encodeURI(val);
		},
		custom_scopes() {
			let scopes = [];
			if (this.oauth.custom_open_id) {
				scopes.push("openid");
			}
			scopes.push(this.oauth.custom_scopes.split(","));
			if (scopes.length !== 0) {
				return `&scope=${scopes.join(" ")}`;
			}
			return "";
		},
		async getOAuthState() {
			// Fetch a state token from the server for CSRF protection
			try {
				const response = await fetch("api/oauth/state");
				const data = await response.json();
				return data.state;
			} catch (e) {
				console.error("Failed to get OAuth state:", e);
				return null;
			}
		},
		async GHlogin() {
			const state = await this.getOAuthState();
			if (!state) return;
			window.location = `https://github.com/login/oauth/authorize?client_id=${this.oauth.gh_client_id}&redirect_uri=${this.encode(`${this.core.domain}/oauth/github`)}&scope=read:user,read:org&state=${state}`;
		},
		async Slacklogin() {
			const state = await this.getOAuthState();
			if (!state) return;
			window.location = `https://slack.com/oauth/authorize?client_id=${this.oauth.slack_client_id}&redirect_uri=${this.encode(`${this.core.domain}/oauth/slack`)}&scope=identity.basic&state=${state}`;
		},
		async Googlelogin() {
			const state = await this.getOAuthState();
			if (!state) return;
			window.location = `https://accounts.google.com/signin/oauth?client_id=${this.oauth.google_client_id}&redirect_uri=${this.encode(`${this.core.domain}/oauth/google`)}&response_type=code&scope=https://www.googleapis.com/auth/userinfo.profile+https://www.googleapis.com/auth/userinfo.email&state=${state}`;
		},
		async Customlogin() {
			const state = await this.getOAuthState();
			if (!state) return;
			window.location = `${this.oauth.custom_endpoint_auth}?client_id=${this.oauth.custom_client_id}&redirect_uri=${this.encode(`${this.core.domain}/oauth/custom`)}&response_type=code${this.custom_scopes()}&state=${state}`;
		},
	},
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
</style>
