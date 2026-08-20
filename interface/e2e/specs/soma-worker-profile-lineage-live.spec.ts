import { expect, test, type Page } from "@playwright/test";
import {
  confirmProposal,
  createOrganization,
  liveTimeoutMs,
  openLiveWorkspace,
  parseJSONIfPossible,
  submitLiveWorkspaceChat,
  type APIEnvelope,
} from "../support/finalization-browser-package";
import { liveAPIHeaders, liveAPIURL } from "../support/live-api-auth";
import { createQAFixtureScope, purgeDeliveryFixture } from "../support/qa-fixture-ownership";

type ConfigRecord = {
  record_id: string;
  digest: string;
  document: { metadata: { id: string; name: string; version: string } };
};

type ProfileSnapshot = {
  id: string;
  version: string;
  digest: string;
  record_id: string;
  tenant_id: string;
  scope: { kind: string; ref: string };
};

type TeamManifest = {
  id: string;
  members: Array<{ profile_ref?: string; profile_snapshot?: ProfileSnapshot }>;
};

type ConfirmationData = {
  run_id?: string;
  verified?: boolean;
  execution_state?: string;
  dispatch_state?: string;
};

async function listProfileRevisions(page: Page, documentID: string) {
  const response = await page.request.get(
    liveAPIURL(`/api/v1/config-documents?kind=WorkerProfile&document_id=${encodeURIComponent(documentID)}&limit=10`),
    { headers: liveAPIHeaders() },
  );
  const parsed = await parseJSONIfPossible<APIEnvelope<ConfigRecord[]>>(response);
  expect(response.ok(), parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBeTruthy();
  return parsed.body?.data ?? [];
}

async function listRuntimeTeams(page: Page) {
  const response = await page.request.get(liveAPIURL("/api/swarm/teams"), { headers: liveAPIHeaders() });
  if (!response.ok()) throw new Error(`List runtime teams failed (${response.status()}): ${await response.text()}`);
  return await response.json() as TeamManifest[];
}

async function proposeAndConfirm(page: Page, request: string, expectedTool: string) {
  const proposal = await submitLiveWorkspaceChat(page, request);
  expect(proposal.response.ok(), proposal.body ? JSON.stringify(proposal.body) : proposal.raw).toBeTruthy();
  expect(
    (proposal.body?.data as { mode?: string })?.mode,
    proposal.body ? JSON.stringify(proposal.body) : proposal.raw,
  ).toBe("proposal");
  expect(JSON.stringify(proposal.body)).toContain(expectedTool);
  const thread = page.getByTestId("soma-conversation-thread");
  await expect(thread.getByText("I can start that.").last()).toBeVisible({ timeout: 30_000 });
  await expect(thread.getByText(/reply.*(?:approve|start).*to begin/i).last()).toBeVisible();
  const completed = await confirmProposal(page);
  expect(completed.response.ok(), completed.body ? JSON.stringify(completed.body) : completed.raw).toBeTruthy();
  return completed;
}

async function saveProfile(page: Page, yaml: string) {
  await proposeAndConfirm(page, `Save this Worker Profile for reuse. Use exactly this YAML:\n\n${yaml}`, "store_config_document");
  await expect(page.getByText("Worker Profile saved", { exact: true }).last()).toBeVisible({ timeout: 30_000 });
}

async function activateCurrentProfile(page: Page, expectedVersion: string, request = "Activate this Worker Profile.") {
  await proposeAndConfirm(page, request, "activate_config_document");
  await expect(page.getByText("Worker Profile active", { exact: true }).last()).toBeVisible({ timeout: 30_000 });
  await expect(page.getByText(new RegExp(`v?${expectedVersion} is active`, "i")).last()).toBeVisible();
}

async function createTeamWithCurrentProfile(page: Page, teamID: string) {
  const accepted = await proposeAndConfirm(
    page,
    `Create a temporary team using this Worker Profile with team_id ${teamID}.`,
    "create_team",
  );
  const acceptedData = accepted.body?.data as ConfirmationData | undefined;
  expect(accepted.response.status(), accepted.body ? JSON.stringify(accepted.body) : accepted.raw).toBe(202);
  expect(acceptedData).toMatchObject({ verified: false, execution_state: "running", dispatch_state: "pending" });
  expect(acceptedData?.run_id).toBeTruthy();
  await expect(page.getByText("Work started", { exact: true }).last()).toBeVisible({ timeout: 30_000 });
  await expect(page.getByText("Soma is thinking...", { exact: true })).toHaveCount(0);
  await expect(page.getByPlaceholder(/Tell Soma what you want/i)).toBeEnabled();
  await expect.poll(async () => (await listRuntimeTeams(page)).some((team) => team.id === teamID), {
    timeout: 30_000,
    message: `${teamID} should be created through Soma`,
  }).toBe(true);
  const team = (await listRuntimeTeams(page)).find((candidate) => candidate.id === teamID);
  expect(team?.members).toHaveLength(1);
  return team?.members[0].profile_snapshot;
}

function workerProfileYAML(
  organizationID: string,
  documentID: string,
  documentName: string,
  version: string,
) {
  return `apiVersion: mycelis.ai/v1
kind: WorkerProfile
metadata:
  id: ${documentID}
  name: ${documentName}
  version: ${version}
  owner_id: qa-browser
  scope: {kind: workspace, ref: ${organizationID}}
  enabled: true
  source: {kind: soma, ref: conversation:worker-profile-live}
  governance: {risk_level: medium, approval_posture: required}
spec:
  description: Reviews retained delivery evidence.
  role: evidence-reviewer
  system_prompt: Review the approved work and return concise retained evidence.
  capability_refs: [store_artifact]
  usage_policy: {selection: soma_or_manual, scope: workspace}
  inputs: [approved_work]
  outputs: [retained_evidence]
  verification_strategy: semantic
  verification_rubric: [The requested evidence is retained]
`;
}

test.skip(({ browserName }) => browserName !== "chromium", "Focused Worker Profile proof runs in Chromium.");

test.describe("Live Soma Worker Profile lineage", () => {
  test.skip(!process.env.PLAYWRIGHT_LIVE_BACKEND, "requires the live authenticated Core and Interface stack");
  test.setTimeout(liveTimeoutMs);

  test("pins A, B, and rolled-back A to exact runtime teams", async ({ page }) => {
    const stamp = Date.now();
    const documentID = `qa-evidence-reviewer-${stamp}`;
    const documentName = `QA evidence reviewer ${stamp}`;
    const teamA = `qa-profile-a-${stamp}`;
    const teamB = `qa-profile-b-${stamp}`;
    const teamA2 = `qa-profile-a2-${stamp}`;
    const fixture = await createQAFixtureScope(page, `worker-profile-lineage-${stamp}`);
    let organizationID: string | undefined;
    let journeyFailure: unknown;

    try {
      organizationID = await createOrganization(page, `Worker Profile Lineage ${stamp}`, { fixtureScopeID: fixture.id });
      await openLiveWorkspace(page, organizationID);

      await saveProfile(page, workerProfileYAML(organizationID, documentID, documentName, "alpha"));
      const [revisionA] = await listProfileRevisions(page, documentID);
      expect(revisionA.document.metadata.version).toBe("alpha");
      await activateCurrentProfile(page, "alpha");
      const snapshotA = await createTeamWithCurrentProfile(page, teamA);
      expect(snapshotA).toMatchObject({
        id: documentID,
        version: "alpha",
        digest: revisionA.digest,
        record_id: revisionA.record_id,
        tenant_id: "default",
        scope: { kind: "workspace", ref: organizationID },
      });

      await saveProfile(page, workerProfileYAML(organizationID, documentID, documentName, "beta"));
      const revisions = await listProfileRevisions(page, documentID);
      const revisionB = revisions.find((record) => record.document.metadata.version === "beta");
      expect(revisionB).toBeTruthy();
      await activateCurrentProfile(page, "beta");
      const snapshotB = await createTeamWithCurrentProfile(page, teamB);
      expect(snapshotB).toMatchObject({ version: "beta", digest: revisionB?.digest, record_id: revisionB?.record_id });
      expect((await listRuntimeTeams(page)).find((team) => team.id === teamA)?.members[0].profile_snapshot).toEqual(snapshotA);

      await activateCurrentProfile(page, "alpha", "Roll back this Worker Profile to version alpha.");
      const snapshotA2 = await createTeamWithCurrentProfile(page, teamA2);
      expect(snapshotA2).toEqual(snapshotA);

      await openLiveWorkspace(page, organizationID);
      const durableTeams = await listRuntimeTeams(page);
      expect(durableTeams.find((team) => team.id === teamA)?.members[0].profile_snapshot).toEqual(snapshotA);
      expect(durableTeams.find((team) => team.id === teamB)?.members[0].profile_snapshot).toEqual(snapshotB);
      expect(durableTeams.find((team) => team.id === teamA2)?.members[0].profile_snapshot).toEqual(snapshotA);
      const thread = page.getByTestId("soma-conversation-thread");
      await expect(thread.getByText(revisionA.record_id, { exact: false })).toHaveCount(0);
      await expect(thread.getByText(revisionA.digest, { exact: false })).toHaveCount(0);
    } catch (error) {
      journeyFailure = error;
      throw error;
    } finally {
      try {
        await purgeDeliveryFixture(page, fixture, { organizationID });
      } catch (error) {
        if (!journeyFailure) throw error;
        if (journeyFailure instanceof Error) journeyFailure.message += ` Cleanup also failed: ${String(error)}`;
      }
    }
  });
});
