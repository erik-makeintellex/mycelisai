import { expect, type Page } from "@playwright/test";
import { storeLiveArtifactWithRetry } from "./live-artifacts";
import { createLiveMissionTeam } from "./live-teams";

export type APIEnvelope<T> = {
  ok?: boolean;
  data?: T;
  error?: string;
};

type OrganizationEnvelope = {
  data?: {
    id?: string;
    name?: string;
  };
};

type ChatEnvelope = {
  ok?: boolean;
  data?: {
    mode?: string;
    payload?: {
      text?: string;
      ask_class?: string;
    };
  };
};

export type TeamLeadWorkflowGroupDraft = {
  name: string;
  goal_statement: string;
  work_mode:
    | "read_only"
    | "propose_only"
    | "execute_with_approval"
    | "execute_bounded";
  coordinator_profile: string;
  allowed_capabilities?: string[];
  initial_member_count?: number;
  recommended_member_limit?: number;
  expansion_policy?: string;
  temporary_addition_guidance?: string;
  expiry_hours?: number;
  summary: string;
};

type TeamLeadExecutionContract = {
  execution_mode?: string;
  team_name?: string;
  coordination_model?: string;
  recommended_team_count?: number;
  initial_member_count?: number;
  recommended_team_member_limit?: number;
  target_outputs?: string[];
  workflow_group?: TeamLeadWorkflowGroupDraft;
};

type TeamLeadGuidanceResponse = {
  headline?: string;
  execution_contract?: TeamLeadExecutionContract;
};

export type GroupRecord = {
  group_id: string;
  name: string;
  goal_statement: string;
  status: string;
};

type ArtifactRecord = {
  id?: string;
  title?: string;
};

export async function parseJSONIfPossible<T>(response: {
  text(): Promise<string>;
}) {
  const raw = await response.text();
  try {
    return { raw, body: JSON.parse(raw) as T };
  } catch {
    return { raw, body: null as T | null };
  }
}

export async function expectExecutionContractOutputs(
  page: Page,
  outputs: string[],
) {
  const targetOutputsSection = page
    .getByText("Target outputs", { exact: true })
    .locator("..");
  for (const output of outputs) {
    await expect(
      targetOutputsSection.getByText(output, { exact: true }).first(),
    ).toBeVisible();
  }
}

export async function gotoWithColdStartRetry(page: Page, path: string) {
  try {
    await page.goto(path, { waitUntil: "domcontentloaded" });
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    const retryable = [
      "net::ERR_ABORTED",
      "net::ERR_CONNECTION_FAILED",
      "ECONNREFUSED",
      "frame was detached",
    ];
    if (!retryable.some((fragment) => message.includes(fragment))) {
      throw error;
    }
    await page.waitForTimeout(1_000);
    await page.goto(path, { waitUntil: "domcontentloaded" });
  }
}

export async function createOrganization(page: Page, name: string) {
  const response = await page.request.post("/api/v1/organizations", {
    data: {
      name,
      purpose: "Live workflow variant verification",
      start_mode: "empty",
    },
  });
  const parsed = await parseJSONIfPossible<OrganizationEnvelope>(response);
  expect(
    response.ok(),
    parsed.body ? JSON.stringify(parsed.body) : parsed.raw,
  ).toBeTruthy();
  expect(parsed.body?.data?.id).toBeTruthy();
  return {
    id: parsed.body!.data!.id!,
    name: parsed.body!.data!.name || name,
  };
}

export async function openWorkspace(page: Page, organizationId: string) {
  await gotoWithColdStartRetry(page, `/organizations/${organizationId}`);
  await page
    .getByPlaceholder(
      /Tell Soma what you want to plan, review, create, or (execute|run)/i,
    )
    .waitFor({ timeout: 30_000 });
}

export async function submitWorkspaceChat(page: Page, content: string) {
  const input = page.getByPlaceholder(
    /Tell Soma what you want to plan, review, create, or (execute|run)/i,
  );
  await input.fill(content);
  const responsePromise = page.waitForResponse(
    (response) => {
      const url = new URL(response.url());
      return (
        response.request().method() === "POST" &&
        url.pathname === "/api/v1/chat"
      );
    },
    { timeout: 120_000 },
  );
  await input.press("Enter");
  const response = await responsePromise;
  const parsed = await parseJSONIfPossible<ChatEnvelope>(response);
  return { response, raw: parsed.raw, body: parsed.body };
}

