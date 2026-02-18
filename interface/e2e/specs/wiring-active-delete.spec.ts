
import { test, expect } from '@playwright/test';

test.describe('Wiring Editor - Agent Deletion', () => {
    test.beforeEach(async ({ page }) => {
        // 1. Navigate to Wiring
        await page.goto('/wiring');

        // 2. Negotiate Intent (Force Draft -> Active transition)
        // Note: Using a unique intent to avoid caching issues if any
        await page.getByPlaceholder('Describe your mission intent...').fill('Debug active delete via playwright ' + Date.now());
        await page.getByRole('button', { name: 'Send' }).click();

        // 3. Wait for Blueprint & Instantiation
        // The previous subagent tests showed this auto-transitions to ACTIVE.
        // We wait for the "ACTIVE" badge to appear.
        await expect(page.getByText(/ACTIVE -/)).toBeVisible({ timeout: 20000 });
    });

    test('should delete an agent in ACTIVE mode (Regression Test H2)', async ({ page }) => {
        // 4. Open Agent Editor (Click first agent node)
        // We target the node by its data-testid or class if available, otherwise by text.
        // Agent nodes usually have text like "agent-01" or a specific role.
        // Let's assume the first agent is reachable via class or text.
        const firstAgent = page.locator('.react-flow__node-agentNode').first();
        await firstAgent.click();

        // 5. Verify Editor Opens
        const deleteBtn = page.locator('button.bg-red-500'); // The trash icon/confirm button
        await expect(deleteBtn).toBeVisible();

        // 6. Click Delete (First Click -> "Confirm?")
        await deleteBtn.click();
        await expect(deleteBtn).toHaveText('Confirm?');

        // 7. Click Confirm (Second Click -> Delete Action)
        await deleteBtn.click();

        // 8. Verify Result
        // The drawer should close
        await expect(deleteBtn).not.toBeVisible();

        // CRITICAL: The agent node should be removed from the canvas.
        // This is where the manual test failed.
        await expect(firstAgent).not.toBeVisible();
    });
});
