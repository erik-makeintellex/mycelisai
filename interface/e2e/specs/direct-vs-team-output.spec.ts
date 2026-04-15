import { expect, test, type Page, type Route } from "@playwright/test";

const organizationId = "org-workflow-variants";
const chatPlaceholder = /Tell .* what you want to plan, review, create, or execute/i;

type GroupRecord = {
    group_id: string;
    name: string;
    goal_statement: string;
    status: string;
    work_mode: string;
    member_user_ids: string[];
    team_ids: string[];
    coordinator_profile: string;
    approval_policy_ref: string;
    expiry: string;
    created_by: string;
    created_at: string;
};

type ArtifactRecord = {
    id: string;
    title: string;
    artifact_type: string;
    content_type: string;
    content?: string;
    file_path?: string;
    team_id: string;
    agent_id: string;
    metadata: Record<string, unknown>;
    status: string;
    created_at: string;
};

async function fulfillJSON(route: Route, status: number, body: unknown) {
    await route.fulfill({
        status,
        contentType: "application/json",
        body: JSON.stringify(body),
    });
}

async function gotoWithColdStartRetry(page: Page, path: string) {
    try {
        await page.goto(path, { waitUntil: "domcontentloaded" });
    } catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        if (!message.includes("net::ERR_ABORTED") && !message.includes("frame was detached")) {
            throw error;
        }
        await page.goto(path, { waitUntil: "domcontentloaded" });
    }
}

async function openOrganization(page: Page) {
    await gotoWithColdStartRetry(page, `/organizations/${organizationId}`);
    await page.getByPlaceholder(chatPlaceholder).waitFor({ timeout: 20_000 });
    await expect(page.getByRole("heading", { name: "Talk with Soma" })).toBeVisible();
}

async function sendWorkspaceMessage(page: Page, content: string) {
    const input = page.getByPlaceholder(chatPlaceholder);
    await input.fill(content);
    await input.press("Enter");
}

