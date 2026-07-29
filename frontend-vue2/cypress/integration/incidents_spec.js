/// <reference types="cypress" />

import "../support/commands";

context("Incidents Tests", () => {
	beforeEach(() => {
		cy.login();
	});

	it("should navigate to service detail", () => {
		cy.visit("/");
		cy.get(".card.index-chart").first().find("button, a").contains("View").click();
		cy.url().should("include", "/service/");
	});

	it("should display service page", () => {
		cy.visit("/");
		cy.get(".card.index-chart").first().find("button, a").contains("View").click();
		cy.get("h3, h4, .card-header").should("be.visible");
	});
});
