import { expect, test } from "@playwright/test";
import { mockRunsBrowserProofApis } from "../support/docs-and-runs-fixtures";
import {
    answerEnvelope,
    lastUserMessage,
    mockOrganizationWorkspace,
    openOrganization,
    sendWorkspaceMessage,
    type RouteResponse,
} from "../support/soma-ui-testing";

test.skip(({ browserName }) => browserName !== "chromium", "V8.3 timeline and causal proof is stabilized in Chromium.");

function blockerEnvelope(text: string): RouteResponse {
    return {
        status: 200,
        body: {
            ok: true,
            data: {
                meta: { source_node: "admin", timestamp: "2026-03-26T16:05:00Z" },
                signal_type: "chat.reply",
                trust_score: 0.4,
                template_id: "chat-to-blocker",
                mode: "blocker",
                payload: {
                    text,
                    tools_used: [],
                    consultations: [],
                    artifacts: [],
                },
            },
        },
    };
}

test.describe("V8.3 Timeline Stage Strip and Causal Diagnostic Surfaces", () => {
    test("renders V8.3 9-stage RunStageProgressStrip on /runs/[id] from mission events and stages projection", async ({ page }) => {
        const runId = "run-v83-stage-1001";
        const failedRunId = "run-v83-failed-1002";
        const missionId = "mission-v83-1001";
        const now = new Date().toISOString();
        let returnServerStageProjection = false;

        await mockRunsBrowserProofApis(page, {
            runId,
            failedRunId,
            missionId,
            now,
        });

        await page.route(`**/api/v1/runs/${runId}/stages**`, async (route) => {
            if (!returnServerStageProjection) {
                await route.fulfill({
                    status: 404,
                    contentType: "application/json",
                    body: JSON.stringify({ ok: false, error: "not projected yet" }),
                });
                return;
            }
            await route.fulfill({
                status: 200,
                contentType: "application/json",
                body: JSON.stringify({
                    ok: true,
                    data: {
                        run_id: runId,
                        stages: [
                            { stage: "intent_received", label: "Intent Received", status: "completed", timestamp: now },
                            { stage: "proposal_generated", label: "Proposal Generated", status: "completed", timestamp: now },
                            { stage: "approval_required", label: "Approval Gate", status: "completed", timestamp: now },
                            { stage: "approval_granted", label: "Approval Granted", status: "completed", timestamp: now },
                            { stage: "mission_running", label: "Mission Running", status: "completed", timestamp: now },
                            { stage: "tool_action_execution", label: "Tool / Action Execution", status: "completed", timestamp: now },
                            { stage: "package_generated", label: "Package Generated", status: "completed", timestamp: now },
                            { stage: "proof_generated", label: "Proof Verified", status: "completed", timestamp: now },
                            { stage: "recovery_next_actions", label: "Recovery & Next Actions", status: "completed", timestamp: now },
                        ],
                    },
                }),
            });
        });

        await page.goto(`/runs/${runId}?tab=events`, { waitUntil: "domcontentloaded" });

        const strip = page.getByTestId("run-stage-progress-strip");
        await expect(strip).toBeVisible({ timeout: 20_000 });
        await expect(page.getByTestId("run-stage-intent_received")).toHaveAttribute("data-stage-status", "completed");
        await expect(page.getByTestId("run-stage-mission_running")).toHaveAttribute("data-stage-status", "active");
        await expect(page.getByTestId("run-stage-package_generated")).toHaveAttribute("data-stage-status", "pending");

        returnServerStageProjection = true;
        await page.getByRole("button", { name: "Refresh events" }).click();

        await expect(page.getByTestId("run-stage-package_generated")).toHaveAttribute("data-stage-status", "completed");
        await expect(page.getByTestId("run-stage-proof_generated")).toHaveAttribute("data-stage-status", "completed");
        await expect(page.getByTestId("run-stage-recovery_next_actions")).toHaveAttribute("data-stage-status", "completed");
    });

    test("renders SomaCausalSummary stage posture, diagnostic channels, and OutputWorkbench stage strip", async ({ page }) => {
        await mockOrganizationWorkspace(page, (requestBody) => {
            const content = lastUserMessage(requestBody);
            if (/trigger blocker/i.test(content)) {
                return blockerEnvelope("Context stream retrieval failed while querying specialist memory.");
            }
            return answerEnvelope("Prepared the V8.3 Codex delivery package with verified proof.", {
                askClass: "governed_artifact",
                artifacts: [
                    {
                        id: "art-v83-task",
                        type: "project_package",
                        title: "Implementation Task",
                        content_type: "text/markdown",
                        content: "# V8.3 Implementation Task\n\nVerified package output.",
                        saved_path: "codex-task-packages/v83-proof/implementation_task.md",
                    },
                    {
                        id: "art-v83-proof",
                        type: "project_package",
                        title: "Package Proof",
                        content_type: "text/markdown",
                        content: "# Proof\n\nAll checks green.",
                        saved_path: "codex-task-packages/v83-proof/proof.md",
                    },
                ],
            });
        });

        await openOrganization(page);

        const causalSummary = page.getByTestId("soma-causal-summary");
        await expect(causalSummary).toBeVisible();
        await expect(causalSummary).toHaveAttribute("data-phase", "idle");

        await sendWorkspaceMessage(page, "Create the V8.3 delivery package for review.");
        await expect(page.getByTestId("output-workbench")).toBeVisible({ timeout: 20_000 });
        await expect(page.getByTestId("output-workbench-stage-strip")).toBeVisible();
        await expect(page.getByTestId("soma-causal-stage-badge")).toContainText("Stage 8 · Package & Proof Verified");

        await sendWorkspaceMessage(page, "Please trigger blocker diagnostic state.");
        await expect(causalSummary).toHaveAttribute("data-phase", "failed", { timeout: 20_000 });
        await expect(page.getByTestId("soma-causal-stage-badge")).toContainText("Stage 9 · Recovery & Next Actions");
        await expect(page.getByTestId("soma-causal-diagnostic")).toContainText("channels: memory.stream, mission_events");
    });
});
