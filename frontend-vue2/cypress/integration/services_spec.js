/// <reference types="cypress" />

import "../support/commands";

context("Services Tests", () => {
	beforeEach(() => {
		cy.login();
	});

	it("should goto services", () => {
		cy.visit("/dashboard/services");
		cy.get("#services_list > tr").should("have.length.at.least", 1);
	});

	it("should create new HTTP service", () => {
		cy.visit("/dashboard/create_service");
		cy.get("#name").clear().type("Cypress HTTP Service");
		cy.get("#service_type").select("http");
		cy.get("#service_url").clear().type("http://localhost:8080");
		cy.get("#service_response_code").clear().type("200");
		cy.get("#service_interval").invoke("val", 60).trigger("change");
		cy.get("#timeout").invoke("val", 10).trigger("change");
		cy.get("#permalink").clear().type("cypress_http_service");
		cy.get("#notify_after").invoke("val", 3).trigger("change");
		cy.get('button[type="submit"]').click();

		cy.visit("/dashboard/services");
		cy.get("#services_list").contains("Cypress HTTP Service");
	});

	it("should confirm new service exists", () => {
		cy.visit("/dashboard/services");
		cy.get("#services_list").contains("Cypress HTTP Service");
	});

	it("should delete the test service", () => {
		cy.visit("/dashboard/services");
		cy.contains("tr", "Cypress HTTP Service").find(".btn-danger").click();
		cy.get("#services_list").should("not.contain", "Cypress HTTP Service");
	});
});
