import { beforeEach, describe, expect, it } from "vitest";
import { fireEvent, screen, waitFor } from "@testing-library/react";
import { mockFetch } from "../setup";
import { jsonResponse, renderTeamLeadInteractionPanel } from "./teamLeadInteractionPanelTestSupport";

describe("TeamLeadInteractionPanel retry flow", () => {
    beforeEach(() => {
        mockFetch.mockReset();
        localStorage.clear();
    });

    it("shows retry guidance when the Soma action request fails", async () => {
        mockFetch.mockResolvedValueOnce(jsonResponse({ ok: false, error: "Team Lead guidance is unavailable right now." }, 500));

        renderTeamLeadInteractionPanel();

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

        renderTeamLeadInteractionPanel();

        fireEvent.click(screen.getByRole("button", { name: /Choose the first priority/i }));
        expect(await screen.findByText("Soma guidance is unavailable")).toBeDefined();

        fireEvent.click(screen.getByRole("button", { name: "Retry Soma action" }));

        await waitFor(() => {
            expect(mockFetch).toHaveBeenCalledTimes(2);
        });
        expect(await screen.findByText("First focus for Northstar Labs")).toBeDefined();
        expect(screen.getByText("Priority steps")).toBeDefined();
    });
});
