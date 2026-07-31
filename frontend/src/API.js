import axios from 'axios'

axios.defaults.withCredentials = true

const tokenKey = 'statping_auth'

// Add response interceptor for better error handling
axios.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response) {
      console.error('API Error:', error.response.status, error.response.data)
    }
    return Promise.reject(error)
  }
)

class Api {
  constructor() {
    this.version = '0.96.0'
  }

  async oauth() {
    return axios.get('api/oauth').then((response) => response.data)
  }

  async core() {
    const core = axios.get('api').then((response) => response.data)
    return core
  }

  async core_save(obj) {
    return axios.post('api/core', obj).then((response) => response.data)
  }

  async oauth_save(obj) {
    return axios.post('api/oauth', obj).then((response) => response.data)
  }

  async setup_save(data) {
    const params = new URLSearchParams(data)
    return axios.post('api/setup', params).then((response) => response.data)
  }

  async services() {
    return axios.get('api/services').then((response) => response.data)
  }

  async service(id) {
    return axios.get(`api/services/${id}`).then((response) => response.data)
  }

  async service_create(data) {
    return axios.post('api/services', data).then((response) => response.data)
  }

  async service_update(data) {
    return axios.post(`api/services/${data.id}`, data).then((response) => response.data)
  }

  async service_hits(id, start, end, group, fill = true) {
    return axios
      .get(`api/services/${id}/hits_data?start=${start}&end=${end}&group=${group}&fill=${fill}`)
      .then((response) => response.data)
  }

  async service_ping(id, start, end, group, fill = true) {
    return axios
      .get(`api/services/${id}/ping_data?start=${start}&end=${end}&group=${group}&fill=${fill}`)
      .then((response) => response.data)
  }

  async service_failures(id, start, end, limit = 999, offset = 0) {
    return axios
      .get(`api/services/${id}/failures?start=${start}&end=${end}&limit=${limit}&offset=${offset}`)
      .then((response) => response.data)
  }

  async service_failures_count(id, start, end) {
    return axios.get(`api/services/${id}/failures_count?start=${start}&end=${end}`).then((response) => response.data)
  }

  async service_failures_data(id, start, end, group, fill = true) {
    return axios
      .get(`api/services/${id}/failure_data?start=${start}&end=${end}&group=${group}&fill=${fill}`)
      .then((response) => response.data)
  }

  async service_failures_delete(service_id) {
    return axios.delete(`api/services/${service_id}/failures`).then((response) => response.data)
  }

  async service_delete(id) {
    return axios.delete(`api/services/${id}`).then((response) => response.data)
  }

  async services_reorder(data) {
    return axios.post('api/reorder/services', data).then((response) => response.data)
  }

  async groups() {
    return axios.get('api/groups').then((response) => response.data)
  }

  async group(id) {
    return axios.get(`api/groups/${id}`).then((response) => response.data)
  }

  async group_delete(id) {
    return axios.delete(`api/groups/${id}`).then((response) => response.data)
  }

  async group_create(data) {
    return axios.post('api/groups', data).then((response) => response.data)
  }

  async group_update(data) {
    return axios.post(`api/groups/${data.id}`, data).then((response) => response.data)
  }

  async groups_reorder(data) {
    return axios.post('api/reorder/groups', data).then((response) => response.data)
  }

  async users() {
    return axios.get('api/users').then((response) => response.data)
  }

  async user(id) {
    return axios.get(`api/users/${id}`).then((response) => response.data)
  }

  async user_create(data) {
    return axios.post('api/users', data).then((response) => response.data)
  }

  async user_update(data) {
    return axios.post(`api/users/${data.id}`, data).then((response) => response.data)
  }

  async user_delete(id) {
    return axios.delete(`api/users/${id}`).then((response) => response.data)
  }

  async messages() {
    return axios.get('api/messages').then((response) => response.data)
  }

  async message(id) {
    return axios.get(`api/messages/${id}`).then((response) => response.data)
  }

  async message_create(data) {
    return axios.post('api/messages', data).then((response) => response.data)
  }

  async message_update(data) {
    return axios.post(`api/messages/${data.id}`, data).then((response) => response.data)
  }

