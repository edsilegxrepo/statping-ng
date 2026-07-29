import { test, expect } from '@playwright/test'

test.describe('Setup Flow', () => {
  test('shows setup page when app is not configured', async ({ page }) => {
    await page.route('**/api/core', (route) => {
      route.fulfill({
        status: 200,
        body: JSON.stringify({ setup: false }),
      })
    })

    await page.goto('/')
    await expect(page).toHaveURL(/\/setup/)
  })

  test('displays setup form with required fields', async ({ page }) => {
    await page.route('**/api/core', (route) => {
      route.fulfill({
        status: 200,
        body: JSON.stringify({ setup: false }),
      })
    })

    await page.goto('/setup')

    await expect(page.locator('input[name="project"]')).toBeVisible()
    await expect(page.locator('input[name="description"]')).toBeVisible()
    await expect(page.locator('input[name="domain"]')).toBeVisible()
    await expect(page.locator('input[name="username"]')).toBeVisible()
    await expect(page.locator('input[name="password"]')).toBeVisible()
  })

  test('completes setup successfully', async ({ page }) => {
    await page.route('**/api/core', (route) => {
      route.fulfill({
        status: 200,
        body: JSON.stringify({ setup: false }),
      })
    })

    await page.route('**/api/setup', (route) => {
      route.fulfill({
        status: 200,
        body: JSON.stringify({ status: 'success' }),
      })
    })

    await page.goto('/setup')

    await page.fill('input[name="project"]', 'Test Status Page')
    await page.fill('input[name="description"]', 'Monitoring all services')
    await page.fill('input[name="domain"]', 'http://localhost:8080')
    await page.fill('input[name="username"]', 'admin')
    await page.fill('input[name="password"]', 'admin123')

    await page.click('button[type="submit"]')
  })
})
