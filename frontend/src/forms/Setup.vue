<template>
    <div class="container col-md-7 col-sm-12 mt-2 sm-container">
        <div class="col-12 mb-5 text-center">
            <h1 class="text-uppercase font-weight-bold text-dark mb-0" style="letter-spacing: 6px; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; font-size: 2.2rem;">Platform Monitoring</h1>
            <hr class="mt-2 mb-0" style="width: 100px; border-top: 3px solid #007bff; margin: 0 auto;">
        </div>

        <div v-if="finished" class="col-12 text-center mt-5">
            <div class="card shadow-sm border-0">
                <div class="card-body p-5">
                    <div class="mb-4">
                        <font-awesome-icon icon="check-circle" class="text-success" size="4x" />
                    </div>
                    <h2 class="font-weight-bold mb-4">Setup Complete!</h2>
                    <p class="text-muted mb-5">Your Enterprise Monitoring Platform is ready. Below are your generated credentials. Please save them now, as they will only be shown once.</p>

                    <div v-if="generatedAdmin" class="bg-light p-4 rounded mb-4 text-left border">
                        <h6 class="text-uppercase text-muted small font-weight-bold mb-3">Administrator Account</h6>
                        <div class="d-flex justify-content-between align-items-center mb-2">
                            <span><strong>Username:</strong> {{setup.username || 'admin'}}</span>
                        </div>
                        <div class="d-flex justify-content-between align-items-center">
                            <span class="text-break"><strong>Password:</strong> <code class="bg-white p-1 rounded border">{{generatedAdmin}}</code></span>
                        </div>
                    </div>

                    <div v-if="generatedSamples && Object.keys(generatedSamples).length > 0" class="bg-light p-4 rounded mb-5 text-left border">
                        <h6 class="text-uppercase text-muted small font-weight-bold mb-3">Sample Users</h6>
                        <div v-for="(pass, user) in generatedSamples" :key="user" class="mb-3 last-child-mb-0">
                            <div class="d-flex justify-content-between align-items-center mb-1">
                                <span><strong>Username:</strong> {{user}}</span>
                            </div>
                            <div class="d-flex justify-content-between align-items-center">
                                <span class="text-break"><strong>Password:</strong> <code class="bg-white p-1 rounded border">{{pass}}</code></span>
                            </div>
                        </div>
                    </div>

                    <button @click="$router.push('/')" class="btn btn-primary btn-lg px-5 shadow-sm">
                        Continue to Dashboard
                    </button>
                </div>
            </div>
        </div>

        <div v-else class="col-12">
            <form @submit.prevent="saveSetup">
                <div class="row">
                    <div class="col-12 col-md-6">
                        <div class="form-group">
                            <label class="text-capitalize">{{ $t('language') }}</label>
                            <select @change="changeLanguages" v-model="setup.language" id="language" class="form-control">
                                <option value="en">English</option>
                                <option value="es">Spanish</option>
                                <option value="fr">French</option>
                                <option value="ru">Russian</option>
                                <option value="de">German</option>
                                <option value="cs">Czech</option>
                                <option value="ja">Japanese</option>
                                <option value="ko">Korean</option>
                                <option value="it">Italian</option>
                                <option value="zh">Chinese</option>
                                <option value="sv">Swedish</option>
                            </select>
                        </div>
                        <div class="form-group">
                            <label class="text-capitalize">{{ $t('db_connection') }}</label>
                            <select @change="canSubmit" v-model="setup.db_connection" id="db_connection" class="form-control">
                                <option value="sqlite">SQLite</option>
                                <option value="postgres">Postgres</option>
                                <option value="mysql">MySQL</option>
                            </select>
                        </div>
                        <div class="row">
                            <div class="col-7 col-md-6">
                                <div v-if="setup.db_connection !== 'sqlite'" class="form-group">
                                    <label class="text-capitalize">{{ $t('db_host') }}</label>
                                    <input @keyup="canSubmit" v-model="setup.db_host" id="db_host" type="text" class="form-control" placeholder="localhost">
                                </div>
                            </div>
                            <div class="col-5 col-md-6">
                                <div v-if="setup.db_connection !== 'sqlite'" class="form-group">
                                    <label class="text-capitalize">{{ $t('db_port') }}</label>
                                    <input @keyup="canSubmit" v-model.number="setup.db_port" id="db_port" type="number" class="form-control" placeholder="5432">
                                </div>
                            </div>
                        </div>
                        <div v-if="setup.db_connection !== 'sqlite'" class="form-group">
                            <label class="text-capitalize">{{ $t('db_username') }}</label>
                            <input @keyup="canSubmit" v-model="setup.db_user" id="db_user" type="text" class="form-control" placeholder="root">
                        </div>
                        <div v-if="setup.db_connection !== 'sqlite'" class="form-group">
                            <label for="db_password" class="text-capitalize">{{ $t('db_password') }}</label>
                            <input @keyup="canSubmit" v-model="setup.db_password" id="db_password" type="password" class="form-control" placeholder="password123">
                        </div>
                        <div v-if="setup.db_connection !== 'sqlite'" class="form-group">
                            <label for="db_database" class="text-capitalize">{{ $t('db_database') }}</label>
                            <input @keyup="canSubmit" v-model="setup.db_database" id="db_database" type="text" class="form-control" placeholder="Database name">
                        </div>

                        <div class="form-group mt-3">
                            <div class="row">
                                <div class="col-9">
                                    <span class="text-left text-capitalize">{{ $t('send_reports') }}</span>
                                </div>
                                <div class="col-3 text-right">
                                    <span @click="setup.send_reports = !!setup.send_reports" class="switch">
                                      <input v-model="setup.send_reports" type="checkbox" name="send_reports" class="switch" id="send_reports" :checked="setup.send_reports">
                                      <label for="send_reports"></label>
                                    </span>
                                </div>
                            </div>
                        </div>

                        <div class="form-group mt-3">
                            <div class="row">
                                <div class="col-9">
                                    <span class="text-left text-capitalize">Sample Data</span>
                                </div>
                                <div class="col-3 text-right">
                                    <span @click="setup.sample_data = !setup.sample_data" class="switch">
                                      <input v-model="setup.sample_data" type="checkbox" name="sample_data" class="switch" id="sample_data" :checked="setup.sample_data">
                                      <label for="sample_data"></label>
                                    </span>
                                </div>
                            </div>
                        </div>

                    </div>

                    <div class="col-12 col-md-6">

                        <div class="form-group">
                            <label class="text-capitalize">{{ $t('project_name') }}</label>
                            <input @keyup="canSubmit" v-model="setup.project" id="project" type="text" class="form-control" placeholder="Work Servers" required>
                        </div>

                        <div class="form-group">
                            <label class="text-capitalize">{{ $t('description') }}</label>
                            <input @keyup="canSubmit" v-model="setup.description" id="description" type="text" class="form-control" placeholder="Monitors all of my work services">
                        </div>

                        <div class="form-group">
                            <label class="text-capitalize" for="domain">{{ $t('domain') }}</label>
                            <input @keyup="canSubmit" v-model="setup.domain" type="text" class="form-control" id="domain" required>
                        </div>

                        <div class="form-group">
                            <label class="text-capitalize">{{ $t('username') }}</label>
                            <input @keyup="canSubmit" v-model="setup.username" id="username" type="text" class="form-control" placeholder="admin" required>
                        </div>

                        <div class="form-group">
                            <label class="text-capitalize">{{ $t('password') }}</label>
                            <input @keyup="canSubmit" v-model="setup.password" id="password" type="password" class="form-control" placeholder="password" required>
                        </div>

                        <div class="form-group">
                            <label class="text-capitalize">{{ $t('confirm_password') }}</label>
                            <input @keyup="canSubmit" v-model="setup.confirm_password" id="password_confirm" type="password" class="form-control" placeholder="password" required>
                            <span v-if="passnomatch" class="small text-danger">Both passwords should match</span>
                            <span v-if="passTooWeak" class="small text-danger d-block">Password must be at least 30 characters and include uppercase, lowercase, and digits</span>
                        </div>

                        <div class="form-group">
                            <label class="text-capitalize">{{ $t('email') }}</label>
                            <input @keyup="canSubmit" v-model="setup.email" id="email" type="text" class="form-control" placeholder="myemail@domain.com" required>
                        </div>
                    </div>

                    <div v-if="error" class="col-12 alert alert-danger">
                        {{error}}
                    </div>

                    <div class="col-12">
                        <button @click.prevent="saveSetup" v-bind:disabled="disabled || loading" type="submit" class="btn btn-primary btn-block" :class="{'btn-primary': !loading, 'btn-default': loading}">
                            <font-awesome-icon v-if="loading" icon="circle-notch" class="mr-2" spin/>{{loading ? $t('loading') : $t('save_settings')}}
                        </button>
                    </div>
                </div>
            </form>

        </div>
    </div>
