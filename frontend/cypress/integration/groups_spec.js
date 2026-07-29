/// <reference types="cypress" />

import "../support/commands";

context("Groups Tests", () => {
	beforeEach(() => {
		cy.login();
	});

	it("should goto groups", () => {
		cy.visit("/dashboard/services");
		cy.get(".sortable_groups > tr").should("have.length.at.least", 1);
	});

	it("should create new Group", () => {
		cy.visit("/dashboard/services");
		cy.get("#title").clear().type("Cypress Test Group");
		cy.get('button[type="submit"]').click();
		cy.contains("Cypress Test Group").should("exist");
	});

	it("should confirm new group exists", () => {
		cy.visit("/dashboard/services");
		cy.contains("Cypress Test Group").should("exist");
	});

	it("should delete the test group", () => {
		cy.visit("/dashboard/services");
		cy.contains("tr", "Cypress Test Group").find(".btn-danger").click();
		cy.contains("Cypress Test Group").should("not.exist");
	});
});
