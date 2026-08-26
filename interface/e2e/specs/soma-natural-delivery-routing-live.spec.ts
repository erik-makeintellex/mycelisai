import { expect, test, type Page } from "@playwright/test";
import {
  type APIEnvelope,
  confirmProposal,
  createOrganization,
  liveAPIGet,
  liveTimeoutMs,
  openLiveWorkspace,
  submitLiveWorkspaceChat,
} from "../support/finalization-browser-package";
import {
  createQAFixtureScope,
  purgeDeliveryFixture,
} from "../support/qa-fixture-ownership";

type NaturalProposalData = {
  mode?: string;
  payload?: {
    text?: string;
    tools_used?: string[];
  };
};
type ConfirmData = { run_id?: string; execution_state?: string; run_status?: string };
type WorkItem = {
  run_id?: string;
  execution_shape?: string;
  state?: string;
  degradation_state?: string;
  last_event?: {
    summary?: string;
    detail?: string;
    recovery?: string;
  };
  output_refs?: Array<{ kind?: string; entrypoint?: string; storage_ref?: string }>;
};

// Local-model multi-file delivery can cross five minutes. Keep the live proof
// below Core's 15-minute recovery deadline while allowing bounded correction.
const naturalDeliveryTimeoutMs = Math.max(liveTimeoutMs, 780_000);

test.describe("Natural Soma delivery routing", () => {
  test.skip(!process.env.PLAYWRIGHT_LIVE_BACKEND, "requires a live Core backend");
  test.setTimeout(naturalDeliveryTimeoutMs);

  test("turns an application outcome ask into governed team delivery", async ({ page }) => {
    const stamp = Date.now();
    const fixture = await createQAFixtureScope(page, `natural-delivery-${stamp}`);
    let organizationID: string | undefined;
    let teamID: string | undefined;
    let runID: string | undefined;
    let journeyFailure: unknown;
    try {
      organizationID = await createOrganization(
        page,
        `Natural Delivery Routing ${stamp}`,
        { fixtureScopeID: fixture.id },
      );
      await openLiveWorkspace(page, organizationID);
      const proposal = await submitLiveWorkspaceChat(
        page,
        `Develop a small playable browser application where clicking a visible Play button increments an on-screen score and Restart resets it. Include validation proof, a direct launch link, and delivery reference ${stamp}.`,
      );

      expect(proposal.response.ok(), proposal.body ? JSON.stringify(proposal.body) : proposal.raw).toBeTruthy();
      const proposalData = proposal.body?.data as NaturalProposalData | undefined;
      expect(proposalData?.mode).toBe("proposal");
      expect(proposalData?.payload?.tools_used).toEqual(
        expect.arrayContaining(["create_team", "write_file", "delegate_task"]),
      );
      await expect(page.getByText(/reply.*(start|approve).*to begin/i).last()).toBeVisible();
      const handoff = page.getByText(/Bring in application-delivery-team-[a-z0-9]+ and keep their work connected to this conversation/i).last();
      await expect(handoff).toBeVisible();
      await expect(page.getByRole("button", { name: /^(Start|Approve)$/i })).toHaveCount(0);

      teamID = (await handoff.textContent())?.match(/application-delivery-team-[a-z0-9]+/i)?.[0];
      expect(teamID).toMatch(/^application-delivery-team-/);
      const confirmationStartedAt = Date.now();
      const confirmed = await confirmProposal(page);
      expect(confirmed.response.request().headers()["x-mycelis-qa-fixture-scope"]).toBe(fixture.id);
      expect(Date.now() - confirmationStartedAt).toBeLessThan(10_000);
      expect(confirmed.response.ok(), confirmed.body ? JSON.stringify(confirmed.body) : confirmed.raw).toBeTruthy();
      const confirmedData = confirmed.body?.data as ConfirmData | undefined;
      expect(confirmedData?.execution_state).toBe("running");
      expect(confirmedData?.run_status).toBe("running");
      runID = confirmedData?.run_id;
      expect(runID).toBeTruthy();
      await expect(page.getByText("Work started", { exact: true }).last()).toBeVisible({ timeout: 30_000 });
      await expect(page.getByPlaceholder(/Tell Soma what you want/i)).toBeEnabled();

      const steeringStartedAt = Date.now();
      const steering = await submitLiveWorkspaceChat(
        page,
        "Also keep the visible Restart control in the finished app.",
      );
      expect(steering.response.ok(), steering.body ? JSON.stringify(steering.body) : steering.raw).toBeTruthy();
      const steeringRequest = steering.response.request().postDataJSON() as {
        active_work_context?: {
          type?: string;
          run_id?: string;
          team_id?: string;
          work_item_id?: string;
          steering_id?: string;
        };
      };
      expect(steeringRequest.active_work_context).toEqual(expect.objectContaining({
        type: "team_work",
        run_id: runID,
        team_id: teamID,
        work_item_id: expect.any(String),
        steering_id: expect.any(String),
      }));
      expect(Date.now() - steeringStartedAt).toBeLessThan(10_000);
      const steeringData = steering.body?.data as NaturalProposalData | undefined;
      expect(steeringData?.payload?.text).toContain("I passed that guidance to the team");
      await expect(page.getByText(/I passed that guidance to the team/i).last()).toBeVisible();
      await expect(page.getByPlaceholder(/Tell Soma what you want/i)).toBeEnabled();

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
      const outputFrame = appPage.locator("iframe").first();
      const appSurface = (await outputFrame.count()) > 0
        ? appPage.frameLocator("iframe").first()
        : appPage;
      await expect(appSurface.getByRole("button", { name: /restart/i })).toBeVisible();
      const appBody = appSurface.locator("body");
      await expect(appBody).toBeVisible({ timeout: 30_000 });
      const beforeInteraction = await appBody.screenshot();
      const primaryControl = appSurface.getByRole("button").filter({ hasNotText: /restart/i }).first();
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
    } catch (error) {
      journeyFailure = error;
      throw error;
    } finally {
      try {
        await purgeDeliveryFixture(page, fixture, { teamID, organizationID, runID });
      } catch (cleanupError) {
        if (!journeyFailure) throw cleanupError;
        if (journeyFailure instanceof Error && cleanupError instanceof Error) {
          journeyFailure.message += ` Cleanup also failed: ${cleanupError.message}`;
        }
      }
    }
  });
});

async function waitForNaturalDelivery(page: import("@playwright/test").Page, teamID: string, runID: string) {
  let latest: WorkItem | undefined;
  try {
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
  } catch (error) {
    const detail = error instanceof Error ? error.message : String(error);
    throw new Error(`${detail}; last work state: ${JSON.stringify(latest ?? {})}`);
  }
  if (latest?.state !== "output_ready") {
    throw new Error(
      `Natural delivery ended ${latest?.state}: ${latest?.degradation_state ?? "unknown"}; ${JSON.stringify(latest?.last_event ?? {})}`,
    );
  }
  return latest;
}
