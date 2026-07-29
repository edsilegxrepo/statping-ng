/// <reference types="cypress" />

import "../support/commands";

context("Users Tests", () => {
	beforeEach(() => {
		cy.login();
	});

	it("should goto users", () => {
		cy.visit("/dashboard/users");
		cy.get("#users_table > tr").should("have.length.at.least", 1);
		cy.get("#users_table").contains("admin");
	});

	it("should create new User", () => {
		cy.visit("/dashboard/users");
		cy.get("#username").clear().type("cypressuser");
		cy.get("#email").clear().type("cypress@example.com");
		cy.get("#password").clear().type("CypressPassword12345678901234");
		cy.get("#password_confirm").clear().type("CypressPassword12345678901234");
		cy.get('button[type="submit"]').click();

		cy.get("#users_table", { timeout: 10000 }).contains("cypressuser");
	});

	it("should confirm new user exists", () => {
		cy.visit("/dashboard/users");
		cy.get("#users_table").contains("cypressuser");
	});

	it("should delete the test user", () => {
		cy.visit("/dashboard/users");
		cy.contains("tr", "cypressuser").find(".delete-user").click();
		cy.get("#users_table").should("not.contain", "cypressuser");
	});
});
