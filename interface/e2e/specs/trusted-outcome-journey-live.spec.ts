import { expect, test } from "@playwright/test";
import {
  type APIEnvelope,
  type GroupRecord,
  confirmProposal,
  createOrganization,
  expectProjectPackageVisible,
  liveAPIGet,
  liveTimeoutMs,
  openLiveWorkspace,
  parseJSONIfPossible,
  removeTarget,
  submitLiveWorkspaceChat,
  targetExists,
} from "../support/finalization-browser-package";
import { liveAPIHeaders, liveAPIURL } from "../support/live-api-auth";

type ConfirmData = {
  run_id?: string;
  run_status?: string;
  verified?: boolean;
  execution_state?: string;
  proof_artifact_id?: string;
  contract_id?: string;
  execution_summary?: { outputs?: Array<Record<string, unknown>> };
};

type ProofRecord = {
  id?: string;
  run_id?: string;
  status?: string;
  proof_class?: string;
  proof_quality?: string;
};

type ContractRecord = {
  id?: string;
  run_id?: string;
  status?: string;
  output_refs?: unknown[];
};

type RunEvent = { event_type?: string; payload?: Record<string, unknown> };
type ArtifactRecord = { title?: string; file_path?: string; artifact_type?: string };
type TeamOutputRef = {
  kind?: string;
  label?: string;
  storage_ref?: string;
  entrypoint?: string;
  proof_ref?: string;
};
type TeamWorkItem = {
  work_item_id?: string;
  team_id?: string;
  run_id?: string;
  execution_shape?: string;
  state?: string;
  degradation_state?: string;
  recovery_options?: string[];
  output_refs?: TeamOutputRef[];
};

test.describe("Trusted Outcome Journey live smoke", () => {
  test.skip(!process.env.PLAYWRIGHT_LIVE_BACKEND, "requires a live Core backend");
  test.setTimeout(liveTimeoutMs);

  test("proves the source-stack Ask to Revisit path with durable proof readback", async ({ page }) => {
    test.slow();
    const stamp = Date.now();
    const teamID = `trusted-outcome-live-${stamp}`;
    const teamName = `Trusted Outcome Live Team ${stamp}`;
    const folder = `groups/${teamID}/generated/first-game`;
    const entrypoint = `${folder}/index.html`;
    const packageTitle = `${teamName} First Playable`;
    const organizationID = await createOrganization(page, `Trusted Outcome Journey ${stamp}`);

    await openLiveWorkspace(page, organizationID);
    try {
      const proposal = await submitLiveWorkspaceChat(
        page,
        [
          `Create a team with team_id ${teamID} named ${teamName}.`,
          "Have that team build the exact first demo deliverable: a playable browser game project package.",
          `Retain it at ${folder} with entrypoint ${entrypoint}.`,
          `Use the package title ${packageTitle}.`,
          "The package metadata must include files index.html, README.md, PROOF.md, and validation notes from opening the browser game.",
          "The self-contained HTML must start its render loop and visibly move the player while ArrowRight is held.",
          "Read the saved entrypoint back and repair it before reporting completion if any required interaction or file is missing.",
          "After approval, return a retained project_package output with entrypoint, folder, files, validation, and proof.",
        ].join(" "),
      );

      expect(proposal.response.ok(), proposal.body ? JSON.stringify(proposal.body) : proposal.raw).toBeTruthy();
      expect(proposal.body?.data?.mode).toBe("proposal");
      await expect(page.getByRole("heading", { name: /Start this\?|Approve this\?/ }).last()).toBeVisible({ timeout: 30_000 });
      await expect(page.getByRole("button", { name: /^(Start|Approve)$/i }).last()).toBeVisible();
      await expect(page.getByText(teamID).last()).toBeVisible();
      await expect(page.getByText(entrypoint).last()).toBeVisible();
      expect(targetExists(entrypoint)).toBeFalsy();

      const confirmed = await confirmProposal(page);
      expect(confirmed.response.ok(), confirmed.body ? JSON.stringify(confirmed.body) : confirmed.raw).toBeTruthy();
      const data = confirmed.body?.data as ConfirmData | undefined;
      expect(data?.verified).toBeFalsy();
      expect(data?.execution_state).toBe("running");
      expect(data?.run_status).toBe("running");
      expect(data?.proof_artifact_id).toBeFalsy();
      expect(data?.run_id).toBeTruthy();

      const workItem = await waitForTeamDelivery(page, teamID, data!.run_id!);
      expect(workItem.state, JSON.stringify(workItem)).toBe("output_ready");
      const projectPackageRef = workItem.output_refs?.find((output) => output.kind === "project_package");
      expect(projectPackageRef, JSON.stringify(workItem.output_refs ?? [])).toBeTruthy();
      expect(projectPackageRef?.storage_ref).toBe(folder);
      expect(projectPackageRef?.entrypoint).toMatch(/index\.html$/);
      expect(projectPackageRef?.proof_ref).toBeTruthy();
      await expect.poll(() => targetExists(entrypoint), { timeout: 60_000 }).toBeTruthy();

      await page.goto(`/dashboard?team_id=${encodeURIComponent(teamID)}`, { waitUntil: "domcontentloaded" });
      await expectProjectPackageVisible(page, { title: packageTitle, entrypoint, folder });

      const outputPagePromise = page.context().waitForEvent("page");
      await page.getByRole("button", { name: new RegExp(`Open app .*${packageTitle}`, "i") }).last().click();
      const outputPage = await outputPagePromise;
      await outputPage.waitForLoadState("domcontentloaded");
      await expect(outputPage).toHaveTitle(packageTitle);
      const canvas = outputPage.locator("canvas").first();
      await expect(canvas).toBeVisible();
      const beforeMove = await canvas.screenshot();
      await outputPage.keyboard.down("ArrowRight");
      try {
        await expect.poll(async () => {
          const whileMoving = await canvas.screenshot();
          return whileMoving.equals(beforeMove);
        }, {
          timeout: 2_000,
          intervals: [100, 150, 250],
          message: "the generated browser package should visibly respond while ArrowRight is held",
        }).toBeFalsy();
      } finally {
        await outputPage.keyboard.up("ArrowRight");
      }
      await outputPage.close();

      await expectProofAndRunReadback(page, data!);
      const group = await expectGroupOutputReadback(page, teamID, packageTitle, entrypoint);
      await expectResourcesRevisit(page, folder, "index.html");
      await expectGroupsRevisit(page, group, packageTitle, entrypoint);
      await expectRunReceiptRevisit(page, data!.run_id!);
    } finally {
      await cleanupLiveJourney(page, teamID);
      removeTarget(entrypoint);
      removeTarget(`${folder}/README.md`);
      removeTarget(`${folder}/PROOF.md`);
      removeTarget(`${folder}/project-package.json`);
    }
  });
});

