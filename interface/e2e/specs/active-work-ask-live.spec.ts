import crypto from "node:crypto";
import { expect, test, type APIResponse } from "@playwright/test";
import { liveAPIHeaders, liveAPIURL } from "../support/live-api-auth";

type TeamWorkItem = {
  work_item_id: string;
  state: string;
  objective: string;
};

function liveGUIProofRequested() {
  return process.env.PLAYWRIGHT_TEAM_WORK_GUI_LIVE === "1";
}

async function skipWhenLivePrerequisiteMissing(response: APIResponse, body: string) {
  test.skip(
    response.status() === 401 || response.status() === 403,
    "BLOCKED: live /teams ask proof needs a valid MYCELIS_API_KEY for Core.",
  );
  test.skip(
    response.status() === 503 && /database not available|connection refused|connectex|no such host|i\/o timeout/i.test(body),
    `BLOCKED: live /teams ask proof needs local Core with PostgreSQL available: ${body}`,
  );
}

test.describe("Active work Ask Team live GUI proof", () => {
  test("submits a bounded ask from /teams without blocking the operator", async ({ page }) => {
    test.skip(
      !liveGUIProofRequested(),
      "BLOCKED: set PLAYWRIGHT_TEAM_WORK_GUI_LIVE=1 with local Core, Interface, NATS, PostgreSQL, and a responsive runtime team.",
    );
    test.slow();
    test.setTimeout(120_000);

    const teamId = process.env.PLAYWRIGHT_TEAM_WORK_API_TEAM_ID ?? "prime-development";
    const sourceWorkItemId = crypto.randomUUID();
    const sourceObjective = `Playwright GUI Ask Team source ${sourceWorkItemId}`;
    const askMarkerId = crypto.randomUUID();
    const askMarker = `gui bounded output ready ${askMarkerId}`;
    const createResponse = await page.request.post(
      liveAPIURL(`/api/v1/teams/${encodeURIComponent(teamId)}/work`),
      {
        headers: liveAPIHeaders(),
        timeout: 10_000,
        data: {
          work_item_id: sourceWorkItemId,
          execution_shape: "delegated_work",
          state: "queued",
          objective: sourceObjective,
          owner: "Playwright GUI proof",
          expected_outputs: ["source row for GUI bounded ask"],
          expected_proof: ["operator can ask the team from /teams"],
          governance_posture: "auto_allowed",
        },
      },
    );
    const createBody = await createResponse.text();
    await skipWhenLivePrerequisiteMissing(createResponse, createBody);
    expect(createResponse.status(), createBody).toBe(201);

    await page.addInitScript(() => {
      window.localStorage.setItem("mycelis-advanced-mode", "true");
    });
    await page.goto("/teams", { waitUntil: "domcontentloaded" });
    await expect(page.getByRole("heading", { name: "Team Lead Workspaces" })).toBeVisible();

    const lane = page.getByTestId("active-work-lane");
    await expect(lane).toBeVisible();
    await expect(lane.getByText("Durable team-work state loaded.")).toBeVisible({ timeout: 20_000 });

    const sourceRow = lane.locator("article").filter({ hasText: sourceObjective });
    await expect(sourceRow).toBeVisible({ timeout: 20_000 });
    await sourceRow.getByRole("button", { name: /ask team/i }).click();
    await sourceRow.getByLabel(`Ask ${sourceObjective}`).fill(`Text-only response. Do not call tools. Reply with exactly one sentence containing ${askMarker}.`);

    const askResponsePromise = page.waitForResponse(
      (response) =>
        response.url().includes(`/api/v1/teams/${encodeURIComponent(teamId)}/work/ask`) &&
        response.request().method() === "POST",
      { timeout: 90_000 },
    );
    await sourceRow.getByRole("button", { name: /^queue ask$/i }).click();
    const askResponse = await askResponsePromise;
    const askBody = await askResponse.text();
    expect([200, 202], askBody).toContain(askResponse.status());
    const askData = JSON.parse(askBody).data as { work_item: TeamWorkItem; reply?: string };
    expect(["queued", "running", "output_ready", "degraded"]).toContain(askData.work_item.state);

    const resultTitle = `Operator asked ${teamId} to continue work on "${sourceObjective}".`;
    const resultRow = lane.locator("article").filter({ hasText: resultTitle });
    await expect(resultRow).toBeVisible({ timeout: 20_000 });
    if (askData.work_item.state === "queued") {
      await expect(resultRow.getByText(/Queued/i).first()).toBeVisible();
      await expect(lane.getByText(/Team ask queued. You can keep working/i)).toBeVisible();
      return;
    }
    if (askData.work_item.state === "running") {
      await expect(resultRow.getByText(/In progress/i).first()).toBeVisible();
      await expect(resultRow.getByText(/Running, output may still change/i).first()).toBeVisible();
      await expect(resultRow.getByText(/Wait for a team status\/result signal/i).first()).toBeVisible();
      return;
    }
    if (askData.work_item.state === "degraded") {
      await expect(resultRow.getByText(/Degraded/i).first()).toBeVisible();
      await expect(resultRow.getByText(/Needs recovery|No retained output yet/i).first()).toBeVisible();
      return;
    }
    await expect(resultRow.getByText(/Output ready/i).first()).toBeVisible();
    await expect(resultRow.getByText("Durable team work", { exact: true }).first()).toBeVisible();
    await expect(resultRow.getByText(/Team response ready/)).toBeVisible();
    expect(askData.reply ?? "").toContain(askMarkerId);
    await expect(resultRow.getByText(new RegExp(askMarkerId))).toBeVisible();
    await expect(resultRow.getByText(/Review the response/i).first()).toBeVisible();
  });
});
