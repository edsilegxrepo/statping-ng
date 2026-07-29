/// <reference types="cypress" />

import "../support/commands";

context("Service Edit Tests", () => {
	beforeEach(() => {
		cy.login();
	});

	it("should navigate to edit service page", () => {
		cy.visit("/dashboard/services");
		cy.get("#services_list > tr").first().find(".btn-outline-secondary").first().click();
		cy.url().should("include", "/dashboard/edit_service/");
	});

	it("should display service edit form", () => {
		cy.visit("/dashboard/services");
		cy.get("#services_list > tr").first().find(".btn-outline-secondary").first().click();
		cy.get("#name").should("not.have.value", "");
	});

	it("should update and restore service name", () => {
		cy.visit("/dashboard/services");
		// Get first service name
		cy.get("#services_list > tr").first().find("td").first().invoke("text").then((originalName) => {
			const trimmedName = originalName.trim();

			// Edit the service
			cy.get("#services_list > tr").first().find(".btn-outline-secondary").first().click();
			cy.get("#name").clear().type("Cypress Renamed Service");
			cy.get('button[type="submit"]').click();

			// Verify renamed
			cy.visit("/dashboard/services");
			cy.get("#services_list").contains("Cypress Renamed Service");

			// Restore original name
			cy.contains("tr", "Cypress Renamed Service").find(".btn-outline-secondary").first().click();
			cy.get("#name").clear().type(trimmedName);
			cy.get('button[type="submit"]').click();

			// Verify restored
			cy.visit("/dashboard/services");
			cy.get("#services_list").contains(trimmedName);
		});
	});
});