async function waitForTeamDelivery(page: import("@playwright/test").Page, teamID: string, runID: string) {
  let latest: TeamWorkItem | undefined;
  await expect.poll(async () => {
    const response = await liveAPIGet(page, `/api/v1/teams/${encodeURIComponent(teamID)}/work?limit=25`);
    if (!response.ok()) return `http_${response.status()}`;
    const items = ((await response.json()) as APIEnvelope<TeamWorkItem[]>).data ?? [];
    latest = items.find((item) => (
      item.run_id === runID
      && item.team_id === teamID
      && item.execution_shape === "delegated_work"
    ));
    return latest?.state ?? "missing";
  }, {
    timeout: 150_000,
    intervals: [500, 1_000, 2_000, 5_000],
    message: `team ${teamID} should finish or expose a recoverable terminal state`,
  }).toMatch(/^(output_ready|degraded|needs_operator)$/);

  expect(latest, `No correlated TeamWorkItem found for run ${runID}`).toBeTruthy();
  if (latest?.state !== "output_ready") {
    throw new Error(
      `Team delivery ended ${latest?.state}: ${latest?.degradation_state ?? "unknown"}; `
      + `recovery=${(latest?.recovery_options ?? []).join(" | ")}`,
    );
  }
  return latest;
}

async function cleanupLiveJourney(page: import("@playwright/test").Page, teamID: string) {
  const groupsResponse = await liveAPIGet(page, "/api/v1/groups");
  if (groupsResponse.ok()) {
    const groups = ((await groupsResponse.json()) as APIEnvelope<GroupRecord[]>).data ?? [];
    const group = groups.find((candidate) => candidate.team_ids?.includes(teamID));
    if (group) {
      await page.request.post(
        liveAPIURL(`/api/v1/groups/${encodeURIComponent(group.group_id)}/clear`),
        {
          headers: {
            ...(liveAPIHeaders() ?? {}),
            "Content-Type": "application/json",
          },
          data: { include_outputs: true },
        },
      );
    }
  }
  await page.request.delete(liveAPIURL(`/api/v1/teams/${encodeURIComponent(teamID)}`), {
    headers: liveAPIHeaders(),
  });
}

