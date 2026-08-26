import { expect, test, type Page } from "@playwright/test";
import {
  type ChatRequestBody,
  type RouteResponse,
  lastUserMessage,
  mockOrganizationWorkspace,
  openOrganization,
  sendWorkspaceMessage,
} from "../support/soma-ui-testing";

const rawJargon = /ConfigDocument|WorkIntent|content_digest|store_config_document|config_document_store_failed|api\.intent|run_id/i;

function outcomeTemplateProposal(action: "store" | "activate" = "store"): RouteResponse {
  const activating = action === "activate";
  return {
    status: 200,
    body: {
      ok: true,
      data: {
        meta: { source_node: "admin", timestamp: "2026-08-17T18:00:00Z" },
        signal_type: "chat.reply",
        trust_score: 0.93,
        template_id: "chat-to-proposal",
        mode: "proposal",
        payload: {
          text: activating
            ? "I can make this reusable outcome active, but I need your approval first."
            : "I can save that reusable outcome, but I need your approval first.",
          tools_used: [activating ? "activate_config_document" : "store_config_document"],
          consultations: [],
          artifacts: [],
          proposal: {
            intent: activating ? "activate_outcome_template" : "save_outcome_template",
            operator_summary: activating
              ? "Use the saved Outcome Template for future launch reviews."
              : "Save a reusable Outcome Template for quarterly launch reviews.",
            expected_result: activating
              ? "The saved template will become active and ready to shape new work."
              : "You will get a saved template that asks for the launch name, owner, and target date.",
            affected_resources: ["Outcome Templates"],
            teams: 0,
            agents: 0,
            tools: [activating ? "activate_config_document" : "store_config_document"],
            risk_level: "medium",
            confirm_token: `confirm-outcome-template-${action}`,
            intent_proof_id: `proof-outcome-template-${action}`,
            approval: {
              approval_required: true,
              approval_reason: "saves_reusable_workflow",
              approval_mode: "required",
              capability_risk: "medium",
              capability_ids: ["config_documents"],
              external_data_use: false,
              estimated_cost: 0,
            },
            execution_mode: "confirm_then_execute",
          },
        },
      },
    },
  };
}

function previewedOutcomeTemplate(): RouteResponse {
  return {
    status: 200,
    body: {
      ok: true,
      data: {
        meta: { source_node: "admin", timestamp: "2026-08-17T18:01:00Z" },
        signal_type: "chat.reply",
        trust_score: 0.94,
        template_id: "chat-to-answer",
        mode: "answer",
        payload: {
          text: "Preview ready. This template will ask for the launch name, owner, and target date. Nothing has been saved yet.",
          tools_used: ["preview_config_document"],
        },
      },
    },
  };
}

function completedOutcomeTemplate(action: "store" | "activate"): RouteResponse {
  const activating = action === "activate";
  const runId = activating ? "run-outcome-template-activate" : "run-outcome-template-store";
  const outputTitle = activating ? "Outcome Template active" : "Outcome Template saved";
  const summary = activating
    ? "The Outcome Template is active and ready to shape new work."
    : "The reusable Outcome Template is saved and remains inactive until you ask Soma to use it.";
  return {
    status: 200,
    body: {
      ok: true,
      data: {
        confirmed: true,
        verified: true,
        execution_state: "verified",
        run_id: runId,
        run_status: "completed",
        execution_summary: {
          execution: {
            shape: "guided_proposal",
            status: "completed",
            summary,
          },
          outputs: [{
            id: activating ? "quarterly-launch-review-active" : "quarterly-launch-review-revision",
            title: outputTitle,
            kind: activating ? "config_activation" : "config_revision",
            output_class: "user_deliverable",
            retained: true,
            proof: { proof_id: `${runId}-proof`, verified: true },
          }],
          proof: { label: "Completion checked", run_id: runId, verified: true },
          audit_recovery: { approval_status: "confirmed", recovery_state: "verified" },
          next_step: {
            label: activating
              ? "Tell Soma what outcome you want to create with this template."
              : "Ask Soma to use this Outcome Template.",
          },
        },
      },
    },
  };
}

async function useNoviceDefaults(page: Page) {
  await page.addInitScript(() => {
    window.localStorage.setItem("mycelis-advanced-mode", "false");
    for (const key of Object.keys(window.localStorage)) {
      if (key.startsWith("mycelis-workspace-chat")) window.localStorage.removeItem(key);
    }
  });
}

async function expectCompactProposal(page: Page, action: "store" | "activate" = "store") {
  await expect(page.getByText("I can start that.")).toBeVisible({ timeout: 20_000 });
  await expect(page.getByText(action === "activate"
    ? "Use the saved Outcome Template for future launch reviews."
    : "Save a reusable Outcome Template for quarterly launch reviews.")).toBeVisible();
  await expect(page.getByText(action === "activate"
    ? /saved template will become active and ready to shape new work/i
    : /saved template that asks for the launch name, owner, and target date/i)).toBeVisible();
  await expect(page.getByText(/or reply.*approve/i)).toBeVisible();
  await expect(page.getByRole("button", { name: /^Details$/i }).last()).toBeVisible();
  await expect(page.getByText(rawJargon)).toHaveCount(0);
}

