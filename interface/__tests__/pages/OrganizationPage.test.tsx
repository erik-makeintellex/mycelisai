import { describe, it, expect, vi, beforeEach } from "vitest";
import { act, fireEvent, render, screen } from "@testing-library/react";
import { mockFetch } from "../setup";

vi.mock("next/navigation", () => ({
    useRouter: () => ({ push: vi.fn(), replace: vi.fn(), back: vi.fn(), prefetch: vi.fn() }),
    usePathname: () => "/organizations/org-123",
}));

import OrganizationPage from "@/app/(app)/organizations/[id]/page";

function jsonResponse(body: unknown, status = 200) {
    return Promise.resolve(new Response(JSON.stringify(body), { status }));
}

describe("OrganizationPage (/organizations/[id])", () => {
    beforeEach(() => {
        mockFetch.mockReset();
    });

    it("renders the AI Organization landing screen with a concrete Team Lead status", async () => {
        mockFetch.mockResolvedValueOnce(new Response(JSON.stringify({ ok: true, data: {
            id: "org-123",
            name: "Northstar Labs",
            purpose: "Ship a focused AI engineering organization for product delivery.",
            start_mode: "template",
            template_id: "engineering-starter",
            template_name: "Engineering Starter",
            team_lead_label: "Team Lead",
            advisor_count: 1,
            department_count: 1,
            specialist_count: 2,
            ai_engine_settings_summary: "Starter defaults included",
            memory_personality_summary: "Prepared for Adaptive Delivery work",
            status: "ready",
            description: "Guided AI Organization for engineering work",
        } })));

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByText("AI Organization Home")).toBeDefined();
        expect(screen.getByText("Northstar Labs")).toBeDefined();
        expect(screen.getByText("Team Lead status")).toBeDefined();
        expect(screen.getByText("Team Lead workspace coming soon")).toBeDefined();
        expect(screen.queryByText(/context shell/i)).toBeNull();
        expect(screen.queryByText(/bounded slice/i)).toBeNull();
        expect(screen.queryByText(/first UI slice/i)).toBeNull();
    });

    it("offers retry guidance when the organization home cannot be loaded", async () => {
        let attempt = 0;
        mockFetch.mockImplementation(() => {
            attempt += 1;
            return attempt === 1
                ? jsonResponse({ ok: false, error: "This AI Organization could not be loaded right now." }, 500)
                : jsonResponse({ ok: true, data: {
                    id: "org-123",
                    name: "Northstar Labs",
                    purpose: "Ship a focused AI engineering organization for product delivery.",
                    start_mode: "empty",
                    team_lead_label: "Team Lead",
                    advisor_count: 0,
                    department_count: 0,
                    specialist_count: 0,
                    ai_engine_settings_summary: "Set up later in Advanced mode",
                    memory_personality_summary: "Set up later in Advanced mode",
                    status: "ready",
                } });
        });

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByText("This AI Organization could not be loaded right now.")).toBeDefined();
        fireEvent.click(screen.getByRole("button", { name: "Retry" }));
        expect(await screen.findByText("Northstar Labs")).toBeDefined();
    });
});
