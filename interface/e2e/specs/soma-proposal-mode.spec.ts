import { expect, test } from "@playwright/test";
import {
    mockOrganizationWorkspace,
    openOrganization,
    proposalEnvelope,
    sendWorkspaceMessage,
} from "../support/soma-ui-testing";

test.skip(({ browserName }) => browserName !== "chromium", "Deep UI testing coverage is stabilized in Chromium for the MVP audit.");

function workIntentProposalEnvelope() {
    const response = proposalEnvelope();
    const proposal = ((response.body as Record<string, any>).data.payload.proposal as Record<string, any>);
    proposal.execution_mode = "schedule_handoff";
    proposal.work_intent = {
        kind: "team_service",
        objective: "Keep a media review lane running for the selected team.",
        cadence: "continuous",
        schedule_summary: "Watch the media inbox every 15 minutes.",
        runtime_posture: "Leave the review lane active until the operator pauses it.",
        bus_scope: "current_team",
        nats_subjects: ["swarm.team.media.signal.status"],
    };
    return response;
}

test.describe("Soma proposal mode", () => {
    test("routes mutating requests through proposal mode and keeps cancel explicit", async ({ page }) => {
        const workspace = await mockOrganizationWorkspace(page, () => proposalEnvelope());

        await openOrganization(page);
        await sendWorkspaceMessage(page, "Create a simple python file named hello_world.py in the workspace.");

        await expect(page.getByText("I can do that.")).toBeVisible({ timeout: 20_000 });
        await expect(page.getByText("Start this?")).toBeVisible();
        await expect(page.getByText("What I will do")).toBeVisible();
        await expect(page.getByText("create a hello_world.py file in your workspace.")).toBeVisible();
        await expect(page.getByText("A new Python file will be saved to workspace/logs/hello_world.py after approval.")).toBeVisible();
        await expect(page.getByText("workspace/logs/hello_world.py", { exact: true })).toHaveCount(0);
        await expect(page.getByText("Ready if you want")).toBeVisible();
        await expect(page.getByText(/RISK MEDIUM/i)).toHaveCount(0);
        await expect(page.getByRole("button", { name: /^Details$/i })).toBeVisible();

        await page.getByRole("button", { name: /^Adjust$/i }).click();
        await expect(page.getByText(/Proposal cancelled\. No action executed\./i)).toBeVisible({ timeout: 20_000 });
        await expect.poll(() => workspace.cancelCalls()).toBe(1);
    });

    test("keeps structured work intent understandable behind Details", async ({ page }) => {
        await mockOrganizationWorkspace(page, () => workIntentProposalEnvelope());

        await openOrganization(page);
        await sendWorkspaceMessage(page, "Keep a media review lane running for this team.");

        await expect(page.getByText("I can do that.")).toBeVisible({ timeout: 20_000 });
        await expect(page.getByText("Start this?")).toBeVisible();
        await expect(page.getByText("swarm.team.media.signal.status")).toHaveCount(0);
        await expect(page.getByText("Team connection")).toHaveCount(0);

        await page.getByRole("button", { name: /^Details$/i }).click();

        await expect(page.getByText("When it runs")).toBeVisible();
        await expect(page.getByText("Keep running")).toBeVisible();
        await expect(page.getByText("Watch the media inbox every 15 minutes.")).toBeVisible();
        await expect(page.getByText("Team connection")).toBeVisible();
        await expect(page.getByText("Current team", { exact: true })).toBeVisible();
        await expect(page.getByText("swarm.team.media.signal.status")).toBeVisible();
    });

    test("keeps failed approved execution reviewable with run proof and recovery copy", async ({ page }) => {
        await mockOrganizationWorkspace(page, () => proposalEnvelope());
        const failedRunId = "run-failed-browser-123456";

        await page.route("**/api/v1/intent/confirm-action", async (route) => {
            await route.fulfill({
                status: 500,
                contentType: "application/json",
                body: JSON.stringify({
                    ok: false,
                    error: "confirmation denied",
                    data: {
                        run_id: failedRunId,
                        execution_summary: {
                            intent: {
                                original: "Create a simple python file named hello_world.py in the workspace.",
                                resolved: "chat-action",
                            },
                            understanding: {
                                summary: "Soma understood the approved file-writing action, but execution failed.",
                            },
                            execution: {
                                shape: "directed_execution",
                                status: "failed",
                                summary: "Approved execution failed before producing trusted output.",
                            },
                            capability_use: [{ id: "write_file", label: "write_file", status: "failed" }],
                            outputs: [],
                            proof: {
                                run_id: failedRunId,
                                proof_class: "failed_run",
                                verified: false,
                            },
                            audit_recovery: {
                                approval_status: "approved",
                                recovery_state: "blocked",
                                blocker: "tool unavailable",
                                degradation: {
                                    code: "approved_execution_failed",
                                    what_failed: "tool unavailable",
                                    trusted_state: "The failed run record remains trusted.",
                                    invalidated_proof: "No completed output should be trusted.",
                                    safe_continuation: "Review the failed run and retry.",
                                    requires_attention: true,
                                },
                            },
                            next_step: {
                                label: "Review failed run",
                                action: "view_run",
                                href: `/api/v1/runs/${failedRunId}`,
                            },
                        },
                    },
                }),
            });
        });

        await openOrganization(page);
        await sendWorkspaceMessage(page, "Create a simple python file named hello_world.py in the workspace.");

        await expect(page.getByText("I can do that.")).toBeVisible({ timeout: 20_000 });
        await page.getByRole("button", { name: /^Start$/i }).click();

        const failureCard = page.getByTestId("execution-summary-card").last();
        await expect(failureCard.getByText("Needs review").first()).toBeVisible({ timeout: 20_000 });
        await expect(failureCard.getByText("Details and proof")).toBeVisible();
        await failureCard.getByText("Details and proof").click();

        await expect(failureCard.getByText("Failed: tool unavailable")).toBeVisible();
        await expect(failureCard.getByText("Still available: The failed run record remains trusted.")).toBeVisible();
        await expect(failureCard.getByText("Not reliable: No completed output should be trusted.")).toBeVisible();
        await expect(failureCard.getByText("Safe next: Review the failed run and retry.")).toBeVisible();
        await expect(failureCard.getByRole("link", { name: /Run run-fail/i }).first()).toHaveAttribute("href", `/runs/${failedRunId}`);
        await expect(page.getByText("Run proof + retained output")).toHaveCount(0);
        await expect(page.getByText("Verified execution proof")).toHaveCount(0);
    });
});
