import { expect, test, type Page, type Route } from "@playwright/test";
import {
  confirmProposal,
  createOrganization,
  liveTimeoutMs,
  openLiveWorkspace,
  submitLiveWorkspaceChat,
} from "../support/finalization-browser-package";
import {
  mockOrganizationWorkspace,
  openOrganization,
  sendWorkspaceMessage,
  type ChatRequestBody,
  type RouteResponse,
} from "../support/soma-ui-testing";

type ToolCallRecord = {
  tool: string;
  arguments: Record<string, unknown>;
};

const mediaTitle = "Local/private media storyboard frame";
const mediaFolder = "workspace/saved-media/media-team-proof";
const mediaPath = `${mediaFolder}/storyboard-frame.png`;
const mediaHref = `/api/v1/workspace/files/view?path=${encodeURIComponent(mediaPath)}`;

const tinyPng =
  "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMB/axX6fQAAAAASUVORK5CYII=";

test.skip(({ browserName }) => browserName !== "chromium", "Media retained-output proof is stabilized in Chromium.");

async function fulfillJSON(route: Route, status: number, body: unknown) {
  await route.fulfill({ status, contentType: "application/json", body: JSON.stringify(body) });
}

function mediaTeamProposal(): RouteResponse {
  return {
    status: 200,
    body: {
      ok: true,
      data: {
        meta: { source_node: "admin", timestamp: "2026-05-27T12:00:00Z" },
        signal_type: "chat.reply",
        trust_score: 0.91,
        template_id: "chat-to-proposal",
        mode: "proposal",
        payload: {
          text: "I can create a local/private media team output and retain it for review.",
          tools_used: ["create_team", "generate_image", "save_cached_image"],
          consultations: [],
          artifacts: [],
          proposal: {
            intent: "create_local_private_media_team_output",
            operator_summary: "create a media team, generate an image through the local/private media provider, and save it into workspace output files.",
            expected_result: `A retained media output will be available at ${mediaPath}.`,
            affected_resources: [mediaPath],
            teams: 1,
            agents: 2,
            tools: ["create_team", "generate_image", "save_cached_image"],
            risk_level: "medium",
            confirm_token: "confirm-media-retained-output",
            intent_proof_id: "proof-media-retained-output",
            approval_required: false,
            approval_reason: "capability_risk",
            approval_mode: "optional",
            capability_risk: "medium",
            capability_ids: ["create_team", "generate_image", "save_cached_image"],
            external_data_use: false,
            estimated_cost: 0,
          },
        },
      },
    },
  };
}

async function mockRetainedMediaExecution(page: Page) {
  const revealCalls: string[] = [];

  await page.route("**/api/v1/intent/confirm-action", async (route) => {
    await fulfillJSON(route, 200, {
      ok: true,
      data: {
        run_id: "run-media-retained-output",
        verified: true,
        execution_state: "verified",
        execution_summary: {
          execution: {
            shape: "directed_execution",
            status: "verified",
            summary: "Local/private media team retained the generated image for operator review.",
          },
          capability_use: {
            tools: [
              { id: "create_team", label: "create_team", status: "verified" },
              { id: "generate_image", label: "generate_image", status: "verified" },
              { id: "save_cached_image", label: "save_cached_image", status: "verified" },
            ],
          },
          outputs: [
            {
              kind: "file",
              title: mediaTitle,
              id: mediaPath,
              href: mediaHref,
              retained: true,
            },
          ],
          proof: {
            run_id: "run-media-retained-output",
            proof_class: "execution_run",
            verified: true,
          },
        },
      },
    });
  });

  await page.route("**/api/v1/workspace/files/reveal?*", async (route) => {
    revealCalls.push(route.request().url());
    await fulfillJSON(route, 200, {
      ok: true,
      data: { workspace_path: mediaPath, folder_path: mediaFolder },
    });
  });

  await page.route("**/api/v1/workspace/files/view?*", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "image/png",
      body: Buffer.from(tinyPng, "base64"),
    });
  });

  return revealCalls;
}

async function mockOutputFilesMCP(page: Page) {
  const calls: ToolCallRecord[] = [];

  await page.route("**/api/v1/mcp/servers", async (route) => {
    await fulfillJSON(route, 200, [
      {
        id: "filesystem-server",
        name: "filesystem",
        status: "connected",
        transport: "stdio",
        tools: [
          { id: "list-directory", name: "list_directory", description: "List workspace files" },
          { id: "read-text-file", name: "read_text_file", description: "Read workspace files" },
        ],
      },
    ]);
  });

  await page.route("**/api/v1/mcp/servers/filesystem-server/tools/*/call", async (route) => {
    const tool = route.request().url().match(/\/tools\/([^/]+)\/call$/)?.[1] ?? "";
    const body = (route.request().postDataJSON() ?? {}) as { arguments?: Record<string, unknown> };
    const args = body.arguments ?? {};
    calls.push({ tool, arguments: args });

    const currentPath = String(args.path ?? "workspace").replaceAll("\\", "/").replace(/\/$/, "");
    let listing = "";
    if (tool === "list_directory" && currentPath === "workspace") {
      listing = "[DIR] saved-media";
    } else if (tool === "list_directory" && currentPath === "workspace/saved-media") {
      listing = "[DIR] media-team-proof";
    } else if (tool === "list_directory" && currentPath === mediaFolder) {
      listing = "[FILE] storyboard-frame.png";
    }

    if (tool === "list_directory") {
      await fulfillJSON(route, 200, { content: [{ type: "text", text: listing }] });
      return;
    }
    if (tool === "read_text_file") {
      await fulfillJSON(route, 200, { content: [{ type: "text", text: "PNG retained media output" }] });
      return;
    }
    await fulfillJSON(route, 404, { error: `unexpected tool ${tool}` });
  });

  return calls;
}

