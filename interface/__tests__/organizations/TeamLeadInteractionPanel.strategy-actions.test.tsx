import { beforeEach, describe, expect, it } from "vitest";
import { fireEvent, screen, waitFor } from "@testing-library/react";
import { mockFetch } from "../setup";
import { jsonResponse, renderTeamLeadInteractionPanel } from "./teamLeadInteractionPanelTestSupport";

describe("TeamLeadInteractionPanel strategy actions", () => {
    beforeEach(() => {
        mockFetch.mockReset();
        localStorage.clear();
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

        renderTeamLeadInteractionPanel();

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

        renderTeamLeadInteractionPanel();

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

        renderTeamLeadInteractionPanel();

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
});