test.skip(({ browserName }) => browserName !== "chromium", "Focused P0.3a acceptance runs in Chromium.");

test.describe("Soma Outcome Template novice journey", () => {
  test("previews, saves, and activates through natural Soma conversation", async ({ page }) => {
    const requests: ChatRequestBody[] = [];
    await useNoviceDefaults(page);
    await mockOrganizationWorkspace(page, (requestBody) => {
      requests.push(requestBody);
      const message = lastUserMessage(requestBody);
      if (/preview/i.test(message)) return previewedOutcomeTemplate();
      if (/use this outcome template/i.test(message)) return outcomeTemplateProposal("activate");
      return outcomeTemplateProposal("store");
    });
    await page.route("**/api/v1/intent/confirm-action", async (route) => {
      const request = route.request().postDataJSON() as { confirm_token?: string };
      const action = request.confirm_token?.endsWith("activate") ? "activate" : "store";
      await route.fulfill({
        ...completedOutcomeTemplate(action),
        contentType: "application/json",
        body: JSON.stringify(completedOutcomeTemplate(action).body),
      });
    });

    await openOrganization(page);
    await sendWorkspaceMessage(
      page,
      "Preview this Outcome Template for quarterly launch reviews. Ask for the launch name, owner, and target date.",
    );
    await expect(page.getByText(/Preview ready.*Nothing has been saved yet/i)).toBeVisible({ timeout: 20_000 });
    await expect(page.getByText(/reply.*approve.*to begin/i)).toHaveCount(0);

    await sendWorkspaceMessage(page, "Save this Outcome Template.");
    await expectCompactProposal(page, "store");
    await sendWorkspaceMessage(page, "approve");
    await expect(page.getByText("Outcome Template saved", { exact: true }).last()).toBeVisible({ timeout: 20_000 });
    await expect(page.getByText(/saved and remains inactive until you ask Soma to use it/i)).toBeVisible();

    await sendWorkspaceMessage(page, "Use this Outcome Template.");
    await expectCompactProposal(page, "activate");
    await sendWorkspaceMessage(page, "approve");
    await expect(page.getByText("Outcome Template active", { exact: true }).last()).toBeVisible({ timeout: 20_000 });
    await expect(page.getByText("The Outcome Template is active and ready to shape new work.", { exact: true })).toBeVisible();
    expect(requests.map(lastUserMessage)).toEqual(expect.arrayContaining([
      expect.stringMatching(/Preview this Outcome Template/i),
      expect.stringMatching(/Save this Outcome Template/i),
      expect.stringMatching(/Use this Outcome Template/i),
    ]));
    await expect(page.getByText(rawJargon)).toHaveCount(0);
  });

  test("turns a failed save into one clear recovery direction", async ({ page }) => {
    await useNoviceDefaults(page);
    await mockOrganizationWorkspace(page, () => outcomeTemplateProposal());
    await page.route("**/api/v1/intent/confirm-action", async (route) => {
      await route.fulfill({
        status: 500,
        contentType: "application/json",
        body: JSON.stringify({
          ok: false,
          error: "config_document_store_failed",
          data: {
            run_id: "run-outcome-template-recovery",
            execution_summary: {
              execution: {
                shape: "native_team",
                status: "failed",
                summary: "The Outcome Template was not saved.",
              },
              outputs: [],
              audit_recovery: {
                approval_status: "approved",
                recovery_state: "blocked",
                blocker: "storage unavailable",
                degradation: {
                  code: "config_document_store_failed",
                  what_failed: "The template could not be saved.",
                  trusted_state: "Your approved wording is still available in this conversation.",
                  invalidated_proof: "No Outcome Template was created.",
                  safe_continuation: "Ask Soma to try again after storage is available.",
                  requires_attention: true,
                },
              },
              next_step: { label: "Try saving the template again" },
            },
          },
        }),
      });
    });

    await openOrganization(page);
    await sendWorkspaceMessage(page, "Soma, create an Outcome Template for quarterly launch reviews.");
    await expectCompactProposal(page);
    await sendWorkspaceMessage(page, "go ahead");

    const result = page.getByTestId("execution-summary-card").last();
    await expect(result.getByText("Needs review")).toBeVisible({ timeout: 20_000 });
    await expect(page.getByText("The template could not be saved.").first()).toBeVisible();
    await result.getByText("Details and proof").click();
    await expect(result.getByText("Still available: Your approved wording is still available in this conversation.")).toBeVisible();
    await expect(result.getByText("Safe next: Ask Soma to try again after storage is available.")).toBeVisible();
    await expect(page.getByText(rawJargon)).toHaveCount(0);
  });
});
