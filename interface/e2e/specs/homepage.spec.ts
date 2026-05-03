import { expect, test } from '@playwright/test';

test.describe('Templated homepage', () => {
    test('homepage loads and primary CTA enters the Soma workflow', async ({ page }) => {
        await page.goto('/');
        await expect(page.getByRole('heading', { name: /Operate AI Organizations through Soma/i })).toBeVisible();
        await expect(page.getByText('Example workspace preview')).toBeVisible();

        await page.getByRole('link', { name: /Start with Soma/i }).first().click();
        await expect(page).toHaveURL(/\/dashboard$/);
        await expect(page.getByRole('heading', { name: /What do you want Soma to do/i })).toBeVisible();
    });

    test('custom homepage links render without breaking routing', async ({ page }) => {
        await page.route('**/api/v1/homepage', async (route) => {
            await route.fulfill({
                contentType: 'application/json',
                body: JSON.stringify({
                    ok: true,
                    data: {
                        enabled: true,
                        brand: { product_name: 'Acme AI', tagline: 'Internal orchestration' },
                        hero: {
                            headline: 'Coordinate work through Soma',
                            subheadline: 'Run governed execution with connected tools.',
                            primary_cta: { label: 'Open Soma', href: '/dashboard' },
                            secondary_cta: { label: 'Docs', href: '/docs' },
                        },
                        sections: [{ title: 'Express intent', body: 'Tell Soma what needs to happen.' }],
                        links: [{ label: 'Support', href: 'https://support.example.com', description: 'Contact the platform team.', external: true }],
                    },
                }),
            });
        });

        await page.goto('/');
        await expect(page.getByRole('heading', { name: 'Coordinate work through Soma' })).toBeVisible();
        await expect(page.getByRole('link', { name: /Support/i })).toHaveAttribute('href', 'https://support.example.com');
    });
});
