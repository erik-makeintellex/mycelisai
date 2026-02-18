import { test, expect } from '@playwright/test';

test.describe('Panel E — Wiring Agent Editor & Two-Click Safety', () => {

    test.beforeEach(async ({ page }) => {
        await page.goto('/wiring');
        await page.waitForLoadState('networkidle');
    });

    test('wiring page loads with ReactFlow canvas', async ({ page }) => {
        // No error overlay
        const errorOverlay = page.locator('nextjs-portal');
        await expect(errorOverlay).not.toBeVisible();

        // ReactFlow is mounted
        const reactFlow = page.locator('.react-flow');
        await expect(reactFlow).toBeVisible();
    });

    test('ReactFlow canvas has dark background (no white)', async ({ page }) => {
        const flowPane = page.locator('.react-flow');
        await expect(flowPane).toBeVisible();
        await expect(flowPane).not.toHaveCSS('background-color', 'rgb(255, 255, 255)');
    });

    test('clicking an agent node opens the editor drawer', async ({ page }) => {
        // Wait for nodes to render
        const agentNodes = page.locator('.react-flow__node');
        const nodeCount = await agentNodes.count();
        if (nodeCount === 0) {
            test.skip();
            return;
        }

        // Click first agent node
        await agentNodes.first().click();

        // Editor drawer should appear (w-96 right panel)
        const drawer = page.locator('.w-96.bg-cortex-surface');
        await expect(drawer).toBeVisible({ timeout: 5000 });

        // Verify drawer header
        await expect(drawer.locator('text=Edit Agent')).toBeVisible();
    });

    test('editor drawer shows agent form fields', async ({ page }) => {
        const agentNodes = page.locator('.react-flow__node');
        if ((await agentNodes.count()) === 0) {
            test.skip();
            return;
        }

        await agentNodes.first().click();
        const drawer = page.locator('.w-96.bg-cortex-surface');
        await expect(drawer).toBeVisible();

        // Agent ID input
        const agentIdInput = drawer.locator('input[placeholder="agent-name"]');
        await expect(agentIdInput).toBeVisible();

        // Role select
        const roleSelect = drawer.locator('select');
        await expect(roleSelect).toBeVisible();

        // Role options
        await expect(roleSelect.locator('option:has-text("cognitive")')).toBeVisible();
        await expect(roleSelect.locator('option:has-text("sensory")')).toBeVisible();
        await expect(roleSelect.locator('option:has-text("actuation")')).toBeVisible();
        await expect(roleSelect.locator('option:has-text("ledger")')).toBeVisible();

        // System prompt textarea
        const promptArea = drawer.locator('textarea');
        await expect(promptArea).toBeVisible();
    });

    test('save button is disabled without agent ID', async ({ page }) => {
        const agentNodes = page.locator('.react-flow__node');
        if ((await agentNodes.count()) === 0) {
            test.skip();
            return;
        }

        await agentNodes.first().click();
        const drawer = page.locator('.w-96.bg-cortex-surface');
        await expect(drawer).toBeVisible();

        // Clear the agent ID input
        const agentIdInput = drawer.locator('input[placeholder="agent-name"]');
        await agentIdInput.clear();

        // Save button should be disabled
        const saveBtn = drawer.locator('button:has-text("Save")');
        await expect(saveBtn).toBeDisabled();
    });

    test('cancel button closes the editor drawer', async ({ page }) => {
        const agentNodes = page.locator('.react-flow__node');
        if ((await agentNodes.count()) === 0) {
            test.skip();
            return;
        }

        await agentNodes.first().click();
        const drawer = page.locator('.w-96.bg-cortex-surface');
        await expect(drawer).toBeVisible();

        // Click cancel
        const cancelBtn = drawer.locator('button:has-text("Cancel")');
        await cancelBtn.click();

        // Drawer should close
        await expect(drawer).not.toBeVisible();
    });

    test('delete button requires two clicks (safety pattern)', async ({ page }) => {
        const agentNodes = page.locator('.react-flow__node');
        if ((await agentNodes.count()) === 0) {
            test.skip();
            return;
        }

        await agentNodes.first().click();
        const drawer = page.locator('.w-96.bg-cortex-surface');
        await expect(drawer).toBeVisible();

        // Find the delete button (footer area, has Trash2 icon)
        // Delete button is the first in the footer row
        const footer = drawer.locator('.border-t.border-cortex-border').last();
        const deleteBtn = footer.locator('button').first();

        // First click — should show "Confirm?" text and pulse
        await deleteBtn.click();
        await expect(deleteBtn).toContainText('Confirm?');
        await expect(deleteBtn).toHaveClass(/animate-pulse/);

        // Wait for 3s timeout — button should revert
        await page.waitForTimeout(3500);
        await expect(deleteBtn).not.toContainText('Confirm?');
        await expect(deleteBtn).not.toHaveClass(/animate-pulse/);
    });

    test('active agent has disabled ID field', async ({ page }) => {
        // Navigate to wiring page — if there's an active mission with agents,
        // their ID fields should be disabled
        const agentNodes = page.locator('.react-flow__node');
        if ((await agentNodes.count()) === 0) {
            test.skip();
            return;
        }

        await agentNodes.first().click();
        const drawer = page.locator('.w-96.bg-cortex-surface');
        await expect(drawer).toBeVisible();

        // Check if "Active" badge is visible
        const activeBadge = drawer.locator('text=Active');
        if (await activeBadge.isVisible()) {
            // ID field should be disabled for active agents
            const agentIdInput = drawer.locator('input[placeholder="agent-name"]');
            await expect(agentIdInput).toBeDisabled();
        }
    });

    test('tag input adds and removes chips', async ({ page }) => {
        const agentNodes = page.locator('.react-flow__node');
        if ((await agentNodes.count()) === 0) {
            test.skip();
            return;
        }

        await agentNodes.first().click();
        const drawer = page.locator('.w-96.bg-cortex-surface');
        await expect(drawer).toBeVisible();

        // Find a tag input (Tools, Inputs, or Outputs)
        const tagInputs = drawer.locator('input[placeholder*="Type + Enter"]');
        if ((await tagInputs.count()) === 0) return;

        const tagInput = tagInputs.first();

        // Type a tag and press Enter
        await tagInput.fill('test-tool');
        await tagInput.press('Enter');

        // Tag chip should appear
        await expect(drawer.locator('text=test-tool')).toBeVisible();

        // Click the X on the chip to remove it
        const chip = drawer.locator('span').filter({ hasText: 'test-tool' });
        const removeBtn = chip.locator('button, svg').first();
        await removeBtn.click();

        // Tag should be gone
        await expect(drawer.locator('text=test-tool')).not.toBeVisible();
    });
});

