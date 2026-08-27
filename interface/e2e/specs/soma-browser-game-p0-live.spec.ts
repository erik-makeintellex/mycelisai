import { expect, test, type Locator, type Page } from "@playwright/test";
import {
  type APIEnvelope,
  chatTimeoutMs,
  attachRetainedPackageEvidence,
  confirmProposal,
  createOrganization,
  expectProjectPackageVisible,
  liveAPIGet,
  liveTimeoutMs,
  openLiveWorkspace,
  parseJSONIfPossible,
  submitLiveWorkspaceChat,
} from "../support/finalization-browser-package";
import {
  createQAFixtureScope,
  purgeDeliveryFixture,
} from "../support/qa-fixture-ownership";

type TeamWorkItem = {
  run_id?: string;
  execution_shape?: string;
  state?: string;
  degradation_state?: string;
  recovery_options?: string[];
  output_refs?: Array<{
    kind?: string;
    storage_ref?: string;
    entrypoint?: string;
  }>;
};

async function canvasSignature(canvasLocator: Locator) {
  return canvasLocator.evaluate((canvas) => {
    const element = canvas as HTMLCanvasElement;
    const context = element.getContext("2d");
    if (!context) return "";
    const sample = context.getImageData(0, 0, element.width, element.height).data;
    let hash = 0;
    for (let index = 0; index < sample.length; index += 113) {
      hash = ((hash << 5) - hash + sample[index]) | 0;
    }
    return String(hash);
  });
}

