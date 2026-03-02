import { test, expect } from '@playwright/test';

const V7_NAV_ENTRIES = [
    { href: '/dashboard', label: 'Workspace' },
    { href: '/automations', label: 'Automations' },
    { href: '/resources', label: 'Resources' },
    { href: '/memory', label: 'Memory' },
    { href: '/docs', label: 'Docs' },
    { href: '/settings', label: 'Settings' },
];

test.describe('V7 Workflow-First Navigation', () => {
    test.beforeEach(async ({ page }) => {
        await page.goto('/dashboard');
        await page.waitForLoadState('domcontentloaded');
        await expect(page.locator('a[href="/dashboard"]').first()).toBeVisible();
    });

    test('ZoneA rail is visible on page load', async ({ page }) => {
        await expect(page.getByRole('link', { name: 'Workspace' })).toBeVisible();
    });

    test('all V7 navigation entries are present', async ({ page }) => {
        for (const entry of V7_NAV_ENTRIES) {
            await expect(page.locator(`a[href="${entry.href}"]`).first()).toBeVisible();
        }
    });

    test('active route gets primary highlight', async ({ page }) => {
        const dashboardLink = page.getByRole('link', { name: 'Workspace' });
        await expect(dashboardLink).toHaveClass(/bg-cortex-primary/);
    });

    test('navigating to each primary route highlights the corresponding nav item', async ({ page }) => {
        const routes: Array<{ href: string; url: RegExp }> = [
            { href: '/automations', url: /\/automations/ },
            { href: '/resources', url: /\/resources/ },
            { href: '/memory', url: /\/memory/ },
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

    test('Advanced toggle shows/hides System tab', async ({ page }) => {
        await page.getByRole('button', { name: 'Advanced: Off' }).click();
        await expect(page.locator('a[href="/system"]')).toBeVisible();
        await page.getByRole('button', { name: 'Advanced: On' }).click();
        await expect(page.locator('a[href="/system"]')).not.toBeVisible();
    });

});
