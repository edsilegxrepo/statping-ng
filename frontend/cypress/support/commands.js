// ***********************************************
// Custom Cypress commands for Statping E2E tests
// ***********************************************

let LOCAL_STORAGE_MEMORY = {};

// Test credentials - password must be 30+ chars with upper, lower, digit
export const TEST_PASSWORD = "AdminPassword123456789012345678";

Cypress.Commands.add("saveLocalStorageCache", () => {
	Object.keys(localStorage).forEach((key) => {
		LOCAL_STORAGE_MEMORY[key] = localStorage[key];
	});
});

Cypress.Commands.add("restoreLocalStorageCache", () => {
	Object.keys(LOCAL_STORAGE_MEMORY).forEach((key) => {
		localStorage.setItem(key, LOCAL_STORAGE_MEMORY[key]);
	});
});

Cypress.Commands.add("clearLocalStorageCache", () => {
	localStorage.clear();
	LOCAL_STORAGE_MEMORY = {};
});

// Login - fresh login each time (session caching causes app state issues)
Cypress.Commands.add("login", (username = "admin", password = "AdminPassword123456789012345678") => {
	cy.visit("/login");
	cy.get("#username").should("be.visible").and("not.be.disabled");
	cy.get("#username").clear().type(username);
	cy.get("#password").clear().type(password);
	cy.get('button[type="submit"]').click();
	cy.url({ timeout: 15000 }).should("not.include", "/login");
	cy.get(".navbar-brand", { timeout: 10000 }).should("be.visible");
});

// Logout helper command
Cypress.Commands.add("logout", () => {
	cy.clearCookies();
	cy.clearLocalStorageCache();
	Cypress.session.clearAllSavedSessions();
	cy.visit("/login");
});

// Wait for API to be ready
Cypress.Commands.add("waitForApi", (timeout = 30000) => {
	cy.request({
		url: "/api",
		timeout: timeout,
		retryOnStatusCodeFailure: true,
	}).then((response) => {
		expect(response.status).to.eq(200);
	});
});
