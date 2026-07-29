/// <reference types="cypress" />

import "../support/commands";

context("Settings Tests", () => {
	beforeEach(() => {
		cy.login();
	});

	it("should confirm notifiers are installed", () => {
		cy.visit("/dashboard/settings");
		cy.get("#notifiers_tabs > a").should("have.length.at.least", 5);
	});

	it("should update Statping settings", () => {
		cy.visit("/dashboard/settings");
		cy.get("#project").clear().type("Cypress Updated");
		cy.get("#description").clear().type("Updated by Cypress E2E tests");
		cy.get("#save_core").click();
	});

	it("should confirm Statping settings were updated", () => {
		cy.visit("/dashboard/settings");
		cy.get("#project").should("have.value", "Cypress Updated");
		cy.get("#description").should("have.value", "Updated by Cypress E2E tests");
	});

	it("should view Cache tab", () => {
		cy.visit("/dashboard/settings");
		cy.get("#v-pills-cache-tab").click();
		cy.get(".card").should("exist");
	});
});
