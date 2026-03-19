import { describe, it, expect, beforeEach } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { mockFetch } from "../setup";
import TeamLeadInteractionPanel from "@/components/organizations/TeamLeadInteractionPanel";

function jsonResponse(body: unknown, status = 200) {
    return Promise.resolve(new Response(JSON.stringify(body), { status }));
}

describe("TeamLeadInteractionPanel", () => {
    beforeEach(() => {
        mockFetch.mockReset();
    });

    it("renders guided Team Lead actions without generic chat wording", () => {
        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        expect(screen.getByText("Work with the Team Lead")).toBeDefined();
        expect(screen.getByRole("button", { name: /Plan next steps for this organization/i })).toBeDefined();
        expect(screen.getByRole("button", { name: /What should I focus on first\?/i })).toBeDefined();
        expect(screen.getByRole("button", { name: /Review my organization setup/i })).toBeDefined();
        expect(screen.queryByText(/chat/i)).toBeNull();
        expect(screen.queryByText(/context shell/i)).toBeNull();
        expect(screen.queryByText(/bounded slice/i)).toBeNull();
    });

    it("triggers a Team Lead action and renders structured guidance", async () => {
        mockFetch.mockResolvedValueOnce(jsonResponse({
            ok: true,
            data: {
                action: "plan_next_steps",
                request_label: "Plan next steps for this organization",
                headline: "Team Lead plan for Northstar Labs",
                summary: "Team Lead recommends a focused first delivery loop.",
                priority_steps: [
                    "Align the first outcome with the AI Organization purpose.",
                    "Use the first Department as the routing layer for work.",
                ],
                suggested_follow_ups: [
                    "Review my organization setup",
                    "What should I focus on first?",
                ],
            },
        }));

        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        fireEvent.click(screen.getByRole("button", { name: /Plan next steps for this organization/i }));

        await waitFor(() => {
            expect(mockFetch).toHaveBeenCalledWith(
                "/api/v1/organizations/org-123/workspace/actions",
                expect.objectContaining({
                    method: "POST",
                }),
            );
        });

        expect(await screen.findByText("Team Lead plan for Northstar Labs")).toBeDefined();
        expect(screen.getByText("Priority steps")).toBeDefined();
        expect(screen.getByText("Keep moving with")).toBeDefined();
        expect(screen.getAllByText("Review my organization setup").length).toBeGreaterThan(0);
    });

    it("shows retry guidance when the Team Lead action request fails", async () => {
        mockFetch.mockResolvedValueOnce(jsonResponse({ ok: false, error: "Team Lead guidance is unavailable right now." }, 500));

        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        fireEvent.click(screen.getByRole("button", { name: /What should I focus on first\?/i }));

        expect(await screen.findByText("Team Lead guidance is unavailable")).toBeDefined();
        expect(screen.getByRole("button", { name: "Retry Team Lead action" })).toBeDefined();
    });
});

