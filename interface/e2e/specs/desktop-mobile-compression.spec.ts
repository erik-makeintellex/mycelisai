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
    await expect(recoveryRow.getByText("Needs recovery")).toBeVisible();
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

  test("System services stay bounded on mobile", async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    await page.goto("/system?tab=services", { waitUntil: "domcontentloaded" });

    await expect(page.getByRole("heading", { name: "System" })).toBeVisible({ timeout: 20_000 });
    await expect(page.getByText(/services online/i)).toBeVisible();
    await expect(page.getByText("Automation Timing", { exact: true })).toBeVisible();
    await expectNoHorizontalOverflow(page);
  });
});