test.describe("Live Soma P0 browser game delivery", () => {
  test.skip(!process.env.PLAYWRIGHT_LIVE_BACKEND, "requires a live Core backend");
  test.describe.configure({ timeout: liveTimeoutMs });

  test("creates and opens an original gothic action-platformer through Soma", async ({ page }, testInfo) => {
    test.slow();
    const stamp = Date.now();
    const teamID = `moonlit-keep-game-${stamp}`;
    const title = "Moonlit Keep First Playable";
    const folder = `groups/${teamID}/generated/package`;
    const entrypoint = `${folder}/index.html`;
    const fixture = await createQAFixtureScope(page, `moonlit-keep-${stamp}`);
    let organizationId: string | undefined;
    let runID: string | undefined;

    try {
      organizationId = await createOrganization(
        page,
        `Moonlit Keep Delivery ${stamp}`,
        { fixtureScopeID: fixture.id },
      );
      await openLiveWorkspace(page, organizationId);
      const proposal = await submitLiveWorkspaceChat(page, [
      `Create a team with team_id ${teamID} named Moonlit Keep Game Team.`,
      `Ask Soma and that team to build an original gothic action-platformer titled ${title}.`,
      `Retain it at ${folder} with entrypoint ${entrypoint}.`,
      `Use the package title ${title}.`,
      "Capture the appeal of classic castle exploration without copying any franchise names, characters, story, music, level layouts, or assets.",
      "Use only code-generated browser graphics, no external assets, and make the objective and controls obvious to a first-time player.",
      "It must include movement, jumping or traversal, collision, an attack, enemies or hazards, health, score, key pickup, locked door, win state, fail state, restart, generated music, and action sounds.",
      "After approval, return a retained project_package output with entrypoint, folder, files, and validation.",
    ].join(" "));

    expect(proposal.response.ok(), proposal.body ? JSON.stringify(proposal.body) : proposal.raw).toBeTruthy();
    expect(proposal.body?.data?.mode).toBe("proposal");
    expect(proposal.body?.data?.payload?.tools_used).toEqual(expect.arrayContaining([
      "create_team",
      "write_file",
      "delegate_task",
    ]));
    const proposalCard = page.getByTestId("soma-proposal").last();
    await expect(proposalCard.getByText("I can start that.")).toBeVisible({ timeout: 30_000 });
    await expect(proposalCard.getByText(title)).toBeVisible();
    await expect(proposalCard.getByText(teamID, { exact: true })).toHaveCount(0);
    await proposalCard.getByRole("button", { name: /^Details$/ }).click();
    await expect(proposalCard.getByText(`team:${teamID}`, { exact: true })).toBeVisible();

      const confirmed = await confirmProposal(page);
    expect(confirmed.response.ok(), confirmed.body ? JSON.stringify(confirmed.body) : confirmed.raw).toBeTruthy();
    expect(confirmed.body?.data?.verified).toBeFalsy();
    expect(confirmed.body?.data?.execution_state).toBe("running");
      expect(confirmed.body?.data?.run_id).toBeTruthy();
      runID = confirmed.body?.data?.run_id;
    await expect(page.getByText(/^(Work queued|Work started)$/).last()).toBeVisible({ timeout: 30_000 });
    await expect(page.getByPlaceholder(/Tell Soma what you want/i)).toBeEnabled();
    const workItem = await waitForGameDelivery(page, teamID, runID!);
    const projectPackage = workItem.output_refs?.find((output) => output.kind === "project_package");
    expect(projectPackage, JSON.stringify(workItem)).toBeTruthy();
    expect(projectPackage?.storage_ref).toBe(folder);
    expect(projectPackage?.entrypoint).toMatch(/index\.html$/);
    for (const supportFile of ["README.md", "PROOF.md"]) {
      const supportResponse = await liveAPIGet(
        page,
        `/api/v1/workspace/files/view?path=${encodeURIComponent(`${folder}/${supportFile}`)}`,
      );
      const parsed = await parseJSONIfPossible(supportResponse);
      expect(supportResponse.ok(), parsed.raw).toBeTruthy();
      expect(parsed.raw).toContain(title);
    }

    await expect(page.getByText("Work complete", { exact: true }).last()).toBeVisible({ timeout: 30_000 });
    await page.goto(`/dashboard?team_id=${encodeURIComponent(teamID)}`, { waitUntil: "domcontentloaded" });
    await expectProjectPackageVisible(page, { title, entrypoint, folder });
    await page.getByRole("button", { name: new RegExp(`Open app .*${title}`, "i") }).last().click();
    await expect(page).toHaveURL(/\/outputs\/view\?/, { timeout: chatTimeoutMs });
    await expect(page.getByRole("heading", { name: title })).toBeVisible();
    const gamePage = page.frameLocator(`iframe[title="${title}"]`);
    const gameCanvas = gamePage.locator("#game");
    await expect(gameCanvas).toBeVisible({ timeout: 30_000 });
    await expect(gamePage.locator("#health")).toHaveText("4");
    await expect(gamePage.locator("#keyState")).toHaveText("No");
    await expect(gamePage.locator("#score")).toHaveText("0");
    await expect(gamePage.locator("#goalState")).toHaveText("Find key");
    await expect(gamePage.getByRole("button", { name: "Restart" })).toBeVisible();
    await expect(gamePage.getByRole("button", { name: "Sound on" })).toBeVisible();

    const html = await gamePage.locator("html").innerHTML();
    for (const required of [
      "const hazards",
      "const keyStart",
      "const gemStarts",
      "window.AudioContext",
      "startMusic()",
      "cue(\"win\")",
      "const door",
      "const relic",
      "state = \"failed\"",
      "state = \"won\"",
      "blockers()",
      "Press R",
    ]) {
      expect(html).toContain(required);
    }

    const before = await canvasSignature(gameCanvas);
    await gameCanvas.click();
    await gamePage.locator("body").press("ArrowRight", { delay: 500 });
    const after = await canvasSignature(gameCanvas);
    expect(after).not.toBe(before);
    } finally {
      try {
        await attachRetainedPackageEvidence(page, testInfo, [
          entrypoint,
          `${folder}/project-package.json`,
          `${folder}/README.md`,
          `${folder}/PROOF.md`,
        ]);
      } finally {
        await purgeDeliveryFixture(page, fixture, { teamID, organizationID: organizationId, runID });
      }
    }
  });
});

async function waitForGameDelivery(page: Page, teamID: string, runID: string) {
  let latest: TeamWorkItem | undefined;
  const refresh = async () => {
    const response = await liveAPIGet(page, `/api/v1/teams/${encodeURIComponent(teamID)}/work?limit=25`);
    if (!response.ok()) return `http_${response.status()}`;
    const items = ((await response.json()) as APIEnvelope<TeamWorkItem[]>).data ?? [];
    latest = items.find((item) => item.run_id === runID && item.execution_shape === "delegated_work");
    return latest?.state ?? "missing";
  };
  await expect.poll(refresh, {
    timeout: 600_000,
    intervals: [500, 1_000, 2_000, 5_000],
    message: `team ${teamID} should deliver the game or expose an honest recovery state`,
  }).toMatch(/^(reviewing|output_ready|degraded|needs_operator)$/);
  if (latest?.state === "reviewing") {
    await expect.poll(refresh, { timeout: 120_000, intervals: [500, 1_000, 2_000, 5_000] })
      .toMatch(/^(output_ready|degraded|needs_operator)$/);
  }
  if (latest?.state !== "output_ready") {
    throw new Error(`Game delivery ended ${latest?.state}: ${latest?.degradation_state ?? "unknown"}; recovery=${latest?.recovery_options?.join(" | ") ?? "none"}`);
  }
  return latest;
}
