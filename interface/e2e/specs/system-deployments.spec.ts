import { expect, test } from "@playwright/test";
import { enableAdvancedMode, expectNoHorizontalOverflow, fulfillJSON } from "../support/finalization-proof";

test.describe.configure({ mode: "serial" });

test.describe("System -> Deployments focused proof", () => {
  test("renders deployment trust posture from mocked API data", async ({ page }) => {
    await enableAdvancedMode(page);
    await page.route("**/api/v1/system/deployments/trust", async (route) => {
      await fulfillJSON(route, 200, {
        ok: true,
        data: {
          deployment_root: "kubernetes/rancher-local",
          execution_root: "core/runtime/execution",
          workspace_root: "core/workspace",
          artifact_root: "core/workspace/artifacts",
          current_commit: "abcdef1234567890",
          image_tag: "mycelis-core:proof",
          chart_version: "0.1.0-proof",
          deployment_lane: "local-k8s-proof",
          endpoint_posture: "interface-proxy-to-core-8081",
          runtime_health: {
            status: "degraded",
            online: 3,
            degraded: 1,
            offline: 0,
            total: 4,
          },
          proof_lane: "headed-chromium-live-backend",
          recovery_posture: "restart Core bridge, then rerun deployment trust proof",
          checked_at: "2026-05-17T12:00:00Z",
        },
      });
    });

    await page.goto("/system?tab=deployments", { waitUntil: "domcontentloaded" });
    await expect(page.getByRole("heading", { name: "System", exact: true })).toBeVisible({ timeout: 20_000 });
    await expect(page.getByRole("button", { name: "Deployments" })).toBeVisible();
    await expect(page.getByRole("heading", { name: "Deployment Trust" })).toBeVisible();
    await expect(page.getByText("kubernetes/rancher-local")).toBeVisible();
    await expect(page.getByText("core/workspace", { exact: true })).toBeVisible();
    await expect(page.getByText("abcdef1234567890")).toBeVisible();
    await expect(page.getByText("interface-proxy-to-core-8081")).toBeVisible();
    await expect(page.getByText("headed-chromium-live-backend")).toBeVisible();
    await expect(page.getByText("restart Core bridge, then rerun deployment trust proof")).toBeVisible();
    await expect(page.getByText("3/4 online, 1 degraded, 0 offline")).toBeVisible();
    await expectNoHorizontalOverflow(page);
  });
});
