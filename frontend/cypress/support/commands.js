const API_URL = Cypress.env('API_URL') || 'http://localhost:8080/api'

Cypress.Commands.add('login', (username = 'admin', password = 'admin') => {
  cy.request({
    method: 'POST',
    url: `${API_URL}/login`,
    body: { username, password },
  }).then((response) => {
    if (response.body && response.body.token) {
      window.localStorage.setItem('statping_auth', response.body.token)
    }
  })
})

Cypress.Commands.add('logout', () => {
  window.localStorage.removeItem('statping_auth')
  cy.visit('/')
})

Cypress.Commands.add('apiRequest', (method, endpoint, body = null) => {
  const token = window.localStorage.getItem('statping_auth')
  const options = {
    method,
    url: `${API_URL}${endpoint}`,
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  }
  if (body) {
    options.body = body
  }
  return cy.request(options)
})

Cypress.Commands.add('createService', (service) => {
  const defaults = {
    name: 'Test Service',
    domain: 'https://example.com',
    type: 'http',
    method: 'GET',
    expected_status: 200,
    check_interval: 30,
    timeout: 30,
    public: true,
    group_id: 0,
    verify_ssl: true,
  }
  return cy.apiRequest('POST', '/services', { ...defaults, ...service })
})

Cypress.Commands.add('deleteService', (serviceId) => {
  return cy.apiRequest('DELETE', `/services/${serviceId}`)
})

Cypress.Commands.add('waitForApp', () => {
  cy.get('#app', { timeout: 10000 }).should('exist')
})
