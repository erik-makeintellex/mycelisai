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
        expect(screen.getByRole("button", { name: /Run a quick strategy check/i })).toBeDefined();
        expect(screen.getByRole("button", { name: /Choose the first priority/i })).toBeDefined();
        expect(screen.getByRole("button", { name: /Review your organization setup/i })).toBeDefined();
        expect(screen.getByText(/Each one updates this workspace immediately/i)).toBeDefined();
        expect(screen.queryByText(/chat/i)).toBeNull();
        expect(screen.queryByText(/context shell/i)).toBeNull();
        expect(screen.queryByText(/bounded slice/i)).toBeNull();
    });

    it("triggers a Team Lead action and renders structured guidance", async () => {
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

        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        fireEvent.click(screen.getByRole("button", { name: /Run a quick strategy check/i }));

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
        expect(screen.getAllByText("Review your organization setup").length).toBeGreaterThan(0);
    });

    it("shows a loading state and prevents duplicate action submission while guidance is loading", async () => {
        let resolveResponse!: (value: Response) => void;
        mockFetch.mockImplementationOnce(
            () =>
                new Promise<Response>((resolve) => {
                    resolveResponse = resolve;
                }),
        );

        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        const actionButton = screen.getByRole("button", { name: /Run a quick strategy check/i });
        fireEvent.click(actionButton);
        fireEvent.click(actionButton);

        expect(await screen.findByText("Team Lead is preparing guidance for this AI Organization...")).toBeDefined();
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

        expect(await screen.findByText("Team Lead plan for Northstar Labs")).toBeDefined();
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

        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        fireEvent.click(screen.getByRole("button", { name: /Run a quick strategy check/i }));

        expect(await screen.findByText("Team Lead for Northstar Labs guidance for Northstar Labs")).toBeDefined();
        expect(screen.getByText("Team Lead for Northstar Labs has guidance ready for Northstar Labs.")).toBeDefined();
        expect(screen.getByText("Turn Northstar Labs into a clear next move.")).toBeDefined();
        expect(screen.queryByText(/debug/i)).toBeNull();
        expect(screen.queryByText(/^contract$/i)).toBeNull();
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

        fireEvent.click(screen.getByRole("button", { name: /Choose the first priority/i }));

        expect(await screen.findByText("Team Lead guidance is unavailable")).toBeDefined();
        expect(screen.getByRole("button", { name: "Retry Team Lead action" })).toBeDefined();
        expect(screen.getByText(/without leaving Northstar Labs/i)).toBeDefined();
        expect(screen.getByText(/choose another guided Team Lead action below/i)).toBeDefined();
    });

    it("retries the same Team Lead action successfully after a failure", async () => {
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

        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        fireEvent.click(screen.getByRole("button", { name: /Choose the first priority/i }));
        expect(await screen.findByText("Team Lead guidance is unavailable")).toBeDefined();

        fireEvent.click(screen.getByRole("button", { name: "Retry Team Lead action" }));

        await waitFor(() => {
            expect(mockFetch).toHaveBeenCalledTimes(2);
        });
        expect(await screen.findByText("First focus for Northstar Labs")).toBeDefined();
        expect(screen.getByText("Priority steps")).toBeDefined();
    });
});


