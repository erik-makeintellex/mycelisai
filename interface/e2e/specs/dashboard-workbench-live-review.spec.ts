import { expect, test, type Page } from "@playwright/test";

const chatPlaceholder = /Tell Soma what you want to plan, review, create, or run/i;

type ConfirmActionBody = {
  ok?: boolean;
  error?: string;
  data?: {
    run_id?: string;
    proof_artifact_id?: string;
    verified?: boolean;
    execution_state?: string;
    execution_summary?: {
      outputs?: Array<{ id?: string; text?: string; title?: string; url?: string } | string>;
      audit_recovery?: unknown;
      next_step?: unknown;
      proof?: unknown;
    };
  };
};

async function pageScrollMetrics(page: Page) {
  return page.evaluate(() => ({
    scrollY: window.scrollY,
    viewportHeight: window.innerHeight,
    documentHeight: document.documentElement.scrollHeight,
  }));
}

async function expectFreshDashboardWithoutStaleContent(page: Page) {
  await page.goto("/dashboard?fresh=1", { waitUntil: "domcontentloaded" });
  await expect(page.getByTestId("soma-operating-surface")).toBeVisible({ timeout: 20_000 });
  await expect(page.getByTestId("soma-environment-entry")).toHaveCount(0);
  await expect(page.getByTestId("central-soma-chat-frame")).toBeVisible();
  await expect(page.getByTestId("soma-conversation-thread").getByText(chatPlaceholder)).toBeVisible();
  await expect(page.getByTestId("output-workbench")).toHaveCount(0);
  await expect(page.getByTestId("soma-workbench-output-digest")).toHaveCount(0);
  await expect(page.getByTestId("soma-workbench-panel-toggle")).toHaveCount(0);
  await expect(page.getByTestId("focused-team-output-dock")).toHaveCount(0);
  await expect(page.getByRole("img", { name: /generated|retained|media|artifact/i })).toHaveCount(0);
  await expect(page.getByText(/Latest:|Open file|Run proof \+ retained output/i)).toHaveCount(0);
}

function outputMatchesTarget(body: ConfirmActionBody, targetPath: string) {
  return (body.data?.execution_summary?.outputs ?? []).some((output) => {
    if (typeof output === "string") return output.includes(targetPath);
    return [output.id, output.text, output.title, output.url].some((value) => value?.includes(targetPath));
  });
}

