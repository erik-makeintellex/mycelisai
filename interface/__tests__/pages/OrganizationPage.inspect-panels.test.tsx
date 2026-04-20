import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { act, fireEvent, render, screen } from "@testing-library/react";
import { mockFetch } from "../setup";
import OrganizationPage from "@/app/(app)/organizations/[id]/page";
import { jsonResponse, resetOrganizationPageStoreState, setupOrganizationFetch } from "./support/OrganizationPage.testSupport";

vi.mock("next/navigation", () => ({
    useRouter: () => ({ push: vi.fn(), replace: vi.fn(), back: vi.fn(), prefetch: vi.fn() }),
    usePathname: () => "/organizations/org-123",
}));

describe("OrganizationPage inspect panel slices", () => {
    beforeEach(() => {
        mockFetch.mockReset();
        resetOrganizationPageStoreState();
    });

    afterEach(() => {
        vi.restoreAllMocks();
        vi.useRealTimers();
    });

    it("opens Automation details from the support column and keeps the Soma workspace visible", async () => {
        setupOrganizationFetch();

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByRole("heading", { name: "Automations" })).toBeDefined();
        fireEvent.click(screen.getAllByRole("button", { name: "Review Automations" })[0]);

        expect(await screen.findByRole("heading", { name: "Automation details" })).toBeDefined();
        expect(screen.getByText("Department readiness review")).toBeDefined();
        expect(screen.getByText("Scheduled")).toBeDefined();
        expect(screen.getAllByText("Team: Platform Department").length).toBeGreaterThan(0);
        expect(screen.getAllByText("What it watches").length).toBeGreaterThan(0);
        expect(screen.getAllByText("How it runs").length).toBeGreaterThan(0);
        expect(screen.getAllByText("Recent outcomes").length).toBeGreaterThan(0);
        expect(screen.getByText("Runs every minute and also after organization setup, Team Lead guidance, AI Engine changes, or Response Style changes.")).toBeDefined();
        expect(screen.getByText("This system runs ongoing reviews and checks to help your organization improve over time.")).toBeDefined();
        expect(screen.getByText("AI Organization Home")).toBeDefined();
        expect(screen.getAllByText(/Team Lead for Northstar Labs/i).length).toBeGreaterThan(0);
        expect(screen.getByText("Create teams with Soma")).toBeDefined();
        expect(screen.queryByText(/loop profile/i)).toBeNull();
        expect(screen.queryByText(/scheduler/i)).toBeNull();
        expect(screen.queryByText(/vector/i)).toBeNull();
        expect(screen.queryByText(/pgvector/i)).toBeNull();
        expect(screen.queryByText(/memory promotion/i)).toBeNull();
    });

    it("shows Automations unavailable without breaking the workspace when definitions cannot be loaded", async () => {
        setupOrganizationFetch({
            automationsHandler: () => jsonResponse({ ok: false, error: "automations unavailable" }, 503),
        });

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByRole("heading", { name: "Automations" })).toBeDefined();
        expect(screen.getByText("Automations unavailable")).toBeDefined();
        fireEvent.click(screen.getAllByRole("button", { name: "Review Automations" })[0]);
        expect(await screen.findByText("Reviews and checks are temporarily unavailable here. The Soma workspace is still ready.")).toBeDefined();
        expect(screen.getByRole("button", { name: "Retry Automations" })).toBeDefined();
        expect(screen.getByText("AI Organization Home")).toBeDefined();
        expect(screen.getAllByText(/Team Lead for Northstar Labs/i).length).toBeGreaterThan(0);
    });

    it("opens Advisor details from the Soma action and keeps the Soma workspace visible", async () => {
        setupOrganizationFetch();

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByText("Create teams with Soma")).toBeDefined();
        fireEvent.click(screen.getAllByRole("button", { name: "Review Advisors" })[0]);

        expect(await screen.findByRole("heading", { name: "Advisor details" })).toBeDefined();
        expect(screen.getByText("Planning Advisor")).toBeDefined();
        expect(screen.getByText("Decision support")).toBeDefined();
        expect(screen.getByText("AI Organization Home")).toBeDefined();
        expect(screen.getAllByText(/Team Lead for Northstar Labs/i).length).toBeGreaterThan(0);
        expect(screen.getByText("Create teams with Soma")).toBeDefined();
    });

    it("opens Department details from the support column and preserves organization context", async () => {
        setupOrganizationFetch();

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByRole("heading", { name: "Departments" })).toBeDefined();
        fireEvent.click(screen.getAllByRole("button", { name: "Open Departments" })[0]);

        expect(await screen.findByRole("heading", { name: "Department details" })).toBeDefined();
        expect(screen.getByText("Platform Department")).toBeDefined();
        expect(screen.getByText("2 Specialists visible here.")).toBeDefined();
        expect(screen.getByText("Using Organization Default: Starter defaults included")).toBeDefined();
        expect(screen.getByText("Agent Type Profiles")).toBeDefined();
        expect(screen.getAllByText("Planner").length).toBeGreaterThan(0);
        expect(screen.getAllByText("Delivery Specialist").length).toBeGreaterThan(0);
        expect(screen.getByText("Type-specific Engine: High Reasoning")).toBeDefined();
        expect(screen.getByText("Using Team Default: Starter defaults included")).toBeDefined();
        expect(screen.getByText("Type-specific Response Style: Structured & Analytical")).toBeDefined();
        expect(screen.getByText("Using Organization or Team Default: Clear & Balanced")).toBeDefined();
        expect(screen.getByText("AI Organization Home")).toBeDefined();
        expect(screen.getAllByText(/Team Lead for Northstar Labs/i).length).toBeGreaterThan(0);

        fireEvent.click(screen.getByRole("button", { name: "Back to Soma" }));
        expect(screen.queryByRole("heading", { name: "Department details" })).toBeNull();
        expect(screen.getByText("Create teams with Soma")).toBeDefined();
    });
});
