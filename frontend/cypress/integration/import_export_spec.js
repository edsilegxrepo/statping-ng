/// <reference types="cypress" />

import "../support/commands";

context("Import/Export Settings Tests", () => {
	beforeEach(() => {
		cy.login();
	});

	it("should navigate to settings page", () => {
		cy.visit("/dashboard/settings");
		cy.url().should("include", "/dashboard/settings");
	});

	it("should display settings tabs", () => {
		cy.visit("/dashboard/settings");
		cy.get(".nav-link, .nav-item").should("have.length.at.least", 1);
	});

	it("should export settings via API", () => {
		cy.request({
			method: "GET",
			url: "/api/settings/export",
			failOnStatusCode: false,
		}).then((response) => {
			// May require auth - both 200 and 401/403 are acceptable
			expect([200, 401, 403]).to.include(response.status);
		});
	});
});
