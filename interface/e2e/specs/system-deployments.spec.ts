import { expect, test, type APIRequestContext } from "@playwright/test";
import { enableAdvancedMode, expectNoHorizontalOverflow, fulfillJSON } from "../support/finalization-proof";
import { liveAPIHeaders, liveAPIURL } from "../support/live-api-auth";

test.describe.configure({ mode: "serial" });

type DeploymentTrustSnapshot = {
  deployment_root: string;
  execution_root: string;
  workspace_root: string;
  artifact_root: string;
  current_commit: string;
  image_tag: string;
  chart_version: string;
  deployment_lane: string;
  endpoint_posture: string;
  runtime_health: {
    status: string;
    online: number;
    degraded: number;
    offline: number;
    total: number;
  };
  proof_lane: string;
  recovery_posture: string;
  checked_at: string;
};

function liveDeploymentProofRequested() {
  return process.env.PLAYWRIGHT_LIVE_BACKEND === "1" || process.env.PLAYWRIGHT_SYSTEM_DEPLOYMENTS_LIVE === "1";
}

async function fetchLiveDeploymentTrust(request: APIRequestContext) {
  try {
    return await request.get(liveAPIURL("/api/v1/system/deployments/trust"), {
      headers: liveAPIHeaders(),
      timeout: 5_000,
    });
  } catch (error) {
    test.skip(true, `BLOCKED: live Core deployment trust endpoint is unreachable: ${String(error)}`);
    throw error;
  }
}

function deploymentValue(snapshot: DeploymentTrustSnapshot, key: keyof DeploymentTrustSnapshot) {
  const value = snapshot[key];
  return typeof value === "string" && value.trim() ? value : "unknown";
}

async function expectDeploymentSnapshotShape(snapshot: DeploymentTrustSnapshot) {
  for (const key of [
    "deployment_root",
    "execution_root",
    "workspace_root",
    "artifact_root",
    "current_commit",
    "image_tag",
    "chart_version",
    "deployment_lane",
    "endpoint_posture",
    "proof_lane",
    "recovery_posture",
    "checked_at",
  ] as const) {
    expect(typeof snapshot[key], `${key} should be a string`).toBe("string");
  }
  expect(snapshot.runtime_health).toEqual(expect.objectContaining({
    status: expect.stringMatching(/^(online|degraded|offline|unknown)$/),
    online: expect.any(Number),
    degraded: expect.any(Number),
    offline: expect.any(Number),
    total: expect.any(Number),
  }));
  expect(JSON.stringify(snapshot)).not.toMatch(/secret|api[_-]?key|bearer\s+/i);
}

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
    await expect(page.getByText("kubernetes/rancher-local").first()).toBeVisible();
    await expect(page.getByText("core/workspace", { exact: true }).first()).toBeVisible();
    await expect(page.getByText("abcdef1234567890")).toBeVisible();
    await expect(page.getByText("interface-proxy-to-core-8081")).toBeVisible();
    await expect(page.getByText("headed-chromium-live-backend")).toBeVisible();
    await expect(page.getByText("restart Core bridge, then rerun deployment trust proof")).toBeVisible();
    await expect(page.getByText("3/4 online, 1 degraded, 0 offline")).toBeVisible();
    await expectNoHorizontalOverflow(page);
  });

  test("renders deployment trust posture from local live Core when available", async ({ page, request }) => {
    test.skip(
      !liveDeploymentProofRequested(),
      "BLOCKED: live System -> Deployments proof needs PLAYWRIGHT_LIVE_BACKEND=1 or PLAYWRIGHT_SYSTEM_DEPLOYMENTS_LIVE=1 with local Core/Interface running.",
    );

    const response = await fetchLiveDeploymentTrust(request);
    const body = await response.text();
    test.skip(
      response.status() === 401 || response.status() === 403,
      "BLOCKED: live Core is reachable but deployment trust proof needs MYCELIS_API_KEY authorization.",
    );
    expect(response.ok(), body).toBeTruthy();

    const payload = JSON.parse(body);
    expect(payload?.ok, body).toBe(true);
    const snapshot = payload.data as DeploymentTrustSnapshot;
    await expectDeploymentSnapshotShape(snapshot);

    await enableAdvancedMode(page);
    await page.goto("/system?tab=deployments", { waitUntil: "domcontentloaded" });
    await expect(page.getByRole("heading", { name: "Deployment Trust" })).toBeVisible({ timeout: 20_000 });
    await expect(page.getByText(deploymentValue(snapshot, "deployment_root")).first()).toBeVisible();
    await expect(page.getByText(deploymentValue(snapshot, "execution_root")).first()).toBeVisible();
    await expect(page.getByText(deploymentValue(snapshot, "endpoint_posture")).first()).toBeVisible();
    await expect(page.getByText(deploymentValue(snapshot, "recovery_posture")).first()).toBeVisible();
    await expect(page.getByText(
      `${snapshot.runtime_health.online}/${snapshot.runtime_health.total} online, ${snapshot.runtime_health.degraded} degraded, ${snapshot.runtime_health.offline} offline`,
    )).toBeVisible();
    await expectNoHorizontalOverflow(page);
  });
});
