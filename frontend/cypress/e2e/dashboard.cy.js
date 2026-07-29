describe('Dashboard', () => {
  beforeEach(() => {
    cy.intercept('GET', '**/api/core', {
      statusCode: 200,
      body: {
        setup: true,
        name: 'Test Status Page',
        description: 'Test Description',
        domain: 'http://localhost:8080',
      },
    }).as('getCore')

    cy.intercept('GET', '**/api/services', {
      statusCode: 200,
      body: [
        { id: 1, name: 'Service 1', online: true, type: 'http' },
        { id: 2, name: 'Service 2', online: false, type: 'tcp' },
      ],
    }).as('getServices')

    cy.intercept('GET', '**/api/groups', {
      statusCode: 200,
      body: [],
    }).as('getGroups')

    cy.intercept('GET', '**/api/users', {
      statusCode: 200,
      body: [{ id: 1, username: 'admin', admin: true }],
    }).as('getUsers')

    cy.intercept('GET', '**/api/messages', {
      statusCode: 200,
      body: [],
    }).as('getMessages')

    cy.intercept('GET', '**/api/notifiers', {
      statusCode: 200,
      body: [],
    }).as('getNotifiers')

    cy.intercept('GET', '**/api/checkins', {
      statusCode: 200,
      body: [],
    }).as('getCheckins')

    cy.window().then((win) => {
      win.localStorage.setItem('statping_auth', 'test-jwt-token')
    })
  })

  it('displays dashboard navigation', () => {
    cy.visit('/dashboard')
    cy.wait('@getCore')

    cy.get('nav, .navbar').should('exist')
    cy.contains('Services').should('be.visible')
  })

  it('shows service list in dashboard', () => {
    cy.visit('/dashboard')
    cy.wait('@getCore')
    cy.wait('@getServices')

    cy.contains('Service 1').should('be.visible')
    cy.contains('Service 2').should('be.visible')
  })

  it('navigates to service edit page', () => {
    cy.intercept('GET', '**/api/services/1', {
      statusCode: 200,
      body: { id: 1, name: 'Service 1', online: true, type: 'http', domain: 'https://example.com' },
    }).as('getService')

    cy.visit('/dashboard')
    cy.wait('@getCore')
    cy.wait('@getServices')

    cy.contains('Service 1').click()
    cy.url().should('match', /\/dashboard\/service\/\d+/)
  })

  it('displays users management section', () => {
    cy.visit('/dashboard/users')
    cy.wait('@getCore')
    cy.wait('@getUsers')

    cy.contains('admin').should('be.visible')
  })

  it('allows creating a new service', () => {
    cy.intercept('POST', '**/api/services', {
      statusCode: 200,
      body: { id: 3, name: 'New Service', type: 'http' },
    }).as('createService')

    cy.visit('/dashboard/create_service')
    cy.wait('@getCore')

    cy.get('input[name="name"]').type('New Service')
    cy.get('input[name="domain"]').type('https://newservice.com')

    cy.get('button[type="submit"]').click()
  })

  it('handles logout correctly', () => {
    cy.visit('/dashboard')
    cy.wait('@getCore')

    cy.contains('Logout').click()

    cy.window().then((win) => {
      expect(win.localStorage.getItem('statping_auth')).to.be.null
    })
  })
})
