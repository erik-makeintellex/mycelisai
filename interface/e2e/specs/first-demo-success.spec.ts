import { expect, test } from "@playwright/test";
import {
  expectProjectPackageVisible,
  firstDemoPackageProposal,
  fulfillJSON,
  type ArtifactRecord,
  type GroupRecord,
} from "../support/finalization-browser-package";
import {
  mockOrganizationWorkspace,
  openOrganization,
  sendWorkspaceMessage,
} from "../support/soma-ui-testing";

test.describe.configure({ mode: "serial" });

test.describe("Canonical first-demo success path", () => {
  test("mocked pre-integration proof retains package metadata, proof link, reload, and Groups output", async ({ page }) => {
    test.setTimeout(90_000);

    const runId = "run-first-demo-package-success";
    const folder = "workspace/generated/coin-runner";
    const entrypoint = `${folder}/index.html`;
    const packageTitle = "Coin Runner Game";
    const groups: GroupRecord[] = [{
      group_id: "group-first-demo-package-success",
      name: "First Demo Game Team retained output",
      work_mode: "propose_only",
      status: "active",
      team_ids: ["first-demo-game-team"],
    }];
    const groupOutputs: ArtifactRecord[] = [{
      id: "artifact-first-demo-package-success",
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
        validation: "Mocked success proof: browser game opened, README was present, and validation notes remained retained.",
      },
      status: "approved",
      created_at: "2026-05-17T12:00:00Z",
    }];

    await mockOrganizationWorkspace(page, () => firstDemoPackageProposal());
    await page.route("**/api/v1/intent/confirm-action", async (route) => {
      await fulfillJSON(route, 200, {
        ok: true,
        data: {
          run_id: runId,
          verified: true,
          execution_state: "verified",
          execution_summary: {
            execution: {
              shape: "directed_execution",
              status: "verified",
              summary: "Mocked first-demo success retained the project package.",
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
              validation: "Mocked success proof: browser game opened, README was present, and validation notes remained retained.",
            }],
            proof: { run_id: runId, proof_class: "execution_run", verified: true },
          },
        },
      });
    });
    await page.route("**/api/v1/groups", async (route) => {
      await fulfillJSON(route, 200, { ok: true, data: groups });
    });
    await page.route("**/api/v1/groups/monitor", async (route) => {
      await fulfillJSON(route, 200, { ok: true, data: { online: true, last_group_id: groups[0].group_id } });
    });
    await page.route(/\/api\/v1\/groups\/([^/]+)\/outputs\?limit=8$/, async (route) => {
      await fulfillJSON(route, 200, { ok: true, data: groupOutputs });
    });
    await page.route("**/api/v1/mcp/servers", async (route) => {
      await fulfillJSON(route, 200, [{
        id: "filesystem-server",
        name: "filesystem",
        status: "connected",
        transport: "stdio",
        tools: [
          { id: "list-directory", name: "list_directory", description: "List workspace files" },
          { id: "read-text-file", name: "read_text_file", description: "Read workspace files" },
        ],
      }]);
    });
    await page.route("**/api/v1/mcp/servers/filesystem-server/tools/*/call", async (route) => {
      const tool = route.request().url().match(/\/tools\/([^/]+)\/call$/)?.[1] ?? "";
      const body = route.request().postDataJSON() as { arguments?: { path?: string } };
      const requestedPath = body.arguments?.path ?? "";
      if (tool === "list_directory" && requestedPath === folder) {
        await fulfillJSON(route, 200, { content: [{ type: "text", text: "[FILE] index.html\n[FILE] README.md\n[FILE] validation-notes.md" }] });
        return;
      }
      if (tool === "read_text_file") {
        await fulfillJSON(route, 200, { content: [{ type: "text", text: "# Coin Runner Game\nMocked package README." }] });
        return;
      }
      await fulfillJSON(route, 404, { error: `unexpected filesystem tool call: ${tool}` });
    });
    await page.context().route(/\/api\/v1\/workspace\/files\/view(?:\?.*)?$/, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "text/html",
        body: "<!doctype html><title>Coin Runner Game</title><h1>Coin Runner Game</h1><canvas id=\"game\"></canvas><p>README.md validation-notes.md</p>",
      });
    });

    await openOrganization(page);
    await sendWorkspaceMessage(page, "Create the exact first-demo playable browser game package with README and validation notes.");
    await expect(page.getByText("RUN CONFIRMATION").last()).toBeVisible({ timeout: 20_000 });
    await page.getByRole("button", { name: /Approve & Execute|Execute|Run/i }).last().click();

    await expectProjectPackageVisible(page, { title: packageTitle, entrypoint, folder });
    await expect(page.locator(`a[href="/runs/${runId}"]`).first()).toBeVisible();
    await page.reload({ waitUntil: "domcontentloaded" });
    await expectProjectPackageVisible(page, { title: packageTitle, entrypoint, folder });

    const outputPagePromise = page.context().waitForEvent("page");
    await page.getByRole("button", { name: `Open file ${packageTitle} in a new browser window` }).last().click();
    const outputPage = await outputPagePromise;
    await outputPage.waitForLoadState("domcontentloaded").catch(() => undefined);
    if (!outputPage.url().includes("/api/v1/workspace/files/view")) {
      await outputPage.goto(`/api/v1/workspace/files/view?path=${encodeURIComponent(entrypoint)}`, { waitUntil: "domcontentloaded" });
    }
    await expect(outputPage).toHaveTitle(packageTitle);
    await expect(outputPage.locator("canvas#game")).toBeVisible();
    await outputPage.reload({ waitUntil: "domcontentloaded" });
    await expect(outputPage.locator("body")).toContainText("validation-notes.md");
    await outputPage.close();

    await page.evaluate(() => window.localStorage.setItem("mycelis-advanced-mode", "true"));
    await page.getByRole("button", { name: /Review output/i }).last().click();
    await page.getByRole("link", { name: `Open ${packageTitle} in Resources` }).last().click();
    await expect(page).toHaveURL(new RegExp(`/resources\\?tab=workspace&path=${encodeURIComponent(folder)}`));
    await expect(page.getByRole("heading", { name: "Resources" })).toBeVisible({ timeout: 20_000 });
    await expect(page.getByText(folder).last()).toBeVisible();
    await expect(page.getByText("index.html")).toBeVisible();
    await expect(page.getByText("README.md")).toBeVisible();

    await page.goto("/groups?group_id=group-first-demo-package-success&advanced=1", { waitUntil: "domcontentloaded" });
    await expect(page.getByRole("heading", { name: "First Demo Game Team retained output" })).toBeVisible({ timeout: 20_000 });
    await page.getByRole("tab", { name: /Outputs/i }).click();
    await expect(page.getByText("Project package")).toBeVisible();
    await expect(page.getByText(entrypoint)).toBeVisible();
    await expect(page.getByText(folder, { exact: true })).toBeVisible();
    await expect(page.getByText("README.md")).toBeVisible();
    await expect(page.getByText(/Mocked success proof/i)).toBeVisible();
  });
});
