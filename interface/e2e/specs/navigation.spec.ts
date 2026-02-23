import { test, expect } from '@playwright/test';

const V7_NAV_ENTRIES = [
    { href: '/dashboard', label: 'Mission Control' },
    { href: '/automations', label: 'Automations' },
    { href: '/resources', label: 'Resources' },
    { href: '/memory', label: 'Memory' },
    { href: '/settings', label: 'Settings' },
];

test.describe('V7 Workflow-First Navigation', () => {

    test.beforeEach(async ({ page }) => {
        await page.goto('/dashboard');
        await page.waitForLoadState('networkidle');
    });

    test('ZoneA rail is visible on page load', async ({ page }) => {
        await expect(page.locator('text=Mycelis')).toBeVisible();
    });

    test('all V7 navigation entries are present', async ({ page }) => {
        for (const entry of V7_NAV_ENTRIES) {
            const link = page.locator(`a[href="${entry.href}"]`);
            await expect(link).toBeVisible();
            await expect(link.locator(`text=${entry.label}`)).toBeVisible();
        }
    });

    test('active route gets primary highlight', async ({ page }) => {
        const dashboardLink = page.locator('a[href="/dashboard"]');
        await expect(dashboardLink).toHaveClass(/bg-cortex-primary/);
    });

    test('navigating to /automations highlights Automations', async ({ page }) => {
        const link = page.locator('a[href="/automations"]');
        await link.click();
        await page.waitForLoadState('networkidle');
        await expect(link).toHaveClass(/bg-cortex-primary/);
        await expect(page.locator('text=Automations').first()).toBeVisible();
    });

    test('navigating to /resources highlights Resources', async ({ page }) => {
        const link = page.locator('a[href="/resources"]');
        await link.click();
        await page.waitForLoadState('networkidle');
        await expect(link).toHaveClass(/bg-cortex-primary/);
        await expect(page.locator('text=Resources').first()).toBeVisible();
    });

    test('navigating to /memory highlights Memory', async ({ page }) => {
        const link = page.locator('a[href="/memory"]');
        await link.click();
        await page.waitForLoadState('networkidle');
        await expect(link).toHaveClass(/bg-cortex-primary/);
    });

    test('inactive nav items have muted text', async ({ page }) => {
        const resourcesLink = page.locator('a[href="/resources"]');
        await expect(resourcesLink).toHaveClass(/text-cortex-text-muted/);
    });

    test('System tab is hidden by default (advanced mode off)', async ({ page }) => {
        const systemLink = page.locator('a[href="/system"]');
        await expect(systemLink).not.toBeVisible();
    });

    test('Advanced toggle shows/hides System tab', async ({ page }) => {
        const advancedToggle = page.locator('text=Advanced: Off');
        await advancedToggle.click();

        const systemLink = page.locator('a[href="/system"]');
        await expect(systemLink).toBeVisible();

        const advancedOnToggle = page.locator('text=Advanced: On');
        await advancedOnToggle.click();
        await expect(systemLink).not.toBeVisible();
    });

    test('nav order matches V7 spec', async ({ page }) => {
        const navLinks = page.locator('a[href^="/"]');
        const hrefs: string[] = [];
        const count = await navLinks.count();
        for (let i = 0; i < count; i++) {
            const href = await navLinks.nth(i).getAttribute('href');
            if (href) hrefs.push(href);
        }

        const dashboardIdx = hrefs.indexOf('/dashboard');
        const automationsIdx = hrefs.indexOf('/automations');
        const resourcesIdx = hrefs.indexOf('/resources');
        const memoryIdx = hrefs.indexOf('/memory');

        expect(dashboardIdx).toBeGreaterThan(-1);
        expect(automationsIdx).toBe(dashboardIdx + 1);
        expect(resourcesIdx).toBe(automationsIdx + 1);
        expect(memoryIdx).toBe(resourcesIdx + 1);
    });

    test('no bg-white leak in navigation rail', async ({ page }) => {
        const railHtml = await page.locator('[class*="flex-col"]').first().innerHTML();
        expect(railHtml).not.toContain('bg-white');
    });

    test('legacy /wiring URL redirects to /automations', async ({ page }) => {
        await page.goto('/wiring');
        await page.waitForLoadState('networkidle');
        expect(page.url()).toContain('/automations');
    });

    test('legacy /teams URL redirects to /automations', async ({ page }) => {
        await page.goto('/teams');
        await page.waitForLoadState('networkidle');
        expect(page.url()).toContain('/automations');
    });

    test('legacy /catalogue URL redirects to /resources', async ({ page }) => {
        await page.goto('/catalogue');
        await page.waitForLoadState('networkidle');
        expect(page.url()).toContain('/resources');
    });

    test('legacy /approvals URL redirects to /automations', async ({ page }) => {
        await page.goto('/approvals');
        await page.waitForLoadState('networkidle');
        expect(page.url()).toContain('/automations');
    });

    test('legacy /telemetry URL redirects to /system', async ({ page }) => {
        await page.goto('/telemetry');
        await page.waitForLoadState('networkidle');
        expect(page.url()).toContain('/system');
    });

    test('legacy /matrix URL redirects to /system', async ({ page }) => {
        await page.goto('/matrix');
        await page.waitForLoadState('networkidle');
        expect(page.url()).toContain('/system');
    });
});
