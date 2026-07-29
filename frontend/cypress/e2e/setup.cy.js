describe('Setup Flow', () => {
  beforeEach(() => {
    cy.clearLocalStorage()
  })

  it('shows setup page when app is not configured', () => {
    cy.intercept('GET', '**/api/core', {
      statusCode: 200,
      body: { setup: false },
    }).as('getCore')

    cy.visit('/')
    cy.wait('@getCore')
    cy.url().should('include', '/setup')
  })

  it('displays setup form with required fields', () => {
    cy.intercept('GET', '**/api/core', {
      statusCode: 200,
      body: { setup: false },
    }).as('getCore')

    cy.visit('/setup')
    cy.wait('@getCore')

    cy.get('input[name="project"]').should('exist')
    cy.get('input[name="description"]').should('exist')
    cy.get('input[name="domain"]').should('exist')
    cy.get('input[name="username"]').should('exist')
    cy.get('input[name="password"]').should('exist')
  })

  it('validates required fields before submission', () => {
    cy.intercept('GET', '**/api/core', {
      statusCode: 200,
      body: { setup: false },
    }).as('getCore')

    cy.visit('/setup')
    cy.wait('@getCore')

    cy.get('input[name="project"]').clear()
    cy.get('button[type="submit"]').click()
    cy.get('input[name="project"]:invalid').should('exist')
  })

  it('completes setup successfully', () => {
    cy.intercept('GET', '**/api/core', {
      statusCode: 200,
      body: { setup: false },
    }).as('getCore')

    cy.intercept('POST', '**/api/setup', {
      statusCode: 200,
      body: { status: 'success' },
    }).as('postSetup')

    cy.visit('/setup')
    cy.wait('@getCore')

    cy.get('input[name="project"]').clear().type('Test Status Page')
    cy.get('input[name="description"]').clear().type('Monitoring all services')
    cy.get('input[name="domain"]').clear().type('http://localhost:8080')
    cy.get('input[name="username"]').clear().type('admin')
    cy.get('input[name="password"]').clear().type('admin123')

    cy.get('button[type="submit"]').click()
    cy.wait('@postSetup')
  })
})
