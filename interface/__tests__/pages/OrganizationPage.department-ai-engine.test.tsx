import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { act, fireEvent, render, screen } from "@testing-library/react";
import { mockFetch } from "../setup";
import OrganizationPage from "@/app/(app)/organizations/[id]/page";
import { resetOrganizationPageStoreState, setupOrganizationFetch } from "./support/OrganizationPage.testSupport";

vi.mock("next/navigation", () => ({
    useRouter: () => ({ push: vi.fn(), replace: vi.fn(), back: vi.fn(), prefetch: vi.fn() }),
    usePathname: () => "/organizations/org-123",
}));

describe("OrganizationPage department AI engine slice", () => {
    beforeEach(() => {
        mockFetch.mockReset();
        resetOrganizationPageStoreState();
    });

    afterEach(() => {
        vi.restoreAllMocks();
        vi.useRealTimers();
    });

    it("shows inherited Department AI Engine state, applies an override, and then reverts to the organization default", async () => {
        setupOrganizationFetch();

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByRole("heading", { name: "Departments" })).toBeDefined();
        fireEvent.click(screen.getAllByRole("button", { name: "Open Departments" })[0]);

        expect(await screen.findByText("Using Organization Default: Starter defaults included")).toBeDefined();
        fireEvent.click(screen.getByRole("button", { name: "Change for this Team" }));
        expect(await screen.findByRole("heading", { name: "Choose an AI Engine for this Team" })).toBeDefined();
        fireEvent.click(screen.getByRole("button", { name: /Balanced/i }));
        fireEvent.click(screen.getByRole("button", { name: "Use selected AI Engine" }));

        expect(await screen.findByText("Overridden: Balanced")).toBeDefined();
        expect(screen.getByRole("button", { name: "Revert to Organization Default" })).toBeDefined();

        fireEvent.click(screen.getByRole("button", { name: "Revert to Organization Default" }));
        expect(await screen.findByText("Using Organization Default: Starter defaults included")).toBeDefined();
    }, 15000);

    it("lets the operator bind an Agent Type AI Engine and then return it to the Team default", async () => {
        setupOrganizationFetch();

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByRole("heading", { name: "Departments" })).toBeDefined();
        fireEvent.click(screen.getAllByRole("button", { name: "Open Departments" })[0]);

        expect(await screen.findByText("Using Team Default: Starter defaults included")).toBeDefined();
        fireEvent.click(screen.getAllByRole("button", { name: "Change for this Agent Type" }).at(-1)!);
        expect(await screen.findByRole("heading", { name: "Choose an AI Engine for this Agent Type" })).toBeDefined();
        fireEvent.click(screen.getByRole("button", { name: /Balanced/i }));
        fireEvent.click(screen.getAllByRole("button", { name: "Use selected AI Engine" }).at(-1)!);

        expect(await screen.findByText("Type-specific Engine: Balanced")).toBeDefined();
        expect(screen.getAllByRole("button", { name: "Use Team Default" }).at(-1)).toBeDefined();

        fireEvent.click(screen.getAllByRole("button", { name: "Use Team Default" }).at(-1)!);
        expect(await screen.findByText("Using Team Default: Starter defaults included")).toBeDefined();
    }, 15000);
});
