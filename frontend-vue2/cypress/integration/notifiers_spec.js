/// <reference types="cypress" />

import "../support/commands";

context("Notifier Tests", () => {
	beforeEach(() => {
		cy.login();
	});

	it("should confirm notifiers are installed", () => {
		cy.visit("/dashboard/settings");
		cy.get("#notifiers_tabs > a").should("have.length.at.least", 5);
	});

	it("should view notifier tabs", () => {
		cy.visit("/dashboard/settings");
		cy.get("#notifiers_tabs > a").first().click();
		cy.get("form").should("exist");
	});
});
