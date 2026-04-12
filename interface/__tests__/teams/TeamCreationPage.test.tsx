import { beforeEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { mockFetch } from "../setup";

const useSearchParamsMock = vi.fn();

vi.mock("next/navigation", () => ({
    useRouter: () => ({
        push: vi.fn(),
        replace: vi.fn(),
        back: vi.fn(),
        prefetch: vi.fn(),
    }),
    usePathname: () => "/teams/create",
    useSearchParams: () => useSearchParamsMock(),
}));

import TeamCreationPage from "@/components/teams/TeamCreationPage";

function jsonResponse(body: unknown, status = 200) {
    return Promise.resolve(new Response(JSON.stringify(body), { status }));
}

describe("TeamCreationPage", () => {
    beforeEach(() => {
        localStorage.clear();
        mockFetch.mockReset();
        useSearchParamsMock.mockReturnValue(new URLSearchParams());
    });

    it("loads the current organization and runs the guided Soma creation workflow", async () => {
        localStorage.setItem("mycelis-last-organization-id", "org-123");
        localStorage.setItem("mycelis-last-organization-name", "Northstar Labs");
        mockFetch
            .mockResolvedValueOnce(
                jsonResponse({
                    ok: true,
                    data: {
                        id: "org-123",
                        name: "Northstar Labs",
                        purpose: "Ship a focused AI organization.",
                        start_mode: "template",
                        team_lead_label: "Launch Lead",
                        advisor_count: 1,
                        department_count: 2,
                        specialist_count: 5,
                        ai_engine_settings_summary: "Balanced",
                        response_contract_summary: "Clear and structured",
                        memory_personality_summary: "Stable continuity",
                        status: "active",
                    },
                }),
            )
            .mockResolvedValueOnce(
                jsonResponse({
                    ok: true,
                    data: {
                        action: "plan_next_steps",
                        request_label: "Run a quick strategy check",
                        headline: "Launch team plan for Northstar Labs",
                        summary: "Soma recommends a launch-focused delivery team.",
                        priority_steps: ["Confirm campaign outcomes", "Define owned deliverables"],
                        suggested_follow_ups: ["Choose the first priority"],
                    },
                }),
            );

        render(<TeamCreationPage />);

        expect(await screen.findByText("Northstar Labs")).toBeDefined();
        expect(screen.getByRole("button", { name: "Marketing launch team" })).toBeDefined();
        expect(screen.getByText(/Default team shape/i)).toBeDefined();
        expect(screen.getByText(/several small teams or lane bundles/i)).toBeDefined();

        fireEvent.click(screen.getByRole("button", { name: "Marketing launch team" }));
        expect(
            (screen.getByLabelText("Tell Soma what team or delivery lane you want to create") as HTMLTextAreaElement).value,
        ).toBe(
            "Create a temporary marketing launch team that can produce campaign copy, a landing page brief, and social rollout assets for a new product launch.",
        );

        fireEvent.click(screen.getByRole("button", { name: "Start team design" }));

        await waitFor(() => {
            expect(mockFetch).toHaveBeenCalledWith(
                "/api/v1/organizations/org-123/workspace/actions",
                expect.objectContaining({ method: "POST" }),
            );
        });

        expect(await screen.findByText("Launch team plan for Northstar Labs")).toBeDefined();
    });

    it("shows recovery guidance when no organization is available yet", async () => {
        render(<TeamCreationPage />);

        expect(screen.getByText("Choose or create an AI Organization first")).toBeDefined();
        expect(screen.getAllByRole("link", { name: "Open Soma workspace" })[0].getAttribute("href")).toBe("/dashboard");
        expect(mockFetch).not.toHaveBeenCalled();
    });
});