async function expectProofAndRunReadback(page: import("@playwright/test").Page, data: ConfirmData) {
  const runResponse = await liveAPIGet(page, `/api/v1/runs/${encodeURIComponent(data.run_id!)}`);
  expect(runResponse.ok(), await runResponse.text()).toBeTruthy();
  const run = ((await runResponse.json()) as APIEnvelope<{ status?: string }>).data;
  expect(run?.status).toBe("completed");

  const proofResponse = await liveAPIGet(page, `/api/v1/trust/proof-artifacts?run_id=${encodeURIComponent(data.run_id!)}&limit=10`);
  expect(proofResponse.ok(), await proofResponse.text()).toBeTruthy();
  const proofRecords = ((await proofResponse.json()) as APIEnvelope<ProofRecord[]>).data ?? [];
  expect(proofRecords.length, JSON.stringify(proofRecords)).toBeGreaterThan(0);
  if (data.proof_artifact_id) {
    expect(proofRecords.some((record) => record.id === data.proof_artifact_id)).toBeTruthy();
  }
  expect(proofRecords.some((record) => record.status === "success" || record.proof_quality === "verified")).toBeTruthy();

  const contractResponse = await liveAPIGet(page, `/api/v1/trust/execution-contracts?run_id=${encodeURIComponent(data.run_id!)}&limit=10`);
  expect(contractResponse.ok(), await contractResponse.text()).toBeTruthy();
  const contracts = ((await contractResponse.json()) as APIEnvelope<ContractRecord[]>).data ?? [];
  expect(contracts.length, JSON.stringify(contracts)).toBeGreaterThan(0);
  if (data.contract_id) expect(contracts.some((contract) => contract.id === data.contract_id)).toBeTruthy();
  expect(contracts.some((contract) => contract.status === "completed")).toBeTruthy();

  const eventResponse = await liveAPIGet(page, `/api/v1/runs/${encodeURIComponent(data.run_id!)}/events`);
  expect(eventResponse.ok(), await eventResponse.text()).toBeTruthy();
  const eventBody = await eventResponse.json() as APIEnvelope<RunEvent[]> | RunEvent[];
  const events = Array.isArray(eventBody) ? eventBody : eventBody.data ?? [];
  expect(Array.isArray(events), JSON.stringify(events)).toBeTruthy();
  expect(events.length, JSON.stringify(events)).toBeGreaterThan(0);
  expect(events.some((event) => event.event_type), JSON.stringify(events)).toBeTruthy();
}

async function expectGroupOutputReadback(
  page: import("@playwright/test").Page,
  teamID: string,
  packageTitle: string,
  entrypoint: string,
) {
  const groupsResponse = await liveAPIGet(page, "/api/v1/groups");
  const parsedGroups = await parseJSONIfPossible<APIEnvelope<GroupRecord[]>>(groupsResponse);
  expect(groupsResponse.ok(), parsedGroups.body ? JSON.stringify(parsedGroups.body) : parsedGroups.raw).toBeTruthy();
  const group = (parsedGroups.body?.data ?? []).find((candidate) => candidate.team_ids?.includes(teamID));
  expect(group, JSON.stringify(parsedGroups.body?.data ?? [])).toBeTruthy();

  const outputsResponse = await liveAPIGet(page, `/api/v1/groups/${encodeURIComponent(group!.group_id)}/outputs?limit=20`);
  expect(outputsResponse.ok(), await outputsResponse.text()).toBeTruthy();
  const outputs = ((await outputsResponse.json()) as APIEnvelope<ArtifactRecord[]>).data ?? [];
  expect(outputs.some((output) => output.title?.includes(packageTitle) || output.file_path === entrypoint)).toBeTruthy();
  return group!;
}

async function expectResourcesRevisit(page: import("@playwright/test").Page, folder: string, fileName: string) {
  const resourcesFolder = `workspace/${folder}`;
  await page.goto(`/resources?tab=workspace&path=${encodeURIComponent(resourcesFolder)}`, { waitUntil: "domcontentloaded" });
  await expect(page.getByRole("heading", { name: "Resources" })).toBeVisible({ timeout: 30_000 });
  await expect(page.getByText(resourcesFolder).last()).toBeVisible({ timeout: 30_000 });
  await expect(page.getByText(fileName).first()).toBeVisible({ timeout: 30_000 });
}

async function expectGroupsRevisit(
  page: import("@playwright/test").Page,
  group: GroupRecord,
  packageTitle: string,
  entrypoint: string,
) {
  await page.goto(`/groups?group_id=${encodeURIComponent(group.group_id)}&panel=outputs&advanced=1`, { waitUntil: "domcontentloaded" });
  await expect(page.getByRole("heading", { name: "Recent outputs" })).toBeVisible({ timeout: 30_000 });
  await page.getByRole("tab", { name: /Outputs/i }).click();
  await expect(page.getByText(packageTitle).first()).toBeVisible({ timeout: 30_000 });
  await expect(page.getByText(entrypoint).first()).toBeVisible();
}

async function expectRunReceiptRevisit(page: import("@playwright/test").Page, runID: string) {
  await page.goto(`/runs/${encodeURIComponent(runID)}?tab=events`, { waitUntil: "domcontentloaded" });
  await expect(page.getByLabel("Run receipt")).toBeVisible({ timeout: 30_000 });
  await expect(page.getByText(/What happened/i)).toBeVisible();
  await expect(page.getByText(/What to trust/i)).toBeVisible();
  await expect(page.getByText(/Next step/i)).toBeVisible();
}
