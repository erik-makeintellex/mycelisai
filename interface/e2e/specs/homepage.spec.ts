import { expect, test } from '@playwright/test';

test.describe('Authenticated front door', () => {
    test('front page requires login for a fresh browser', async ({ browser }, testInfo) => {
        const context = await browser.newContext({
            baseURL: String(testInfo.project.use.baseURL),
            storageState: { cookies: [], origins: [] },
        });
        const page = await context.newPage();
        await page.goto('/');
        await expect(page).toHaveURL(/\/login\?next=%2F$/);
        await expect(page.getByRole('heading', { name: /Sign in to operate Mycelis/i })).toBeVisible();
        await expect(page.getByText(/Personal Gmail accounts are rejected/i)).toBeVisible();
        await context.close();
    });

    test('local owner login enters the Soma workflow', async ({ browser }, testInfo) => {
        const context = await browser.newContext({
            baseURL: String(testInfo.project.use.baseURL),
            storageState: { cookies: [], origins: [] },
        });
        const page = await context.newPage();
        await page.goto('/login');
        await page.getByLabel(/Local admin username/i).fill(process.env.MYCELIS_LOCAL_ADMIN_USERNAME || 'admin');
        await page.getByLabel(/Password or local API key/i).fill(process.env.MYCELIS_LOCAL_ADMIN_PASSWORD || process.env.MYCELIS_API_KEY || 'playwright-admin');
        await page.getByRole('button', { name: /Sign in as local admin/i }).click();
        await expect(page).toHaveURL(/\/dashboard$/);
        await expect(page.getByRole('heading', { name: /What do you want Soma to do/i })).toBeVisible();
        await expect(page.getByTestId('soma-environment-entry')).toBeVisible();
        await expect(page.getByText('Signed in').first()).toBeVisible();
        await expect(page.getByText(/Soma environment/i).first()).toBeVisible();
        await context.close();
    });

    test('authenticated root redirects into Soma instead of public marketing', async ({ page }) => {
        await page.goto('/');
        await expect(page).toHaveURL(/\/dashboard$/);
        await expect(page.getByRole('heading', { name: /What do you want Soma to do/i })).toBeVisible();
        await expect(page.getByTestId('soma-environment-entry')).toBeVisible();
    });

    test('dashboard first paint honors the stored workspace theme', async ({ page }) => {
        await page.addInitScript(() => {
            window.localStorage.setItem('mycelis-user-settings', JSON.stringify({
                assistantName: 'Soma',
                theme: 'midnight-cortex',
            }));
        });
        await page.route('**/api/v1/user/settings', async (route) => {
            await route.fulfill({
                json: {
                    ok: true,
                    data: {
                        assistant_name: 'Soma',
                        theme: 'midnight-cortex',
                    },
                },
            });
        });

        await page.goto('/dashboard', { waitUntil: 'domcontentloaded' });

        await expect(page.locator('html')).toHaveAttribute('data-theme', 'midnight-cortex');
        await expect.poll(async () => page.evaluate(() =>
            window.getComputedStyle(document.documentElement).getPropertyValue('--color-cortex-bg').trim(),
        )).toBe('#0f1116');
    });

    test('dashboard first paint defaults to the dark workspace theme without hydration noise', async ({ page }) => {
        const hydrationMessages: string[] = [];
        page.on('console', (message) => {
            if (!['error', 'warning'].includes(message.type())) return;
            const text = message.text();
            if (/hydration|hydrating|did not match|text content does not match/i.test(text)) {
                hydrationMessages.push(text);
            }
        });

        await page.addInitScript(() => {
            window.localStorage.removeItem('mycelis-user-settings');
            window.localStorage.setItem('mycelis-advanced-mode', 'true');
        });
        await page.route('**/api/v1/user/settings', async (route) => {
            await route.fulfill({
                json: {
                    ok: true,
                    data: {
                        assistant_name: 'Soma',
                        theme: 'midnight-cortex',
                    },
                },
            });
        });

        await page.goto('/dashboard', { waitUntil: 'domcontentloaded' });

        await expect(page.locator('html')).toHaveAttribute('data-theme', 'midnight-cortex');
        await expect(page.getByRole('button', { name: /Advanced: On/i })).toBeVisible();
        await expect.poll(async () => page.evaluate(() =>
            window.getComputedStyle(document.documentElement).getPropertyValue('--color-cortex-bg').trim(),
        )).toBe('#0f1116');
        await expect(page.getByRole('heading', { name: /What do you want Soma to do/i })).toBeVisible();
        expect(hydrationMessages).toEqual([]);
    });
});
