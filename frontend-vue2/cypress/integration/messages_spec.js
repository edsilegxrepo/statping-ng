/// <reference types="cypress" />

import "../support/commands";

context("Announcements Tests", () => {
	beforeEach(() => {
		cy.login();
	});

	it("should goto messages", () => {
		cy.visit("/dashboard/messages");
		cy.get("tbody > tr").should("have.length.at.least", 1);
	});

	it("should create Message", () => {
		cy.visit("/dashboard/messages");
		cy.get("#title").clear().type("Cypress Test Message");
		cy.get("#description").clear().type("This message was created by Cypress!");
		cy.get('button[type="submit"]').click();
		cy.contains("Cypress Test Message").should("exist");
	});

	it("should confirm new Message exists", () => {
		cy.visit("/dashboard/messages");
		cy.contains("Cypress Test Message").should("exist");
	});

	it("should delete Message", () => {
		cy.visit("/dashboard/messages");
		cy.contains("tr", "Cypress Test Message").find(".btn-danger").click();
		cy.contains("Cypress Test Message").should("not.exist");
	});
});
