import crypto from "node:crypto";
import { expect, test, type APIRequestContext, type APIResponse } from "@playwright/test";
import { liveAPIHeaders, liveAPIURL } from "../support/live-api-auth";

test.describe.configure({ mode: "serial" });

type TeamWorkItem = {
  work_item_id: string;
  team_id: string;
  objective: string;
  state: string;
  execution_shape: string;
  expected_outputs: string[];
  expected_proof: string[];
  version: string;
};

function liveTeamWorkProofRequested() {
  return process.env.PLAYWRIGHT_TEAM_WORK_API === "1" || process.env.PLAYWRIGHT_ACTIVE_WORK_API_LIVE === "1";
}

async function getTeamWork(request: APIRequestContext, teamId: string) {
  try {
    return await request.get(liveAPIURL(`/api/v1/teams/${encodeURIComponent(teamId)}/work?limit=25`), {
      headers: liveAPIHeaders(),
      timeout: 5_000,
    });
  } catch (error) {
    test.skip(true, `BLOCKED: live Core team-work API is unreachable: ${String(error)}`);
    throw error;
  }
}

async function skipWhenLivePrerequisiteMissing(response: APIResponse, body: string) {
  test.skip(
    response.status() === 401 || response.status() === 403,
    "BLOCKED: active-work API proof needs MYCELIS_API_KEY authorization for the live Core.",
  );
  test.skip(
    response.status() === 503 && /database not available|connection refused|connectex|no such host|i\/o timeout/i.test(body),
    `BLOCKED: active-work API proof needs local Core with PostgreSQL available: ${body}`,
  );
}

function unpackItems(payload: unknown): TeamWorkItem[] {
  const maybeEnvelope = payload as { data?: unknown };
  const raw = Array.isArray(maybeEnvelope?.data) ? maybeEnvelope.data : payload;
  return Array.isArray(raw) ? raw as TeamWorkItem[] : [];
}

function expectTeamWorkContract(item: TeamWorkItem, teamId: string) {
  expect(item).toEqual(expect.objectContaining({
    work_item_id: expect.any(String),
    team_id: teamId,
    objective: expect.any(String),
    state: expect.stringMatching(/^(new|briefed|queued|running|paused|reviewing|output_ready|degraded|needs_operator|archived)$/),
    execution_shape: expect.stringMatching(/^(create_team|delegated_work|deliverable)$/),
    expected_outputs: expect.any(Array),
    expected_proof: expect.any(Array),
    version: "v1",
  }));
}

test.describe("Active work TeamWorkItem API contract", () => {
  test("live API creates and reads a durable TeamWorkItem for the active work lane", async ({ request }) => {
    test.skip(
      !liveTeamWorkProofRequested(),
      "BLOCKED: active-work API live proof needs PLAYWRIGHT_TEAM_WORK_API=1 or PLAYWRIGHT_ACTIVE_WORK_API_LIVE=1 with local Core and migrated team-work tables.",
    );

    const teamId = process.env.PLAYWRIGHT_TEAM_WORK_API_TEAM_ID ?? "local-source-proof-team";
    const initialResponse = await getTeamWork(request, teamId);
    const initialBody = await initialResponse.text();
    await skipWhenLivePrerequisiteMissing(initialResponse, initialBody);
    expect(initialResponse.ok(), initialBody).toBeTruthy();

    const workItemId = crypto.randomUUID();
    const objective = `Playwright local-source active-work proof ${workItemId}`;
    const createResponse = await request.post(liveAPIURL(`/api/v1/teams/${encodeURIComponent(teamId)}/work`), {
      headers: liveAPIHeaders(),
      timeout: 5_000,
      data: {
        work_item_id: workItemId,
        execution_shape: "deliverable",
        state: "queued",
        objective,
        scope: ["local source proof lane"],
        owner: "deployment-proof",
        expected_outputs: ["durable TeamWorkItem row"],
        expected_proof: ["POST created the work item", "GET returned the work item"],
        capability_requirements: ["migration 042_team_work_spine"],
        governance_posture: "required",
      },
    });
    const createBody = await createResponse.text();
    expect(createResponse.status(), createBody).toBe(201);
    const created = JSON.parse(createBody).data as TeamWorkItem;
    expectTeamWorkContract(created, teamId);
    expect(created.work_item_id).toBe(workItemId);
    expect(created.objective).toBe(objective);

    const listResponse = await getTeamWork(request, teamId);
    const listBody = await listResponse.text();
    expect(listResponse.ok(), listBody).toBeTruthy();
    const items = unpackItems(JSON.parse(listBody));
    expect(items.length, listBody).toBeGreaterThan(0);
    const persisted = items.find((item) => item.work_item_id === workItemId);
    expect(persisted, listBody).toBeTruthy();
    expectTeamWorkContract(persisted as TeamWorkItem, teamId);
  });
});
