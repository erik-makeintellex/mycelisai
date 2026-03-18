import { describe, it, expect, vi, beforeEach } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { mockFetch } from "../setup";

const push = vi.fn();

vi.mock("next/navigation", () => ({
    useRouter: () => ({ push, replace: vi.fn(), back: vi.fn(), prefetch: vi.fn() }),
}));

import DashboardPage from "@/app/(app)/dashboard/page";

describe("Dashboard Page (V8 AI Organization entry flow)", () => {
    beforeEach(() => {
        push.mockReset();
    });

    it("renders the Create AI Organization entry flow", async () => {
        mockFetch
            .mockResolvedValueOnce(new Response(JSON.stringify({ ok: true, data: [{
                id: "engineering-starter",
                name: "Engineering Starter",
                description: "Guided AI Organization for engineering work",
                organization_type: "AI Organization starter",
                team_lead_label: "Team Lead",
                advisor_count: 1,
                department_count: 1,
                specialist_count: 2,
                ai_engine_settings_summary: "Starter defaults included",
                memory_personality_summary: "Prepared for Adaptive Delivery work",
            }] })))
            .mockResolvedValueOnce(new Response(JSON.stringify({ ok: true, data: [] })));

        render(<DashboardPage />);

        expect(await screen.findByRole("heading", { name: "Create AI Organization" })).toBeDefined();
        expect(screen.getByRole("button", { name: /Start from template/i })).toBeDefined();
        expect(screen.getByRole("button", { name: /Start empty/i })).toBeDefined();

        fireEvent.click(screen.getByRole("button", { name: /Start from template/i }));
        expect(await screen.findByText("Engineering Starter")).toBeDefined();
        expect(screen.getByText("AI Organization starter")).toBeDefined();
    });

    it("submits an empty-start organization and routes into the context shell", async () => {
        mockFetch
            .mockResolvedValueOnce(new Response(JSON.stringify({ ok: true, data: [] })))
            .mockResolvedValueOnce(new Response(JSON.stringify({ ok: true, data: [] })))
            .mockResolvedValueOnce(new Response(JSON.stringify({ ok: true, data: {
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
            } }), { status: 201 }));

        render(<DashboardPage />);

        expect(await screen.findByRole("heading", { name: "Create AI Organization" })).toBeDefined();
        fireEvent.click(screen.getByRole("button", { name: /Start empty/i }));
        fireEvent.change(screen.getByLabelText("AI Organization name"), { target: { value: "Northstar Labs" } });
        fireEvent.change(screen.getByLabelText("Purpose"), { target: { value: "Ship a focused AI engineering organization for product delivery." } });
        fireEvent.click(screen.getByRole("button", { name: "Create AI Organization" }));

        await waitFor(() => {
            expect(push).toHaveBeenCalledWith("/organizations/org-123");
        });
    });

    it("shows recent AI Organizations when summaries exist", async () => {
        mockFetch
            .mockResolvedValueOnce(new Response(JSON.stringify({ ok: true, data: [] })))
            .mockResolvedValueOnce(new Response(JSON.stringify({ ok: true, data: [{
                id: "org-42",
                name: "Atlas",
                purpose: "Resume me later",
                start_mode: "empty",
                team_lead_label: "Team Lead",
                advisor_count: 0,
                department_count: 0,
                specialist_count: 0,
                ai_engine_settings_summary: "Set up later in Advanced mode",
                memory_personality_summary: "Set up later in Advanced mode",
                status: "ready",
            }] })));

        render(<DashboardPage />);

        expect(await screen.findByText("Recent AI Organizations")).toBeDefined();
        expect(screen.getByText("Atlas")).toBeDefined();
        fireEvent.click(screen.getByRole("button", { name: "Open AI Organization" }));
        expect(push).toHaveBeenCalledWith("/organizations/org-42");
    });
});
