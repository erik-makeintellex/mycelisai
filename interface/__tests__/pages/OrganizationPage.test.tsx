import { describe, it, expect, vi, beforeEach } from "vitest";
import { act, fireEvent, render, screen } from "@testing-library/react";
import { mockFetch } from "../setup";

vi.mock("next/navigation", () => ({
    useRouter: () => ({ push: vi.fn(), replace: vi.fn(), back: vi.fn(), prefetch: vi.fn() }),
    usePathname: () => "/organizations/org-123",
}));

import OrganizationPage from "@/app/(app)/organizations/[id]/page";

const organizationHome = {
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
};

function jsonResponse(body: unknown, status = 200) {
    return Promise.resolve(new Response(JSON.stringify(body), { status }));
}

describe("OrganizationPage (/organizations/[id])", () => {
    beforeEach(() => {
        mockFetch.mockReset();
    });

    it("renders a Team Lead-first organization workspace with action cards and no generic or dev wording", async () => {
        mockFetch.mockResolvedValueOnce(new Response(JSON.stringify({ ok: true, data: organizationHome })));

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByText("AI Organization Home")).toBeDefined();
        expect(screen.getByText("Team Lead ready")).toBeDefined();
        expect(screen.getByText("Team Lead for Northstar Labs")).toBeDefined();
        expect(screen.getByText("What I can help with")).toBeDefined();
        expect(screen.getByRole("button", { name: /Review Advisors/i })).toBeDefined();
        expect(screen.getByRole("button", { name: /Open Departments/i })).toBeDefined();
        expect(screen.getByRole("button", { name: /Ask Team Lead to plan next steps/i })).toBeDefined();
        expect(screen.getByRole("button", { name: /Review AI Engine Settings/i })).toBeDefined();
        expect(screen.getByRole("button", { name: /Review Memory & Personality/i })).toBeDefined();
        expect(screen.getByText("Team Lead workspace")).toBeDefined();
        expect(screen.getByText("Recommended next steps")).toBeDefined();
        expect(screen.queryByText(/context shell/i)).toBeNull();
        expect(screen.queryByText(/bounded slice/i)).toBeNull();
        expect(screen.queryByText(/implementation slice/i)).toBeNull();
        expect(screen.queryByText(/contract/i)).toBeNull();
        expect(screen.queryByText(/New Chat/i)).toBeNull();
        expect(screen.queryByText(/generic chat/i)).toBeNull();
    });

    it("keeps the organization frame visible while moving from home into the Team Lead workspace focus", async () => {
        mockFetch.mockResolvedValueOnce(new Response(JSON.stringify({ ok: true, data: organizationHome })));

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByText("Team Lead workspace")).toBeDefined();
        fireEvent.click(screen.getByRole("button", { name: /Review Advisors/i }));
        expect(screen.getByText("Review the advisor picture")).toBeDefined();
        expect(screen.getByText("1 Advisor ready to guide")).toBeDefined();
        expect(screen.getByText("Northstar Labs")).toBeDefined();
        expect(screen.getByText("AI Organization Home")).toBeDefined();

        fireEvent.click(screen.getByRole("button", { name: /Ask Team Lead to plan next steps/i }));
        expect(screen.getByText("Plan the next steps for Northstar Labs")).toBeDefined();
        expect(screen.getByText("Recommended next steps")).toBeDefined();
    });

    it("offers retry guidance when the organization home cannot be loaded", async () => {
        let attempt = 0;
        mockFetch.mockImplementation(() => {
            attempt += 1;
            return attempt === 1
                ? jsonResponse({ ok: false, error: "This AI Organization could not be loaded right now." }, 500)
                : jsonResponse({ ok: true, data: organizationHome });
        });

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByText("This AI Organization could not be loaded right now.")).toBeDefined();
        fireEvent.click(screen.getByRole("button", { name: "Retry" }));
        expect(await screen.findByText("Team Lead ready")).toBeDefined();
    });
});
