import { beforeEach, describe, expect, it } from "vitest";
import { fireEvent, screen } from "@testing-library/react";
import { mockFetch } from "../setup";
import { jsonResponse, renderTeamLeadInteractionPanel } from "./teamLeadInteractionPanelTestSupport";

describe("TeamLeadInteractionPanel retained package continuity", () => {
    beforeEach(() => {
        mockFetch.mockReset();
        localStorage.clear();
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

        renderTeamLeadInteractionPanel();

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
