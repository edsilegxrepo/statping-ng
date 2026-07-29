describe('Index Page (Public Status)', () => {
  beforeEach(() => {
    cy.intercept('GET', '**/api/core', {
      statusCode: 200,
      body: {
        setup: true,
        name: 'Test Status Page',
        description: 'Monitoring all services',
        domain: 'http://localhost:8080',
        footer: '',
      },
    }).as('getCore')

    cy.intercept('GET', '**/api/services', {
      statusCode: 200,
      body: [
        {
          id: 1,
          name: 'Google',
          domain: 'https://google.com',
          type: 'http',
          online: true,
          status: 'online',
          latency: 45,
          stats: { uptime: 99.9 },
        },
        {
          id: 2,
          name: 'GitHub',
          domain: 'https://github.com',
          type: 'http',
          online: true,
          status: 'online',
          latency: 120,
          stats: { uptime: 99.5 },
        },
        {
          id: 3,
          name: 'Offline Service',
          domain: 'https://example.com',
          type: 'http',
          online: false,
          status: 'offline',
          latency: 0,
          stats: { uptime: 85.0 },
        },
      ],
    }).as('getServices')

    cy.intercept('GET', '**/api/groups', {
      statusCode: 200,
      body: [],
    }).as('getGroups')

    cy.intercept('GET', '**/api/messages', {
      statusCode: 200,
      body: [],
    }).as('getMessages')
  })

  it('displays the status page title and description', () => {
    cy.visit('/')
    cy.wait('@getCore')
    cy.wait('@getServices')

    cy.contains('Test Status Page').should('be.visible')
    cy.contains('Monitoring all services').should('be.visible')
  })

  it('shows list of services', () => {
    cy.visit('/')
    cy.wait('@getCore')
    cy.wait('@getServices')

    cy.contains('Google').should('be.visible')
    cy.contains('GitHub').should('be.visible')
    cy.contains('Offline Service').should('be.visible')
  })

  it('displays service status indicators', () => {
    cy.visit('/')
    cy.wait('@getCore')
    cy.wait('@getServices')

    cy.get('.bg-success, .badge-success, .text-success').should('have.length.at.least', 2)
    cy.get('.bg-danger, .badge-danger, .text-danger').should('have.length.at.least', 1)
  })

  it('navigates to service details on click', () => {
    cy.intercept('GET', '**/api/services/1', {
      statusCode: 200,
      body: {
        id: 1,
        name: 'Google',
        domain: 'https://google.com',
        type: 'http',
        online: true,
      },
    }).as('getService')

    cy.visit('/')
    cy.wait('@getCore')
    cy.wait('@getServices')

    cy.contains('Google').click()
    cy.url().should('include', '/service/')
  })

  it('displays announcements when present', () => {
    cy.intercept('GET', '**/api/messages', {
      statusCode: 200,
      body: [
        {
          id: 1,
          title: 'Scheduled Maintenance',
          description: 'System will be down for maintenance',
          start_on: new Date().toISOString(),
          end_on: new Date(Date.now() + 86400000).toISOString(),
        },
      ],
    }).as('getMessagesWithAnnouncement')

    cy.visit('/')
    cy.wait('@getCore')
    cy.wait('@getServices')
    cy.wait('@getMessagesWithAnnouncement')

    cy.contains('Scheduled Maintenance').should('be.visible')
  })
})
