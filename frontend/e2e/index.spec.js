import { expect, test } from "@playwright/test";

test.describe("Index Page (Public Status)", () => {
	test.beforeEach(async ({ page }) => {
		await page.route("**/api/core", (route) => {
			route.fulfill({
				status: 200,
				body: JSON.stringify({
					setup: true,
					name: "Test Status Page",
					description: "Monitoring all services",
					domain: "http://localhost:8080",
					footer: "",
				}),
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
						status: "online",
						latency: 45,
						stats: { uptime: 99.9 },
					},
					{
						id: 2,
						name: "GitHub",
						domain: "https://github.com",
						type: "http",
						online: true,
						status: "online",
						latency: 120,
						stats: { uptime: 99.5 },
					},
					{
						id: 3,
						name: "Offline Service",
						domain: "https://example.com",
						type: "http",
						online: false,
						status: "offline",
						latency: 0,
						stats: { uptime: 85.0 },
					},
				]),
			});
		});

		await page.route("**/api/groups", (route) => {
			route.fulfill({ status: 200, body: JSON.stringify([]) });
		});

		await page.route("**/api/messages", (route) => {
			route.fulfill({ status: 200, body: JSON.stringify([]) });
		});
	});

	test("displays the status page title and description", async ({ page }) => {
		await page.goto("/");

		await expect(page.getByText("Test Status Page")).toBeVisible();
		await expect(page.getByText("Monitoring all services")).toBeVisible();
	});

	test("shows list of services", async ({ page }) => {
		await page.goto("/");

		await expect(page.getByText("Google")).toBeVisible();
		await expect(page.getByText("GitHub")).toBeVisible();
		await expect(page.getByText("Offline Service")).toBeVisible();
	});

	test("navigates to service details on click", async ({ page }) => {
		await page.route("**/api/services/1", (route) => {
			route.fulfill({
				status: 200,
				body: JSON.stringify({
					id: 1,
					name: "Google",
					domain: "https://google.com",
					type: "http",
					online: true,
				}),
			});
		});

		await page.goto("/");
		await page.getByText("Google").click();

		await expect(page).toHaveURL(/\/service\//);
	});

	test("displays announcements when present", async ({ page }) => {
		await page.route("**/api/messages", (route) => {
			route.fulfill({
				status: 200,
				body: JSON.stringify([
					{
						id: 1,
						title: "Scheduled Maintenance",
						description: "System will be down for maintenance",
						start_on: new Date().toISOString(),
						end_on: new Date(Date.now() + 86400000).toISOString(),
					},
				]),
			});
		});

		await page.goto("/");

		await expect(page.getByText("Scheduled Maintenance")).toBeVisible();
	});
});
