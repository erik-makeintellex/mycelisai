import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { act, fireEvent, render, screen } from "@testing-library/react";
import { mockFetch } from "../setup";
import OrganizationPage from "@/app/(app)/organizations/[id]/page";
import type { ResponseContractProfileId } from "@/lib/organizations";
import {
    applyAgentTypeResponseContract,
    jsonResponse,
    organizationHome,
    resetOrganizationPageStoreState,
    setupOrganizationFetch,
} from "./support/OrganizationPage.testSupport";

vi.mock("next/navigation", () => ({
    useRouter: () => ({ push: vi.fn(), replace: vi.fn(), back: vi.fn(), prefetch: vi.fn() }),
    usePathname: () => "/organizations/org-123",
}));

describe("OrganizationPage department response style slice", () => {
    beforeEach(() => {
        mockFetch.mockReset();
        resetOrganizationPageStoreState();
    });

    afterEach(() => {
        vi.restoreAllMocks();
        vi.useRealTimers();
    });

    it("lets the operator bind an Agent Type Response Style and then return it to the Organization / Team default", async () => {
        setupOrganizationFetch();

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByRole("heading", { name: "Departments" })).toBeDefined();
        fireEvent.click(screen.getAllByRole("button", { name: "Open Departments" })[0]);

        expect(await screen.findByText("Using Organization or Team Default: Clear & Balanced")).toBeDefined();
        fireEvent.click(screen.getAllByRole("button", { name: "Change Response Style for this Agent Type" }).at(-1)!);
        expect(await screen.findByRole("heading", { name: "Choose a Response Style for this Agent Type" })).toBeDefined();
        fireEvent.click(screen.getByRole("button", { name: /Warm & Supportive/i }));
        fireEvent.click(screen.getByRole("button", { name: "Use selected Response Style" }));

        expect(await screen.findByText("Type-specific Response Style: Warm & Supportive")).toBeDefined();
        expect(screen.getAllByRole("button", { name: "Use Organization / Team Default" }).at(-1)).toBeDefined();

        fireEvent.click(screen.getAllByRole("button", { name: "Use Organization / Team Default" }).at(-1)!);
        expect(await screen.findByText("Using Organization or Team Default: Clear & Balanced")).toBeDefined();
    }, 10000);

    it("keeps a type-bound Response Style stable when the organization default changes and reapplies inheritance after revert", async () => {
        setupOrganizationFetch();

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByRole("heading", { name: "Departments" })).toBeDefined();
        fireEvent.click(screen.getAllByRole("button", { name: "Open Departments" })[0]);
        fireEvent.click(screen.getAllByRole("button", { name: "Change Response Style for this Agent Type" }).at(-1)!);
        fireEvent.click(screen.getByRole("button", { name: /Warm & Supportive/i }));
        fireEvent.click(screen.getByRole("button", { name: "Use selected Response Style" }));
        expect(await screen.findByText("Type-specific Response Style: Warm & Supportive")).toBeDefined();

        fireEvent.click(screen.getByRole("button", { name: "Back to Soma" }));
        fireEvent.click(screen.getAllByRole("button", { name: "Review Response Style" }).at(-1)!);
        fireEvent.click(screen.getByRole("button", { name: "Change Response Style" }));
        fireEvent.click(screen.getByRole("button", { name: /Concise & Direct/i }));
        fireEvent.click(screen.getByRole("button", { name: "Use selected Response Style" }));
        expect(await screen.findByText("Current profile: Concise & Direct.")).toBeDefined();

        fireEvent.click(screen.getByRole("button", { name: "Back to Soma" }));
        fireEvent.click(screen.getAllByRole("button", { name: "Open Departments" })[0]);
        expect(await screen.findByText("Type-specific Response Style: Warm & Supportive")).toBeDefined();

        fireEvent.click(screen.getAllByRole("button", { name: "Use Organization / Team Default" }).at(-1)!);
        expect(await screen.findByText("Using Organization or Team Default: Concise & Direct")).toBeDefined();
    }, 10000);

    it("shows retry guidance when changing an Agent Type Response Style fails and recovers on retry", async () => {
        let attempts = 0;
        setupOrganizationFetch({
            agentTypeResponseContractUpdateHandler: (body) => {
                attempts += 1;
                if (attempts === 1) {
                    return jsonResponse({ ok: false, error: "Agent Type Response Style update is unavailable right now." }, 500);
                }

                return jsonResponse({
                    ok: true,
                    data: applyAgentTypeResponseContract(
                        organizationHome,
                        "platform",
                        "delivery-specialist",
                        String(body.profile_id ?? "") as ResponseContractProfileId,
                        "Warm & Supportive",
                    ),
                });
            },
        });

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByRole("heading", { name: "Departments" })).toBeDefined();
        fireEvent.click(screen.getAllByRole("button", { name: "Open Departments" })[0]);
        fireEvent.click(screen.getAllByRole("button", { name: "Change Response Style for this Agent Type" }).at(-1)!);
        fireEvent.click(screen.getByRole("button", { name: /Warm & Supportive/i }));
        fireEvent.click(screen.getByRole("button", { name: "Use selected Response Style" }));

        expect(await screen.findByText("Unable to update this Agent Type Response Style")).toBeDefined();
        expect(screen.getByText("Agent Type Response Style update is unavailable right now.")).toBeDefined();
        expect(screen.getByRole("button", { name: "Retry Response Style change" })).toBeDefined();
        expect(screen.getByText("AI Organization Home")).toBeDefined();
        expect(screen.getAllByText(/Team Lead for Northstar Labs/i).length).toBeGreaterThan(0);

        fireEvent.click(screen.getByRole("button", { name: "Retry Response Style change" }));
        expect(await screen.findByText("Type-specific Response Style: Warm & Supportive")).toBeDefined();
    }, 20000);
});
