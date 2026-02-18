import { test, expect } from '@playwright/test';

test.describe('Phase 5.4 - Command Deck Proposals', () => {

    test('Manifest Team interaction triggers proposal workflow', async ({ page }) => {
        // 1. Navigate to Command Deck (Mission Control)
        await page.goto('http://localhost:3000/');

        // 2. Locate "Manifest Team" or "New Proposal" Action
        // Looking for a button that initiates the creation flow
        const manifestBtn = page.getByRole('button', { name: /Manifest Team|New Proposal/i });
        await expect(manifestBtn).toBeVisible();
        await manifestBtn.click();

        // 3. Assert: Modal/Form Appearance
        const modal = page.locator('[role="dialog"]');
        await expect(modal).toBeVisible();

        // 4. Fill Proposal Form
        await modal.getByPlaceholder(/Team Name|Proposal Title/i).fill('Topological Audit Team');
        await modal.getByPlaceholder(/Intent|Description/i).fill('Audit the node topology for disconnected graphs.');

        // 5. Submit Proposal
        // Listen for the API call that should happen on submit
        const proposalRequestPromise = page.waitForResponse(response =>
            response.url().includes('/api/v1/proposals') && response.request().method() === 'POST'
        );

        await modal.getByRole('button', { name: /Submit|Manifest/i }).click();

        const response = await proposalRequestPromise;
        expect(response.status()).toBe(201); // Created

        // 6. Assert: UI Feedback
        // The modal should close or a success toast should appear
        await expect(modal).not.toBeVisible();
        await expect(page.locator('text=Proposal Created')).toBeVisible(); // Optional, depends on UI implementation
    });

});