test.describe('Panel E — Circuit Board Mission Actions', () => {

    test.beforeEach(async ({ page }) => {
        await page.goto('/wiring');
        await page.waitForLoadState('networkidle');
    });

    test('mission top bar shows metadata when blueprint loaded', async ({ page }) => {
        // The top bar only appears when a blueprint is present
        // Check if it's there — if not, we're in empty state
        const topBar = page.locator('text=/\\d+ team/');
        const visible = await topBar.isVisible().catch(() => false);
        if (!visible) {
            test.skip();
            return;
        }

        // Mission ID shown
        await expect(page.locator('text=/mission-/')).toBeVisible();

        // Team and agent counts
        await expect(page.locator('text=/\\d+ team/')).toBeVisible();
        await expect(page.locator('text=/\\d+ agent/')).toBeVisible();
    });

    test('draft mode shows Discard button (single click)', async ({ page }) => {
        const discardBtn = page.locator('button:has-text("Discard")');
        const visible = await discardBtn.isVisible().catch(() => false);
        if (!visible) {
            test.skip();
            return;
        }

        // Draft status badge
        await expect(page.locator('text=DRAFT')).toBeVisible();

        // Discard is single-click (no confirm needed)
        // We just verify the button exists — clicking would clear the draft
    });

    test('active mode shows Terminate button with two-click safety', async ({ page }) => {
        const terminateBtn = page.locator('button:has-text("Terminate")');
        const visible = await terminateBtn.isVisible().catch(() => false);
        if (!visible) {
            test.skip();
            return;
        }

        // First click — confirm state
        await terminateBtn.click();
        await expect(page.locator('button:has-text("Confirm Terminate?")')).toBeVisible();
        await expect(page.locator('button:has-text("Confirm Terminate?")')).toHaveClass(/animate-pulse/);

        // Wait for auto-revert (3s)
        await page.waitForTimeout(3500);
        await expect(page.locator('button:has-text("Terminate")')).toBeVisible();
        await expect(page.locator('button:has-text("Confirm Terminate?")')).not.toBeVisible();
    });
});