test.describe("Direct Soma vs team-managed output package", () => {
    test("proves direct answer, compact team package, multi-lane package, and reload-safe retained output review", async ({
        page,
    }) => {
        const groups: GroupRecord[] = [];
        const groupArtifacts: Record<string, ArtifactRecord[]> = {};
        let actionsCallCount = 0;

        await page.addInitScript(() => {
            window.localStorage.setItem("mycelis-last-organization-id", "org-workflow-variants");
            window.localStorage.setItem("mycelis-last-organization-name", "Northstar Labs");
        });

        await page.route("**/api/v1/user/me", async (route) => {
            await fulfillJSON(route, 200, {
                ok: true,
                data: {
                    id: "operator-1",
                    name: "Operator",
                    email: "operator@example.test",
                },
            });
        });

        await page.route("**/api/v1/services/status", async (route) => {
            await fulfillJSON(route, 200, {
                ok: true,
                data: [
                    { name: "nats", status: "online" },
                    { name: "postgres", status: "online" },
                    { name: "reactive", status: "online" },
                ],
            });
        });

        await page.route(`**/api/v1/organizations/${organizationId}/home`, async (route) => {
            await fulfillJSON(route, 200, {
                ok: true,
                data: {
                    id: organizationId,
                    name: "Northstar Labs",
                    purpose: "Run a product-ready Soma-first workflow with durable retained outputs.",
                    start_mode: "template",
                    team_lead_label: "Team Lead",
                    advisor_count: 1,
                    department_count: 2,
                    specialist_count: 4,
                    ai_engine_settings_summary: "Balanced",
                    response_contract_summary: "Clear and structured",
                    memory_personality_summary: "Stable continuity",
                    status: "ready",
                },
            });
        });

        await page.route("**/api/v1/chat", async (route) => {
            const requestBody = route.request().postDataJSON() as {
                messages?: Array<{ content?: string }>;
            };
            const content = requestBody?.messages?.[requestBody.messages.length - 1]?.content ?? "";

            if (content.includes("Resume the release-readiness work")) {
                await fulfillJSON(route, 200, {
                    ok: true,
                    data: {
                        meta: { source_node: "admin", timestamp: "2026-04-15T20:10:00Z" },
                        signal_type: "chat.reply",
                        trust_score: 0.95,
                        template_id: "chat-to-answer",
                        mode: "answer",
                        payload: {
                            text: "Already done: planning lane package, validation checklist, and review summary are retained. Remaining: confirm the live Windows operator validation pass. Next owner: Review lane lead.",
                            tools_used: [],
                            consultations: [],
                            artifacts: [],
                        },
                    },
                });
                return;
            }

            await fulfillJSON(route, 200, {
                ok: true,
                data: {
                    meta: { source_node: "admin", timestamp: "2026-04-15T20:00:00Z" },
                    signal_type: "chat.reply",
                    trust_score: 0.93,
                    template_id: "chat-to-answer",
                    mode: "answer",
                    payload: {
                        text: "Use the self-hosted Kubernetes lane with an explicit Windows AI endpoint, then run the Windows browser validation flow against the retained output and continuity checks.",
                        tools_used: [],
                        consultations: [],
                        artifacts: [],
                    },
                },
            });
        });

        await page.route(`**/api/v1/organizations/${organizationId}/workspace/actions`, async (route) => {
            actionsCallCount += 1;
            if (actionsCallCount === 1) {
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
                return;
            }

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
                        target_outputs: [
                            "Planning lane package",
                            "Validation lane checklist",
                            "Review lane summary",
                        ],
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

            const groupId = groups.length === 0 ? "group-compact-release" : "group-multilane-release";
            const isCompact = groups.length === 0;
            const created: GroupRecord = {
                group_id: groupId,
                name: isCompact
                    ? "Release Readiness Team temporary workflow"
                    : "Release Readiness Workflow temporary workflow",
                goal_statement: isCompact
                    ? "Produce a compact release-readiness package for the Windows self-hosted lane."
                    : "Split release-readiness work into planning, validation, and review lanes.",
                work_mode: "propose_only",
                member_user_ids: ["owner"],
                team_ids: isCompact
                    ? ["release-lead", "validation-lead"]
                    : ["planning-lead", "validation-lead", "review-lead"],
                coordinator_profile: isCompact ? "Release Readiness lead" : "Release workflow coordinator",
                approval_policy_ref: "browser-proof",
                status: "active",
                expiry: "2026-04-18T18:00:00Z",
                created_by: "owner",
                created_at: "2026-04-15T20:01:00Z",
            };
            groups.push(created);
            groupArtifacts[groupId] = isCompact
                ? [
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
                  ]
                : [
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
                    group_id: groupId,
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

        await sendWorkspaceMessage(
            page,
            "Give me the shortest practical recommendation for how to validate this Windows self-hosted release lane.",
        );
        await expect(
            page.getByText(
                "Use the self-hosted Kubernetes lane with an explicit Windows AI endpoint, then run the Windows browser validation flow against the retained output and continuity checks.",
            ),
        ).toBeVisible();

        await gotoWithColdStartRetry(page, "/teams/create");
        await expect(page.getByRole("heading", { name: "Create a team through Soma" })).toBeVisible();

        const requestField = page.getByLabel("Tell Soma what team or delivery lane you want to create");
        await requestField.fill(
            "Create the smallest useful team to produce a release-readiness package for this Windows self-hosted lane.",
        );
        await page.getByRole("button", { name: "Start team design" }).click();

        await expect(page.getByText("Compact release-readiness package", { exact: true })).toBeVisible();
        await expect(page.getByText("Release Readiness Team", { exact: true })).toBeVisible();
        await expect(page.getByText("Validation checklist", { exact: true })).toBeVisible();
        await expect(page.getByText("Deployment recommendation", { exact: true })).toBeVisible();
        await expect(page.getByText("Risk review", { exact: true })).toBeVisible();

        await page.getByRole("button", { name: "Create temporary workflow group" }).click();
        const openCompactLink = page.getByRole("link", {
            name: "Open Release Readiness Team temporary workflow",
        });
        await expect(openCompactLink).toHaveAttribute("href", "/groups?group_id=group-compact-release");
        await openCompactLink.click();

        await expect(
            page.getByRole("heading", { name: "Release Readiness Team temporary workflow" }),
        ).toBeVisible();
        await expect(page.getByTestId("groups-output-summary")).toContainText("2 outputs");
        await expect(page.getByText("Validation checklist", { exact: true })).toBeVisible();
        await expect(page.getByText("Risk review", { exact: true })).toBeVisible();
        await page.getByRole("button", { name: "Archive temporary group" }).click();
        await expect(page.getByText("Archived temporary group", { exact: true })).toBeVisible();
        await page.reload({ waitUntil: "domcontentloaded" });
        await expect(page.getByText("Archived temporary group", { exact: true })).toBeVisible();
        await expect(page.getByText("Validation checklist", { exact: true })).toBeVisible();
        await expect(page.getByText("Risk review", { exact: true })).toBeVisible();

        await gotoWithColdStartRetry(page, "/teams/create");
        await expect(page.getByRole("heading", { name: "Create a team through Soma" })).toBeVisible();
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
        const openMultiLaneLink = page.getByRole("link", {
            name: "Open Release Readiness Workflow temporary workflow",
        });
        await expect(openMultiLaneLink).toHaveAttribute("href", "/groups?group_id=group-multilane-release");
        await openMultiLaneLink.click();

        await expect(
            page.getByRole("heading", { name: "Release Readiness Workflow temporary workflow" }),
        ).toBeVisible();
        await expect(page.getByTestId("groups-output-summary")).toContainText("3 outputs");
        await expect(page.getByText("Planning lane package", { exact: true })).toBeVisible();
        await expect(page.getByText("Validation lane checklist", { exact: true })).toBeVisible();
        await expect(page.getByText("Review lane summary", { exact: true })).toBeVisible();
        await page.getByRole("button", { name: "Archive temporary group" }).click();
        await expect(page.getByText("Archived temporary group", { exact: true })).toBeVisible();
        await page.reload({ waitUntil: "domcontentloaded" });
        await expect(page.getByText("Planning lane package", { exact: true })).toBeVisible();
        await expect(page.getByText("Validation lane checklist", { exact: true })).toBeVisible();
        await expect(page.getByText("Review lane summary", { exact: true })).toBeVisible();

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
