import { expect, test } from "@playwright/test";
import {
  createOrganization,
  liveTimeoutMs,
  openLiveWorkspace,
  submitLiveWorkspaceChat,
} from "../support/finalization-browser-package";

test.describe("Natural Soma delivery routing", () => {
  test.skip(!process.env.PLAYWRIGHT_LIVE_BACKEND, "requires a live Core backend");
  test.setTimeout(liveTimeoutMs);

  test("turns an application outcome ask into governed team delivery", async ({ page }) => {
    const stamp = Date.now();
    const organizationID = await createOrganization(page, `Natural Delivery Routing ${stamp}`);

    await openLiveWorkspace(page, organizationID);
    const proposal = await submitLiveWorkspaceChat(
      page,
      "Develop a small playable browser application with a clear objective, controls, restart, validation proof, and a direct launch link.",
    );

    expect(proposal.response.ok(), proposal.body ? JSON.stringify(proposal.body) : proposal.raw).toBeTruthy();
    expect(proposal.body?.data?.mode).toBe("proposal");
    expect(proposal.body?.data?.payload?.tools_used).toEqual(
      expect.arrayContaining(["create_team", "write_file", "delegate_task"]),
    );
    await expect(page.getByRole("heading", { name: /Start this\?|Approve this\?/ }).last()).toBeVisible();
    await expect(page.getByText(/Hand the work to application-delivery-team-[a-z0-9]+ through the team bus/i).last()).toBeVisible();
    await expect(page.getByRole("button", { name: /^(Start|Approve)$/i }).last()).toBeVisible();
  });
});
