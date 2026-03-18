import { describe, it, expect, vi, beforeEach } from "vitest";
import { act, render, screen } from "@testing-library/react";
import { mockFetch } from "../setup";

vi.mock("next/navigation", () => ({
    useRouter: () => ({ push: vi.fn(), replace: vi.fn(), back: vi.fn(), prefetch: vi.fn() }),
    usePathname: () => "/organizations/org-123",
}));

import OrganizationPage from "@/app/(app)/organizations/[id]/page";

describe("OrganizationPage (/organizations/[id])", () => {
    beforeEach(() => {
        mockFetch.mockReset();
    });

    it("renders the AI Organization context shell", async () => {
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
        expect(screen.getByText("Team Lead workspace is next")).toBeDefined();
        expect(screen.getByText("Engineering Starter")).toBeDefined();
    });
});
