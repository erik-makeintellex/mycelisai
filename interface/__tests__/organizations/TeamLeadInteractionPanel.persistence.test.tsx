import { beforeEach, describe, expect, it } from "vitest";
import { fireEvent, screen } from "@testing-library/react";
import { mockFetch } from "../setup";
import { jsonResponse, renderTeamLeadInteractionPanel } from "./teamLeadInteractionPanelTestSupport";

describe("TeamLeadInteractionPanel persistence", () => {
    beforeEach(() => {
        mockFetch.mockReset();
        localStorage.clear();
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

        const { unmount } = renderTeamLeadInteractionPanel();

        fireEvent.change(screen.getByLabelText("Tell Soma what team or delivery lane you want to create"), {
            target: { value: "Help me choose the first priority for this launch." },
        });
        fireEvent.click(screen.getByRole("button", { name: "Start team design" }));

        expect(await screen.findByText("First focus for Northstar Labs")).toBeDefined();
        unmount();

        renderTeamLeadInteractionPanel();

        expect(screen.getByDisplayValue("Help me choose the first priority for this launch.")).toBeDefined();
        expect(screen.getByText("First focus for Northstar Labs")).toBeDefined();
        expect(screen.getByText("You asked Soma to help with")).toBeDefined();
    });
});
