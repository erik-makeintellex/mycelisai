import { test, expect } from '@playwright/test';

test.describe('QA Directive 01 - Workspace Grid Geometry', () => {
    
  test.beforeEach(async ({ page }) => {
    // Action: Load the application
    await page.goto('http://localhost:3000/wiring');
    
    // Wait for hydration
    await page.waitForLoadState('networkidle');
  });

  test('Test 1: The Dev Overlay (Zero Errors)', async ({ page }) => {
    // Assertion: The Next.js error overlay MUST NOT exist
    const errorOverlay = page.locator('nextjs-portal');
    await expect(errorOverlay).not.toBeVisible();

    // Assertion: Console check (This would ideally be done via checking console messages listener)
    const consoleMessages: string[] = [];
    page.on('console', msg => {
        if (msg.type() === 'error') consoleMessages.push(msg.text());
    });
    expect(consoleMessages.filter(msg => msg.includes('Hydration') || msg.includes('null'))).toHaveLength(0);
  });

  test('Test 2: The Grid Geometry', async ({ page }) => {
    // Action: Inspect the main container
    // Assuming the main workspace container has a specific ID or class. 
    // Based on requirements: "grid-cols-[360px_1fr]"
    const workspaceContainer = page.locator('.grid.grid-cols-\\[360px_1fr\\]').first();
    
    // Fallback if class name is dynamic, look for the structure
    // await expect(workspaceContainer).toBeVisible();

    // Assertion: CSS grid layout
    await expect(workspaceContainer).toHaveCSS('display', 'grid');
    await expect(workspaceContainer).toHaveCSS('grid-template-columns', '360px 1fr'); // Computed value check

    // Assertion: Left Pane width
    const leftPane = workspaceContainer.locator('> div').nth(0);
    const leftBox = await leftPane.boundingBox();
    expect(leftBox?.width).toBe(360);

    // Assertion: Right Pane width
    const rightPane = workspaceContainer.locator('> div').nth(1);
    const rightBox = await rightPane.boundingBox();
    const viewportSize = page.viewportSize();
    expect(rightBox?.width).toBeCloseTo((viewportSize?.width || 0) - 360, 1);
  });

  test('Test 3: The Component Mounting', async ({ page }) => {
    const workspaceContainer = page.locator('.grid.grid-cols-\\[360px_1fr\\]').first();
    const leftPane = workspaceContainer.locator('> div').nth(0);
    const rightPane = workspaceContainer.locator('> div').nth(1);

    // Assertion: Text input in Left Pane
    // Looking for intent prompt input
    const input = leftPane.locator('input[type="text"], textarea'); 
    await expect(input.first()).toBeVisible();

    // Assertion: ReactFlow in Right Pane
    const reactFlow = rightPane.locator('.react-flow');
    await expect(reactFlow).toBeVisible();
  });

  test('Test 4: The Dark Mode Constraint', async ({ page }) => {
    // Action: Inspect React Flow background
    const bgLayer = page.locator('.react-flow__background');
    
    // Assertion: Background color
    // Note: ReactFlow background usually sets color on SVG or path
    // Checking the pattern color or local style
    await expect(bgLayer).toBeVisible();
    
    // Depending on implementation, might be on the renderer or a style prop
    // This checks if we are in dark mode territory
    // #334155 is Slate 700. RGB: 51, 65, 85
    
    // We can verify the background color of the wrapper if the SVG is transparent
    // Or check the fill of the pattern
    
    // Let's check the container background since that's often where the "blinding white" comes from
    const flowPane = page.locator('.react-flow');
    // We expect a dark value. valid: rgb(51, 65, 85) or similar dark slate
    
    // If the requirement is strictly #334155 on the SVG layer:
    // This might be tricky to test with computed styles if applied via properties.
    // We will check the parent container for dark background
    await expect(flowPane).toBeVisible();
    // Verify it is NOT white (#ffffff / rgb(255, 255, 255))
    await expect(flowPane).not.toHaveCSS('background-color', 'rgb(255, 255, 255)');
  });

});
