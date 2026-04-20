import { beforeEach, describe, expect, it } from "vitest";
import { fireEvent, screen } from "@testing-library/react";
import { mockFetch } from "../setup";
import { renderTeamLeadInteractionPanel } from "./teamLeadInteractionPanelTestSupport";

describe("TeamLeadInteractionPanel guidance copy", () => {
    beforeEach(() => {
        mockFetch.mockReset();
        localStorage.clear();
    });

    it("renders guided Soma actions without generic chat wording", () => {
        renderTeamLeadInteractionPanel();

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
        renderTeamLeadInteractionPanel({
            promptSuggestions: [
                {
                    label: "Marketing launch team",
                    prompt: "Create a temporary marketing launch team for a new product rollout.",
                },
            ],
        });

        fireEvent.click(screen.getByRole("button", { name: "Marketing launch team" }));

        expect(
            (screen.getByLabelText("Tell Soma what team or delivery lane you want to create") as HTMLTextAreaElement).value,
        ).toBe("Create a temporary marketing launch team for a new product rollout.");
    });
});
