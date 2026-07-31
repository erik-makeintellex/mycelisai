import { expect, test, type Page } from "@playwright/test";
import {
  type APIEnvelope,
  type GroupRecord,
  confirmProposal,
  createOrganization,
  liveAPIGet,
  liveTimeoutMs,
  openLiveWorkspace,
  submitLiveWorkspaceChat,
} from "../support/finalization-browser-package";
import { liveAPIHeaders, liveAPIURL } from "../support/live-api-auth";

type NaturalProposalData = {
  mode?: string;
  payload?: {
    tools_used?: string[];
  };
};
type ConfirmData = { run_id?: string; execution_state?: string; run_status?: string };
type WorkItem = {
  run_id?: string;
  execution_shape?: string;
  state?: string;
  degradation_state?: string;
  output_refs?: Array<{ kind?: string; entrypoint?: string; storage_ref?: string }>;
};

const naturalDeliveryTimeoutMs = Math.max(liveTimeoutMs, 420_000);

test.describe("Natural Soma delivery routing", () => {
  test.skip(!process.env.PLAYWRIGHT_LIVE_BACKEND, "requires a live Core backend");
  test.setTimeout(naturalDeliveryTimeoutMs);

  test("turns an application outcome ask into governed team delivery", async ({ page }) => {
    const stamp = Date.now();
    const organizationID = await createOrganization(page, `Natural Delivery Routing ${stamp}`);

    await openLiveWorkspace(page, organizationID);
    const proposal = await submitLiveWorkspaceChat(
      page,
      "Develop a small playable browser application where clicking a visible Play button increments an on-screen score and Restart resets it. Include validation proof and a direct launch link.",
    );

    expect(proposal.response.ok(), proposal.body ? JSON.stringify(proposal.body) : proposal.raw).toBeTruthy();
    const proposalData = proposal.body?.data as NaturalProposalData | undefined;
    expect(proposalData?.mode).toBe("proposal");
    expect(proposalData?.payload?.tools_used).toEqual(
      expect.arrayContaining(["create_team", "write_file", "delegate_task"]),
    );
    await expect(page.getByRole("heading", { name: /Start this\?|Approve this\?/ }).last()).toBeVisible();
    const handoff = page.getByText(/Hand the work to application-delivery-team-[a-z0-9]+ through the team bus/i).last();
    await expect(handoff).toBeVisible();
    await expect(page.getByRole("button", { name: /^(Start|Approve)$/i }).last()).toBeVisible();

    const teamID = (await handoff.textContent())?.match(/application-delivery-team-[a-z0-9]+/i)?.[0];
    expect(teamID).toMatch(/^application-delivery-team-/);

    try {
      const confirmed = await confirmProposal(page);
      expect(confirmed.response.ok(), confirmed.body ? JSON.stringify(confirmed.body) : confirmed.raw).toBeTruthy();
      const confirmedData = confirmed.body?.data as ConfirmData | undefined;
      expect(confirmedData?.execution_state).toBe("running");
      expect(confirmedData?.run_status).toBe("running");
      await expect(page.getByText("Execution started", { exact: true }).last()).toBeVisible({ timeout: 30_000 });

      const completed = await waitForNaturalDelivery(page, teamID!, confirmedData!.run_id!);
      expect(completed.output_refs?.some((output) => output.kind === "project_package"), JSON.stringify(completed)).toBeTruthy();
      await expect(page.getByText("Work complete", { exact: true }).last()).toBeVisible({ timeout: 30_000 });
      const directOpen = page.getByRole("link", { name: "Open app" }).last();
      await expect(directOpen).toBeVisible();

      const pageErrors: string[] = [];
      const recordPageErrors = (candidate: Page) => {
        candidate.on("pageerror", (error) => pageErrors.push(error.message));
      };
      recordPageErrors(page);
      page.context().on("page", recordPageErrors);
      const appPagePromise = page.waitForEvent("popup", { timeout: 5_000 }).catch(() => null);
      await directOpen.click();
      const appPage = (await appPagePromise) ?? page;
      await appPage.waitForLoadState("domcontentloaded");
      await expect(appPage.getByRole("button", { name: /restart/i })).toBeVisible();
      const appBody = appPage.locator("body");
      await expect(appBody).toBeVisible({ timeout: 30_000 });
      const beforeInteraction = await appBody.screenshot();
      const primaryControl = appPage.getByRole("button").filter({ hasNotText: /restart/i }).first();
      if (await primaryControl.count()) {
        await primaryControl.click();
      } else {
        await appPage.keyboard.press("ArrowRight");
      }
      await expect.poll(async () => (await appBody.screenshot()).equals(beforeInteraction), {
        timeout: 5_000,
        intervals: [100, 200, 500],
        message: "the generated browser app should visibly respond to its primary documented control",
      }).toBeFalsy();
      expect(pageErrors).toEqual([]);
      page.context().off("page", recordPageErrors);
      if (appPage !== page) await appPage.close();
    } finally {
      await cleanupNaturalDelivery(page, teamID!);
    }
  });
});

async function waitForNaturalDelivery(page: import("@playwright/test").Page, teamID: string, runID: string) {
  let latest: WorkItem | undefined;
  await expect.poll(async () => {
    const response = await liveAPIGet(page, `/api/v1/teams/${encodeURIComponent(teamID)}/work?limit=25`);
    if (!response.ok()) return `http_${response.status()}`;
    const items = ((await response.json()) as APIEnvelope<WorkItem[]>).data ?? [];
    latest = items.find((item) => item.run_id === runID && item.execution_shape === "delegated_work");
    return latest?.state ?? "missing";
  }, {
    timeout: naturalDeliveryTimeoutMs - 60_000,
    intervals: [500, 1_000, 2_000, 5_000],
    message: `natural delivery team ${teamID} should return a usable output or an honest blocker`,
  }).toMatch(/^(output_ready|degraded|needs_operator)$/);
  if (latest?.state !== "output_ready") {
    throw new Error(`Natural delivery ended ${latest?.state}: ${latest?.degradation_state ?? "unknown"}`);
  }
  return latest;
}

async function cleanupNaturalDelivery(page: import("@playwright/test").Page, teamID: string) {
  const groupsResponse = await liveAPIGet(page, "/api/v1/groups");
  if (groupsResponse.ok()) {
    const groups = ((await groupsResponse.json()) as APIEnvelope<GroupRecord[]>).data ?? [];
    const group = groups.find((candidate) => candidate.team_ids?.includes(teamID));
    if (group) {
      await page.request.post(liveAPIURL(`/api/v1/groups/${encodeURIComponent(group.group_id)}/clear`), {
        headers: { ...(liveAPIHeaders() ?? {}), "Content-Type": "application/json" },
        data: { include_outputs: true },
      });
    }
  }
  await page.request.delete(liveAPIURL(`/api/v1/teams/${encodeURIComponent(teamID)}`), {
    headers: liveAPIHeaders(),
  });
}
