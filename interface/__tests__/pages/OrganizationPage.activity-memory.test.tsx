import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { act, fireEvent, render, screen } from "@testing-library/react";
import { mockFetch } from "../setup";
import OrganizationPage from "@/app/(app)/organizations/[id]/page";
import {
    jsonResponse,
    resetOrganizationPageStoreState,
    setupOrganizationFetch,
} from "./support/OrganizationPage.testSupport";

vi.mock("next/navigation", () => ({
    useRouter: () => ({ push: vi.fn(), replace: vi.fn(), back: vi.fn(), prefetch: vi.fn() }),
    usePathname: () => "/organizations/org-123",
}));

describe("OrganizationPage activity and memory slices", () => {
    beforeEach(() => {
        mockFetch.mockReset();
        resetOrganizationPageStoreState();
    });

    afterEach(() => {
        vi.restoreAllMocks();
        vi.useRealTimers();
    });

    it("shows a clean empty Recent Activity state when no checks have run yet", async () => {
        setupOrganizationFetch({
            loopActivityHandler: () => jsonResponse({ ok: true, data: [] }),
        });

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByRole("heading", { name: "Recent Activity" })).toBeDefined();
        expect(screen.getByText("No recent activity yet")).toBeDefined();
        expect(screen.getByText("This is where reviews, checks, and updates will appear as your AI Organization starts operating.")).toBeDefined();
        expect(screen.getByText("Take a guided Soma action to start creating visible movement here.")).toBeDefined();
    });

    it("shows Activity unavailable without breaking the workspace when recent updates cannot be loaded", async () => {
        setupOrganizationFetch({
            loopActivityHandler: () => jsonResponse({ ok: false, error: "activity unavailable" }, 503),
        });

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByRole("heading", { name: "Recent Activity" })).toBeDefined();
        expect(screen.getByText("Activity unavailable")).toBeDefined();
        expect(screen.getByText("Recent reviews and updates are not available right now. The Soma workspace is still ready.")).toBeDefined();
        expect(screen.getByRole("button", { name: "Retry Recent Activity" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "Soma for Northstar Labs" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "Advisors" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "Departments" })).toBeDefined();
    });

    it("shows a clean empty learning state when no recent highlights are available yet", async () => {
        setupOrganizationFetch({
            learningInsightsHandler: () => jsonResponse({ ok: true, data: [] }),
        });

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByRole("heading", { name: "What the Organization Is Retaining" })).toBeDefined();
        expect(screen.getByText("No retained patterns yet")).toBeDefined();
        expect(screen.getByText("This is where reusable patterns, continuity cues, and stronger working habits will appear in plain language.")).toBeDefined();
        expect(screen.getByText("Ordinary planning chat stays in working continuity until a stronger reusable pattern emerges or you intentionally save something for later recall.")).toBeDefined();
    });

    it("shows Memory & Continuity updates unavailable without breaking the workspace when insights cannot be loaded", async () => {
        setupOrganizationFetch({
            learningInsightsHandler: () => jsonResponse({ ok: false, error: "learning updates unavailable" }, 503),
        });

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByRole("heading", { name: "What the Organization Is Retaining" })).toBeDefined();
        expect(screen.getByText("Memory & Continuity updates unavailable")).toBeDefined();
        expect(screen.getByText("Recent retained patterns are not available right now. The Soma workspace is still ready.")).toBeDefined();
        expect(screen.getByRole("button", { name: "Retry Memory & Continuity" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "Soma for Northstar Labs" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "Advisors" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "Departments" })).toBeDefined();
        expect(screen.queryByText(/vector/i)).toBeNull();
        expect(screen.queryByText(/pgvector/i)).toBeNull();
        expect(screen.queryByText(/memory promotion/i)).toBeNull();
    }, 15000);
});
