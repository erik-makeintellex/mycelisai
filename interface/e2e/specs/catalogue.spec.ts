import { test, expect } from '@playwright/test';

const copiedProfileName = /^Research Specialist custom-[0-9a-f]{5}$/;

async function removeInterruptedTestCopies(page: import('@playwright/test').Page) {
    const response = await page.request.get('/api/v1/catalogue/agents');
    expect(response.ok()).toBeTruthy();
    const profiles = (await response.json()) as Array<{ id: string; name: string; source?: string }>;
    for (const profile of profiles) {
        if (profile.source === 'user' && copiedProfileName.test(profile.name)) {
            const deletion = await page.request.delete(`/api/v1/catalogue/agents/${profile.id}`);
            expect(deletion.ok()).toBeTruthy();
        }
    }
}

test.describe('Worker Profiles Page (/catalogue)', () => {
    test.skip(({ browserName }) => browserName === 'firefox', 'Firefox catalogue coverage is currently unstable; keep Chromium as the baseline browser for now.');

    test.beforeEach(async ({ page }) => {
        await page.goto('/catalogue', { waitUntil: 'domcontentloaded' });
        await expect(page).toHaveURL(/\/resources\?tab=roles$/);
    });

    test('page loads without errors', async ({ page }) => {
        await expect(page.getByRole('heading', { name: 'Resources' })).toBeVisible();
        await expect(page.getByText('Worker profiles', { exact: true })).toBeVisible();
        const errorOverlay = page.locator('nextjs-portal');
        await expect(errorOverlay).not.toBeVisible();
    });

    test('worker profile library exposes its current controls', async ({ page }) => {
        await expect(page.getByRole('tab', { name: /Ready-made/ })).toBeVisible();
        await expect(page.getByRole('tab', { name: /My profiles/ })).toBeVisible();
        await expect(page.getByRole('button', { name: 'New profile' })).toBeVisible();
    });

    test('create agent button is visible', async ({ page }) => {
        const createButton = page.getByRole('button', { name: 'New profile' });
        await expect(createButton).toBeVisible();
        await createButton.click();
        await expect(page.getByText('New worker profile', { exact: true }).last()).toBeVisible();
        await expect(page.getByRole('button', { name: 'Create profile' })).toBeVisible();
    });

    test('ready-made profile can be inspected, copied, edited, and cleaned up', async ({ page }) => {
        await removeInterruptedTestCopies(page);
        await page.reload({ waitUntil: 'domcontentloaded' });
        await expect(page.getByText('Research Specialist', { exact: true })).toBeVisible();
        await page.getByText('Research Specialist', { exact: true }).click();
        const dialog = page.getByRole('dialog', { name: 'Research Specialist profile' });
        await expect(dialog).toBeVisible();
        await expect(dialog.getByRole('button', { name: 'Copy profile' })).toBeVisible();

        await dialog.getByRole('tab', { name: 'Access & context' }).click();
        await expect(dialog.getByText('web_search', { exact: true })).toBeVisible();
        await expect(dialog.getByLabel('Context type 1')).toHaveValue('public_web');
        await expect(dialog.getByLabel('Context access 1')).toHaveValue('search');

        const createResponse = page.waitForResponse((response) => response.url().includes('/api/v1/catalogue/agents') && response.request().method() === 'POST');
        await dialog.getByRole('button', { name: 'Copy profile' }).click();
        expect((await createResponse).status()).toBe(201);
        const copiedDialog = page.getByRole('dialog', { name: /Research Specialist custom-[0-9a-f]{5} profile/ });
        await expect(copiedDialog.getByRole('button', { name: 'Save changes' })).toBeVisible();
        const copiedName = (await copiedDialog.getByRole('heading').textContent())?.replace(' profile', '') ?? '';
        await copiedDialog.getByRole('button', { name: 'Cancel' }).click();

        await expect(page.getByRole('tab', { name: /My profiles/ })).toHaveAttribute('aria-selected', 'true');
        const copyCard = page.getByText(copiedName, { exact: true }).first();
        await expect(copyCard).toBeVisible();

        page.once('dialog', (confirmation) => confirmation.accept());
        await copyCard.locator('xpath=ancestor::article').hover();
        const deleteResponse = page.waitForResponse((response) => response.url().includes('/api/v1/catalogue/agents/') && response.request().method() === 'DELETE');
        await page.getByRole('button', { name: `Delete ${copiedName}` }).click();
        expect((await deleteResponse).status()).toBe(200);
        await expect(page.getByText(copiedName ?? '', { exact: true })).not.toBeVisible();
    });

    test('no bg-white leak on catalogue page', async ({ page }) => {
        const body = await page.content();
        expect(body).not.toContain('bg-white');
    });
});
