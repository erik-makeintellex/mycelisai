import { expect, test } from "@playwright/test";
import {
  type APIEnvelope,
  type GroupRecord,
  confirmProposal,
  createOrganization,
  expectProjectPackageMetadata,
  expectProjectPackageVisible,
  liveAPIGet,
  liveTimeoutMs,
  openLiveWorkspace,
  parseJSONIfPossible,
  removeTarget,
  submitLiveWorkspaceChat,
  targetExists,
} from "../support/finalization-browser-package";

type ConfirmData = {
  run_id?: string;
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
          "Ask Soma for the exact first demo deliverable: a playable browser game project package.",
          `Retain it at ${folder} with entrypoint ${entrypoint}.`,
          "The package metadata must include files index.html, README.md, PROOF.md, and validation notes from opening the browser game.",
          "After approval, return a retained project_package output with entrypoint, folder, files, validation, and proof.",
        ].join(" "),
      );

      expect(proposal.response.ok(), proposal.body ? JSON.stringify(proposal.body) : proposal.raw).toBeTruthy();
      expect(proposal.body?.data?.mode).toBe("proposal");
      await expect(page.getByRole("heading", { name: /Start this\?|Run this\?/ }).last()).toBeVisible({ timeout: 30_000 });
      await expect(page.getByRole("button", { name: /^(Start|Run)$/i }).last()).toBeVisible();
      await expect(page.getByText(teamID).last()).toBeVisible();
      await expect(page.getByText(entrypoint).last()).toBeVisible();
      expect(targetExists(entrypoint)).toBeFalsy();

      const confirmed = await confirmProposal(page);
      expect(confirmed.response.ok(), confirmed.body ? JSON.stringify(confirmed.body) : confirmed.raw).toBeTruthy();
      const data = confirmed.body?.data as ConfirmData | undefined;
      expect(data?.verified).toBeTruthy();
      expect(data?.execution_state).toBe("verified");
      expect(data?.run_id).toBeTruthy();

      const outputs = data?.execution_summary?.outputs ?? [];
      const projectPackage = outputs.find((output) => output.kind === "project_package");
      expect(projectPackage, JSON.stringify(outputs)).toBeTruthy();
      expectProjectPackageMetadata(projectPackage!, { title: packageTitle, entrypoint, folder });
      await expect.poll(() => targetExists(entrypoint), { timeout: 30_000 }).toBeTruthy();
      await expectProjectPackageVisible(page, { title: packageTitle, entrypoint, folder });

      const outputPagePromise = page.context().waitForEvent("page");
      await page.getByRole("button", { name: new RegExp(`Open file .*${packageTitle}`, "i") }).last().click();
      const outputPage = await outputPagePromise;
      await outputPage.waitForLoadState("domcontentloaded");
      await expect(outputPage).toHaveTitle(packageTitle);
      await expect(outputPage.locator("body")).toContainText(/score|start|play|restart|game/i);
      await outputPage.close();

      await expectProofAndRunReadback(page, data!);
      const group = await expectGroupOutputReadback(page, teamID, packageTitle, entrypoint);
      await expectResourcesRevisit(page, folder, "index.html");
      await expectGroupsRevisit(page, group, packageTitle, entrypoint);
      await expectRunReceiptRevisit(page, data!.run_id!);
    } finally {
      removeTarget(entrypoint);
      removeTarget(`${folder}/README.md`);
      removeTarget(`${folder}/PROOF.md`);
      removeTarget(`${folder}/project-package.json`);
    }
  });
});

async function expectProofAndRunReadback(page: import("@playwright/test").Page, data: ConfirmData) {
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
