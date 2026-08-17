import { expect, test } from "@playwright/test";
import {
  createOrganization,
  liveTimeoutMs,
  openLiveWorkspace,
  submitLiveWorkspaceChat,
} from "../support/finalization-browser-package";
import {
  createQAFixtureScope,
  purgeDeliveryFixture,
} from "../support/qa-fixture-ownership";

type PreviewData = {
  mode?: string;
  payload?: {
    text?: string;
    tools_used?: string[];
  };
};

const previewDocument = `apiVersion: mycelis.ai/v1
kind: OutcomeTemplate
metadata:
  id: quarterly-launch-review
  name: Quarterly launch review
  version: "1"
  owner_id: soma
  scope: {kind: workspace, ref: primary}
  enabled: true
  source: {kind: soma, ref: conversation:live-preview}
  governance: {risk_level: low, approval_posture: optional}
spec:
  defaults:
    target_outcome: Review a quarterly launch
    delivery_form: Decision-ready launch review
    acceptance_evidence: [Owner and target date are present]
`;

test.describe("Live Soma Outcome Template preview", () => {
  test.skip(!process.env.PLAYWRIGHT_LIVE_BACKEND, "requires a live Core backend");
  test.setTimeout(liveTimeoutMs);

  test("uses the real read-only preview path without creating retained configuration", async ({ page }) => {
    const stamp = Date.now();
    const fixture = await createQAFixtureScope(page, `outcome-template-preview-${stamp}`);
    let organizationID: string | undefined;
    let journeyFailure: unknown;
    try {
      organizationID = await createOrganization(
        page,
        `Outcome Template Preview ${stamp}`,
        { fixtureScopeID: fixture.id },
      );
      await openLiveWorkspace(page, organizationID);
      const result = await submitLiveWorkspaceChat(
        page,
        `Preview this Outcome Template. Use exactly this YAML and do not save it:\n\n${previewDocument}`,
      );

      expect(result.response.ok(), result.body ? JSON.stringify(result.body) : result.raw).toBeTruthy();
      const data = result.body?.data as PreviewData | undefined;
      expect(data?.mode).toBe("answer");
      expect(data?.payload?.tools_used).toContain("preview_config_document");
      expect(data?.payload?.text?.trim()).toBeTruthy();
      await expect(page.getByPlaceholder(/Tell Soma what you want/i)).toBeEnabled();
      await expect(page.getByText(/reply.*approve.*to begin/i)).toHaveCount(0);
      await expect(page.getByText(/ConfigDocument|WorkIntent|content_digest/i)).toHaveCount(0);
    } catch (error) {
      journeyFailure = error;
      throw error;
    } finally {
      try {
        await purgeDeliveryFixture(page, fixture, { organizationID });
      } catch (cleanupError) {
        if (!journeyFailure) throw cleanupError;
        if (journeyFailure instanceof Error && cleanupError instanceof Error) {
          journeyFailure.message += ` Cleanup also failed: ${cleanupError.message}`;
        }
      }
    }
  });
});
