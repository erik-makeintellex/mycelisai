import { expect, test } from "@playwright/test";
import {
  enableAdvancedMode,
  expectNoHorizontalOverflow,
  mockTeamsWorkspace,
} from "../support/finalization-proof";

test.describe.configure({ mode: "serial" });

test.describe("Desktop/mobile compression proof", () => {
  test.beforeEach(async ({ page }) => {
    await enableAdvancedMode(page);
    await mockTeamsWorkspace(page);
  });

  test("Teams active work remains scannable on desktop", async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.goto("/teams", { waitUntil: "domcontentloaded" });

    await expect(page.getByRole("heading", { name: "Team Lead Workspaces" })).toBeVisible({ timeout: 20_000 });
    const activeLane = page.getByTestId("active-work-lane");
    const firstDemoRow = activeLane.locator("article").filter({ hasText: "First Demo Game Team" });
    await expect(activeLane).toBeVisible();
    await expect(page.getByText("First Demo Game Team").first()).toBeVisible();
    await expect(firstDemoRow.getByText("Output ready", { exact: true })).toBeVisible();
    await expect(firstDemoRow.getByText("Durable team work").first()).toBeVisible();
    await expect(firstDemoRow.getByText("Projection fallback")).toHaveCount(0);
    await expect(firstDemoRow.getByRole("link", { name: /Run proof/i })).toHaveAttribute("href", /\/runs\/run-first-demo/);
    await expect(firstDemoRow.getByRole("link", { name: /Coin Runner package/i })).toHaveAttribute(
      "href",
      "/api/v1/workspace/files/view?path=generated%2Fcoin-runner%2Findex.html",
    );
    await expect(firstDemoRow.getByText("1 package retained")).toBeVisible();
    await expect(firstDemoRow.getByText("Proof available")).toBeVisible();
    await expect(firstDemoRow.getByRole("button", { name: /Ask team/i })).toBeVisible();

    const recoveryRow = activeLane.locator("article").filter({ hasText: "Recover failed package proof" });
    await expect(recoveryRow).toBeVisible();
    await expect(recoveryRow.getByText("Degraded", { exact: true })).toBeVisible();
    await expect(recoveryRow.getByText("Needs recovery", { exact: true })).toBeVisible();
    await expect(recoveryRow.getByText("Recovery: Retry with retained run context")).toBeVisible();
    await expect(recoveryRow.getByRole("button", { name: /Recover/i })).toBeVisible();
    await expectNoHorizontalOverflow(page);
  });

  test("Teams active work compresses without horizontal overflow on mobile", async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    await page.goto("/teams", { waitUntil: "domcontentloaded" });

    await expect(page.getByRole("heading", { name: "Team Lead Workspaces" })).toBeVisible({ timeout: 20_000 });
    await expect(page.getByTestId("active-work-lane")).toBeVisible();
    await expect(page.getByText("First Demo Game Team").first()).toBeVisible();
    await expect(page.getByRole("button", { name: /Ask team/i }).first()).toBeVisible();
    await expectNoHorizontalOverflow(page);
  });

  test("Dashboard focused team keeps output access compact without stacked pre-chat tiles", async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.goto("/dashboard?team_id=active-demo-team", { waitUntil: "domcontentloaded" });

    await expect(page.getByTestId("soma-operating-surface")).toBeVisible({ timeout: 20_000 });
    await expect(page.getByTestId("central-soma-chat-frame")).toBeVisible();
    await expect(page.getByTestId("soma-context-focus-bar")).toHaveCount(0);
    await expect(page.getByTestId("focused-team-output-dock")).toHaveCount(0);
    const switcher = page.getByTestId("soma-team-context-switcher");
    await expect(switcher).toBeVisible();
    await expect(switcher).toContainText("Working in");
    await expect(switcher).toContainText("First Demo Game Team");
    const digest = page.getByTestId("soma-workbench-output-digest");
    await expect(digest).toBeVisible();
    await expect(digest.getByText("Coin Runner package")).toBeVisible();
    await expect(digest.getByRole("button", { name: /Open local folder/i })).toBeVisible();
    await expect(page.getByTestId("soma-workbench-panel-toggle")).toHaveAttribute("aria-expanded", "false");
    await expectNoHorizontalOverflow(page);
  });

  test("Dashboard Soma composer remains reachable when current work is visible", async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 560 });
    await page.goto("/dashboard?team_id=active-demo-team", { waitUntil: "domcontentloaded" });

    await expect(page.getByTestId("soma-current-work-lane")).toBeVisible({ timeout: 20_000 });
    const input = page.getByTestId("central-soma-chat-frame").locator("textarea").first();
    await input.click();
    await input.fill("Verify the dashboard composer remains reachable.");
    const reachability = await input.evaluate((node) => {
      const rect = node.getBoundingClientRect();
      const centerX = rect.left + rect.width / 2;
      const centerY = rect.top + rect.height / 2;
      const target = document.elementFromPoint(centerX, centerY);
      return {
        bottom: rect.bottom,
        centerReceivesInput: target === node || Boolean(target?.closest("textarea")),
        viewportHeight: window.innerHeight,
      };
    });

    expect(reachability.bottom).toBeLessThanOrEqual(reachability.viewportHeight);
    expect(reachability.centerReceivesInput).toBe(true);
    await expectNoHorizontalOverflow(page);
  });

  test("System services stay bounded on mobile", async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    await page.goto("/system?tab=services", { waitUntil: "domcontentloaded" });

    await expect(page.getByRole("heading", { name: "System" })).toBeVisible({ timeout: 20_000 });
    await expect(page.getByText(/services online/i)).toBeVisible();
    await expect(page.getByText("Automation Timing", { exact: true })).toBeVisible();
    await expectNoHorizontalOverflow(page);
  });
});
