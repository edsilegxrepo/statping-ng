/// <reference types="cypress" />

context("Setup Process", () => {
	it("should be not be setup yet", () => {
		cy.request(`/api`).then((response) => {
			expect(response.body).to.have.property("setup", false);
		});
	});

	it("should setup Statping with SQLite", () => {
		cy.visit("/setup", { failOnStatusCode: false });
		cy.get("#db_connection").select("sqlite");
		cy.get("#project").clear().type("Demo Tester");
		cy.get("#description").clear().type("This is a test from Cypress!");
		cy.get("#domain").clear().type("http://localhost:8080");
		cy.get("#email").clear().type("admin@example.com");
		cy.get("#username").clear().type("admin");
		// Password must be 30+ chars with upper, lower, digit
		cy.get("#password").clear().type("AdminPassword123456789012345678");
		cy.get("#password_confirm").clear().type("AdminPassword123456789012345678");
		// Enable sample data so tests have services/groups/stats to work with
		cy.get('label[for="sample_data"]').click({ force: true });
		// Wait for button to be enabled
		cy.get('button[type="submit"]').should('not.be.disabled', { timeout: 5000 });
		cy.get('button[type="submit"]').click();

		// Wait for setup to complete - shows confirmation page with "Continue to Dashboard"
		cy.contains('Setup Complete!', { timeout: 60000 }).should('be.visible');
		cy.contains('Continue to Dashboard').click();

		// Now should be on dashboard
		cy.url({ timeout: 10000 }).should('not.include', '/setup');
		cy.get("#title", { timeout: 10000 }).should("contain", "Demo Tester");
	});

	it("should show dashboard after setup", () => {
		cy.visit("/");
		cy.get("#title", { timeout: 10000 }).should("contain", "Demo Tester");
	});

	it("should be completely setup", () => {
		cy.request(`/api`).then((response) => {
			expect(response.body).to.have.property("setup", true);
			expect(response.body).to.have.property("domain", "http://localhost:8080");
		});
	});

	it("should be able to Login", () => {
		cy.visit("/login");
		cy.get("#username", { timeout: 10000 }).should("not.be.disabled");
		cy.get("#username").clear().type("admin");
		cy.get("#password").clear().type("AdminPassword123456789012345678");
		cy.get('button[type="submit"]').click();

		// Verify login succeeded by checking for admin-only UI elements
		cy.get(".navbar-brand", { timeout: 10000 }).should("be.visible");
		cy.contains("Logout", { timeout: 10000 }).should("be.visible");
		cy.contains("Settings").should("be.visible");
		cy.getCookies().should("have.length.at.least", 1);
	});
});
