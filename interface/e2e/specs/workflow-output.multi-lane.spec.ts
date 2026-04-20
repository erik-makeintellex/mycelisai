import { expect, test } from "@playwright/test";
import {
    fulfillJSON,
    gotoWithColdStartRetry,
    installWorkflowOutputShell,
    openOrganization,
    type ArtifactRecord,
    type GroupRecord,
} from "../support/workflow-output";

test.describe("Workflow output multi-lane package", () => {
    test("keeps the planning, validation, and review lanes visible", async ({ page }) => {
        test.setTimeout(180_000);
        const groups: GroupRecord[] = [];
        const groupArtifacts: Record<string, ArtifactRecord[]> = {};

        await installWorkflowOutputShell(page);

        await page.route("**/api/v1/organizations/org-workflow-variants/workspace/actions", async (route) => {
            await fulfillJSON(route, 200, {
                ok: true,
                data: {
                    action: "plan_next_steps",
                    request_label: "Split into workflow lanes",
                    headline: "Multi-lane release-readiness workflow",
                    summary: "Soma split the work into planning, validation, and review lanes so each lane stays compact and resumable.",
                    priority_steps: ["Keep each lane small.", "Retain the lane outputs for resume."],
                    suggested_follow_ups: ["Launch the lane bundle"],
                    execution_contract: {
                        execution_mode: "native_team",
                        owner_label: "Coordinated lane bundle",
                        team_name: "Release Readiness Workflow",
                        summary: "Planning lane, validation lane, and review lane each keep a separate output contract.",
                        target_outputs: ["Planning lane package", "Validation lane checklist", "Review lane summary"],
                        workflow_group: {
                            name: "Release Readiness Workflow temporary workflow",
                            goal_statement: "Split release-readiness work into planning, validation, and review lanes.",
                            work_mode: "propose_only",
                            coordinator_profile: "Release workflow coordinator",
                            allowed_capabilities: ["team.coordinate", "artifact.review"],
                            expiry_hours: 72,
                            summary: "Launch a multi-lane workflow group for retained planning, validation, and review outputs.",
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
                group_id: "group-multilane-release",
                name: "Release Readiness Workflow temporary workflow",
                goal_statement: "Split release-readiness work into planning, validation, and review lanes.",
                work_mode: "propose_only",
                member_user_ids: ["owner"],
                team_ids: ["planning-lead", "validation-lead", "review-lead"],
                coordinator_profile: "Release workflow coordinator",
                approval_policy_ref: "browser-proof",
                status: "active",
                expiry: "2026-04-18T18:00:00Z",
                created_by: "owner",
                created_at: "2026-04-15T20:01:00Z",
            };
            groups.push(created);
            groupArtifacts[created.group_id] = [
                {
                    id: "artifact-planning-package",
                    team_id: "planning-lead",
                    agent_id: "planning-lead",
                    artifact_type: "document",
                    title: "Planning lane package",
                    content_type: "text/markdown",
                    content: "# Planning lane package\n\n- Scope\n- Acceptance criteria",
                    metadata: {},
                    status: "approved",
                    created_at: "2026-04-15T20:04:00Z",
                },
                {
                    id: "artifact-validation-checklist",
                    team_id: "validation-lead",
                    agent_id: "validation-lead",
                    artifact_type: "document",
                    title: "Validation lane checklist",
                    content_type: "text/markdown",
                    content: "# Validation lane checklist\n\n- Runtime gate\n- Windows browser pass",
                    metadata: {},
                    status: "approved",
                    created_at: "2026-04-15T20:05:00Z",
                },
                {
                    id: "artifact-review-summary",
                    team_id: "review-lead",
                    agent_id: "review-lead",
                    artifact_type: "document",
                    title: "Review lane summary",
                    content_type: "text/markdown",
                    content: "# Review lane summary\n\n- Remaining risk\n- Resume recommendation",
                    metadata: {},
                    status: "approved",
                    created_at: "2026-04-15T20:06:00Z",
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
                    published_count: 2,
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
        await page.getByRole("button", { name: "Create teams with Soma" }).click();
        await expect(page.getByRole("heading", { name: "Create teams with Soma" })).toBeVisible();

        const requestField = page.getByLabel("Tell Soma what team or delivery lane you want to create");
        await requestField.fill(
            "This is broad. Split it into compact lanes for planning, deployment validation, and review.",
        );
        await page.getByRole("button", { name: "Start team design" }).click();

        await expect(page.getByText("Multi-lane release-readiness workflow", { exact: true })).toBeVisible();
        await expect(page.getByText("Release Readiness Workflow", { exact: true })).toBeVisible();
        await expect(page.getByText("Planning lane package", { exact: true })).toBeVisible();
        await expect(page.getByText("Validation lane checklist", { exact: true })).toBeVisible();
        await expect(page.getByText("Review lane summary", { exact: true })).toBeVisible();

        await page.getByRole("button", { name: "Create temporary workflow group" }).click();
        const openMultiLaneLink = page.getByRole("link", { name: "Open Release Readiness Workflow temporary workflow" });
        await expect(openMultiLaneLink).toHaveAttribute("href", "/groups?group_id=group-multilane-release");
        await gotoWithColdStartRetry(page, "/groups");
        await expect(page.getByRole("heading", { name: "Create, review, and coordinate focused groups." })).toBeVisible();
        await page.getByRole("button", { name: "Release Readiness Workflow temporary workflow" }).click();

        await expect(page.getByRole("heading", { name: "Release Readiness Workflow temporary workflow" })).toBeVisible();
        await expect(page.getByTestId("groups-output-summary")).toContainText("3 outputs");
        await expect(page.getByText("Planning lane package", { exact: true })).toBeVisible();
        await expect(page.getByText("Validation lane checklist", { exact: true })).toBeVisible();
        await expect(page.getByText("Review lane summary", { exact: true })).toBeVisible();
    });
});
