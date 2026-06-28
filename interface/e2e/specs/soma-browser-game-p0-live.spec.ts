import { expect, test, type Page } from "@playwright/test";
import {
  chatTimeoutMs,
  confirmProposal,
  createOrganization,
  expectProjectPackageMetadata,
  liveAPIGet,
  liveTimeoutMs,
  openLiveWorkspace,
  parseJSONIfPossible,
  submitLiveWorkspaceChat,
} from "../support/finalization-browser-package";

async function canvasSignature(page: Page) {
  return page.locator("#game").evaluate((canvas) => {
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

  test("creates and opens a playable console-era browser game through Soma", async ({ page }) => {
    test.slow();
    const stamp = Date.now();
    const teamID = `p0-console-game-${stamp}`;
    const title = "P0 Console Browser Game Team First Playable";
    const folder = `groups/${teamID}/generated/first-game`;
    const entrypoint = `${folder}/index.html`;
    const organizationId = await createOrganization(page, `P0 Game Delivery ${stamp}`);

    await openLiveWorkspace(page, organizationId);
    const proposal = await submitLiveWorkspaceChat(page, [
      `Create a team with team_id ${teamID} named P0 Console Browser Game Team.`,
      "Ask Soma and that team to build a playable classic console-era browser game project package.",
      "Use only code-generated browser graphics, no external assets, and keep it small enough to inspect.",
      "It must include movement, collision, enemies or hazards, health, score, key pickup, locked door, win state, fail state, restart, generated music, and action sounds.",
      "After approval, return a retained project_package output with entrypoint, folder, files, and validation.",
    ].join(" "));

    expect(proposal.response.ok(), proposal.body ? JSON.stringify(proposal.body) : proposal.raw).toBeTruthy();
    expect(proposal.body?.data?.mode).toBe("proposal");
    expect(proposal.body?.data?.payload?.tools_used).toEqual(["create_team", "write_file"]);
    await expect(page.getByText("I can start that.").last()).toBeVisible({ timeout: 30_000 });
    await expect(page.getByText(teamID).last()).toBeVisible();
    await expect(page.getByText(entrypoint).last()).toBeVisible();

    const confirmed = await confirmProposal(page);
    expect(confirmed.response.ok(), confirmed.body ? JSON.stringify(confirmed.body) : confirmed.raw).toBeTruthy();
    expect(confirmed.body?.data?.verified).toBeTruthy();
    expect(confirmed.body?.data?.execution_state).toBe("verified");
    expect(confirmed.body?.data?.run_id).toBeTruthy();
    const projectPackage = (confirmed.body?.data?.execution_summary?.outputs ?? [])
      .find((output) => output.kind === "project_package");
    expect(projectPackage).toBeTruthy();
    expectProjectPackageMetadata(projectPackage!, { title, entrypoint, folder });
    expect(projectPackage?.validation).toMatch(/movement|collision|hazards|key|door|win\/fail|restart|music|audio/i);
    expect(projectPackage?.files ?? []).toEqual(expect.arrayContaining(["README.md", "PROOF.md"]));
    for (const supportFile of ["README.md", "PROOF.md"]) {
      const supportResponse = await liveAPIGet(
        page,
        `/api/v1/workspace/files/view?path=${encodeURIComponent(`${folder}/${supportFile}`)}`,
      );
      const parsed = await parseJSONIfPossible(supportResponse);
      expect(supportResponse.ok(), parsed.raw).toBeTruthy();
      expect(parsed.raw).toContain(title);
    }

    await expect(page.getByText("P0 Console Browser Game Team").last()).toBeVisible();
    await expect(page.getByText(`groups/${teamID}`).last()).toBeVisible();
    const gamePagePromise = page.context().waitForEvent("page", { timeout: chatTimeoutMs });
    await page.getByRole("button", { name: /Open file P0 Console Browser Game Team/i }).last().click();
    const gamePage = await gamePagePromise;
    await gamePage.waitForLoadState("domcontentloaded");
    await expect(gamePage).toHaveTitle(title);
    await expect(gamePage.locator("#game")).toBeVisible({ timeout: 30_000 });
    await expect(gamePage.locator("#health")).toHaveText("4");
    await expect(gamePage.locator("#keyState")).toHaveText("No");
    await expect(gamePage.locator("#score")).toHaveText("0");
    await expect(gamePage.locator("#goalState")).toHaveText("Find key");
    await expect(gamePage.getByRole("button", { name: "Restart" })).toBeVisible();
    await expect(gamePage.getByRole("button", { name: "Sound on" })).toBeVisible();

    const html = await gamePage.content();
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

    const before = await canvasSignature(gamePage);
    await gamePage.locator("#game").click();
    await gamePage.keyboard.down("ArrowRight");
    await gamePage.waitForTimeout(500);
    await gamePage.keyboard.up("ArrowRight");
    const after = await canvasSignature(gamePage);
    expect(after).not.toBe(before);
    await gamePage.close();
  });
});
