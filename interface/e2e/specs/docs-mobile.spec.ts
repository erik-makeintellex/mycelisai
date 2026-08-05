import { expect, test } from '@playwright/test';

test.describe('Mobile Docs', () => {
    test('uses a readable list-to-article flow without nested overflow', async ({ page }) => {
        await page.setViewportSize({ width: 390, height: 844 });
        await page.route('**/docs-api', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    sections: [
                        {
                            section: 'Start here',
                            docs: [
                                {
                                    slug: 'mobile-start',
                                    label: 'Getting Started',
                                    path: 'docs/user/getting-started.md',
                                    description: 'Learn the main workspace flow.',
                                },
                                {
                                    slug: 'mobile-soma',
                                    label: 'Working With Soma',
                                    path: 'docs/user/working-with-soma.md',
                                    description: 'Ask, review, and continue work.',
                                },
                            ],
                        },
                    ],
                }),
            });
        });
        await page.route('**/docs-api/mobile-start', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    slug: 'mobile-start',
                    label: 'Getting Started',
                    content: '# Getting Started\n\nChoose a guide from the documentation list.',
                }),
            });
        });
        await page.route('**/docs-api/mobile-soma', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    slug: 'mobile-soma',
                    label: 'Working With Soma',
                    content: '# Working With Soma\n\nA readable article with `workspace/generated/a-very-long-deliverable-name-that-must-wrap.html`.\n\n| Outcome | Deliverable | Verification |\n| --- | --- | --- |\n| A long running project | A retained application package | Proof remains linked to the outcome |\n\n```text\nthis-intentionally-wide-code-line-scrolls-locally-without-widening-the-article-pane\n```',
                }),
            });
        });

        await page.goto('/docs', { waitUntil: 'domcontentloaded' });

        const navigation = page.getByTestId('docs-navigation-pane');
        const article = page.getByTestId('docs-article-pane');
        await expect(navigation).toBeVisible();
        await expect(article).toBeHidden();
        await page.getByRole('button', { name: 'Working With Soma' }).click();
        await expect(article).toBeVisible();
        await expect(navigation).toBeHidden();
        await expect(page.getByRole('button', { name: 'All docs' })).toBeVisible();
        await expect(page.getByRole('heading', { name: 'Working With Soma' })).toBeVisible();

        const overflow = await article.evaluate((pane) => {
            const offenders = Array.from(pane.querySelectorAll<HTMLElement>('*'))
                .filter((node) => node.clientWidth > 0 && node.scrollWidth > node.clientWidth + 1)
                .filter((node) => !node.closest('pre, [data-testid="docs-table-scroll"]'))
                .map((node) => ({ tag: node.tagName, testid: node.dataset.testid, className: node.className }));
            return {
                paneOverflow: pane.scrollWidth - pane.clientWidth,
                offenders,
            };
        });
        expect(overflow.paneOverflow).toBeLessThanOrEqual(1);
        expect(overflow.offenders).toEqual([]);

        await page.getByRole('button', { name: 'All docs' }).click();
        await expect(navigation).toBeVisible();
        await expect(article).toBeHidden();
        const navigationOverflow = await navigation.evaluate((pane) => pane.scrollWidth - pane.clientWidth);
        expect(navigationOverflow).toBeLessThanOrEqual(1);
    });
});
