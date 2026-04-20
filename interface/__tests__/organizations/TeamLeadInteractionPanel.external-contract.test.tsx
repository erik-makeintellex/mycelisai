import { beforeEach, describe, expect, it } from "vitest";
import { fireEvent, screen } from "@testing-library/react";
import { mockFetch } from "../setup";
import { jsonResponse, renderTeamLeadInteractionPanel } from "./teamLeadInteractionPanelTestSupport";

describe("TeamLeadInteractionPanel external workflow contract", () => {
    beforeEach(() => {
        mockFetch.mockReset();
        localStorage.clear();
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

        renderTeamLeadInteractionPanel();

        fireEvent.change(screen.getByLabelText("Tell Soma what team or delivery lane you want to create"), {
            target: { value: "Create an n8n workflow contract for inbound leads." },
        });
        fireEvent.click(screen.getByRole("button", { name: "Start team design" }));

        expect((await screen.findAllByText("External workflow contract")).length).toBeGreaterThan(0);
        expect(screen.getByText("n8n workflow contract")).toBeDefined();
        expect(screen.getByText("Normalized workflow result")).toBeDefined();
    });
});
