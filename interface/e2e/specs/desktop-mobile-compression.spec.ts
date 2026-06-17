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

  test("Teams setup surface remains scannable on desktop", async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.goto("/teams", { waitUntil: "domcontentloaded" });

    await expect(page.getByRole("heading", { name: "Team Lead Workspaces" })).toBeVisible({ timeout: 20_000 });
    await expect(page.getByText("Specialize new teams through Soma")).toBeVisible();
    await expect(page.getByText("Soma team-member templates")).toBeVisible();
    await expect(page.getByRole("link", { name: "Open guided team creation" })).toHaveAttribute("href", "/teams/create");
    await expect(page.getByRole("link", { name: "Open Soma workspace" }).first()).toHaveAttribute("href", "/dashboard");
    await expect(page.getByRole("link", { name: "Review outputs" }).first()).toHaveAttribute("href", "/groups");
    await expect(page.getByRole("link", { name: "Configure event rules" }).first()).toHaveAttribute("href", "/automations?tab=triggers");
    await expectNoHorizontalOverflow(page);
  });

  test("Teams setup surface compresses without horizontal overflow on mobile", async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    await page.goto("/teams", { waitUntil: "domcontentloaded" });

    await expect(page.getByRole("heading", { name: "Team Lead Workspaces" })).toBeVisible({ timeout: 20_000 });
    await expect(page.getByText("Specialize new teams through Soma")).toBeVisible();
    await expect(page.getByRole("link", { name: "Open guided team creation" })).toBeVisible();
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

  test("Dashboard work review opens a focused inbox panel", async ({ page }) => {
    await page.setViewportSize({ width: 1366, height: 720 });
    await page.goto("/dashboard?team_id=degraded-proof-team", { waitUntil: "domcontentloaded" });

    await expect(page.getByTestId("soma-current-work-lane")).toBeVisible({ timeout: 20_000 });
    const toggle = page.getByTestId("soma-workbench-panel-toggle");
    await expect(toggle).toContainText("Review work");
    await expect(toggle).toHaveAttribute("aria-expanded", "false");
    await toggle.click();

    const panel = page.getByTestId("soma-workbench-side-rail");
    await expect(panel).toHaveAttribute("aria-hidden", "false");
    await expect(panel.getByText("Review work")).toBeVisible();
    await expect(panel.getByText(/Understand the item/i)).toBeVisible();
    await expect(panel.getByRole("tab", { name: /Work/i })).toHaveCount(0);
    await expect(panel.getByRole("link", { name: /Open inbox/i })).toHaveAttribute("href", "/teams?view=work");
    await expect(panel.getByLabel("Review queue summary")).toBeVisible();
    await expect(panel.getByLabel("Needs decision: 1")).toBeVisible();
    await expect(panel.getByText("Team work needs recovery")).toBeVisible();
    await expect(panel.getByRole("button", { name: /Retry recovery/i })).toBeVisible();
    await expect(panel.getByRole("button", { name: /Clear from review/i })).toBeVisible();
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
