import { test, expect } from '@playwright/test';

const NAV_ENTRIES = [
    { href: '/', label: 'Mission Control' },
    { href: '/dashboard', label: 'Dashboard' },
    { href: '/architect', label: 'Swarm Architect' },
    { href: '/matrix', label: 'Cognitive Matrix' },
    { href: '/wiring', label: 'Neural Wiring' },
    { href: '/teams', label: 'Team Management' },
    { href: '/catalogue', label: 'Agent Catalogue' },
    { href: '/marketplace', label: 'Skills Market' },
    { href: '/memory', label: 'Memory' },
    { href: '/telemetry', label: 'System Status' },
    { href: '/approvals', label: 'Governance' },
    { href: '/settings', label: 'Settings' },
];

test.describe('Panel D — ZoneA Rail Navigation', () => {

    test.beforeEach(async ({ page }) => {
        await page.goto('/');
        await page.waitForLoadState('networkidle');
    });

    test('ZoneA rail is visible on page load', async ({ page }) => {
        // Rail has the Mycelis logo and brand name
        await expect(page.locator('text=Mycelis')).toBeVisible();
    });

    test('all navigation entries are present', async ({ page }) => {
        for (const entry of NAV_ENTRIES) {
            const link = page.locator(`a[href="${entry.href}"]`);
            await expect(link).toBeVisible();
            // Label visible at md+ breakpoint
            await expect(link.locator(`text=${entry.label}`)).toBeVisible();
        }
    });

    test('active route gets primary highlight', async ({ page }) => {
        // Home route should be active on /
        const homeLink = page.locator('a[href="/"]');
        await expect(homeLink).toHaveClass(/bg-cortex-primary/);
        await expect(homeLink).toHaveClass(/text-white/);
    });

    test('navigating to /teams highlights Team Management', async ({ page }) => {
        const teamsLink = page.locator('a[href="/teams"]');
        await teamsLink.click();
        await page.waitForLoadState('networkidle');

        // Teams link should now be active
        await expect(teamsLink).toHaveClass(/bg-cortex-primary/);
        await expect(teamsLink).toHaveClass(/text-white/);

        // Page content loads
        await expect(page.locator('text=Team Management')).toBeVisible();
    });

    test('navigating to /wiring highlights Neural Wiring', async ({ page }) => {
        const wiringLink = page.locator('a[href="/wiring"]');
        await wiringLink.click();
        await page.waitForLoadState('networkidle');

        await expect(wiringLink).toHaveClass(/bg-cortex-primary/);
    });

    test('navigating to /catalogue highlights Agent Catalogue', async ({ page }) => {
        const catalogueLink = page.locator('a[href="/catalogue"]');
        await catalogueLink.click();
        await page.waitForLoadState('networkidle');

        await expect(catalogueLink).toHaveClass(/bg-cortex-primary/);
    });

    test('inactive nav items have muted text', async ({ page }) => {
        // Non-active links should have muted styling
        const wiringLink = page.locator('a[href="/wiring"]');
        await expect(wiringLink).toHaveClass(/text-cortex-text-muted/);
    });

    test('nav order matches design spec', async ({ page }) => {
        // All nav links in the sidebar
        const navLinks = page.locator('a[href^="/"], a[href="/"]');

        // Collect hrefs in order
        const hrefs: string[] = [];
        const count = await navLinks.count();
        for (let i = 0; i < count; i++) {
            const href = await navLinks.nth(i).getAttribute('href');
            if (href) hrefs.push(href);
        }

        // Verify /teams comes after /wiring and before /catalogue
        const wiringIdx = hrefs.indexOf('/wiring');
        const teamsIdx = hrefs.indexOf('/teams');
        const catalogueIdx = hrefs.indexOf('/catalogue');

        expect(wiringIdx).toBeGreaterThan(-1);
        expect(teamsIdx).toBeGreaterThan(-1);
        expect(catalogueIdx).toBeGreaterThan(-1);
        expect(teamsIdx).toBe(wiringIdx + 1);
        expect(catalogueIdx).toBe(teamsIdx + 1);
    });

    test('no bg-white leak in navigation rail', async ({ page }) => {
        const railHtml = await page.locator('nav, [class*="flex-col"]').first().innerHTML();
        expect(railHtml).not.toContain('bg-white');
    });
});
