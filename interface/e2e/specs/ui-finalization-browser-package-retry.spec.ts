import { expect, test } from "@playwright/test";
import {
  firstDemoPackageProposal,
  fulfillJSON,
  type ArtifactRecord,
  type GroupRecord,
  expectProjectPackageVisible,
} from "../support/finalization-browser-package";
import {
  mockOrganizationWorkspace,
  openOrganization,
  sendWorkspaceMessage,
  type ChatRequestBody,
} from "../support/soma-ui-testing";

test.describe("UI finalization first-demo degraded retry proof", () => {
  test("mocked package failure preserves failed proof, then retry retains package metadata", async ({ page }) => {
    test.setTimeout(90_000);
    let confirmCalls = 0;
    const failedRunId = "run-first-demo-package-failed";
    const retryRunId = "run-first-demo-package-retry";
    const folder = "workspace/generated/coin-runner";
    const entrypoint = `${folder}/index.html`;
    const packageTitle = "Coin Runner Game";
    const groups: GroupRecord[] = [{
      group_id: "group-first-demo-package",
      name: "First Demo Game Team temporary workflow",
      work_mode: "propose_only",
      status: "active",
      team_ids: ["first-demo-game-team"],
    }];
    const groupOutputs: ArtifactRecord[] = [{
      id: "artifact-first-demo-package",
      title: packageTitle,
      artifact_type: "project_package",
      team_id: "first-demo-game-team",
      agent_id: "first-demo-game-team",
      content_type: "application/vnd.mycelis.project+json",
      file_path: entrypoint,
      metadata: {
        entrypoint,
        folder,
        files: ["index.html", "README.md", "validation-notes.md"],
        validation: "Mocked retry proof: live forced failure is not yet available, but the browser game opened and retained metadata remained reviewable.",
      },
      status: "approved",
      created_at: "2026-05-16T20:20:00Z",
    }];

    await mockOrganizationWorkspace(page, (_requestBody: ChatRequestBody) => firstDemoPackageProposal());
    await page.route("**/api/v1/intent/confirm-action", async (route) => {
      confirmCalls += 1;
      if (confirmCalls === 1) {
        await fulfillJSON(route, 500, {
          ok: false,
          error: "mocked package write failed",
          data: {
            run_id: failedRunId,
            execution_summary: {
              execution: {
                shape: "directed_execution",
                status: "failed",
                summary: "Mocked forced-failure gap coverage: live package forced failure is not currently supported.",
              },
              capability_use: [{ id: "write_file", label: "write_file", status: "failed" }],
              outputs: [],
              proof: { run_id: failedRunId, proof_class: "failed_run", verified: false },
              audit_recovery: {
                approval_status: "approved",
                recovery_state: "blocked",
                blocker: "mocked package write failed",
                retryable: true,
                degradation: {
                  code: "mocked_first_demo_package_forced_failure_gap",
                  what_failed: "mocked package write failed",
                  trusted_state: "The failed run record remains trusted, but no package output is trusted.",
                  invalidated_proof: "README, validation notes, entrypoint, folder, and files metadata were not retained by the failed run.",
                  safe_continuation: "Resubmit the same package request and approve the retry.",
                  requires_attention: true,
                },
              },
            },
          },
        });
        return;
      }
      await fulfillJSON(route, 200, {
        ok: true,
        data: {
          run_id: retryRunId,
          verified: true,
          execution_state: "verified",
          execution_summary: {
            execution: {
              shape: "directed_execution",
              status: "verified",
              summary: "Mocked retry retained the first-demo project package.",
            },
            outputs: [{
              kind: "project_package",
              title: packageTitle,
              id: folder,
              href: `/api/v1/workspace/files/view?path=${encodeURIComponent(entrypoint)}`,
              retained: true,
              entrypoint,
              folder,
              files: ["index.html", "README.md", "validation-notes.md"],
              validation: "Mocked retry proof: live forced failure is not yet available, but the browser game opened and retained metadata remained reviewable.",
            }],
            proof: { run_id: retryRunId, proof_class: "execution_run", verified: true },
          },
        },
      });
    });

    await page.route("**/api/v1/groups", async (route) => {
      await fulfillJSON(route, 200, { ok: true, data: groups });
    });
    await page.route(/\/api\/v1\/groups\/([^/]+)\/outputs\?limit=8$/, async (route) => {
      await fulfillJSON(route, 200, { ok: true, data: groupOutputs });
    });
    await page.context().route(/\/api\/v1\/workspace\/files\/view(?:\?.*)?$/, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "text/html",
        body: "<!doctype html><title>Coin Runner Game</title><h1>Coin Runner Game</h1><canvas id=\"game\"></canvas><p>README.md validation-notes.md</p>",
      });
    });

    const firstDemoAsk = "Create the exact first-demo playable browser game package with README and validation notes.";
    await openOrganization(page);
    await sendWorkspaceMessage(page, firstDemoAsk);
    await expect(page.getByText("PROPOSED ACTION").last()).toBeVisible({ timeout: 20_000 });
    await page.getByRole("button", { name: /Approve & Execute|Execute|Run/i }).last().click();
    const failureCard = page.getByTestId("execution-summary-card").last();
    await expect(failureCard.getByText("Needs review").first()).toBeVisible({ timeout: 20_000 });
    await expect(failureCard.getByText("Review request, proof, and recovery")).toBeVisible();
    await failureCard.getByText("Review request, proof, and recovery").click();
    await expect(page.getByText("mocked package write failed").last()).toBeVisible();
    await expect(page.getByText("Still available: The failed run record remains trusted, but no package output is trusted.").last()).toBeVisible();
    await expect(page.getByText("Not reliable: README, validation notes, entrypoint, folder, and files metadata were not retained by the failed run.").last()).toBeVisible();
    await expect(page.getByText("Safe next: Resubmit the same package request and approve the retry.").last()).toBeVisible();
    await expect(page.getByText("Run proof + retained output")).toHaveCount(0);

    await sendWorkspaceMessage(page, firstDemoAsk);
    await expect(page.getByText("PROPOSED ACTION").last()).toBeVisible({ timeout: 20_000 });
    await page.getByRole("button", { name: /Approve & Execute|Execute|Run/i }).last().click();
    await expectProjectPackageVisible(page, { title: packageTitle, entrypoint, folder });
    await expect(page.locator(`a[href="/runs/${retryRunId}"]`).first()).toBeVisible();

    const outputPagePromise = page.context().waitForEvent("page");
    await page.getByRole("button", { name: `Open Game ${packageTitle} in a new browser window` }).last().click();
    const outputPage = await outputPagePromise;
    await outputPage.waitForLoadState("domcontentloaded").catch(() => undefined);
    if (!outputPage.url().includes("/api/v1/workspace/files/view")) {
      await outputPage.goto(`/api/v1/workspace/files/view?path=${encodeURIComponent(entrypoint)}`, { waitUntil: "domcontentloaded" });
    }
    await expect(outputPage).toHaveTitle(packageTitle);
    await expect(outputPage.locator("body")).toContainText("README.md");
    await outputPage.reload({ waitUntil: "domcontentloaded" });
    await expect(outputPage.locator("body")).toContainText("validation-notes.md");
    await outputPage.close();

    await page.goto("/groups?group_id=group-first-demo-package", { waitUntil: "domcontentloaded" });
    await expect(page.getByRole("heading", { name: "First Demo Game Team temporary workflow" })).toBeVisible({ timeout: 20_000 });
    await expect(page.getByText("Project package")).toBeVisible();
    await expect(page.getByText(entrypoint)).toBeVisible();
    await expect(page.getByText(folder, { exact: true })).toBeVisible();
    await expect(page.getByText("README.md")).toBeVisible();
    await expect(page.getByText(/Mocked retry proof: live forced failure is not yet available/i)).toBeVisible();
  });
});
