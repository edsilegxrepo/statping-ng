import { test, expect } from '@playwright/test'

test.describe('Login Flow', () => {
  test.beforeEach(async ({ page }) => {
    await page.route('**/api/core', (route) => {
      route.fulfill({
        status: 200,
        body: JSON.stringify({
          setup: true,
          name: 'Test Status Page',
          description: 'Test Description',
          domain: 'http://localhost:8080',
        }),
      })
    })
  })

  test('displays login page', async ({ page }) => {
    await page.goto('/login')

    await expect(page.locator('input[name="username"]')).toBeVisible()
    await expect(page.locator('input[name="password"]')).toBeVisible()
    await expect(page.locator('button[type="submit"]')).toBeVisible()
  })

  test('shows error on invalid credentials', async ({ page }) => {
    await page.route('**/api/login', (route) => {
      route.fulfill({
        status: 401,
        body: JSON.stringify({ error: 'Invalid credentials' }),
      })
    })

    await page.goto('/login')

    await page.fill('input[name="username"]', 'wronguser')
    await page.fill('input[name="password"]', 'wrongpass')
    await page.click('button[type="submit"]')

    await expect(page.locator('.alert-danger, .text-danger')).toBeVisible()
  })

  test('redirects to dashboard on successful login', async ({ page }) => {
    await page.route('**/api/login', (route) => {
      route.fulfill({
        status: 200,
        body: JSON.stringify({ token: 'test-jwt-token', admin: true }),
      })
    })

    await page.route('**/api/services', (route) => {
      route.fulfill({ status: 200, body: JSON.stringify([]) })
    })

    await page.route('**/api/groups', (route) => {
      route.fulfill({ status: 200, body: JSON.stringify([]) })
    })

    await page.goto('/login')

    await page.fill('input[name="username"]', 'admin')
    await page.fill('input[name="password"]', 'admin')
    await page.click('button[type="submit"]')

    await expect(page).toHaveURL(/\/dashboard/)
  })

  test('persists authentication in localStorage', async ({ page }) => {
    await page.route('**/api/login', (route) => {
      route.fulfill({
        status: 200,
        body: JSON.stringify({ token: 'test-jwt-token', admin: true }),
      })
    })

    await page.goto('/login')

    await page.fill('input[name="username"]', 'admin')
    await page.fill('input[name="password"]', 'admin')
    await page.click('button[type="submit"]')

    const token = await page.evaluate(() => localStorage.getItem('statping_auth'))
    expect(token).toBe('test-jwt-token')
  })
})
