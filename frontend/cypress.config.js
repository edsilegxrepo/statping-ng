const { defineConfig } = require('cypress')

module.exports = defineConfig({
  projectId: 'wmw54a',
  env: {
    DB_HOST: 'localhost',
    DB_USER: 'root',
    DB_DATABASE: 'root',
    DB_PORT: '5432',
    DB_PASS: 'password123',
    GO_ENV: 'production'
  },
  e2e: {
    baseUrl: 'http://localhost:8080',
    specPattern: 'cypress/integration/**/*.js',
    supportFile: 'cypress/support/index.js',
    chromeWebSecurity: false,
    defaultCommandTimeout: 30000,
    requestTimeout: 30000,
    watchForFileChanges: false,
  },
  video: false,  // Disable video for faster runs
})
