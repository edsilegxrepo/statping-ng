describe('Service Management', () => {
  beforeEach(() => {
    cy.intercept('GET', '**/api/core', {
      statusCode: 200,
      body: {
        setup: true,
        name: 'Test Status Page',
        domain: 'http://localhost:8080',
      },
    }).as('getCore')

    cy.intercept('GET', '**/api/groups', {
      statusCode: 200,
      body: [{ id: 1, name: 'Web Services' }],
    }).as('getGroups')

    cy.intercept('GET', '**/api/services', {
      statusCode: 200,
      body: [
        {
          id: 1,
          name: 'Google',
          domain: 'https://google.com',
          type: 'http',
          online: true,
          check_interval: 60,
          timeout: 30,
        },
      ],
    }).as('getServices')

    cy.intercept('GET', '**/api/checkins', {
      statusCode: 200,
      body: [],
    }).as('getCheckins')

    cy.window().then((win) => {
      win.localStorage.setItem('statping_auth', 'test-jwt-token')
    })
  })

  describe('Service Creation', () => {
    it('creates an HTTP service', () => {
      cy.intercept('POST', '**/api/services', {
        statusCode: 200,
        body: { id: 2, name: 'New HTTP Service', type: 'http' },
      }).as('createService')

      cy.visit('/dashboard/create_service')
      cy.wait('@getCore')

      cy.get('input[name="name"]').type('New HTTP Service')
      cy.get('input[name="domain"]').type('https://httpbin.org/get')
      cy.get('select[name="type"]').select('http')
      cy.get('input[name="check_interval"]').clear().type('60')
      cy.get('input[name="timeout"]').clear().type('30')

      cy.get('button[type="submit"]').click()
      cy.wait('@createService')
    })

    it('creates a TCP service', () => {
      cy.intercept('POST', '**/api/services', {
        statusCode: 200,
        body: { id: 3, name: 'DNS Check', type: 'tcp' },
      }).as('createService')

      cy.visit('/dashboard/create_service')
      cy.wait('@getCore')

      cy.get('input[name="name"]').type('DNS Check')
      cy.get('input[name="domain"]').type('8.8.8.8')
      cy.get('select[name="type"]').select('tcp')
      cy.get('input[name="port"]').type('53')

      cy.get('button[type="submit"]').click()
      cy.wait('@createService')
    })

    it('validates required fields', () => {
      cy.visit('/dashboard/create_service')
      cy.wait('@getCore')

      cy.get('button[type="submit"]').click()
      cy.get('input[name="name"]:invalid').should('exist')
    })
  })

  describe('Service Editing', () => {
    beforeEach(() => {
      cy.intercept('GET', '**/api/services/1', {
        statusCode: 200,
        body: {
          id: 1,
          name: 'Google',
          domain: 'https://google.com',
          type: 'http',
          method: 'GET',
          expected_status: 200,
          check_interval: 60,
          timeout: 30,
          public: true,
          verify_ssl: true,
        },
      }).as('getService')
    })

    it('loads existing service data', () => {
      cy.visit('/dashboard/service/1/edit')
      cy.wait('@getCore')
      cy.wait('@getService')

      cy.get('input[name="name"]').should('have.value', 'Google')
      cy.get('input[name="domain"]').should('have.value', 'https://google.com')
    })

    it('updates service successfully', () => {
      cy.intercept('PUT', '**/api/services/1', {
        statusCode: 200,
        body: { id: 1, name: 'Google Updated' },
      }).as('updateService')

      cy.visit('/dashboard/service/1/edit')
      cy.wait('@getCore')
      cy.wait('@getService')

      cy.get('input[name="name"]').clear().type('Google Updated')
      cy.get('button[type="submit"]').click()

      cy.wait('@updateService')
    })
  })

  describe('Service Deletion', () => {
    it('deletes a service with confirmation', () => {
      cy.intercept('DELETE', '**/api/services/1', {
        statusCode: 200,
        body: { status: 'success' },
      }).as('deleteService')

      cy.intercept('GET', '**/api/services/1', {
        statusCode: 200,
        body: { id: 1, name: 'Google', type: 'http' },
      }).as('getService')

      cy.visit('/dashboard/service/1')
      cy.wait('@getCore')
      cy.wait('@getService')

      cy.contains('Delete').click()
      cy.get('.modal').should('be.visible')
      cy.get('.modal').contains('Delete').click()

      cy.wait('@deleteService')
    })
  })

  describe('Service Details', () => {
    beforeEach(() => {
      cy.intercept('GET', '**/api/services/1', {
        statusCode: 200,
        body: {
          id: 1,
          name: 'Google',
          domain: 'https://google.com',
          type: 'http',
          online: true,
          latency: 45,
          stats: { uptime: 99.9, failures: 2 },
        },
      }).as('getService')

      cy.intercept('GET', '**/api/services/1/hits*', {
        statusCode: 200,
        body: [],
      }).as('getHits')

      cy.intercept('GET', '**/api/services/1/failures*', {
        statusCode: 200,
        body: [],
      }).as('getFailures')
    })

    it('displays service statistics', () => {
      cy.visit('/service/1')
      cy.wait('@getCore')
      cy.wait('@getService')

      cy.contains('Google').should('be.visible')
      cy.contains('99.9').should('be.visible')
    })

    it('shows latency chart', () => {
      cy.visit('/service/1')
      cy.wait('@getCore')
      cy.wait('@getService')

      cy.get('.apexcharts-canvas, [class*="chart"]').should('exist')
    })
  })
})
