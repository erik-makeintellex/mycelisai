import crypto from "node:crypto";
import { expect, test, type APIResponse } from "@playwright/test";
import { liveAPIHeaders, liveAPIURL } from "../support/live-api-auth";

type TeamDetail = {
  id?: string;
  name?: string;
};

function liveRailTargetProofRequested() {
  return process.env.PLAYWRIGHT_RAIL_TARGET_GUI_LIVE === "1";
}

async function skipWhenLivePrerequisiteMissing(response: APIResponse, body: string) {
  test.skip(
    response.status() === 401 || response.status() === 403,
    "BLOCKED: rail target GUI proof needs a valid MYCELIS_API_KEY for Core.",
  );
  test.skip(
    response.status() === 503 && /database not available|connection refused|connectex|no such host|i\/o timeout/i.test(body),
    `BLOCKED: rail target GUI proof needs local Core with PostgreSQL available: ${body}`,
  );
}

function teamDetailsFromPayload(payload: unknown): TeamDetail[] {
  const data = (payload as { data?: unknown })?.data ?? payload;
  return Array.isArray(data) ? data as TeamDetail[] : [];
}

test.describe("Dashboard right-rail target live proof", () => {
  test("opens API-backed recovery work from Outcome Vault", async ({ page }) => {
    test.skip(
      !liveRailTargetProofRequested(),
      "BLOCKED: set PLAYWRIGHT_RAIL_TARGET_GUI_LIVE=1 with local Core, Interface, PostgreSQL, and at least one team.",
    );
    test.slow();
    test.setTimeout(90_000);

    const teamsResponse = await page.request.get(liveAPIURL("/api/v1/teams/detail"), {
      headers: liveAPIHeaders(),
      timeout: 10_000,
    });
    const teamsBody = await teamsResponse.text();
    await skipWhenLivePrerequisiteMissing(teamsResponse, teamsBody);
    expect(teamsResponse.ok(), teamsBody).toBeTruthy();
    const team = teamDetailsFromPayload(JSON.parse(teamsBody)).find((entry) => entry.id);
    if (!team?.id) {
      test.skip(true, "BLOCKED: rail target GUI proof needs at least one runtime team.");
      return;
    }

    const teamId = team.id;
    const workItemId = crypto.randomUUID();
    const objective = "Recover targeted rail proof";
    const createResponse = await page.request.post(
      liveAPIURL(`/api/v1/teams/${encodeURIComponent(teamId)}/work`),
      {
        headers: liveAPIHeaders(),
        timeout: 10_000,
        data: {
          work_item_id: workItemId,
          execution_shape: "deliverable",
          state: "degraded",
          objective,
          owner: "Playwright rail target proof",
          expected_outputs: ["quiet right-rail target link"],
          expected_proof: ["Dashboard rail opens the matching review item"],
          recovery_options: ["retry after reviewing target proof"],
          needs_operator: true,
          degradation_state: "playwright_target_probe",
          governance_posture: "required",
        },
      },
    );
    const createBody = await createResponse.text();
    await skipWhenLivePrerequisiteMissing(createResponse, createBody);
    expect(createResponse.status(), createBody).toBe(201);

    try {
      await page.addInitScript(() => {
        window.localStorage.setItem("mycelis-advanced-mode", "true");
      });
      await page.goto(`/dashboard?team_id=${encodeURIComponent(teamId)}`, { waitUntil: "domcontentloaded" });
      await expect(page.getByTestId("soma-outcome-vault")).toBeVisible({ timeout: 20_000 });
      await expect(page.getByText("Outcome Vault")).toBeVisible();

      const targetReference = `recovery:${workItemId}`;
      const railLink = page.locator(`[data-target-reference="${targetReference}"]`).first();
      await expect(railLink).toBeVisible({ timeout: 30_000 });
      await expect(railLink).toHaveAttribute("data-target-type", "recovery");
      await expect(railLink).toHaveAttribute("data-target-id", workItemId);
      await expect(railLink).toHaveAttribute("href", `/teams?view=work&work_item_id=${workItemId}`);
      await expect(railLink).toContainText("Work needs attention");

      const dashboardText = await page.getByTestId("soma-outcome-vault").innerText();
      expect(dashboardText).not.toContain(workItemId);
      expect(dashboardText).not.toContain(targetReference);

      await railLink.click();
      await expect(page).toHaveURL(new RegExp(`/teams\\?view=work&work_item_id=${workItemId}`));
      await expect(page.getByRole("heading", { name: "Recovery and Review", exact: true })).toBeVisible();
      await expect(page.locator(`[data-work-item-id="${workItemId}"]`)).toBeVisible({ timeout: 30_000 });
      await expect(page.getByText(`Opened "${objective}" from Outcome Vault.`)).toBeVisible();
      await expect(page.getByLabel(`Review details for ${objective}`)).toBeVisible();

      const reviewText = await page.getByTestId("work-review-inbox").innerText();
      expect(reviewText).not.toContain(workItemId);
      expect(reviewText).not.toContain(targetReference);
    } finally {
      await page.request.post(
        liveAPIURL(`/api/v1/teams/${encodeURIComponent(teamId)}/work/${encodeURIComponent(workItemId)}/actions`),
        {
          headers: liveAPIHeaders(),
          timeout: 10_000,
          data: {
            action: "archive",
            summary: "Clear Playwright rail target proof item after live browser validation.",
            actor_ref: "playwright",
            source_kind: "workspace_ui",
            source_channel: "e2e.dashboard_rail_target",
            payload_kind: "operator_action",
          },
        },
      ).catch(() => undefined);
    }
  });
});
