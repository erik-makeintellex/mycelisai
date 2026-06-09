import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { act, fireEvent, render, screen } from "@testing-library/react";
import { mockFetch } from "../setup";
import OrganizationPage from "@/app/(app)/organizations/[id]/page";
import { resetOrganizationPageStoreState, setupOrganizationFetch } from "./support/OrganizationPage.testSupport";

vi.mock("next/navigation", () => ({
    useRouter: () => ({ push: vi.fn(), replace: vi.fn(), back: vi.fn(), prefetch: vi.fn() }),
    usePathname: () => "/organizations/org-123",
}));

describe("OrganizationPage home workspace slices", () => {
    beforeEach(() => {
        mockFetch.mockReset();
        resetOrganizationPageStoreState();
    });

    afterEach(() => {
        vi.restoreAllMocks();
        vi.useRealTimers();
    });

    it("renders a Soma-primary organization workspace with guided actions and no generic or dev wording", async () => {
        vi.spyOn(Date, "now").mockReturnValue(new Date("2026-03-19T18:00:00Z").valueOf());
        setupOrganizationFetch();

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByText("AI Organization Home")).toBeDefined();
        expect(screen.getByLabelText("Organization breadcrumb")).toBeDefined();
        expect(screen.getByRole("link", { name: "AI Organizations" }).getAttribute("href")).toBe("/dashboard");
        expect(screen.getByText("Soma ready")).toBeDefined();
        expect(screen.getByRole("heading", { name: "Soma for Northstar Labs" })).toBeDefined();
        expect(screen.getAllByText(/Team Lead for Northstar Labs/i).length).toBeGreaterThan(0);
        expect(screen.getByText("Inspect the current organization")).toBeDefined();
        expect(screen.getByTestId("soma-operating-surface")).toBeDefined();
        expect(screen.getByRole("heading", { name: "What do you want Soma to do?" })).toBeDefined();
        expect(screen.getByTestId("mission-chat")).toBeDefined();
        expect(screen.queryByRole("button", { name: "Work panel" })).toBeNull();
        expect(screen.queryByText("Direct")).toBeNull();
        expect(screen.queryByTitle(/Broadcast mode/i)).toBeNull();
        expect(screen.queryByText("Create teams with Soma")).toBeNull();
        expect(screen.queryByText("Open team design lane")).toBeNull();
        expect(screen.queryByRole("heading", { name: "Quick Checks" })).toBeNull();
        expect(screen.getByRole("heading", { name: "Advisors" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "Departments" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "Automations" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "Recent Activity" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "What the Organization Is Retaining" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "AI Engine Settings" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "Response Style" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "Memory & Continuity" })).toBeDefined();
        expect(screen.getByText("Advisor support")).toBeDefined();
        expect(screen.getByText("Visible specialist roles")).toBeDefined();
        expect(screen.getByText("Your AI Organization is actively working through recent reviews, checks, and updates in the background.")).toBeDefined();
        expect(screen.getByText("Department check")).toBeDefined();
        expect(screen.getByText("Specialist review")).toBeDefined();
        expect(screen.getAllByText("2 minutes ago").length).toBeGreaterThan(0);
        expect(screen.getAllByText("5 minutes ago").length).toBeGreaterThan(0);
        expect(screen.getAllByText("No issues detected").length).toBeGreaterThan(0);
        expect(screen.getAllByText("2 items flagged").length).toBeGreaterThan(0);
        expect(screen.getAllByText("Platform Department is building a steadier execution lane for the organization.").length).toBeGreaterThan(0);
        expect(screen.getAllByText("Team: Platform Department").length).toBeGreaterThan(0);
        expect(screen.getAllByText("Strong").length).toBeGreaterThan(0);
        expect(screen.getAllByText("Emerging").length).toBeGreaterThan(0);
        expect(screen.getAllByText("Planning review").length).toBeGreaterThan(0);
        expect(screen.getByText("Department readiness review • Scheduled")).toBeDefined();
        expect(screen.getByText("Agent type readiness review • Event-driven")).toBeDefined();
        expect(screen.getAllByText("What this affects").length).toBeGreaterThan(0);
        expect(screen.getByText("Response style")).toBeDefined();
        expect(screen.getByText("Planning depth")).toBeDefined();
        expect(screen.getByText("Tone")).toBeDefined();
        expect(screen.getByText("Structure")).toBeDefined();
        expect(screen.getByText("Verbosity")).toBeDefined();
        expect(screen.getByText("Durable memory recall")).toBeDefined();
        expect(screen.getByText("Temporary planning continuity")).toBeDefined();
        expect(screen.getByText("Plan something")).toBeDefined();
        expect(screen.getByText("Research something")).toBeDefined();
        expect(screen.getByText("Create something")).toBeDefined();
        expect(screen.getByText("Review something")).toBeDefined();
        expect(screen.getByText("Configure tools")).toBeDefined();
        expect(screen.getAllByRole("button", { name: "Review Advisors" }).length).toBeGreaterThan(0);
        expect(screen.getAllByRole("button", { name: "Open Departments" }).length).toBeGreaterThan(0);
        expect(screen.getAllByRole("button", { name: "Review Automations" }).length).toBeGreaterThan(0);
        expect(screen.getAllByRole("button", { name: "Review AI Engine Settings" }).length).toBeGreaterThan(0);
        expect(screen.getAllByRole("button", { name: "Review Response Style" }).length).toBeGreaterThan(0);
        expect(screen.queryByText(/context shell/i)).toBeNull();
        expect(screen.queryByText(/bounded slice/i)).toBeNull();
        expect(screen.queryByText(/implementation slice/i)).toBeNull();
        expect(screen.queryByText(/contract/i)).toBeNull();
        expect(screen.queryByText(/loop profile/i)).toBeNull();
        expect(screen.queryByText(/New Chat/i)).toBeNull();
        expect(screen.queryByText(/generic chat/i)).toBeNull();
        expect(screen.queryByText(/scheduler/i)).toBeNull();
    }, 15000);

    it("opens organization evidence details without a competing team-design front door", async () => {
        setupOrganizationFetch();

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByTestId("soma-operating-surface")).toBeDefined();
        expect(screen.queryByRole("button", { name: "Open team design lane" })).toBeNull();

        fireEvent.click(screen.getAllByRole("button", { name: "Open Departments" })[0]);
        expect(await screen.findByRole("heading", { name: "Department details" })).toBeDefined();
    }, 15000);
});
