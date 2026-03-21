import { test, expect } from '@playwright/test';

const DEFAULT_NAV_ENTRIES = [
    { href: '/dashboard', label: 'AI Organization' },
    { href: '/automations', label: 'Automations' },
    { href: '/docs', label: 'Docs' },
    { href: '/settings', label: 'Settings' },
];

const ADVANCED_NAV_ENTRIES = [
    { href: '/resources', label: 'Resources' },
    { href: '/memory', label: 'Memory' },
    { href: '/system', label: 'System' },
];

test.describe('V8.1 Soma-primary Navigation', () => {
    test.beforeEach(async ({ page }) => {
        await page.goto('/dashboard');
        await page.waitForLoadState('domcontentloaded');
        await expect(page.locator('a[href="/dashboard"]').first()).toBeVisible();
    });

    test('ZoneA rail is visible on page load', async ({ page }) => {
        await expect(page.locator('a[href="/dashboard"]').first()).toBeVisible();
    });

    test('default MVP navigation entries are present', async ({ page }) => {
        for (const entry of DEFAULT_NAV_ENTRIES) {
            await expect(page.locator(`a[href="${entry.href}"]`).first()).toBeVisible();
        }
        for (const entry of ADVANCED_NAV_ENTRIES) {
            await expect(page.locator(`a[href="${entry.href}"]`)).toHaveCount(0);
        }
    });

    test('active route gets primary highlight', async ({ page }) => {
        const dashboardLink = page.locator('a[href="/dashboard"]').first();
        await expect(dashboardLink).toHaveClass(/bg-cortex-primary/);
    });

    test('navigating to each primary route highlights the corresponding nav item', async ({ page }) => {
        const routes: Array<{ href: string; url: RegExp }> = [
            { href: '/automations', url: /\/automations/ },
            { href: '/docs', url: /\/docs/ },
            { href: '/settings', url: /\/settings/ },
        ];

        for (const route of routes) {
            await page.goto(route.href, { waitUntil: 'domcontentloaded' });
            const link = page.locator(`a[href="${route.href}"]`).first();
            await expect(page).toHaveURL(route.url);
            await expect(link).toHaveClass(/bg-cortex-primary/);
        }
    });

    test('System tab is hidden by default (advanced mode off)', async ({ page }) => {
        await expect(page.locator('a[href="/system"]')).not.toBeVisible();
    });

    test('Advanced toggle shows and hides advanced support routes', async ({ page }) => {
        await page.getByRole('button', { name: 'Advanced: Off' }).click();
        for (const entry of ADVANCED_NAV_ENTRIES) {
            await expect(page.locator(`a[href="${entry.href}"]`).first()).toBeVisible();
        }
        await page.getByRole('button', { name: 'Advanced: On' }).click();
        for (const entry of ADVANCED_NAV_ENTRIES) {
            await expect(page.locator(`a[href="${entry.href}"]`)).toHaveCount(0);
        }
    });
});
