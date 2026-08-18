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
import {
  createQAFixtureScope,
  purgeDeliveryFixture,
} from "../support/qa-fixture-ownership";

type ConfigDocumentRecord = {
  record_id: string;
  digest: string;
  document: {
    metadata: {
      id: string;
      name: string;
      version: string;
      source: { kind: string; ref: string };
    };
  };
};

type WorkIntentSnapshot = {
  id?: string;
  version?: string;
  digest?: string;
};

type LiveChatData = {
  mode?: string;
  payload?: {
    proposal?: { tools?: string[] };
    execution_summary?: {
      work_intent?: { outcome_template_snapshot?: WorkIntentSnapshot };
    };
  };
};

type LiveConfirmData = {
  verified?: boolean;
  execution_state?: string;
  execution_summary?: {
    work_intent?: { outcome_template_snapshot?: WorkIntentSnapshot };
  };
};

const rawRuntimeVocabulary = /ConfigDocument|WorkIntent|content_digest|store_config_document|activate_config_document|record_id|run_id|api\.intent|immutable configuration revision/i;

async function listTemplateRevisions(page: Page, documentID: string) {
  const response = await page.request.get(
    liveAPIURL(`/api/v1/config-documents?kind=OutcomeTemplate&document_id=${encodeURIComponent(documentID)}&limit=10`),
    { headers: liveAPIHeaders() },
  );
  const parsed = await parseJSONIfPossible<APIEnvelope<ConfigDocumentRecord[]>>(response);
  expect(response.ok(), parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBeTruthy();
  return parsed.body?.data ?? [];
}

async function expectNaturalApproval(page: Page) {
  const thread = page.getByTestId("soma-conversation-thread");
  await expect(thread.getByText("I can start that.").last()).toBeVisible({ timeout: 30_000 });
  await expect(thread.getByText(/reply.*(?:approve|start).*to begin/i).last()).toBeVisible();
  await expect.soft(thread.getByText(rawRuntimeVocabulary)).toHaveCount(0);
}

async function expectSynchronousReceiptLanguage(page: Page) {
  const thread = page.getByTestId("soma-conversation-thread");
  await expect(thread.getByText(/Work started|still running|work bus|handed this to/i)).toHaveCount(0);
}

test.skip(({ browserName }) => browserName !== "chromium", "Focused P0.3a live acceptance runs in Chromium.");

test.describe("Live retained Soma Outcome Template journey", () => {
  test.skip(!process.env.PLAYWRIGHT_LIVE_BACKEND, "requires the live authenticated Core and Interface stack");
  test.setTimeout(liveTimeoutMs);

  test("saves, activates, reloads, applies, and removes only its own revision", async ({ page }) => {
    const stamp = Date.now();
    const documentID = `qa-launch-brief-${stamp}`;
    const documentName = `QA launch readiness ${stamp}`;
    const version = `qa-${stamp}`;
    const sourceRef = `conversation:qa-${stamp}`;
    const fixture = await createQAFixtureScope(page, `outcome-template-retained-${stamp}`);
    let organizationID: string | undefined;
    let revisionID: string | undefined;
    let journeyFailure: unknown;

    try {
      organizationID = await createOrganization(
        page,
        `Outcome Template Retained ${stamp}`,
        { fixtureScopeID: fixture.id },
      );
      const document = `apiVersion: mycelis.ai/v1
kind: OutcomeTemplate
metadata:
  id: ${documentID}
  name: ${documentName}
  version: "${version}"
  owner_id: qa-browser
  scope: {kind: workspace, ref: ${organizationID}}
  enabled: true
  source: {kind: soma, ref: ${sourceRef}}
  governance: {risk_level: medium, approval_posture: required}
spec:
  intent_kind: project
  defaults:
    target_outcome: Produce a decision-ready launch readiness brief
    audience: Launch owner and approver
    essential_behavior: [Name the launch, owner, and target date]
    quality_bar: Concise and decision-ready
    delivery_form: Launch readiness document
    constraints: [Keep the recommendation to one page]
    acceptance_evidence: [Launch name, owner, and target date are present]
  question_limit: 3
`;

      await openLiveWorkspace(page, organizationID);
      const saveProposal = await submitLiveWorkspaceChat(
        page,
        `Save this Outcome Template for future launch readiness briefs. Use exactly this YAML:\n\n${document}`,
      );
      expect(saveProposal.response.ok(), saveProposal.body ? JSON.stringify(saveProposal.body) : saveProposal.raw).toBeTruthy();
      const saveData = saveProposal.body?.data as LiveChatData | undefined;
      expect(saveData?.mode).toBe("proposal");
      expect(saveData?.payload?.proposal?.tools).toContain("store_config_document");
      expect(saveData?.payload?.proposal?.tools).not.toContain("write_file");
      await expectNaturalApproval(page);

      const saved = await confirmProposal(page);
      expect(saved.response.ok(), saved.body ? JSON.stringify(saved.body) : saved.raw).toBeTruthy();
      expect((saved.body?.data as LiveConfirmData | undefined)?.verified).toBe(true);
      await expect(page.getByText("Outcome Template saved", { exact: true }).last()).toBeVisible({ timeout: 30_000 });
      await expect(page.getByText(/saved but not active/i).last()).toBeVisible();
      await expectSynchronousReceiptLanguage(page);

      await expect.poll(async () => (await listTemplateRevisions(page, documentID)).length, {
        timeout: 30_000,
        message: "the approved Outcome Template revision should be retained",
      }).toBe(1);
      const [storedRevision] = await listTemplateRevisions(page, documentID);
      revisionID = storedRevision.record_id;
      expect(storedRevision.document.metadata).toMatchObject({
        id: documentID,
        name: documentName,
        version,
        source: { kind: "soma", ref: sourceRef },
      });
      const thread = page.getByTestId("soma-conversation-thread");
      await expect(thread.getByText(revisionID, { exact: false })).toHaveCount(0);
      await expect(thread.getByText(storedRevision.digest, { exact: false })).toHaveCount(0);
      await expect(page.getByRole("button", { name: /Details|Inspect/i }).last()).toBeVisible();

      const activationProposal = await submitLiveWorkspaceChat(
        page,
        `Activate the ${documentName} Outcome Template I just saved.`,
      );
      expect(
        activationProposal.response.ok(),
        activationProposal.body ? JSON.stringify(activationProposal.body) : activationProposal.raw,
      ).toBeTruthy();
      const activationData = activationProposal.body?.data as LiveChatData | undefined;
      expect(activationData?.mode).toBe("proposal");
      expect(activationData?.payload?.proposal?.tools).toContain("activate_config_document");
      await expectNaturalApproval(page);

      const activated = await confirmProposal(page);
      expect(activated.response.ok(), activated.body ? JSON.stringify(activated.body) : activated.raw).toBeTruthy();
      expect((activated.body?.data as LiveConfirmData | undefined)?.verified).toBe(true);
      await expect(page.getByText("Outcome Template active", { exact: true }).last()).toBeVisible({ timeout: 30_000 });
      await expect(page.getByText(/active for this workspace/i).last()).toBeVisible();
      await expectSynchronousReceiptLanguage(page);

      await openLiveWorkspace(page, organizationID);
      await expect(page.getByText(/active for this workspace/i).last()).toBeVisible({ timeout: 30_000 });
      const durableRevisions = await listTemplateRevisions(page, documentID);
      expect(durableRevisions).toHaveLength(1);
      expect(durableRevisions[0]).toMatchObject({ record_id: revisionID, digest: storedRevision.digest });

      const workProposal = await submitLiveWorkspaceChat(
        page,
        `Use the ${documentName} Outcome Template to write a launch readiness document for Aurora, owned by Maya, with a target date of September 30, 2026.`,
      );
      expect(workProposal.response.ok(), workProposal.body ? JSON.stringify(workProposal.body) : workProposal.raw).toBeTruthy();
      const workData = workProposal.body?.data as LiveChatData | undefined;
      expect(workData?.mode).toBe("proposal");
      expect(workData?.payload?.proposal?.tools).not.toContain("activate_config_document");
      const proposedSnapshot = workData?.payload?.execution_summary?.work_intent?.outcome_template_snapshot;
      expect(proposedSnapshot).toMatchObject({ id: documentID, version, digest: storedRevision.digest });
      await expect(page.getByText(`Using ${documentName} v${version} to shape this work.`, { exact: true }).last()).toBeVisible();
      await expectNaturalApproval(page);

      const completedWork = await confirmProposal(page);
      expect(
        completedWork.response.ok(),
        completedWork.body ? JSON.stringify(completedWork.body) : completedWork.raw,
      ).toBeTruthy();
      const completedData = completedWork.body?.data as LiveConfirmData | undefined;
      expect(completedData?.verified).toBe(true);
      expect(completedData?.execution_state).toBe("verified");
      expect(completedData?.execution_summary?.work_intent?.outcome_template_snapshot).toEqual(proposedSnapshot);

      await openLiveWorkspace(page, organizationID);
      await expect(page.getByText(/active for this workspace/i).last()).toBeVisible({ timeout: 30_000 });
      await expect(page.getByText(`Using ${documentName} v${version} to shape this work.`, { exact: true }).last()).toBeVisible();
      await expect(page.getByText(/Aurora.*Maya.*September 30, 2026/i).last()).toBeVisible();
      await expect.soft(page.getByTestId("soma-conversation-thread").getByText(rawRuntimeVocabulary)).toHaveCount(0);
    } catch (error) {
      journeyFailure = error;
      throw error;
    } finally {
      const cleanupErrors: Error[] = [];
      try {
        const revisions = await listTemplateRevisions(page, documentID);
        if (revisions.length > 1) {
          throw new Error(`Refusing cleanup: expected at most one QA revision, found ${revisions.length}`);
        }
        revisionID = revisionID ?? revisions[0]?.record_id;
      } catch (error) {
        cleanupErrors.push(error instanceof Error ? error : new Error(String(error)));
      }
      try {
        await purgeDeliveryFixture(page, fixture, { organizationID });
        expect(await listTemplateRevisions(page, documentID)).toHaveLength(0);
      } catch (error) {
        cleanupErrors.push(error instanceof Error ? error : new Error(String(error)));
      }
      if (cleanupErrors.length > 0) {
        const cleanupMessage = cleanupErrors.map((error) => error.message).join("; ");
        if (!journeyFailure) throw new Error(cleanupMessage);
        if (journeyFailure instanceof Error) journeyFailure.message += ` Cleanup also failed: ${cleanupMessage}`;
      }
    }
  });
});
