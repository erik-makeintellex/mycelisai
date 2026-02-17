import { test, expect } from '@playwright/test';

let AxeBuilder: any;
try {
    AxeBuilder = require('@axe-core/playwright').default;
} catch {
    // axe-core not installed — tests will be skipped
}

test.describe('Accessibility Baseline', () => {
    test.skip(!AxeBuilder, '@axe-core/playwright not installed');

    test('dashboard has no critical a11y violations', async ({ page }) => {
        await page.goto('/');
        await page.waitForLoadState('networkidle');

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
        await page.waitForLoadState('networkidle');

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
        await page.waitForLoadState('networkidle');

        const results = await new AxeBuilder({ page })
            .withTags(['wcag2a', 'wcag2aa'])
            .analyze();

        const critical = results.violations.filter(
            (v: any) => v.impact === 'critical'
        );
        expect(critical).toHaveLength(0);
    });
});
