import { expect, test } from "@playwright/test";
import { liveAPIHeaders, liveAPIURL } from "../support/live-api-auth";

test.describe.configure({ mode: "serial" });

test.describe("Active work TeamWorkItem API contract", () => {
  test("live API exposes durable TeamWorkItem records for the active work lane", async ({ request }) => {
    test.skip(
      process.env.PLAYWRIGHT_TEAM_WORK_API !== "1",
      "BLOCKED: TeamWorkItem API live proof needs a live Core with migrated team-work tables and a seeded/created team.",
    );

    const teamId = process.env.PLAYWRIGHT_TEAM_WORK_API_TEAM_ID ?? "first-demo-game-team";
    const response = await request.get(liveAPIURL(`/api/v1/teams/${encodeURIComponent(teamId)}/work`), {
      headers: liveAPIHeaders(),
    });
    expect(response.ok(), await response.text()).toBeTruthy();
    const payload = await response.json();
    const items = Array.isArray(payload?.data) ? payload.data : payload;
    expect(Array.isArray(items)).toBeTruthy();

    for (const item of items) {
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
  });
});
