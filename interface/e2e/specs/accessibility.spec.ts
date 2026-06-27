import AxeBuilder from '@axe-core/playwright';
import { test, expect } from '@playwright/test';

async function analyzeAppWorkspace(page: import('@playwright/test').Page) {
    return new AxeBuilder({ page })
        .include('[data-testid="app-workspace"]')
        .withTags(['wcag2a', 'wcag2aa'])
        .disableRules(['color-contrast'])
        .analyze();
}

test.describe('Accessibility Baseline', () => {
    test.skip(({ browserName }) => browserName === 'firefox', 'Firefox accessibility coverage is currently unstable for these routes; keep Chromium as the baseline browser for now.');
    test.slow();

    test('dashboard has no critical a11y violations', async ({ page }) => {
        await page.goto('/dashboard', { waitUntil: 'domcontentloaded' });
        await expect(page.getByRole('heading', { name: /Talk to Soma/i })).toBeVisible();

        const results = await analyzeAppWorkspace(page);

        const critical = results.violations.filter(
            (v: any) => v.impact === 'critical'
        );
        expect(critical).toHaveLength(0);
    });

    test('automations page has no critical a11y violations', async ({ page }) => {
        await page.goto('/automations', { waitUntil: 'domcontentloaded' });
        await expect(page.getByRole('heading', { name: 'Automations' })).toBeVisible();

        const results = await analyzeAppWorkspace(page);

        const critical = results.violations.filter(
            (v: any) => v.impact === 'critical'
        );
        expect(critical).toHaveLength(0);
    });

    test('settings page has no critical a11y violations', async ({ page }) => {
        await page.goto('/settings', { waitUntil: 'domcontentloaded' });
        await expect(page.getByRole('heading', { name: 'Settings' })).toBeVisible();

        const results = await analyzeAppWorkspace(page);

        const critical = results.violations.filter(
            (v: any) => v.impact === 'critical'
        );
        expect(critical).toHaveLength(0);
    });
});
