import { beforeEach, describe, expect, it } from "vitest";
import { fireEvent, screen, waitFor } from "@testing-library/react";
import { mockFetch } from "../setup";
import { jsonResponse, renderTeamLeadInteractionPanel } from "./teamLeadInteractionPanelTestSupport";

describe("TeamLeadInteractionPanel compact team defaults", () => {
    beforeEach(() => {
        mockFetch.mockReset();
        localStorage.clear();
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

        renderTeamLeadInteractionPanel();

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
});
