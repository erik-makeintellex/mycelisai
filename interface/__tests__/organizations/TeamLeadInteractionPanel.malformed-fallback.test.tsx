import { beforeEach, describe, expect, it } from "vitest";
import { fireEvent, screen } from "@testing-library/react";
import { mockFetch } from "../setup";
import { jsonResponse, renderTeamLeadInteractionPanel } from "./teamLeadInteractionPanelTestSupport";

describe("TeamLeadInteractionPanel malformed fallback", () => {
    beforeEach(() => {
        mockFetch.mockReset();
        localStorage.clear();
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

        renderTeamLeadInteractionPanel();

        fireEvent.click(screen.getByRole("button", { name: /Run a quick strategy check/i }));

        expect(await screen.findByText("Soma for Northstar Labs guidance for Northstar Labs")).toBeDefined();
        expect(screen.getByText("Soma for Northstar Labs has guidance ready for Northstar Labs.")).toBeDefined();
        expect(screen.getByText("Turn Northstar Labs into a clear next move.")).toBeDefined();
        expect(screen.queryByText(/debug/i)).toBeNull();
        expect(screen.queryByText(/^contract$/i)).toBeNull();
    });
});
