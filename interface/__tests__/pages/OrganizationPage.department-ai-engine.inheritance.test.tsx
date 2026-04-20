import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { act, fireEvent, render, screen } from "@testing-library/react";
import { mockFetch } from "../setup";
import OrganizationPage from "@/app/(app)/organizations/[id]/page";
import type { OrganizationAIEngineProfileId } from "@/lib/organizations";
import {
    applyAgentTypeAIEngine,
    jsonResponse,
    organizationHome,
    resetOrganizationPageStoreState,
    setupOrganizationFetch,
} from "./support/OrganizationPage.testSupport";

vi.mock("next/navigation", () => ({
    useRouter: () => ({ push: vi.fn(), replace: vi.fn(), back: vi.fn(), prefetch: vi.fn() }),
    usePathname: () => "/organizations/org-123",
}));

describe("OrganizationPage department AI engine inheritance slice", () => {
    beforeEach(() => {
        mockFetch.mockReset();
        resetOrganizationPageStoreState();
    });

    afterEach(() => {
        vi.restoreAllMocks();
        vi.useRealTimers();
    });

    it("keeps a Department override when the organization AI Engine changes and reapplies inheritance after revert", async () => {
        setupOrganizationFetch();

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByRole("heading", { name: "Departments" })).toBeDefined();
        fireEvent.click(screen.getAllByRole("button", { name: "Open Departments" })[0]);
        fireEvent.click(screen.getByRole("button", { name: "Change for this Team" }));
        fireEvent.click(screen.getByRole("button", { name: /High Reasoning/i }));
        fireEvent.click(screen.getByRole("button", { name: "Use selected AI Engine" }));
        expect(await screen.findByText("Overridden: High Reasoning")).toBeDefined();

        fireEvent.click(screen.getByRole("button", { name: "Back to Soma" }));
        fireEvent.click(screen.getAllByRole("button", { name: "Review AI Engine Settings" })[0]);
        fireEvent.click(screen.getByRole("button", { name: "Change AI Engine" }));
        fireEvent.click(screen.getByRole("button", { name: /Fast & Lightweight/i }));
        fireEvent.click(screen.getByRole("button", { name: "Use selected AI Engine" }));
        expect(await screen.findByText("Current profile: Fast & Lightweight.")).toBeDefined();

        fireEvent.click(screen.getByRole("button", { name: "Back to Soma" }));
        fireEvent.click(screen.getAllByRole("button", { name: "Open Departments" })[0]);
        expect(await screen.findByText("Overridden: High Reasoning")).toBeDefined();

        fireEvent.click(screen.getByRole("button", { name: "Revert to Organization Default" }));
        expect(await screen.findByText("Using Organization Default: Fast & Lightweight")).toBeDefined();
    }, 20000);

    it("shows retry guidance when changing an Agent Type AI Engine fails and recovers on retry", async () => {
        let attempts = 0;
        setupOrganizationFetch({
            agentTypeAIEngineUpdateHandler: (body) => {
                attempts += 1;
                if (attempts === 1) {
                    return jsonResponse({ ok: false, error: "Agent Type AI Engine update is unavailable right now." }, 500);
                }

                return jsonResponse({
                    ok: true,
                    data: applyAgentTypeAIEngine(organizationHome, "platform", "delivery-specialist", String(body.profile_id ?? "") as OrganizationAIEngineProfileId, "Balanced"),
                });
            },
        });

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByRole("heading", { name: "Departments" })).toBeDefined();
        fireEvent.click(screen.getAllByRole("button", { name: "Open Departments" })[0]);
        fireEvent.click(screen.getAllByRole("button", { name: "Change for this Agent Type" }).at(-1)!);
        fireEvent.click(screen.getByRole("button", { name: /Balanced/i }));
        fireEvent.click(screen.getAllByRole("button", { name: "Use selected AI Engine" }).at(-1)!);

        expect(await screen.findByText("Unable to update this Agent Type AI Engine")).toBeDefined();
        expect(screen.getByText("Agent Type AI Engine update is unavailable right now.")).toBeDefined();
        expect(screen.getByText("AI Organization Home")).toBeDefined();
        expect(screen.getAllByText(/Team Lead for Northstar Labs/i).length).toBeGreaterThan(0);

        fireEvent.click(screen.getAllByRole("button", { name: "Use selected AI Engine" }).at(-1)!);
        expect(await screen.findByText("Type-specific Engine: Balanced")).toBeDefined();
    }, 15000);
});