export async function openTeamCreation(page: Page, organizationId: string) {
  await gotoWithColdStartRetry(
    page,
    `/teams/create?organization_id=${encodeURIComponent(organizationId)}`,
  );
  await expect(
    page.getByRole("heading", { name: "Create a team through Soma" }),
  ).toBeVisible({ timeout: 30_000 });
}

export async function submitTeamDesign(
  page: Page,
  organizationId: string,
  prompt: string,
) {
  await page
    .getByLabel("Tell Soma what team or delivery lane you want to create")
    .fill(prompt);
  const [response] = await Promise.all([
    page.waitForResponse(
      (candidate) => {
        const url = new URL(candidate.url());
        return (
          candidate.request().method() === "POST" &&
          url.pathname ===
            `/api/v1/organizations/${organizationId}/workspace/actions`
        );
      },
      { timeout: 120_000 },
    ),
    page.getByRole("button", { name: "Start team design" }).click(),
  ]);
  const parsed =
    await parseJSONIfPossible<APIEnvelope<TeamLeadGuidanceResponse>>(response);
  expect(
    response.ok(),
    parsed.body ? JSON.stringify(parsed.body) : parsed.raw,
  ).toBeTruthy();
  expect(
    parsed.body?.ok,
    parsed.body ? JSON.stringify(parsed.body) : parsed.raw,
  ).toBe(true);
  expect(parsed.body?.data?.execution_contract).toBeTruthy();
  return parsed.body!.data!;
}

export async function createLiveTeamIDs(page: Page, count: number) {
  const ids: string[] = [];
  for (let index = 0; index < count; index += 1) {
    const liveTeam = await createLiveMissionTeam(page);
    ids.push(liveTeam.teamID);
  }
  return ids;
}

export async function createLiveGroup(
  page: Page,
  draft: TeamLeadWorkflowGroupDraft,
  teamIDs: string[],
) {
  const uniqueSuffix = `${Date.now()}-${Math.floor(Math.random() * 10_000)}`;
  const expiry =
    typeof draft.expiry_hours === "number" && draft.expiry_hours > 0
      ? new Date(Date.now() + draft.expiry_hours * 60 * 60 * 1000).toISOString()
      : new Date(Date.now() + 72 * 60 * 60 * 1000).toISOString();
  const response = await page.request.post("/api/v1/groups", {
    data: {
      name: `${draft.name} ${uniqueSuffix}`,
      goal_statement: draft.goal_statement,
      work_mode: draft.work_mode,
      allowed_capabilities: draft.allowed_capabilities ?? [],
      member_user_ids: ["owner"],
      team_ids: teamIDs,
      coordinator_profile: draft.coordinator_profile,
      approval_policy_ref: "browser-live-proof",
      expiry,
    },
  });
  const parsed = await parseJSONIfPossible<APIEnvelope<GroupRecord>>(response);
  expect(
    response.status(),
    parsed.body ? JSON.stringify(parsed.body) : parsed.raw,
  ).toBe(201);
  expect(
    response.ok(),
    parsed.body ? JSON.stringify(parsed.body) : parsed.raw,
  ).toBeTruthy();
  expect(
    parsed.body?.ok,
    parsed.body ? JSON.stringify(parsed.body) : parsed.raw,
  ).toBe(true);
  expect(parsed.body?.data?.group_id).toBeTruthy();
  return parsed.body!.data!;
}

export async function storeLiveArtifact(
  page: Page,
  teamID: string,
  title: string,
  agentID: string,
  content: string,
) {
  return await storeLiveArtifactWithRetry<ArtifactRecord>(
    page,
    {
      team_id: teamID,
      agent_id: agentID,
      artifact_type: "document",
      title,
      content_type: "text/markdown",
      content,
      metadata: { source: "workflow-variants-live-backend.spec.ts" },
      status: "approved",
    },
    "workflow variants live backend",
  );
}
