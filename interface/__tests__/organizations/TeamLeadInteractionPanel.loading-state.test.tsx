import { beforeEach, describe, expect, it } from "vitest";
import { fireEvent, screen } from "@testing-library/react";
import { mockFetch } from "../setup";
import { renderTeamLeadInteractionPanel } from "./teamLeadInteractionPanelTestSupport";

describe("TeamLeadInteractionPanel loading state", () => {
    beforeEach(() => {
        mockFetch.mockReset();
        localStorage.clear();
    });

    it("shows a loading state and prevents duplicate action submission while guidance is loading", async () => {
        let resolveResponse!: (value: Response) => void;
        mockFetch.mockImplementationOnce(
            () =>
                new Promise<Response>((resolve) => {
                    resolveResponse = resolve;
                }),
        );

        renderTeamLeadInteractionPanel();

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
});
