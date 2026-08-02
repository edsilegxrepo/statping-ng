<template>
  <div class="col-12">
    <div class="row mb-4">
      <div class="col-12">
        <h2 class="font-weight-bold mb-2">Administration Guide</h2>
        <p class="text-muted">Quick reference for managing the monitoring platform</p>
      </div>
    </div>

    <div class="row">
      <!-- Services Management -->
      <div class="col-lg-6 mb-4">
        <div class="card h-100 border-0 shadow-sm">
          <div class="card-header bg-primary text-white">
            <font-awesome-icon icon="server" class="me-2" />
            Services
          </div>
          <div class="card-body">
            <p class="text-muted small">Create and manage monitored endpoints</p>
            <table class="table table-sm small mb-0">
              <tbody>
                <tr>
                  <td class="text-nowrap"><router-link to="/dashboard/services" class="help-link"><font-awesome-icon icon="list" class="me-1" />Services</router-link></td>
                  <td>View all services, status, and quick actions</td>
                </tr>
                <tr>
                  <td class="text-nowrap"><router-link to="/dashboard/create_service" class="help-link"><font-awesome-icon icon="plus" class="me-1" />Create Service</router-link></td>
                  <td>Add HTTP, TCP, UDP, gRPC, or ICMP monitors</td>
                </tr>
                <tr>
                  <td class="text-nowrap"><router-link to="/dashboard" class="help-link"><font-awesome-icon icon="tachometer-alt" class="me-1" />Dashboard</router-link></td>
                  <td>Overview with charts and recent activity</td>
                </tr>
                <tr>
                  <td class="text-nowrap"><router-link to="/dashboard/services" class="help-link"><font-awesome-icon icon="layer-group" class="me-1" />Groups</router-link></td>
                  <td>Organize services by team, environment, or function</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- Per-Service Features -->
      <div class="col-lg-6 mb-4">
        <div class="card h-100 border-0 shadow-sm">
          <div class="card-header bg-danger text-white">
            <font-awesome-icon icon="heartbeat" class="me-2" />
            Per-Service Features
          </div>
          <div class="card-body">
            <p class="text-muted small">Access via <router-link to="/dashboard/services" class="help-link">Services</router-link> page → select a service</p>
            <table class="table table-sm small mb-0">
              <tbody>
                <tr>
                  <td class="text-nowrap"><router-link :to="'/dashboard/service/' + (firstServiceId || 1) + '/failures'" class="help-link"><font-awesome-icon icon="exclamation-circle" class="me-1" />Failures</router-link></td>
                  <td>View failure history, error messages, and timestamps</td>
                </tr>
                <tr>
                  <td class="text-nowrap"><router-link :to="'/dashboard/service/' + (firstServiceId || 1) + '/checkins'" class="help-link"><font-awesome-icon icon="clock" class="me-1" />Checkins</router-link></td>
                  <td>Dead man's switch for cron jobs and batch tasks</td>
                </tr>
                <tr>
                  <td class="text-nowrap"><router-link :to="'/dashboard/service/' + (firstServiceId || 1) + '/incidents'" class="help-link"><font-awesome-icon icon="flag" class="me-1" />Incidents</router-link></td>
                  <td>Track and communicate service-specific issues</td>
                </tr>
                <tr>
                  <td class="text-nowrap"><router-link :to="'/dashboard/edit_service/' + (firstServiceId || 1)" class="help-link"><font-awesome-icon icon="bell" class="me-1" />Notifiers</router-link></td>
                  <td>Enable/disable notification channels per service</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- Users & Authentication -->
      <div class="col-lg-6 mb-4">
        <div class="card h-100 border-0 shadow-sm">
          <div class="card-header bg-success text-white">
            <font-awesome-icon icon="users" class="me-2" />
            Users & Authentication
          </div>
          <div class="card-body">
            <p class="text-muted small">Manage access and authentication methods</p>
            <table class="table table-sm small mb-0">
              <tbody>
                <tr>
                  <td class="text-nowrap"><router-link to="/dashboard/users" class="help-link"><font-awesome-icon icon="users" class="me-1" />Users</router-link></td>
                  <td>Create, edit, delete user accounts</td>
                </tr>
                <tr>
                  <td class="text-nowrap"><router-link to="/dashboard/settings" class="help-link"><font-awesome-icon icon="key" class="me-1" />Authentication</router-link></td>
                  <td>OAuth providers (GitHub, Google, Slack, Custom)</td>
                </tr>
                <tr>
                  <td class="text-nowrap"><router-link to="/dashboard/settings" class="help-link"><font-awesome-icon icon="sitemap" class="me-1" />LDAP</router-link></td>
                  <td>Enterprise directory integration (AD/LDAP)</td>
                </tr>
                <tr>
                  <td class="text-nowrap"><router-link to="/dashboard/settings" class="help-link"><font-awesome-icon icon="shield-alt" class="me-1" />Forward Auth</router-link></td>
                  <td>SSO via Authelia, Authentik, Keycloak</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- Notifications -->
      <div class="col-lg-6 mb-4">
        <div class="card h-100 border-0 shadow-sm">
          <div class="card-header bg-warning text-dark">
            <font-awesome-icon icon="bell" class="me-2" />
            Notifications
          </div>
          <div class="card-body">
            <p class="text-muted small">Alert channels for service failures</p>
            <table class="table table-sm small mb-0">
              <tbody>
                <tr>
                  <td class="text-nowrap"><router-link to="/dashboard/settings" class="help-link"><font-awesome-icon icon="cog" class="me-1" />Notifier Settings</router-link></td>
                  <td>Configure each notification channel</td>
                </tr>
                <tr>
                  <td><strong>Channels:</strong></td>
                  <td>Email, Slack, Discord, Teams, Telegram, Webhooks, Twilio, Pushover, ntfy</td>
                </tr>
                <tr>
                  <td class="text-nowrap"><router-link to="/dashboard/settings" class="help-link"><font-awesome-icon icon="envelope" class="me-1" />Daily Digest</router-link></td>
                  <td>Scheduled summary email report</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- Announcements -->
      <div class="col-lg-6 mb-4">
        <div class="card h-100 border-0 shadow-sm">
          <div class="card-header bg-info text-white">
            <font-awesome-icon icon="bullhorn" class="me-2" />
            Announcements
          </div>
          <div class="card-body">
            <p class="text-muted small">Communicate maintenance and incidents</p>
            <table class="table table-sm small mb-0">
              <tbody>
                <tr>
                  <td class="text-nowrap"><router-link to="/dashboard/messages" class="help-link"><font-awesome-icon icon="comment-alt" class="me-1" />Announcements</router-link></td>
                  <td>Create global or service-specific messages</td>
                </tr>
                <tr>
                  <td class="text-nowrap"><strong>Scheduled</strong></td>
                  <td>Set start/end times for planned maintenance</td>
                </tr>
                <tr>
                  <td class="text-nowrap"><strong>Incidents</strong></td>
                  <td>Per-service incidents (via service edit page)</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- API Reference -->
      <div class="col-lg-6 mb-4">
        <div class="card h-100 border-0 shadow-sm">
          <div class="card-header bg-purple text-white" style="background-color: #6f42c1;">
            <font-awesome-icon icon="code" class="me-2" />
            API Reference
          </div>
          <div class="card-body">
            <p class="text-muted small">REST API for automation and integrations</p>
            <table class="table table-sm small mb-0">
              <tbody>
                <tr>
                  <td class="text-nowrap"><strong>Base URL</strong></td>
                  <td><code>/api</code></td>
                </tr>
                <tr>
                  <td class="text-nowrap"><strong>Auth Header</strong></td>
                  <td><code>Authorization: Bearer &lt;API_KEY&gt;</code></td>
                </tr>
                <tr>
                  <td colspan="2" class="pt-2"><strong>Common Endpoints:</strong></td>
                </tr>
                <tr>
                  <td><code>GET /api/services</code></td>
                  <td>List all services</td>
                </tr>
                <tr>
                  <td><code>GET /api/services/:id</code></td>
                  <td>Get service details</td>
                </tr>
                <tr>
                  <td><code>POST /api/services</code></td>
                  <td>Create new service</td>
                </tr>
                <tr>
                  <td><code>GET /api/core</code></td>
                  <td>System information</td>
                </tr>
              </tbody>
            </table>
            <p class="text-muted small mt-2 mb-0"><strong>API Key:</strong> Found in <router-link to="/dashboard/settings" class="help-link">Settings</router-link> → API Secret</p>
          </div>
        </div>
      </div>

      <!-- Settings Overview -->
      <div class="col-12 mb-4">
        <div class="card border-0 shadow-sm">
          <div class="card-header bg-secondary text-white">
            <font-awesome-icon icon="cog" class="me-2" />
            Settings
          </div>
          <div class="card-body">
            <div class="row">
              <div class="col-md-3">
                <h6 class="font-weight-bold"><router-link to="/dashboard/settings" class="help-link">General</router-link></h6>
                <ul class="small text-muted ps-3">
                  <li>Site name, description, domain</li>
                  <li>Session timeout</li>
                  <li>API secret key</li>
                </ul>
              </div>
              <div class="col-md-3">
                <h6 class="font-weight-bold"><router-link to="/dashboard/settings" class="help-link">Theme</router-link></h6>
                <ul class="small text-muted ps-3">
                  <li>Custom CSS injection</li>
                  <li>Override default styles</li>
                  <li>Brand customization</li>
                </ul>
              </div>
              <div class="col-md-3">
                <h6 class="font-weight-bold"><router-link to="/dashboard/settings" class="help-link">Log Shipping</router-link></h6>
                <ul class="small text-muted ps-3">
                  <li>Loki, Elasticsearch</li>
                  <li>Splunk, Cribl</li>
                  <li>Custom labels</li>
                </ul>
              </div>
              <div class="col-md-3">
                <h6 class="font-weight-bold"><router-link to="/dashboard/import" class="help-link">Import/Export</router-link></h6>
                <ul class="small text-muted ps-3">
                  <li>Backup services</li>
                  <li>Migrate config</li>
                  <li>Bulk import</li>
                </ul>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Tools & Logs -->
      <div class="col-12 mb-4">
        <div class="card border-0 shadow-sm">
          <div class="card-header bg-dark text-white">
            <font-awesome-icon icon="tools" class="me-2" />
            Tools & Diagnostics
          </div>
          <div class="card-body">
            <div class="row">
              <div class="col-md-4">
                <router-link to="/dashboard/logs" class="help-link d-block mb-2">
                  <font-awesome-icon icon="file-alt" class="me-2" />Application Logs
                </router-link>
                <p class="small text-muted mb-0">View real-time application logs for troubleshooting</p>
              </div>
              <div class="col-md-4">
                <router-link to="/dashboard/import" class="help-link d-block mb-2">
                  <font-awesome-icon icon="file-import" class="me-2" />Import Services
                </router-link>
                <p class="small text-muted mb-0">Bulk import services from YAML/JSON configuration</p>
              </div>
              <div class="col-md-4">
                <router-link to="/" class="help-link d-block mb-2">
                  <font-awesome-icon icon="globe" class="me-2" />Public Status Page
                </router-link>
                <p class="small text-muted mb-0">View the public-facing status page</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Quick Tips -->
      <div class="col-12 mb-4">
        <div class="card border-0 shadow-sm bg-light">
          <div class="card-body">
            <h5 class="font-weight-bold mb-3">
              <font-awesome-icon icon="lightbulb" class="text-warning me-2" />
              Quick Tips
            </h5>
            <div class="row small">
              <div class="col-md-6">
                <ul class="text-muted ps-3 mb-0">
                  <li class="mb-2"><strong>Test notifiers</strong> before relying on them - each has a "Test" button</li>
                  <li class="mb-2"><strong>Use groups</strong> to organize services by team, environment, or function</li>
                  <li class="mb-2"><strong>Check <router-link to="/dashboard/logs">Logs</router-link></strong> when troubleshooting issues</li>
                </ul>
              </div>
              <div class="col-md-6">
                <ul class="text-muted ps-3 mb-0">
                  <li class="mb-2"><strong><router-link to="/dashboard/import">Import/Export</router-link></strong> services for backup and migration</li>
                  <li class="mb-2"><strong>API access</strong> - all actions available via REST API with Bearer token</li>
                  <li class="mb-2"><strong>Environment badge</strong> - set STATPING_ENV=PROD|QA|DEV to display</li>
                </ul>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { computed } from 'vue'
import { useMainStore } from '@/stores/main'

export default {
  name: "AdminHelp",
  setup() {
    const store = useMainStore()
    const firstServiceId = computed(() => {
      const services = store.services || []
      return services.length > 0 ? services[0].id : 1
    })
    return { firstServiceId }
  }
};
</script>

<style scoped>
.card-header {
  font-weight: 600;
  font-size: 0.95rem;
}
.table td {
  vertical-align: middle;
  border-top: 1px solid #f0f0f0;
}
.table tr:first-child td {
  border-top: none;
}
.help-link {
  color: #007bff;
  text-decoration: none;
  font-weight: 600;
}
.help-link:hover {
  color: #0056b3;
  text-decoration: underline;
}
</style>
