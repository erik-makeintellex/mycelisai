import AxeBuilder from '@axe-core/playwright';
import { test, expect } from '@playwright/test';

test.describe('Accessibility Baseline', () => {
    test('dashboard has no critical a11y violations', async ({ page }) => {
        await page.goto('/');
        await page.waitForLoadState('domcontentloaded');

        const results = await new AxeBuilder({ page })
            .withTags(['wcag2a', 'wcag2aa'])
            .analyze();

        const critical = results.violations.filter(
            (v: any) => v.impact === 'critical'
        );
        expect(critical).toHaveLength(0);
    });

    test('teams page has no critical a11y violations', async ({ page }) => {
        await page.goto('/teams');
        await page.waitForLoadState('domcontentloaded');

        const results = await new AxeBuilder({ page })
            .withTags(['wcag2a', 'wcag2aa'])
            .analyze();

        const critical = results.violations.filter(
            (v: any) => v.impact === 'critical'
        );
        expect(critical).toHaveLength(0);
    });

    test('wiring page has no critical a11y violations', async ({ page }) => {
        await page.goto('/wiring');
        await page.waitForLoadState('domcontentloaded');

        const results = await new AxeBuilder({ page })
            .withTags(['wcag2a', 'wcag2aa'])
            .analyze();

        const critical = results.violations.filter(
            (v: any) => v.impact === 'critical'
        );
        expect(critical).toHaveLength(0);
    });
});