  async message_delete(id) {
    return axios.delete(`api/messages/${id}`).then((response) => response.data)
  }

  async notifiers() {
    return axios.get('api/notifiers').then((response) => response.data)
  }

  async notifier(method) {
    return axios.get(`api/notifier/${method}`).then((response) => response.data)
  }

  async notifier_save(data) {
    return axios.post(`api/notifier/${data.method}`, data).then((response) => response.data)
  }

  async notifier_test(method, data) {
    return axios.post(`api/notifier/${method}/test`, data).then((response) => response.data)
  }

  async notifier_logs(method, start, end) {
    return axios.get(`api/notifier/${method}/logs?start=${start}&end=${end}`).then((response) => response.data)
  }

  async incident_update(data) {
    return axios.post(`api/incidents/${data.id}`, data).then((response) => response.data)
  }

  async incident_create(data) {
    return axios.post('api/incidents', data).then((response) => response.data)
  }

  async incident_create_update(incident_id, data) {
    return axios.post(`api/incidents/${incident_id}/updates`, data).then((response) => response.data)
  }

  async incident_delete(id) {
    return axios.delete(`api/incidents/${id}`).then((response) => response.data)
  }

  async incidents(service_id) {
    return axios.get(`api/services/${service_id}/incidents`).then((response) => response.data)
  }

  async incidents_service(service_id) {
    return axios.get(`api/services/${service_id}/incidents`).then((response) => response.data)
  }

  async all_incidents() {
    return axios.get('api/incidents').then((response) => response.data)
  }

  async checkin_create(data) {
    return axios.post(`api/services/${data.service_id}/checkins`, data).then((response) => response.data)
  }

  async checkin_update(data) {
    return axios.post(`api/checkins/${data.id}`, data).then((response) => response.data)
  }

  async checkin_delete(id) {
    return axios.delete(`api/checkins/${id}`).then((response) => response.data)
  }

  async checkins() {
    return axios.get('api/checkins').then((response) => response.data)
  }

  async token() {
    return axios.get('api/users/token').then((response) => response.data)
  }

  async check_token() {
    return axios.get('api/users/token').then((response) => response.data)
  }

  async logs() {
    return axios.get('api/logs').then((response) => response.data)
  }

  async logs_last() {
    return axios.get('api/logs/last').then((response) => response.data)
  }

  async renew() {
    return axios.get('api/renew').then((response) => response.data)
  }

  async configs() {
    return axios.get('api/settings/configs').then((response) => response.data)
  }

  async configs_save(data) {
    return axios.post('api/settings/configs', data).then((response) => response.data)
  }

  async cache() {
    return axios.get('api/cache').then((response) => response.data)
  }

  async theme() {
    return axios.get('api/theme').then((response) => response.data)
  }

  async theme_save(data) {
    return axios.post('api/theme', data).then((response) => response.data)
  }

  async ldap() {
    return axios.get('api/ldap').then((response) => response.data)
  }

  async ldap_save(data) {
    return axios.post('api/ldap', data).then((response) => response.data)
  }

  async ldap_test(data) {
    return axios.post('api/ldap/test', data).then((response) => response.data)
  }

  async ldap_templates() {
    return axios.get('api/ldap/templates').then((response) => response.data)
  }

  async digest() {
    return axios.get('api/digest').then((response) => response.data)
  }

  async digest_save(data) {
    return axios.post('api/digest', data).then((response) => response.data)
  }

  async digest_test() {
    return axios.post('api/digest/test').then((response) => response.data)
  }

  async digest_smtp_test() {
    return axios.post('api/digest/smtp-test').then((response) => response.data)
  }

  async settings_export() {
    return axios.get('api/settings/export').then((response) => response.data)
  }

  async settings_import(data) {
    const params = new URLSearchParams(data)
    return axios.post('api/settings/import', params).then((response) => response.data)
  }

  async login(username, password) {
    const params = new URLSearchParams({ username, password })
    return axios.post('api/login', params).then((response) => response.data)
  }

  async logout() {
    return axios.get('api/logout').then((response) => response.data)
  }

}

export default new Api()
