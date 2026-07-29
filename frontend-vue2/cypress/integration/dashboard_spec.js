/// <reference types="cypress" />

import "../support/commands";

context("Dashboard Tests", () => {
	beforeEach(() => {
		cy.login();
	});

	it("should display dashboard with service cards", () => {
		cy.visit("/");
		cy.get("#title").should("be.visible");
		cy.get(".card.index-chart").should("have.length.at.least", 1);
	});

	it("should display service status badges", () => {
		cy.visit("/");
		cy.get(".card.index-chart").first().within(() => {
			cy.get(".badge").should("exist");
			cy.get(".badge").invoke("text").should("match", /ONLINE|OFFLINE/);
		});
	});

	it("should display group headers", () => {
		cy.visit("/");
		cy.get("h4, .group-header, [class*='group']").should("have.length.at.least", 1);
	});

	it("should have clickable service cards", () => {
		cy.visit("/");
		cy.get(".card.index-chart").first().find("button, a").contains("View").click();
		cy.url().should("include", "/service/");
	});

	it("should display service detail page with charts", () => {
		cy.visit("/");
		cy.get(".card.index-chart").first().find("button, a").contains("View").click();
		cy.url().should("include", "/service/");
		cy.get("h3, h4, .service-name, .card-header").should("be.visible");
		cy.get(".apexcharts-canvas, [class*='chart'], svg").should("exist");
	});

	it("should navigate back to dashboard from service detail", () => {
		cy.visit("/");
		cy.get(".card.index-chart").first().find("button, a").contains("View").click();
		cy.url().should("include", "/service/");
		cy.get(".navbar-brand").first().click();
		cy.url().should("eq", Cypress.config().baseUrl + "/");
	});

	it("should Logout successfully", () => {
		cy.visit("/dashboard");
		cy.get(".nav-link").contains("Logout").click();
		cy.url().should("include", "/login");
	});

	it("should redirect to login when not authenticated", () => {
		cy.clearCookies();
		Cypress.session.clearAllSavedSessions();
		cy.visit("/dashboard/services");
		cy.url().should("include", "/login");
	});

	it("should show failed login message with wrong credentials", () => {
		cy.clearCookies();
		Cypress.session.clearAllSavedSessions();
		cy.visit("/login");
		cy.get("#username").clear().type("admin");
		cy.get("#password").clear().type("wrongpassword");
		cy.get('button[type="submit"]').click();
		cy.get(".alert, [class*='error'], [class*='alert']").should("be.visible");
		cy.url().should("include", "/login");
	});
});
