import { test, expect } from '@playwright/test';

const DEFAULT_NAV_ENTRIES = [
    { href: '/dashboard', label: 'AI Organization', testId: 'nav-dashboard' },
    { href: '/automations', label: 'Automations', testId: 'nav-automations' },
    { href: '/docs', label: 'Docs', testId: 'nav-docs' },
    { href: '/settings', label: 'Settings', testId: 'nav-settings' },
];

const ADVANCED_NAV_ENTRIES = [
    { href: '/resources', label: 'Resources', testId: 'nav-resources' },
    { href: '/memory', label: 'Memory', testId: 'nav-memory' },
    { href: '/system', label: 'System', testId: 'nav-system' },
];

test.describe('V8.1 Soma-primary Navigation', () => {
    test.describe.configure({ mode: 'serial' });

    test.beforeEach(async ({ page }) => {
        await page.goto('/dashboard', { waitUntil: 'domcontentloaded' });
        await page.evaluate(() => {
            window.localStorage.setItem('mycelis-advanced-mode', 'false');
        });
        await page.reload({ waitUntil: 'domcontentloaded' });
        await expect(page.getByTestId('nav-dashboard')).toBeVisible();
    });

    test('ZoneA rail is visible on page load', async ({ page }) => {
        await expect(page.getByTestId('nav-dashboard')).toBeVisible();
    });

    test('default MVP navigation entries are present', async ({ page }) => {
        for (const entry of DEFAULT_NAV_ENTRIES) {
            await expect(page.getByTestId(entry.testId)).toBeVisible();
        }
        for (const entry of ADVANCED_NAV_ENTRIES) {
            await expect(page.getByTestId(entry.testId)).toHaveCount(0);
        }
    });

    test('active route gets primary highlight', async ({ page }) => {
        const dashboardLink = page.getByTestId('nav-dashboard');
        await expect(dashboardLink).toHaveClass(/bg-cortex-primary/);
    });

    for (const route of [
        { href: '/automations', testId: 'nav-automations', url: /\/automations/, label: 'automations' },
        { href: '/settings', testId: 'nav-settings', url: /\/settings/, label: 'settings' },
    ]) {
        test(`navigating to ${route.label} highlights the corresponding nav item`, async ({ page, browserName }) => {
            test.slow();
            test.skip(
                browserName === 'webkit',
                'WebKit intermittently self-navigates back to /dashboard on the managed Next dev server; Chromium/Firefox plus rail unit tests cover this highlight contract.',
            );
            await page.goto(route.href, { waitUntil: 'domcontentloaded' });
            await expect(page).toHaveURL(route.url);
            const link = page.getByTestId(route.testId);
            await expect(link).toBeVisible();
            await expect(link).toHaveClass(/bg-cortex-primary/);
        });
    }

    test('System tab is hidden by default (advanced mode off)', async ({ page }) => {
        await expect(page.getByTestId('nav-system')).toHaveCount(0);
    });

    test('Advanced toggle flips visible state labels', async ({ page }) => {
        const toggle = page.getByRole('button', { name: /Advanced:/ });
        await expect(page.getByTestId('nav-resources')).toHaveCount(0);
        await toggle.click();
        await page.waitForFunction(() => window.localStorage.getItem('mycelis-advanced-mode') === 'true');
        await page.reload({ waitUntil: 'domcontentloaded' });
        await expect(page.getByTestId('nav-resources')).toBeVisible();
        await expect(page.getByRole('button', { name: /Advanced:/ })).toContainText('Advanced: On');
        await page.getByRole('button', { name: /Advanced:/ }).click();
        await page.waitForFunction(() => window.localStorage.getItem('mycelis-advanced-mode') === 'false');
        await page.reload({ waitUntil: 'domcontentloaded' });
        await expect(page.getByTestId('nav-resources')).toHaveCount(0);
    });
});