async function enableAdvancedMode(page: Page) {
  await page.addInitScript(() => {
    window.localStorage.setItem("mycelis-advanced-mode", "true");
  });
}

test.describe("Soma media retained output proof", () => {
  test("mocked browser proof shows, opens, and locates a retained local/private media output", async ({ page }) => {
    const revealCalls = await mockRetainedMediaExecution(page);
    await mockOutputFilesMCP(page);
    await mockOrganizationWorkspace(page, (_requestBody: ChatRequestBody) => mediaTeamProposal());

    await openOrganization(page);
    await sendWorkspaceMessage(page, "Create a local/private media team output and retain the image for review.");

    await expect(page.getByText("PROPOSED ACTION").last()).toBeVisible({ timeout: 20_000 });
    await expect(page.getByText(mediaPath).last()).toBeVisible();
    await page.getByRole("button", { name: /Approve & Execute|Execute/i }).last().click();

    await expect(page.getByText("Run proof + retained output").last()).toBeVisible({ timeout: 20_000 });
    await expect(page.getByText(mediaTitle).last()).toBeVisible();
    await expect(page.getByRole("img", { name: mediaTitle })).toBeVisible();
    await expect(page.getByRole("link", { name: mediaTitle }).last()).toHaveAttribute("href", mediaHref);

    const outputPagePromise = page.context().waitForEvent("page");
    await page.getByRole("button", { name: `Open ${mediaTitle} in a new browser window` }).last().click();
    const outputPage = await outputPagePromise;
    await outputPage.waitForLoadState("domcontentloaded").catch(() => undefined);
    expect(outputPage.url()).toContain("/api/v1/workspace/files/view");
    expect(outputPage.url()).toContain(encodeURIComponent(mediaPath));
    await outputPage.close();

    await page.getByRole("button", { name: `Open local folder for ${mediaTitle}` }).last().click();
    await expect.poll(() => revealCalls.some((url) => url.includes(encodeURIComponent(mediaPath)))).toBe(true);

    await enableAdvancedMode(page);
    await page.goto("/resources?tab=workspace", { waitUntil: "domcontentloaded" });
    await expect(page.getByRole("heading", { name: "Advanced Resources" })).toBeVisible({ timeout: 15_000 });
    await expect(page.getByRole("button", { name: /Output Files/i })).toBeVisible();
    await page.getByRole("button", { name: "Open folder saved-media" }).click();
    await page.getByRole("button", { name: "Open folder media-team-proof" }).click();
    await expect(page.getByText("storyboard-frame.png")).toBeVisible();
  });

  test("live Core creates a local/private media team output and exposes retained output controls", async ({ page }) => {
    test.skip(
      !process.env.PLAYWRIGHT_LIVE_MEDIA_RETAINED_OUTPUT,
      "requires live Core plus configured local/private media provider; set PLAYWRIGHT_LIVE_MEDIA_RETAINED_OUTPUT=1",
    );
    test.setTimeout(liveTimeoutMs);
    test.slow();

    const stamp = Date.now();
    const teamID = `qa-media-team-${stamp}`;
    const organizationId = await createOrganization(page, `QA Media Retained Output ${stamp}`);

    await openLiveWorkspace(page, organizationId);
    const proposal = await submitLiveWorkspaceChat(
      page,
      [
        `Create a compact local/private media team with team_id ${teamID}.`,
        "Generate one image with the configured local/private media provider.",
        "Save the retained output under saved-media and return the saved workspace file as a retained output.",
        "The response must expose the saved media output so the operator can open it and open its local folder.",
      ].join(" "),
    );
    expect(proposal.response.ok(), proposal.body ? JSON.stringify(proposal.body) : proposal.raw).toBeTruthy();
    expect(proposal.body?.data?.mode).toBe("proposal");
    await expect(page.getByText("PROPOSED ACTION").last()).toBeVisible({ timeout: 30_000 });

    const confirmed = await confirmProposal(page);
    expect(confirmed.response.ok(), confirmed.body ? JSON.stringify(confirmed.body) : confirmed.raw).toBeTruthy();
    expect(confirmed.body?.data?.verified).toBeTruthy();
    const outputs = confirmed.body?.data?.execution_summary?.outputs ?? [];
    const mediaOutput = outputs.find((output) => (
      output.retained
      && (output.kind === "file" || output.kind === "image")
      && typeof output.href === "string"
      && output.href.includes("/api/v1/workspace/files/view")
    ));
    expect(mediaOutput, JSON.stringify(outputs)).toBeTruthy();

    const outputLabel = mediaOutput!.title || mediaOutput!.id || "Team output";
    await expect(page.getByText("Run proof + retained output").last()).toBeVisible({ timeout: 30_000 });
    await expect(page.getByText(outputLabel).last()).toBeVisible();
    await expect(page.getByRole("button", { name: new RegExp(`Open .*${outputLabel.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}.*browser window`) }).last()).toBeVisible();
    await expect(page.getByRole("button", { name: `Open local folder for ${outputLabel}` }).last()).toBeVisible();
  });
});
