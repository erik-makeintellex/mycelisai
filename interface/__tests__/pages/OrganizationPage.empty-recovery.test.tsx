import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { act, fireEvent, render, screen } from "@testing-library/react";
import { mockFetch } from "../setup";
import OrganizationPage from "@/app/(app)/organizations/[id]/page";
import {
    jsonResponse,
    organizationHome,
    resetOrganizationPageStoreState,
    setupOrganizationFetch,
} from "./support/OrganizationPage.testSupport";

vi.mock("next/navigation", () => ({
    useRouter: () => ({ push: vi.fn(), replace: vi.fn(), back: vi.fn(), prefetch: vi.fn() }),
    usePathname: () => "/organizations/org-123",
}));

describe("OrganizationPage empty and recovery slices", () => {
    beforeEach(() => {
        mockFetch.mockReset();
        resetOrganizationPageStoreState();
    });

    afterEach(() => {
        vi.restoreAllMocks();
        vi.useRealTimers();
    });

    it("shows inspect-only Advisor and Department summaries when the organization starts empty", async () => {
        setupOrganizationFetch({
            homeHandler: () =>
                jsonResponse({
                    ok: true,
                    data: {
                        ...organizationHome,
                        name: "Skylight Works",
                        start_mode: "empty",
                        template_id: undefined,
                        template_name: undefined,
                        advisor_count: 0,
                        department_count: 0,
                        specialist_count: 0,
                        ai_engine_settings_summary: "Set up later in Advanced mode",
                        memory_personality_summary: "Set up later in Advanced mode",
                        departments: [],
                    },
                }),
            automationsHandler: () => jsonResponse({ ok: true, data: [] }),
        });

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByRole("heading", { name: "Advisors" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "Departments" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "Automations" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "AI Engine Settings" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "Response Style" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "Memory & Continuity" })).toBeDefined();
        expect(screen.getAllByText("Inspect only").length).toBeGreaterThan(0);
        expect(screen.getByText("Review support appears here")).toBeDefined();
        expect(screen.getAllByText("Try reviewing your organization setup").length).toBeGreaterThan(0);
        expect(screen.getByText("Reviews appear here")).toBeDefined();
        expect(screen.getByText("Started from Empty")).toBeDefined();
        expect(screen.getByText("The current AI Engine Settings keep the organization on a simple starter profile until deeper tuning is needed.")).toBeDefined();
        expect(screen.getByText("Memory & Continuity stay on a simple starter posture so Soma can keep working continuity without turning every conversation into durable memory.")).toBeDefined();
    });

    it("offers retry guidance when the organization home cannot be loaded", async () => {
        let attempt = 0;
        setupOrganizationFetch({
            homeHandler: () => {
                attempt += 1;
                return attempt === 1
                    ? jsonResponse({ ok: false, error: "This AI Organization could not be loaded right now." }, 500)
                    : jsonResponse({ ok: true, data: organizationHome });
            },
        });

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByText("This AI Organization could not be loaded right now.")).toBeDefined();
        fireEvent.click(screen.getByRole("button", { name: "Retry" }));
        expect(await screen.findByText("Soma ready")).toBeDefined();
    });
});
