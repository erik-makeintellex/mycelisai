import { expect, test } from '@playwright/test';

test.describe('Live documentation authority', () => {
    test('serves the canonical PRD and current acceptance guide from the real manifest', async ({ page }) => {
        await page.goto('/docs?doc=mycelis-canonical-prd', { waitUntil: 'domcontentloaded' });

        const canonicalHeading = page.getByRole('heading', { name: 'Mycelis Canonical PRD' });
        if (!await canonicalHeading.isVisible({ timeout: 5_000 }).catch(() => false)) {
            await page.reload({ waitUntil: 'domcontentloaded' });
        }
        await expect(canonicalHeading).toBeVisible({ timeout: 20_000 });
        await expect(page.getByText('Mycelis Canonical PRD').first()).toBeVisible();
        await expect(page.getByText('Architecture Overview')).toHaveCount(0);

        await page.getByText('User Acceptance', { exact: true }).click();
        await expect(page).toHaveURL(/\/docs\?doc=user-acceptance$/);
        await expect(page.getByRole('heading', { name: 'User Acceptance Runbook' })).toBeVisible();
        await expect(page.getByRole('heading', { name: 'Trusted Outcome Walkthrough' })).toBeVisible();
    });

    for (const viewport of [
        { name: 'desktop', width: 1366, height: 768 },
        { name: 'compact', width: 390, height: 844 },
    ]) {
        test(`${viewport.name} navigates to the bounded G4/E10 contract`, async ({ page }, testInfo) => {
            const errors: string[] = [];
            page.on('pageerror', (error) => errors.push(error.message));
            page.on('console', (message) => {
                if (message.type() === 'error') errors.push(message.text());
            });
            page.on('response', (response) => {
                if (response.status() >= 500) errors.push(`${response.status()} ${response.url()}`);
            });

            await page.setViewportSize({ width: viewport.width, height: viewport.height });
            await page.goto('/docs?doc=mycelis-canonical-prd', { waitUntil: 'domcontentloaded' });
            if (viewport.name === 'compact') {
                await page.getByRole('button', { name: 'All docs' }).click();
            }
            await page.getByText('G4/E10 Invocation Contract', { exact: true }).click();
            await expect(page).toHaveURL(/\/docs\?doc=g4-e10-invocation$/);
            await testInfo.attach(`g4-e10-${viewport.name}-target`, {
                body: JSON.stringify({ url: page.url() }),
                contentType: 'application/json',
            });
            await expect(page.getByRole('heading', { name: 'G4/E10 — Minimum Durable Invocation Acceptance Contract' })).toBeVisible();
            const article = page.getByTestId('docs-article-pane');
            await expect(article).toContainText('Only the counting capability is admitted by this packet.');
            await expect(article).toContainText('outside this certification');
            await expect(article).toContainText('unknown_effect');
            const overflow = await page.evaluate(() =>
                Math.max(document.documentElement.scrollWidth, document.body.scrollWidth) - window.innerWidth,
            );
            expect(overflow).toBeLessThanOrEqual(4);
            await testInfo.attach(`g4-e10-${viewport.name}`, {
                body: await page.screenshot({ fullPage: true }),
                contentType: 'image/png',
            });
            const boundedScope = article.getByText('Only the counting capability is admitted by this packet.', { exact: false });
            await boundedScope.scrollIntoViewIfNeeded();
            await expect(boundedScope).toBeVisible();
            await testInfo.attach(`g4-e10-${viewport.name}-scope`, {
                body: await page.screenshot(),
                contentType: 'image/png',
            });
            expect(errors).toEqual([]);
        });
    }
});
