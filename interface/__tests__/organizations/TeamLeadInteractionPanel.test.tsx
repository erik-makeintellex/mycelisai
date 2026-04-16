import { describe, it, expect, beforeEach } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { mockFetch } from "../setup";
import TeamLeadInteractionPanel from "@/components/organizations/TeamLeadInteractionPanel";

function jsonResponse(body: unknown, status = 200) {
    return Promise.resolve(new Response(JSON.stringify(body), { status }));
}

describe("TeamLeadInteractionPanel", () => {
    beforeEach(() => {
        mockFetch.mockReset();
        localStorage.clear();
    });

    it("renders guided Soma actions without generic chat wording", () => {
        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                somaName="Soma for Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        expect(screen.getByText("Create teams with Soma")).toBeDefined();
        expect(screen.getByText("Team creation lane")).toBeDefined();
        expect(screen.getByRole("button", { name: "Start team design" })).toBeDefined();
        expect(screen.getByLabelText("Tell Soma what team or delivery lane you want to create")).toBeDefined();
        expect(screen.getByRole("button", { name: /Run a quick strategy check/i })).toBeDefined();
        expect(screen.getByRole("button", { name: /Choose the first priority/i })).toBeDefined();
        expect(screen.getByRole("button", { name: /Review your organization setup/i })).toBeDefined();
        expect(screen.getByText(/Each one should produce a clearer team-creation direction/i)).toBeDefined();
        expect(screen.queryByText(/chat/i)).toBeNull();
        expect(screen.queryByText(/context shell/i)).toBeNull();
        expect(screen.queryByText(/bounded slice/i)).toBeNull();
    });

    it("lets the operator seed the prompt from guided suggestion chips", () => {
        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                somaName="Soma for Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
                promptSuggestions={[
                    {
                        label: "Marketing launch team",
                        prompt: "Create a temporary marketing launch team for a new product rollout.",
                    },
                ]}
            />,
        );

        fireEvent.click(screen.getByRole("button", { name: "Marketing launch team" }));

        expect(
            (screen.getByLabelText("Tell Soma what team or delivery lane you want to create") as HTMLTextAreaElement).value,
        ).toBe("Create a temporary marketing launch team for a new product rollout.");
    });

    it("triggers a Soma action and renders structured guidance", async () => {
        mockFetch.mockResolvedValueOnce(jsonResponse({
            ok: true,
            data: {
                action: "plan_next_steps",
                request_label: "Run a quick strategy check",
                headline: "Team Lead plan for Northstar Labs",
                summary: "Team Lead recommends a clear next move for Northstar Labs.",
                priority_steps: [
                    "Align the first outcome with the AI Organization purpose.",
                    "Use the first Department as the routing layer for work.",
                ],
                suggested_follow_ups: [
                    "Review your organization setup",
                    "Choose the first priority",
                ],
            },
        }));

        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                somaName="Soma for Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        fireEvent.click(screen.getByRole("button", { name: /Run a quick strategy check/i }));

        await waitFor(() => {
            expect(mockFetch).toHaveBeenCalledWith(
                "/api/v1/organizations/org-123/workspace/actions",
                expect.objectContaining({
                    method: "POST",
                }),
            );
        });

        expect(await screen.findByText("Soma plan for Northstar Labs")).toBeDefined();
        expect(screen.getByText("Priority steps")).toBeDefined();
        expect(screen.getByText("Keep moving with")).toBeDefined();
        expect(screen.getAllByText("Review your organization setup").length).toBeGreaterThan(0);
    });

    it("uses the visible Soma prompt to choose a first guided action and shows what the operator asked for", async () => {
        mockFetch.mockResolvedValueOnce(jsonResponse({
            ok: true,
            data: {
                action: "focus_first",
                request_label: "Choose the first priority",
                headline: "First focus for Northstar Labs",
                summary: "Start by confirming the first outcome this AI Organization should deliver.",
                priority_steps: [
                    "Confirm the first priority for Northstar Labs.",
                    "Use that priority to create visible movement across the workspace.",
                ],
                suggested_follow_ups: [
                    "Run a quick strategy check",
                    "Review your organization setup",
                ],
            },
        }));

        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                somaName="Soma for Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        fireEvent.change(screen.getByLabelText("Tell Soma what team or delivery lane you want to create"), {
            target: { value: "Help me choose the first priority for this launch." },
        });
        fireEvent.click(screen.getByRole("button", { name: "Start team design" }));

        await waitFor(() => {
            expect(mockFetch).toHaveBeenCalledWith(
                "/api/v1/organizations/org-123/workspace/actions",
                expect.objectContaining({
                    method: "POST",
                    body: JSON.stringify({
                        action: "focus_first",
                        request_context: "Help me choose the first priority for this launch.",
                    }),
                }),
            );
        });

        expect(await screen.findByText("You asked Soma to help with")).toBeDefined();
        expect(screen.getAllByText("Help me choose the first priority for this launch.")).toHaveLength(2);
        expect(screen.getByText("First focus for Northstar Labs")).toBeDefined();
    });

    it("renders a native team execution path for image-oriented team requests", async () => {
        mockFetch.mockResolvedValueOnce(jsonResponse({
            ok: true,
            data: {
                action: "plan_next_steps",
                request_label: "Plan next steps for this organization",
                headline: "Team Lead plan for Northstar Labs",
                summary: "Team Lead recommends moving Northstar Labs into a focused first delivery loop.",
                priority_steps: [
                    "Align the first outcome with the AI Organization purpose.",
                ],
                suggested_follow_ups: [
                    "Review your organization setup",
                ],
                execution_contract: {
                    execution_mode: "native_team",
                    owner_label: "Native Mycelis team",
                    team_name: "Creative Delivery Team",
                    summary: "Use a bounded creative team inside Northstar Labs so Soma can shape the work and return the generated image as a managed artifact.",
                    target_outputs: [
                        "Reviewable image artifact",
                        "Short concept note",
                    ],
                    workstreams: [
                        {
                            label: "Creative direction lane",
                            owner_label: "Creative Delivery Team lead",
                            status: "ACTIVE",
                            summary: "Shape the visual direction before generation starts.",
                            next_step: "Lock the concept direction and hand it to the generation specialist.",
                            target_outputs: ["Short concept note"],
                        },
                        {
                            label: "Artifact generation lane",
                            owner_label: "Image generation specialist",
                            status: "NEXT",
                            summary: "Generate the image artifact and keep it reviewable.",
                            next_step: "Produce the first artifact candidate for review.",
                            target_outputs: ["Reviewable image artifact"],
                        },
                    ],
                    workflow_group: {
                        name: "Creative Delivery Team temporary workflow",
                        goal_statement: "Create a creative team to generate a launch hero image.",
                        work_mode: "propose_only",
                        coordinator_profile: "Creative Delivery Team lead",
                        allowed_capabilities: ["content.plan", "artifact.review"],
                        expiry_hours: 72,
                        summary: "Launch a temporary workflow group for Creative Delivery Team.",
                    },
                },
            },
        }));

        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                somaName="Soma for Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        fireEvent.change(screen.getByLabelText("Tell Soma what team or delivery lane you want to create"), {
            target: { value: "Create a creative team to generate a launch hero image." },
        });
        fireEvent.click(screen.getByRole("button", { name: "Start team design" }));

        await waitFor(() => {
            expect(mockFetch).toHaveBeenCalledWith(
                "/api/v1/organizations/org-123/workspace/actions",
                expect.objectContaining({
                    body: JSON.stringify({
                        action: "plan_next_steps",
                        request_context: "Create a creative team to generate a launch hero image.",
                    }),
                }),
            );
        });

        expect(await screen.findByText("Execution path")).toBeDefined();
        expect(screen.getAllByText("Native Mycelis team").length).toBeGreaterThan(0);
        expect(screen.getByText("Creative Delivery Team")).toBeDefined();
        expect(screen.getAllByText("Reviewable image artifact").length).toBeGreaterThan(0);
        expect(screen.getByText("Working together now")).toBeDefined();
        expect(screen.getByText("Creative direction lane")).toBeDefined();
        expect(screen.getByText("Artifact generation lane")).toBeDefined();
        expect(screen.getByRole("button", { name: "Create temporary workflow group" })).toBeDefined();
    });

    it("teaches compact team defaults and renders broad orchestration hints when the ask spans multiple functions", async () => {
        mockFetch.mockResolvedValueOnce(jsonResponse({
            ok: true,
            data: {
                action: "plan_next_steps",
                request_label: "Plan next steps for this organization",
                headline: "Team Lead plan for Northstar Labs",
                summary: "Team Lead recommends a broad multi-team launch approach for Northstar Labs.",
                priority_steps: [
                    "Split the launch into focused lanes.",
                ],
                suggested_follow_ups: [
                    "Review your organization setup",
                ],
                execution_contract: {
                    execution_mode: "native_team",
                    owner_label: "Native Mycelis team",
                    team_name: "Launch Coordination Team",
                    summary: "Split the work across several compact teams and coordinate the handoffs over NATS.",
                    target_outputs: [
                        "Launch plan",
                        "Website brief",
                        "Media brief",
                    ],
                    recommended_team_shape: "3-6 people per team",
                    coordination_model: "multi-team bundle",
                    recommended_team_count: 3,
                    recommended_team_member_limit: 6,
                    workstreams: [
                        {
                            label: "Planning lane",
                            owner_label: "Planning lane lead",
                            status: "ACTIVE",
                            summary: "Define the scope and retained package.",
                            next_step: "Publish the retained planning package.",
                            target_outputs: ["Launch plan"],
                        },
                        {
                            label: "Validation lane",
                            owner_label: "Validation lane lead",
                            status: "NEXT",
                            summary: "Turn the plan into proof steps.",
                            next_step: "Capture the first validation pass.",
                            target_outputs: ["Website brief"],
                        },
                        {
                            label: "Review lane",
                            owner_label: "Review lane lead",
                            status: "NEXT",
                            summary: "Review the outputs together and name the next owner.",
                            next_step: "Call out the next follow-through step.",
                            target_outputs: ["Media brief"],
                        },
                    ],
                },
            },
        }));

        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                somaName="Soma for Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        fireEvent.change(screen.getByLabelText("Tell Soma what team or delivery lane you want to create"), {
            target: { value: "Create a company-wide launch across marketing, web, support, and leadership." },
        });
        fireEvent.click(screen.getByRole("button", { name: "Start team design" }));

        await waitFor(() => {
            expect(mockFetch).toHaveBeenCalledWith(
                "/api/v1/organizations/org-123/workspace/actions",
                expect.objectContaining({
                    method: "POST",
                    body: JSON.stringify({
                        action: "plan_next_steps",
                        request_context: "Create a company-wide launch across marketing, web, support, and leadership.",
                    }),
                }),
            );
        });

        expect(await screen.findByText("Compact team default")).toBeDefined();
        expect(screen.getByText(/several small teams or lane bundles/i)).toBeDefined();
        expect(screen.getByText(/Council plus NATS/i)).toBeDefined();
        expect(screen.getByText("Compact orchestration hints")).toBeDefined();
        expect(screen.getByText("3-6 people per team")).toBeDefined();
        expect(screen.getByText("multi-team bundle")).toBeDefined();
        expect(screen.getByText("3 teams")).toBeDefined();
        expect(screen.getByText("member limit 6")).toBeDefined();
        expect(screen.getByText("Planning lane")).toBeDefined();
        expect(screen.getByText("Validation lane")).toBeDefined();
        expect(screen.getByText("Review lane")).toBeDefined();
    });

    it("creates a temporary workflow group directly from Soma guidance and links to the created group", async () => {
        mockFetch
            .mockResolvedValueOnce(jsonResponse({
                ok: true,
                data: {
                    action: "plan_next_steps",
                    request_label: "Plan next steps for this organization",
                    headline: "Team Lead plan for Northstar Labs",
                    summary: "Team Lead recommends moving Northstar Labs into a focused first delivery loop.",
                    priority_steps: [
                        "Align the first outcome with the AI Organization purpose.",
                    ],
                    suggested_follow_ups: [
                        "Review your organization setup",
                    ],
                    execution_contract: {
                        execution_mode: "native_team",
                        owner_label: "Native Mycelis team",
                        team_name: "Marketing Launch Team",
                        summary: "Use a bounded marketing team inside Northstar Labs.",
                        target_outputs: [
                            "Launch plan",
                            "Messaging brief",
                            "Campaign asset list",
                        ],
                        workflow_group: {
                            name: "Marketing Launch Team temporary workflow",
                            goal_statement: "Create a temporary marketing launch team for a new product rollout.",
                            work_mode: "propose_only",
                            coordinator_profile: "Marketing Launch Team lead",
                            allowed_capabilities: ["team.coordinate", "artifact.review"],
                            expiry_hours: 72,
                            summary: "Launch a temporary workflow group for Marketing Launch Team.",
                        },
                    },
                },
            }))
            .mockResolvedValueOnce(jsonResponse({
                ok: true,
                data: {
                    group_id: "group-temp-launch",
                    name: "Marketing Launch Team temporary workflow",
                },
            }, 201));

        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                somaName="Soma for Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        fireEvent.change(screen.getByLabelText("Tell Soma what team or delivery lane you want to create"), {
            target: { value: "Create a temporary marketing launch team for a new product rollout." },
        });
        fireEvent.click(screen.getByRole("button", { name: "Start team design" }));

        expect(await screen.findByRole("button", { name: "Create temporary workflow group" })).toBeDefined();
        expect(screen.getByText("This ask looks compact enough for one focused team.")).toBeDefined();
        fireEvent.click(screen.getByRole("button", { name: "Create temporary workflow group" }));

        await waitFor(() => {
            expect(mockFetch).toHaveBeenLastCalledWith(
                "/api/v1/groups",
                expect.objectContaining({
                    method: "POST",
                    body: expect.stringContaining("\"name\":\"Marketing Launch Team temporary workflow\""),
                }),
            );
        });

        expect(await screen.findByText(/Soma launched Marketing Launch Team temporary workflow/i)).toBeDefined();
        expect(screen.getByRole("link", { name: "Open Marketing Launch Team temporary workflow" }).getAttribute("href")).toBe("/groups?group_id=group-temp-launch");
    });

    it("renders an external workflow contract path without stripping the contract wording", async () => {
        mockFetch.mockResolvedValueOnce(jsonResponse({
            ok: true,
            data: {
                action: "plan_next_steps",
                request_label: "Plan next steps for this organization",
                headline: "Team Lead plan for Northstar Labs",
                summary: "Team Lead recommends moving Northstar Labs into a focused first delivery loop.",
                priority_steps: [
                    "Align the first outcome with the AI Organization purpose.",
                ],
                suggested_follow_ups: [
                    "Review your organization setup",
                ],
                execution_contract: {
                    execution_mode: "external_workflow_contract",
                    owner_label: "External workflow contract",
                    external_target: "n8n workflow contract",
                    summary: "This request is best handled as an external workflow contract so Mycelis can keep the result return clear.",
                    target_outputs: [
                        "Normalized workflow result",
                        "Linked artifact or execution note",
                    ],
                },
            },
        }));

        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                somaName="Soma for Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        fireEvent.change(screen.getByLabelText("Tell Soma what team or delivery lane you want to create"), {
            target: { value: "Create an n8n workflow contract for inbound leads." },
        });
        fireEvent.click(screen.getByRole("button", { name: "Start team design" }));

        expect((await screen.findAllByText("External workflow contract")).length).toBeGreaterThan(0);
        expect(screen.getByText("n8n workflow contract")).toBeDefined();
        expect(screen.getByText("Normalized workflow result")).toBeDefined();
    });

    it("lets Start with Soma run a quick strategy check even when the prompt is blank", async () => {
        mockFetch.mockResolvedValueOnce(jsonResponse({
            ok: true,
            data: {
                action: "plan_next_steps",
                request_label: "Run a quick strategy check",
                headline: "Soma plan for Northstar Labs",
                summary: "Soma recommends the clearest next move for Northstar Labs.",
                priority_steps: [
                    "Align the first outcome with the AI Organization purpose.",
                ],
                suggested_follow_ups: [
                    "Choose the first priority",
                ],
            },
        }));

        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                somaName="Soma for Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        fireEvent.click(screen.getByRole("button", { name: "Start team design" }));

        await waitFor(() => {
            expect(mockFetch).toHaveBeenCalledWith(
                "/api/v1/organizations/org-123/workspace/actions",
                expect.objectContaining({
                    method: "POST",
                    body: JSON.stringify({ action: "plan_next_steps" }),
                }),
            );
        });

        expect(await screen.findByText("Soma plan for Northstar Labs")).toBeDefined();
    });

    it("shows a loading state and prevents duplicate action submission while guidance is loading", async () => {
        let resolveResponse!: (value: Response) => void;
        mockFetch.mockImplementationOnce(
            () =>
                new Promise<Response>((resolve) => {
                    resolveResponse = resolve;
                }),
        );

        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                somaName="Soma for Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        const actionButton = screen.getByRole("button", { name: /Run a quick strategy check/i });
        fireEvent.click(actionButton);
        fireEvent.click(actionButton);

        expect(await screen.findByText("Soma is preparing guidance for this AI Organization...")).toBeDefined();
        expect(actionButton.hasAttribute("disabled")).toBe(true);
        expect(mockFetch).toHaveBeenCalledTimes(1);

        resolveResponse(
            new Response(
                JSON.stringify({
                    ok: true,
                    data: {
                        action: "plan_next_steps",
                        request_label: "Run a quick strategy check",
                        headline: "Team Lead plan for Northstar Labs",
                        summary: "Team Lead recommends a clear next move for Northstar Labs.",
                        priority_steps: ["Align the first outcome with the AI Organization purpose."],
                        suggested_follow_ups: ["Review your organization setup"],
                    },
                }),
                { status: 200 },
            ),
        );

        expect(await screen.findByText("Soma plan for Northstar Labs")).toBeDefined();
    });

    it("persists the Soma draft and last guidance across remount for the same organization", async () => {
        mockFetch.mockResolvedValueOnce(jsonResponse({
            ok: true,
            data: {
                action: "focus_first",
                request_label: "Choose the first priority",
                headline: "First focus for Northstar Labs",
                summary: "Start by confirming the first outcome this AI Organization should deliver.",
                priority_steps: [
                    "Confirm the first priority for Northstar Labs.",
                ],
                suggested_follow_ups: [
                    "Review your organization setup",
                ],
            },
        }));

        const { unmount } = render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                somaName="Soma for Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        fireEvent.change(screen.getByLabelText("Tell Soma what team or delivery lane you want to create"), {
            target: { value: "Help me choose the first priority for this launch." },
        });
        fireEvent.click(screen.getByRole("button", { name: "Start team design" }));

        expect(await screen.findByText("First focus for Northstar Labs")).toBeDefined();
        unmount();

        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                somaName="Soma for Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        expect(screen.getByDisplayValue("Help me choose the first priority for this launch.")).toBeDefined();
        expect(screen.getByText("First focus for Northstar Labs")).toBeDefined();
        expect(screen.getByText("You asked Soma to help with")).toBeDefined();
    });

    it("renders readable fallback guidance when the backend returns a partial malformed payload", async () => {
        mockFetch.mockResolvedValueOnce(jsonResponse({
            ok: true,
            data: {
                action: "plan_next_steps",
                request_label: "system: debug trace",
                headline: "",
                summary: "{debug:true}",
                priority_steps: [
                    "debug: internal trace",
                    "Turn Northstar Labs into a clear next move.",
                ],
                suggested_follow_ups: [
                    "contract",
                    "Choose the first priority",
                ],
            },
        }));

        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                somaName="Soma for Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        fireEvent.click(screen.getByRole("button", { name: /Run a quick strategy check/i }));

        expect(await screen.findByText("Soma for Northstar Labs guidance for Northstar Labs")).toBeDefined();
        expect(screen.getByText("Soma for Northstar Labs has guidance ready for Northstar Labs.")).toBeDefined();
        expect(screen.getByText("Turn Northstar Labs into a clear next move.")).toBeDefined();
        expect(screen.queryByText(/debug/i)).toBeNull();
        expect(screen.queryByText(/^contract$/i)).toBeNull();
    });

    it("shows retry guidance when the Soma action request fails", async () => {
        mockFetch.mockResolvedValueOnce(jsonResponse({ ok: false, error: "Team Lead guidance is unavailable right now." }, 500));

        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                somaName="Soma for Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        fireEvent.click(screen.getByRole("button", { name: /Choose the first priority/i }));

        expect(await screen.findByText("Soma guidance is unavailable")).toBeDefined();
        expect(screen.getByRole("button", { name: "Retry Soma action" })).toBeDefined();
        expect(screen.getByText(/without leaving Northstar Labs/i)).toBeDefined();
        expect(screen.getByText(/choose another guided Soma action below/i)).toBeDefined();
    });

    it("retries the same Soma action successfully after a failure", async () => {
        mockFetch
            .mockResolvedValueOnce(jsonResponse({ ok: false, error: "Team Lead guidance is unavailable right now." }, 500))
            .mockResolvedValueOnce(jsonResponse({
                ok: true,
                data: {
                    action: "focus_first",
                    request_label: "Choose the first priority",
                    headline: "First focus for Northstar Labs",
                    summary: "Start by confirming the first outcome this AI Organization should deliver.",
                    priority_steps: [
                        "Use 1 Department as the first routing layer for work.",
                        "Use 1 Advisor when the Team Lead needs review or decision support.",
                    ],
                    suggested_follow_ups: [
                        "Run a quick strategy check",
                        "Review your organization setup",
                    ],
                },
            }));

        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                somaName="Soma for Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        fireEvent.click(screen.getByRole("button", { name: /Choose the first priority/i }));
        expect(await screen.findByText("Soma guidance is unavailable")).toBeDefined();

        fireEvent.click(screen.getByRole("button", { name: "Retry Soma action" }));

        await waitFor(() => {
            expect(mockFetch).toHaveBeenCalledTimes(2);
        });
        expect(await screen.findByText("First focus for Northstar Labs")).toBeDefined();
        expect(screen.getByText("Priority steps")).toBeDefined();
    });

    it("renders retained package continuity guidance without offering a temporary group launch", async () => {
        mockFetch.mockResolvedValueOnce(jsonResponse({
            ok: true,
            data: {
                action: "resume_retained_package",
                request_label: "Resume retained package continuity",
                headline: "Retained package continuity for Northstar Labs",
                summary: "Team Lead resumes the retained package for Northstar Labs so completed work stays durable and the next step stays explicit after a reboot or reload.",
                priority_steps: [
                    "Open the retained package and confirm the latest durable outputs.",
                    "Record what is already complete, what remains, and who owns the next step.",
                ],
                suggested_follow_ups: [
                    "Run a quick strategy check",
                    "Review your organization setup",
                ],
                execution_contract: {
                    execution_mode: "continuity_resume",
                    owner_label: "Release Workflow Coordinator",
                    summary: "Resume Northstar Labs from the Release Readiness Workflow retained package, confirm the finished outputs, and keep the next owner explicit instead of rebuilding finished work.",
                    continuity_label: "Release Readiness Workflow",
                    continuity_summary: "Release Readiness Workflow is archived and already retains Review lane summary and Validation lane checklist.",
                    resume_checkpoint: "Open Release Readiness Workflow and continue with Review Lead after reviewing Review lane summary.",
                    target_outputs: [
                        "Review lane summary",
                        "Validation lane checklist",
                    ],
                    workstreams: [
                        {
                            label: "Completed work lane",
                            owner_label: "Release Workflow Coordinator",
                            status: "COMPLETE",
                            summary: "Use durable outputs like Review lane summary and Validation lane checklist as the completed baseline already captured in Release Readiness Workflow.",
                            next_step: "Keep Review lane summary and Validation lane checklist linked to Release Readiness Workflow so the restart point stays anchored in retained work.",
                            target_outputs: ["Review lane summary"],
                        },
                        {
                            label: "Continuity briefing lane",
                            owner_label: "Release Workflow Coordinator",
                            status: "ACTIVE",
                            summary: "Turn Release Readiness Workflow into a readable continuity brief so the rebooted workspace can resume without rebuilding finished work.",
                            next_step: "Publish the retained package summary for Release Readiness Workflow and highlight Review Lead as the next owner.",
                            target_outputs: ["Validation lane checklist"],
                        },
                        {
                            label: "Next-step handoff lane",
                            owner_label: "Review Lead",
                            status: "NEXT",
                            summary: "Hand the retained package to Review Lead with the finished outputs and the remaining step made explicit.",
                            next_step: "Continue from Review lane summary after Review Lead reviews the retained outputs.",
                            target_outputs: ["Validation lane checklist"],
                        },
                    ],
                    workflow_group: {
                        group_id: "group-retained",
                        name: "Release Readiness Workflow",
                        goal_statement: "Resume the retained package for Northstar Labs after a reboot or reload.",
                        work_mode: "resume_continuity",
                        coordinator_profile: "release-workflow-coordinator",
                        allowed_capabilities: ["artifact.review", "team.coordinate"],
                        expiry_hours: 4,
                        summary: "Reopen the Release Readiness Workflow retained package, review Review lane summary and Validation lane checklist, and continue with Review Lead as the next owner.",
                    },
                },
            },
        }));

        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                somaName="Soma for Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        fireEvent.click(screen.getByRole("button", { name: /Resume retained package/i }));

        expect((await screen.findAllByText("Retained package continuity")).length).toBeGreaterThan(0);
        expect(screen.getByText("Open Release Readiness Workflow and continue with Review Lead after reviewing Review lane summary.")).toBeDefined();
        expect(screen.getAllByText("Review lane summary").length).toBeGreaterThan(0);
        expect(screen.getByText("Completed work lane")).toBeDefined();
        expect(screen.getByText("Continuity briefing lane")).toBeDefined();
        expect(screen.getByText("Next-step handoff lane")).toBeDefined();
        expect(screen.getByText(/review-first/i)).toBeDefined();
        expect(screen.getByRole("link", { name: "Open retained package" }).getAttribute("href")).toBe("/groups?group_id=group-retained");
        expect(screen.queryByRole("button", { name: "Create temporary workflow group" })).toBeNull();
    });
});


