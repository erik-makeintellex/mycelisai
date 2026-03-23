import AxeBuilder from '@axe-core/playwright';
import { test, expect } from '@playwright/test';

test.describe('Accessibility Baseline', () => {
    test.slow();

    test('dashboard has no critical a11y violations', async ({ page }) => {
        await page.goto('/');
        await page.waitForLoadState('domcontentloaded');
        await expect(page.getByRole('link', { name: 'Create AI Organization' }).first()).toBeVisible();

        const results = await new AxeBuilder({ page })
            .withTags(['wcag2a', 'wcag2aa'])
            .analyze();

        const critical = results.violations.filter(
            (v: any) => v.impact === 'critical'
        );
        expect(critical).toHaveLength(0);
    });

    test('automations page has no critical a11y violations', async ({ page }) => {
        await page.goto('/automations');
        await page.waitForLoadState('domcontentloaded');
        await expect(page.getByRole('heading', { name: 'Automations' })).toBeVisible();

        const results = await new AxeBuilder({ page })
            .withTags(['wcag2a', 'wcag2aa'])
            .analyze();

        const critical = results.violations.filter(
            (v: any) => v.impact === 'critical'
        );
        expect(critical).toHaveLength(0);
    });

    test('settings page has no critical a11y violations', async ({ page }) => {
        await page.goto('/settings');
        await page.waitForLoadState('domcontentloaded');
        await expect(page.getByRole('heading', { name: 'Settings' })).toBeVisible();

        const results = await new AxeBuilder({ page })
            .withTags(['wcag2a', 'wcag2aa'])
            .analyze();

        const critical = results.violations.filter(
            (v: any) => v.impact === 'critical'
        );
        expect(critical).toHaveLength(0);
    });
});
