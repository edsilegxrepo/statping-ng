import { expect, test } from "@playwright/test";

test.describe("Service Management", () => {
	test.beforeEach(async ({ page }) => {
		await page.route("**/api/core", (route) => {
			route.fulfill({
				status: 200,
				body: JSON.stringify({
					setup: true,
					name: "Test Status Page",
					domain: "http://localhost:8080",
				}),
			});
		});

		await page.route("**/api/groups", (route) => {
			route.fulfill({
				status: 200,
				body: JSON.stringify([{ id: 1, name: "Web Services" }]),
			});
		});

		await page.route("**/api/services", (route) => {
			route.fulfill({
				status: 200,
				body: JSON.stringify([
					{
						id: 1,
						name: "Google",
						domain: "https://google.com",
						type: "http",
						online: true,
						check_interval: 60,
						timeout: 30,
					},
				]),
			});
		});

		await page.route("**/api/checkins", (route) => {
			route.fulfill({ status: 200, body: JSON.stringify([]) });
		});

		await page.evaluate(() => {
			localStorage.setItem("statping_auth", "test-jwt-token");
		});
	});

	test("creates an HTTP service", async ({ page }) => {
		await page.route("**/api/services", (route, request) => {
			if (request.method() === "POST") {
				route.fulfill({
					status: 200,
					body: JSON.stringify({
						id: 2,
						name: "New HTTP Service",
						type: "http",
					}),
				});
			} else {
				route.fulfill({ status: 200, body: JSON.stringify([]) });
			}
		});

		await page.goto("/dashboard/create_service");

		await page.fill('input[name="name"]', "New HTTP Service");
		await page.fill('input[name="domain"]', "https://httpbin.org/get");
		await page.selectOption('select[name="type"]', "http");

		await page.click('button[type="submit"]');
	});

	test("loads existing service data for editing", async ({ page }) => {
		await page.route("**/api/services/1", (route) => {
			route.fulfill({
				status: 200,
				body: JSON.stringify({
					id: 1,
					name: "Google",
					domain: "https://google.com",
					type: "http",
					method: "GET",
					expected_status: 200,
					check_interval: 60,
					timeout: 30,
					public: true,
					verify_ssl: true,
				}),
			});
		});

		await page.goto("/dashboard/service/1/edit");

		await expect(page.locator('input[name="name"]')).toHaveValue("Google");
		await expect(page.locator('input[name="domain"]')).toHaveValue(
			"https://google.com",
		);
	});

	test("displays service statistics", async ({ page }) => {
		await page.route("**/api/services/1", (route) => {
			route.fulfill({
				status: 200,
				body: JSON.stringify({
					id: 1,
					name: "Google",
					domain: "https://google.com",
					type: "http",
					online: true,
					latency: 45,
					stats: { uptime: 99.9, failures: 2 },
				}),
			});
		});

		await page.route("**/api/services/1/hits*", (route) => {
			route.fulfill({ status: 200, body: JSON.stringify([]) });
		});

		await page.route("**/api/services/1/failures*", (route) => {
			route.fulfill({ status: 200, body: JSON.stringify([]) });
		});

		await page.goto("/service/1");

		await expect(page.getByText("Google")).toBeVisible();
		await expect(page.getByText("99.9")).toBeVisible();
	});
});
