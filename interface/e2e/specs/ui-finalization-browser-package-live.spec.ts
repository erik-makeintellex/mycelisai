import { expect, test } from "@playwright/test";
import {
  type APIEnvelope,
  type GroupRecord,
  confirmProposal,
  createOrganization,
  expectProjectPackageMetadata,
  expectProjectPackageVisible,
  liveTimeoutMs,
  liveAPIGet,
  openLiveWorkspace,
  parseJSONIfPossible,
  removeTarget,
  submitLiveWorkspaceChat,
  targetExists,
} from "../support/finalization-browser-package";

test.describe("UI finalization exact browser package live proof", () => {
  test("creates the first-demo package, opens proof, reloads output, and verifies Groups", async ({ page }) => {
    test.skip(!process.env.PLAYWRIGHT_LIVE_BACKEND, "requires a live Core backend");
    test.setTimeout(liveTimeoutMs);
    test.slow();

    const stamp = Date.now();
    const teamID = `first-demo-game-team-${stamp}`;
    const teamName = "First Demo Game Team";
    const folder = `groups/${teamID}/generated/first-game`;
    const entrypoint = `${folder}/index.html`;
    const packageTitle = `${teamName} First Playable`;
    const organizationId = await createOrganization(page, `First Demo UI Finalization ${stamp}`);

    await openLiveWorkspace(page, organizationId);
    try {
      const proposal = await submitLiveWorkspaceChat(
        page,
        [
          `Create a team with team_id ${teamID} named ${teamName}.`,
          "Ask Soma for the exact first demo deliverable: a playable browser game project package.",
          `Retain it at ${folder} with entrypoint ${entrypoint}.`,
          "The package metadata must include files index.html and README.md plus validation notes from opening the browser game.",
          "After approval, return a retained project_package output with entrypoint, folder, files, and validation.",
        ].join(" "),
      );

      expect(proposal.response.ok(), proposal.body ? JSON.stringify(proposal.body) : proposal.raw).toBeTruthy();
      expect(proposal.body?.data?.mode).toBe("proposal");
      await expect(page.getByText("PROPOSED ACTION").last()).toBeVisible({ timeout: 30_000 });
      await expect(page.getByText(teamID).last()).toBeVisible();
      expect(targetExists(entrypoint)).toBeFalsy();

      const confirmed = await confirmProposal(page);
      expect(confirmed.response.ok(), confirmed.body ? JSON.stringify(confirmed.body) : confirmed.raw).toBeTruthy();
      expect(confirmed.body?.data?.verified).toBeTruthy();
      expect(confirmed.body?.data?.execution_state).toBe("verified");
      const runID = confirmed.body?.data?.run_id;
      expect(runID).toBeTruthy();

      const outputs = confirmed.body?.data?.execution_summary?.outputs ?? [];
      const projectPackage = outputs.find((output) => output.kind === "project_package");
      expect(projectPackage, JSON.stringify(outputs)).toBeTruthy();
      expectProjectPackageMetadata(projectPackage!, { title: packageTitle, entrypoint, folder });

      await expect.poll(() => targetExists(entrypoint), {
        timeout: 30_000,
        message: `expected backend workspace entrypoint ${entrypoint} to exist after approval`,
      }).toBeTruthy();
      await expectProjectPackageVisible(page, { title: packageTitle, entrypoint, folder });

      const runProofLink = page.getByRole("link", { name: `Run ${runID!.slice(0, 8)}` }).last();
      await expect(runProofLink).toHaveAttribute("href", `/runs/${runID}`);
      await runProofLink.click();
      await expect(page).toHaveURL(new RegExp(`/runs/${runID}$`));
      await page.goBack({ waitUntil: "domcontentloaded" });
      await expectProjectPackageVisible(page, { title: packageTitle, entrypoint, folder });

      const outputPagePromise = page.context().waitForEvent("page");
      await page.getByRole("button", { name: `Open Game ${packageTitle} in a new browser window` }).last().click();
      const outputPage = await outputPagePromise;
      await outputPage.waitForLoadState("domcontentloaded");
      await expect(outputPage).toHaveTitle(packageTitle);
      await expect(outputPage.locator("canvas#game")).toBeVisible({ timeout: 30_000 });
      await outputPage.reload({ waitUntil: "domcontentloaded" });
      await expect(outputPage.locator("canvas#game")).toBeVisible({ timeout: 30_000 });
      await outputPage.close();

      const groupsResponse = await liveAPIGet(page, "/api/v1/groups");
      const parsedGroups = await parseJSONIfPossible<APIEnvelope<GroupRecord[]>>(groupsResponse);
      expect(groupsResponse.ok(), parsedGroups.body ? JSON.stringify(parsedGroups.body) : parsedGroups.raw).toBeTruthy();
      const group = (parsedGroups.body?.data ?? []).find((candidate) => candidate.team_ids?.includes(teamID));
      expect(group, JSON.stringify(parsedGroups.body?.data ?? [])).toBeTruthy();

      await page.goto(`/groups?group_id=${encodeURIComponent(group!.group_id)}`, { waitUntil: "domcontentloaded" });
      await expect(page.getByRole("heading", { name: group!.name })).toBeVisible({ timeout: 30_000 });
      await expect(page.getByText(teamID).first()).toBeVisible();
      await expect(page.getByText("Project package").first()).toBeVisible({ timeout: 30_000 });
      await expect(page.getByText(entrypoint).first()).toBeVisible();
      await expect(page.getByText(folder, { exact: true }).first()).toBeVisible();
      await expect(page.getByText("README.md").first()).toBeVisible();
      await expect(page.getByText(/browser|validation|opened|play/i).first()).toBeVisible();
    } finally {
      removeTarget(entrypoint);
    }
  });
});
