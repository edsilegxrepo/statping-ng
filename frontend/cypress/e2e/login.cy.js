describe('Login Flow', () => {
  beforeEach(() => {
    cy.clearLocalStorage()
    cy.intercept('GET', '**/api/core', {
      statusCode: 200,
      body: {
        setup: true,
        name: 'Test Status Page',
        description: 'Test Description',
        domain: 'http://localhost:8080',
      },
    }).as('getCore')
  })

  it('displays login page', () => {
    cy.visit('/login')
    cy.wait('@getCore')

    cy.get('input[name="username"]').should('exist')
    cy.get('input[name="password"]').should('exist')
    cy.get('button[type="submit"]').should('exist')
  })

  it('shows error on invalid credentials', () => {
    cy.intercept('POST', '**/api/login', {
      statusCode: 401,
      body: { error: 'Invalid credentials' },
    }).as('loginFailed')

    cy.visit('/login')
    cy.wait('@getCore')

    cy.get('input[name="username"]').type('wronguser')
    cy.get('input[name="password"]').type('wrongpass')
    cy.get('button[type="submit"]').click()

    cy.wait('@loginFailed')
    cy.get('.alert-danger, .text-danger').should('be.visible')
  })

  it('redirects to dashboard on successful login', () => {
    cy.intercept('POST', '**/api/login', {
      statusCode: 200,
      body: { token: 'test-jwt-token', admin: true },
    }).as('loginSuccess')

    cy.intercept('GET', '**/api/services', {
      statusCode: 200,
      body: [],
    }).as('getServices')

    cy.visit('/login')
    cy.wait('@getCore')

    cy.get('input[name="username"]').type('admin')
    cy.get('input[name="password"]').type('admin')
    cy.get('button[type="submit"]').click()

    cy.wait('@loginSuccess')
    cy.url().should('include', '/dashboard')
  })

  it('persists authentication in localStorage', () => {
    cy.intercept('POST', '**/api/login', {
      statusCode: 200,
      body: { token: 'test-jwt-token', admin: true },
    }).as('loginSuccess')

    cy.visit('/login')
    cy.wait('@getCore')

    cy.get('input[name="username"]').type('admin')
    cy.get('input[name="password"]').type('admin')
    cy.get('button[type="submit"]').click()

    cy.wait('@loginSuccess')
    cy.window().then((win) => {
      expect(win.localStorage.getItem('statping_auth')).to.equal('test-jwt-token')
    })
  })
})
