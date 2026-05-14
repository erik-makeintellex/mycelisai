import { expect, test, type Page } from "@playwright/test";

type RouteLike = {
  fulfill: (options: {
    status: number;
    contentType: string;
    body: string;
  }) => Promise<void>;
};

async function fulfillJSON(route: RouteLike, status: number, body: unknown) {
  await route.fulfill({
    status,
    contentType: "application/json",
    body: JSON.stringify(body),
  });
}

function isoDaysFromNow(days: number): string {
  const date = new Date();
  date.setUTCDate(date.getUTCDate() + days);
  date.setUTCHours(12, 0, 0, 0);
  return date.toISOString();
}

async function mockBroadcastProofWorkspace(page: Page) {
  await page.route("**/api/v1/groups", async (route) => {
    await fulfillJSON(route, 200, {
      ok: true,
      data: [
        {
          group_id: "group-broadcast-proof",
          name: "Broadcast Proof Lane",
          goal_statement: "Prove group broadcast summary visibility.",
          work_mode: "execute_with_approval",
          member_user_ids: ["owner"],
          team_ids: ["launch-lead", "design-lead"],
          coordinator_profile: "proof lead",
          approval_policy_ref: "browser-proof",
          status: "active",
          expiry: isoDaysFromNow(7),
          created_by: "owner",
          created_at: isoDaysFromNow(-1),
        },
      ],
    });
  });
  await page.route("**/api/v1/groups/monitor", async (route) => {
    await fulfillJSON(route, 200, {
      ok: true,
      data: { status: "online", published_count: 1 },
    });
  });
  await page.route(
    "**/api/v1/groups/group-broadcast-proof/outputs?limit=8",
    async (route) => {
      await fulfillJSON(route, 200, { ok: true, data: [] });
    },
  );
  await page.route(
    "**/api/v1/groups/group-broadcast-proof/broadcast",
    async (route) => {
      await fulfillJSON(route, 202, {
        ok: true,
        data: {
          group_id: "group-broadcast-proof",
          status: "queued",
          execution_summary: {
            intent: { original: "Send a browser-visible launch proof." },
            execution: {
              shape: "team_execution",
              status: "running",
              summary: "Broadcast queued to 2 team command lanes.",
            },
            capability_use: { teams: ["launch-lead", "design-lead"] },
            proof: {
              label: "Audit audit-broadcast-1",
              audit_event_id: "audit-broadcast-1",
            },
            next_step: "Watch retained group outputs for team responses.",
          },
        },
      });
    },
  );
}

test.describe("Groups broadcast proof visibility", () => {
  test("surfaces execution summary and audit proof after broadcast", async ({
    page,
  }) => {
    await mockBroadcastProofWorkspace(page);
    await page.goto("/groups", { waitUntil: "domcontentloaded" });

    await expect(
      page.getByRole("heading", { name: "Broadcast Proof Lane" }),
    ).toBeVisible();
    await page
      .getByLabel("Broadcast message")
      .fill("Send a browser-visible launch proof.");
    await page.getByRole("button", { name: "Broadcast to group" }).click();

    const summary = page.getByTestId("groups-broadcast-execution-summary");
    await expect(summary).toContainText("Directed execution");
    await expect(summary).toContainText(
      "Broadcast queued to 2 team command lanes.",
    );
    await expect(summary).toContainText("Audit audit-broadcast-1");
  });
});
