import { beforeEach, describe, expect, it } from "vitest";
import { fireEvent, screen, waitFor } from "@testing-library/react";
import { mockFetch } from "../setup";
import { jsonResponse, renderTeamLeadInteractionPanel } from "./teamLeadInteractionPanelTestSupport";

describe("TeamLeadInteractionPanel native team contract", () => {
    beforeEach(() => {
        mockFetch.mockReset();
        localStorage.clear();
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

        renderTeamLeadInteractionPanel();

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

        renderTeamLeadInteractionPanel();

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
});
