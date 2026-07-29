import { defineConfig } from 'cypress'

export default defineConfig({
  projectId: 'wmw54a',
  env: {
    API_URL: 'http://localhost:8080/api',
  },
  e2e: {
    baseUrl: 'http://localhost:8888',
    specPattern: 'cypress/e2e/**/*.cy.js',
    supportFile: 'cypress/support/e2e.js',
    chromeWebSecurity: false,
    defaultCommandTimeout: 30000,
    requestTimeout: 30000,
    watchForFileChanges: false,
    viewportWidth: 1280,
    viewportHeight: 720,
  },
  video: false,
})
