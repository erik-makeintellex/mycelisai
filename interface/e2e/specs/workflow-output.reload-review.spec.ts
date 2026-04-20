import { expect, test } from "@playwright/test";
import {
    fulfillJSON,
    installWorkflowOutputShell,
    openOrganization,
    sendWorkspaceMessage,
    type ArtifactRecord,
    type GroupRecord,
    gotoWithColdStartRetry,
} from "../support/workflow-output";

test.describe("Workflow output reload and retained review", () => {
    test("keeps retained outputs visible after reload and re-review", async ({ page }) => {
        test.setTimeout(180_000);
        const groups: GroupRecord[] = [];
        const groupArtifacts: Record<string, ArtifactRecord[]> = {};

        await installWorkflowOutputShell(page);

        await page.route("**/api/v1/chat", async (route) => {
            const requestBody = route.request().postDataJSON() as {
                messages?: Array<{ content?: string }>;
            };
            const content = requestBody?.messages?.[requestBody.messages.length - 1]?.content ?? "";

            const text = content.includes("Resume the release-readiness work")
                ? "Already done: planning lane package, validation checklist, and review summary are retained. Remaining: confirm the live Windows operator validation pass. Next owner: Review lane lead."
                : "Use the self-hosted Kubernetes lane with an explicit Windows AI endpoint, then run the Windows browser validation flow against the retained output and continuity checks.";

            await route.fulfill({
                status: 200,
                contentType: "application/json",
                body: JSON.stringify({
                    ok: true,
                    data: {
                        meta: { source_node: "admin", timestamp: "2026-04-15T20:10:00Z" },
                        signal_type: "chat.reply",
                        trust_score: 0.95,
                        template_id: "chat-to-answer",
                        mode: "answer",
                        payload: {
                            text,
                            tools_used: [],
                            consultations: [],
                            artifacts: [],
                        },
                    },
                }),
            });
        });

        await page.route("**/api/v1/organizations/org-workflow-variants/workspace/actions", async (route) => {
            await fulfillJSON(route, 200, {
                ok: true,
                data: {
                    action: "plan_next_steps",
                    request_label: "Create a compact release-readiness team",
                    headline: "Compact release-readiness package",
                    summary: "Soma recommends one focused team with a clear lead, one validation package, and one risk review.",
                    priority_steps: ["Confirm the validation scope.", "Keep the output package retained."],
                    suggested_follow_ups: ["Launch the compact team"],
                    execution_contract: {
                        execution_mode: "native_team",
                        owner_label: "Native Mycelis team",
                        team_name: "Release Readiness Team",
                        summary: "Use one compact team for planning, validation, and review.",
                        target_outputs: ["Validation checklist", "Deployment recommendation", "Risk review"],
                        workflow_group: {
                            name: "Release Readiness Team temporary workflow",
                            goal_statement: "Produce a compact release-readiness package for the Windows self-hosted lane.",
                            work_mode: "propose_only",
                            coordinator_profile: "Release Readiness lead",
                            allowed_capabilities: ["team.coordinate", "artifact.review"],
                            expiry_hours: 72,
                            summary: "Launch a temporary workflow group for the compact release-readiness package.",
                        },
                    },
                },
            });
        });

        await page.route("**/api/v1/groups", async (route) => {
            if (route.request().method() === "GET") {
                await fulfillJSON(route, 200, { ok: true, data: groups });
                return;
            }

            const created: GroupRecord = {
                group_id: "group-compact-release",
                name: "Release Readiness Team temporary workflow",
                goal_statement: "Produce a compact release-readiness package for the Windows self-hosted lane.",
                work_mode: "propose_only",
                member_user_ids: ["owner"],
                team_ids: ["release-lead", "validation-lead"],
                coordinator_profile: "Release Readiness lead",
                approval_policy_ref: "browser-proof",
                status: "active",
                expiry: "2026-04-18T18:00:00Z",
                created_by: "owner",
                created_at: "2026-04-15T20:01:00Z",
            };
            groups.push(created);
            groupArtifacts[created.group_id] = [
                {
                    id: "artifact-compact-checklist",
                    team_id: "release-lead",
                    agent_id: "release-lead",
                    artifact_type: "document",
                    title: "Validation checklist",
                    content_type: "text/markdown",
                    content: "# Validation checklist\n\n- Runtime health\n- Browser proof\n- Recovery notes",
                    metadata: {},
                    status: "approved",
                    created_at: "2026-04-15T20:02:00Z",
                },
                {
                    id: "artifact-compact-risk-review",
                    team_id: "validation-lead",
                    agent_id: "validation-lead",
                    artifact_type: "document",
                    title: "Risk review",
                    content_type: "text/markdown",
                    content: "# Risk review\n\n- Endpoint drift\n- Resume path validation",
                    metadata: {},
                    status: "approved",
                    created_at: "2026-04-15T20:03:00Z",
                },
            ];

            await fulfillJSON(route, 201, {
                ok: true,
                data: {
                    group_id: created.group_id,
                    name: created.name,
                },
            });
        });

        await page.route("**/api/v1/groups/monitor", async (route) => {
            await fulfillJSON(route, 200, {
                ok: true,
                data: {
                    status: "online",
                    published_count: 1,
                    last_group_id: groups[groups.length - 1]?.group_id ?? "",
                    last_message: "Prepare retained workflow outputs",
                    last_published_at: "2026-04-15T20:07:00Z",
                },
            });
        });

        await page.route(/\/api\/v1\/groups\/([^/]+)\/status$/, async (route) => {
            if (route.request().method() !== "PATCH") {
                await fulfillJSON(route, 405, { ok: false, error: "method not allowed" });
                return;
            }
            const groupId = route.request().url().match(/\/api\/v1\/groups\/([^/]+)\/status$/)?.[1];
            const target = groups.find((group) => group.group_id === groupId);
            if (!target) {
                await fulfillJSON(route, 404, { ok: false, error: "group not found" });
                return;
            }
            target.status = "archived";
            await fulfillJSON(route, 200, { ok: true, data: target });
        });

        await page.route(/\/api\/v1\/groups\/([^/]+)\/outputs\?limit=8$/, async (route) => {
            const groupId = route.request().url().match(/\/api\/v1\/groups\/([^/]+)\/outputs\?limit=8$/)?.[1];
            await fulfillJSON(route, 200, {
                ok: true,
                data: groupId ? groupArtifacts[groupId] ?? [] : [],
            });
        });

        await openOrganization(page);
        await sendWorkspaceMessage(
            page,
            "Give me the shortest practical recommendation for how to validate this Windows self-hosted release lane.",
        );
        await expect(
            page.getByText(
                "Use the self-hosted Kubernetes lane with an explicit Windows AI endpoint, then run the Windows browser validation flow against the retained output and continuity checks.",
            ),
        ).toBeVisible();

        await page.getByRole("button", { name: "Create teams with Soma" }).click();
        await expect(page.getByRole("heading", { name: "Create teams with Soma" })).toBeVisible();

        const requestField = page.getByLabel("Tell Soma what team or delivery lane you want to create");
        await requestField.fill(
            "Create the smallest useful team to produce a release-readiness package for this Windows self-hosted lane.",
        );
        await page.getByRole("button", { name: "Start team design" }).click();

        await expect(page.getByText("Compact release-readiness package", { exact: true })).toBeVisible();
        await page.getByRole("button", { name: "Create temporary workflow group" }).click();

        const openCompactLink = page.getByRole("link", { name: "Open Release Readiness Team temporary workflow" });
        await expect(openCompactLink).toHaveAttribute("href", "/groups?group_id=group-compact-release");
        await gotoWithColdStartRetry(page, "/groups?group_id=group-compact-release");
        await expect(page.getByRole("heading", { name: "Create, review, and coordinate focused groups." })).toBeVisible();
        await page.getByRole("button", { name: "Release Readiness Team temporary workflow" }).click();
        await expect(page.getByRole("heading", { name: "Release Readiness Team temporary workflow" })).toBeVisible();
        await page.getByRole("button", { name: "Archive temporary group" }).click();
        await expect(page.getByText("Archived temporary group", { exact: true })).toBeVisible();
        await page.reload({ waitUntil: "domcontentloaded" });
        await expect(page.getByText("Archived temporary group", { exact: true })).toBeVisible();
        await expect(page.getByText("Validation checklist", { exact: true })).toBeVisible();
        await expect(page.getByText("Risk review", { exact: true })).toBeVisible();

        await openOrganization(page);
        await sendWorkspaceMessage(
            page,
            "Resume the release-readiness work from the retained package and show me what is already done, what remains, and which lane or lead owns the next step.",
        );
        await expect(
            page.getByText(
                "Already done: planning lane package, validation checklist, and review summary are retained. Remaining: confirm the live Windows operator validation pass. Next owner: Review lane lead.",
            ),
        ).toBeVisible();
    });
});