</template>

<script>
  import Api from "../API";

  export default {
  name: 'Setup',
  data () {
    return {
      error: null,
      loading: false,
      disabled: true,
      passnomatch: false,
      passTooWeak: false,
      finished: false,
      generatedAdmin: "",
      generatedSamples: {},
      setup: {
        language: "en",
        db_connection: "sqlite",
        db_host: "",
        db_port: "",
        db_user: "",
        db_password: "",
        db_database: "",
        project: "",
        description: "",
        domain: "",
        username: "",
        password: "",
        confirm_password: "",
        sample_data: false,
        send_reports: false,
        email: "",
      }
    }
  },
  async created() {
    const core = await Api.core()
    if (core.setup) {
        if (!this.$store.getters.hasPublicData) {
            await this.$store.dispatch('loadRequired')
        }
        this.$router.push('/')
    }
  },
  mounted() {
    this.changeLanguages()
    this.setup.domain = window.location.protocol + "//" + window.location.hostname + (window.location.port ? ":"+window.location.port : "")
  },
  methods: {
    changeLanguages() {
      this.$i18n.locale = this.setup.language
    },
      canSubmit() {
          this.error = null
          const s = this.setup
        if (s.confirm_password.length > 0 && s.confirm_password !== s.password) {
          this.passnomatch = true
        } else {
          this.passnomatch = false
        }

        if (s.password.length > 0) {
            const hasUpper = /[A-Z]/.test(s.password);
            const hasLower = /[a-z]/.test(s.password);
            const hasDigit = /[0-9]/.test(s.password);
            if (s.password.length < 30 || !hasUpper || !hasLower || !hasDigit) {
                this.passTooWeak = true
            } else {
                this.passTooWeak = false
            }
        } else {
            this.passTooWeak = false
        }

          if (s.db_connection !== 'sqlite') {
              if (!s.db_host || !s.db_port || !s.db_user || !s.db_password || !s.db_database) {
                  this.disabled = true
                  return
              }
          }
          if (!s.project || !s.domain || !s.username || !s.password || !s.confirm_password || !s.email || this.passTooWeak) {
              this.disabled = true
              return
          }
          if (s.password !== s.confirm_password) {
              this.disabled = true
              return
          }
          this.disabled = false
      },
    async saveSetup() {
      this.loading = true
      let resp
      try {
        resp = await Api.setup_save(this.setup)
      } catch(e) {
        resp = {status: 'error', error: e.response.data.error}
      }
      if (resp.status === 'error') {
        this.error = resp.error
        this.loading = false
        return
      }

      this.generatedAdmin = resp.admin_password
      this.generatedSamples = resp.sample_passwords
      this.finished = true

      await this.$store.dispatch('loadCore')
      await this.$store.dispatch('loadRequired')

      this.loading = false
      if (!this.generatedAdmin && (!this.generatedSamples || Object.keys(this.generatedSamples).length === 0)) {
          this.$router.push('/')
      }
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
.last-child-mb-0:last-child {
    margin-bottom: 0 !important;
}
</style>