test.describe("Dashboard workbench live review", () => {
  test("fresh business-owner dashboard stays clean and approved proposal creates a retained output", async ({ page }, testInfo) => {
    test.skip(!process.env.PLAYWRIGHT_LIVE_BACKEND, "requires the live local Core/Interface stack");
    test.setTimeout(180_000);

    const stamp = Date.now();
    const targetPath = `generated/business-owner-flow-${stamp}/owner-note.md`;

    await expectFreshDashboardWithoutStaleContent(page);
    await page.screenshot({ path: testInfo.outputPath("fresh-dashboard-clean.png"), fullPage: true });

    const input = page.getByPlaceholder(chatPlaceholder);
    await input.fill(
      [
        `Create a markdown file at ${targetPath}.`,
        'Put exactly "# Business Owner Flow\n\nThe approval path must return output or one clear recovery action."',
        "Return retained output and proof.",
      ].join(" "),
    );
    const chatResponse = page.waitForResponse(
      (response) => response.url().includes("/api/v1/chat") && response.request().method() === "POST",
      { timeout: 120_000 },
    );
    await input.press("Enter");
    const chat = await chatResponse;
    const chatRaw = await chat.text();
    expect(chat.ok(), chatRaw).toBeTruthy();

    await expect(page.getByText("I can start that.").last()).toBeVisible({ timeout: 45_000 });

    const confirmResponse = page.waitForResponse(
      (response) => response.url().includes("/api/v1/intent/confirm-action") && response.request().method() === "POST",
      { timeout: 120_000 },
    );
    await input.fill("approve");
    await input.press("Enter");
    await expect(page.getByText(/Starting|Action completed|Result saved|work bus/i).last()).toBeVisible({ timeout: 15_000 });
    const confirmed = await confirmResponse;
    const confirmedRaw = await confirmed.text();
    const confirmedBody = JSON.parse(confirmedRaw) as ConfirmActionBody;

    expect(confirmed.ok(), confirmedRaw).toBeTruthy();
    expect(confirmedBody.data?.run_id, confirmedRaw).toBeTruthy();
    expect(
      outputMatchesTarget(confirmedBody, targetPath) || Boolean(confirmedBody.data?.proof_artifact_id),
      confirmedRaw,
    ).toBeTruthy();
    await expect(page.getByText(/Action completed|Result saved|Latest output|retained output|verified/i).last()).toBeVisible({ timeout: 45_000 });
    const resultStage = page.getByTestId("soma-result-first-stage");
    await expect(resultStage).toBeVisible({ timeout: 30_000 });
    await expect(resultStage.getByTestId("soma-continuation-sidecar")).toBeVisible();
    await expect(page.getByTestId("soma-workbench-output-digest")).toHaveCount(0);
    await expect(page.getByTestId("soma-workbench-panel-toggle")).toHaveCount(0);
    const workbench = resultStage.getByTestId("output-workbench");
    await expect(workbench.getByText(/owner-note\.md/i).last()).toBeVisible({ timeout: 30_000 });
    await expect(workbench.getByText(/Latest output/i).last()).toBeVisible({ timeout: 15_000 });
    await expect(workbench.getByRole("button", { name: /Open file .*owner-note\.md/i }).last()).toBeVisible({ timeout: 15_000 });
    await workbench.getByText("Details and proof").last().click();
    await expect(workbench.getByRole("button", { name: /Open local folder .*owner-note\.md/i }).last()).toBeVisible({ timeout: 15_000 });
  });

  test("uses Soma and keeps retained content, proof, and continuation in the result-first stage", async ({ page }, testInfo) => {
    test.skip(!process.env.PLAYWRIGHT_LIVE_BACKEND, "requires the live local Core/Interface stack");
    test.setTimeout(180_000);

    const stamp = Date.now();
    const targetPath = `generated/workbench-review-${stamp}/operator-note.md`;

    await page.goto("/dashboard", { waitUntil: "domcontentloaded" });
    await expect(page.getByTestId("soma-operating-surface")).toBeVisible();
    await expect(page.getByTestId("soma-environment-entry")).toHaveCount(0);
    await expect(page.getByTestId("central-soma-chat-frame")).toBeVisible();
    await expect(page.getByTestId("soma-workbench-panel-toggle")).toHaveCount(0);
    await expect(page.getByTestId("soma-workbench-side-rail")).toHaveCount(0);
    await page.screenshot({ path: testInfo.outputPath("dashboard-initial.png"), fullPage: true });

    const initialPageMetrics = await pageScrollMetrics(page);
    console.log("initial page metrics", initialPageMetrics);

    await page.mouse.wheel(0, 900);
    const afterPageWheel = await pageScrollMetrics(page);
    console.log("after page wheel metrics", afterPageWheel);
    await page.evaluate(() => window.scrollTo(0, 0));

    const input = page.getByPlaceholder(chatPlaceholder);
    await input.fill(
      [
        `Create a markdown file at ${targetPath}.`,
        'containing "# Workbench Review Note\n\n- Keep generated output visible near Soma.\n- Keep proof and folder access beside the output.\n- Keep prior work in an internal rail instead of page scroll."',
        "Return the retained output with proof and folder access.",
      ].join(" "),
    );
    const chatResponse = page.waitForResponse(
      (response) => response.url().includes("/api/v1/chat") && response.request().method() === "POST",
      { timeout: 120_000 },
    );
    await input.press("Enter");
    const chat = await chatResponse;
    expect(chat.ok(), await chat.text()).toBeTruthy();

    await expect(page.getByText("I can start that.").last()).toBeVisible({ timeout: 45_000 });
    await page.screenshot({ path: testInfo.outputPath("dashboard-proposal.png"), fullPage: true });

    const confirmResponse = page.waitForResponse(
      (response) => response.url().includes("/api/v1/intent/confirm-action") && response.request().method() === "POST",
      { timeout: 120_000 },
    );
    await input.fill("approve");
    await input.press("Enter");
    await expect(page.getByText(/Starting|Action completed|Result saved|work bus/i).last()).toBeVisible({ timeout: 15_000 });
    const confirmed = await confirmResponse;
    const confirmedRaw = await confirmed.text();
    expect(confirmed.ok(), confirmedRaw).toBeTruthy();
    const confirmedBody = JSON.parse(confirmedRaw) as {
      data?: {
        proof_artifact_id?: string;
        execution_summary?: {
          outputs?: Array<{
            id?: string;
            proof_artifact_id?: string;
            proof?: {
              proof_id?: string;
              path_boundary_status?: string;
              readback_status?: string;
              checksum_algorithm?: string;
              checksum?: string;
            };
          }>;
        };
      };
    };
    const outputRef = confirmedBody.data?.execution_summary?.outputs?.find((output) => output.id === targetPath);
    expect(outputRef, confirmedRaw).toBeTruthy();
    expect(outputRef?.proof_artifact_id).toBe(confirmedBody.data?.proof_artifact_id);
    expect(outputRef?.proof?.proof_id).toBe(confirmedBody.data?.proof_artifact_id);
    expect(outputRef?.proof?.path_boundary_status).toBe("verified");
    expect(outputRef?.proof?.readback_status).toBe("verified");
    expect(outputRef?.proof?.checksum_algorithm).toBe("sha256");
    expect(outputRef?.proof?.checksum).toMatch(/^[a-f0-9]{64}$/);

    await expect(page.getByText(/Result saved|Latest output/i).last()).toBeVisible({ timeout: 45_000 });
    const resultStage = page.getByTestId("soma-result-first-stage");
    await expect(resultStage).toBeVisible({ timeout: 30_000 });
    await expect(resultStage.getByTestId("soma-continuation-sidecar")).toBeVisible();
    await expect(page.getByTestId("soma-workbench-output-digest")).toHaveCount(0);
    await expect(page.getByTestId("soma-workbench-panel-toggle")).toHaveCount(0);
    const workbench = resultStage.getByTestId("output-workbench");
    await expect(workbench.getByText(/operator-note\.md/i).last()).toBeVisible({ timeout: 30_000 });
    await expect(workbench.getByText(/Latest output/i).last()).toBeVisible({ timeout: 15_000 });
    await expect(workbench.getByRole("button", { name: /Open file .*operator-note\.md/i }).last()).toBeVisible({ timeout: 15_000 });
    await workbench.getByText("Details and proof").last().click();
    const folderButton = workbench.getByRole("button", { name: /Open local folder .*operator-note\.md/i }).last();
    await expect(folderButton).toBeVisible({ timeout: 15_000 });
    await folderButton.click();
    await expect(folderButton).toContainText(/Folder opened|Folder blocked/i, { timeout: 15_000 });

    await expect(workbench.getByText(/Guided proposal/i)).toHaveCount(0);
    await workbench.getByText("Verification details").last().click();
    await expect(workbench.getByText("path verified").last()).toBeVisible({ timeout: 30_000 });
    await expect(workbench.getByText("readback verified").last()).toBeVisible({ timeout: 30_000 });
    await expect(workbench.getByText(/sha256 [a-f0-9]{12}/i).last()).toBeVisible({ timeout: 30_000 });
    await page.screenshot({ path: testInfo.outputPath("dashboard-after-output.png"), fullPage: true });
  });
});
