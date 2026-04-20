import { test, expect, type Page } from '@playwright/test';

const DEFAULT_NAV_ENTRIES = [
    { href: '/dashboard', label: 'Soma', testId: 'nav-dashboard' },
    { href: '/docs', label: 'Docs', testId: 'nav-docs' },
    { href: '/settings', label: 'Settings', testId: 'nav-settings' },
];

const ADVANCED_NAV_ENTRIES = [
    { href: '/resources', label: 'Resources', testId: 'nav-resources' },
    { href: '/memory', label: 'Memory', testId: 'nav-memory' },
    { href: '/system', label: 'System', testId: 'nav-system' },
];

async function openDashboard(page: Page) {
    for (let attempt = 0; attempt < 3; attempt += 1) {
        try {
            await page.goto('/dashboard', { waitUntil: 'domcontentloaded' });
            await page.waitForLoadState('domcontentloaded');
            return;
        } catch (error) {
            const message = error instanceof Error ? error.message : String(error);
            const canRetry =
                message.includes('ERR_ABORTED') ||
                message.includes('ERR_NETWORK_CHANGED') ||
                message.includes('frame was detached') ||
                message.includes('interrupted by another navigation') ||
                message.includes('chrome-error://chromewebdata/');
            if (!canRetry || attempt === 2) {
                throw error;
            }
            await page.waitForTimeout(500);
        }
    }
}

test.describe('V8.1 Soma-primary Navigation', () => {
    test.describe.configure({ mode: 'serial' });

    test.beforeEach(async ({ page }) => {
        await openDashboard(page);
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
        { href: '/docs', testId: 'nav-docs', url: /\/docs/, label: 'docs' },
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

    test('Advanced mode persistence flips visible state labels', async ({ page }) => {
        await expect(page.getByTestId('nav-resources')).toHaveCount(0);
        await page.evaluate(() => {
            window.localStorage.setItem('mycelis-advanced-mode', 'true');
        });
        await openDashboard(page);
        await expect(page.getByTestId('nav-resources')).toBeVisible();
        await expect(page.getByRole('button', { name: 'Advanced: On' })).toBeVisible();
        await page.evaluate(() => {
            window.localStorage.setItem('mycelis-advanced-mode', 'false');
        });
        await openDashboard(page);
        await expect(page.getByTestId('nav-resources')).toHaveCount(0);
    });
});
