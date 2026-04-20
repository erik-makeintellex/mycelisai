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

describe("OrganizationPage AI engine and output model slices", () => {
    beforeEach(() => {
        mockFetch.mockReset();
        resetOrganizationPageStoreState();
    });

    afterEach(() => {
        vi.restoreAllMocks();
        vi.useRealTimers();
    });

    it("opens AI Engine Settings details and shows organization, team, and role scope labels", async () => {
        setupOrganizationFetch();

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByRole("heading", { name: "AI Engine Settings" })).toBeDefined();
        fireEvent.click(screen.getAllByRole("button", { name: "Review AI Engine Settings" })[0]);

        expect(await screen.findByRole("heading", { name: "AI Engine Settings details" })).toBeDefined();
        expect(screen.getByRole("button", { name: "Change AI Engine" })).toBeDefined();
        expect(screen.getByText("Organization-wide AI engine")).toBeDefined();
        expect(screen.getByText("Team defaults")).toBeDefined();
        expect(screen.getByText("Specific role overrides")).toBeDefined();
        expect(screen.getByText("Current profile: Starter defaults included.")).toBeDefined();
        expect(screen.getByText("Departments start from the organization-wide AI engine unless a team-specific setting appears here.")).toBeDefined();
        expect(screen.getByText("No specific role overrides are visible in this workspace right now.")).toBeDefined();
        expect(screen.getByText("AI Organization Home")).toBeDefined();
        expect(screen.getByText("Create teams with Soma")).toBeDefined();
    });

    it("lets the operator change the organization AI Engine through a guided selection flow", async () => {
        setupOrganizationFetch();

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByRole("heading", { name: "AI Engine Settings" })).toBeDefined();
        fireEvent.click(screen.getAllByRole("button", { name: "Review AI Engine Settings" })[0]);
        fireEvent.click(screen.getByRole("button", { name: "Change AI Engine" }));

        expect(await screen.findByRole("heading", { name: "Choose an AI Engine profile" })).toBeDefined();
        expect(screen.getByText("Balanced")).toBeDefined();
        expect(screen.getByText("High Reasoning")).toBeDefined();
        fireEvent.click(screen.getByRole("button", { name: /High Reasoning/i }));
        fireEvent.click(screen.getByRole("button", { name: "Use selected AI Engine" }));

        expect(await screen.findByText("Current profile: High Reasoning.")).toBeDefined();
        expect(screen.getByText("The current AI Engine Settings profile is high reasoning and shapes how the organization responds, plans, and carries work forward.")).toBeDefined();
        expect(screen.getByText("AI Organization Home")).toBeDefined();
        expect(screen.getByText("Create teams with Soma")).toBeDefined();
    }, 15000);

    it("shows retry guidance when changing the AI Engine fails and recovers on retry", async () => {
        let attempts = 0;
        setupOrganizationFetch({
            aiEngineUpdateHandler: (body) => {
                attempts += 1;
                if (attempts === 1) {
                    return jsonResponse({ ok: false, error: "AI Engine update is unavailable right now." }, 500);
                }

                return jsonResponse({
                    ok: true,
                    data: {
                        ...organizationHome,
                        ai_engine_profile_id: body.profile_id,
                        ai_engine_settings_summary: "Balanced",
                    },
                });
            },
        });

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByRole("heading", { name: "AI Engine Settings" })).toBeDefined();
        fireEvent.click(screen.getAllByRole("button", { name: "Review AI Engine Settings" })[0]);
        fireEvent.click(screen.getByRole("button", { name: "Change AI Engine" }));
        fireEvent.click(screen.getByRole("button", { name: /Balanced/i }));
        fireEvent.click(screen.getByRole("button", { name: "Use selected AI Engine" }));

        expect(await screen.findByText("Unable to update AI Engine Settings")).toBeDefined();
        expect(screen.getByText("AI Engine update is unavailable right now.")).toBeDefined();
        expect(screen.getByRole("button", { name: "Retry AI Engine change" })).toBeDefined();
        expect(screen.getByText("AI Organization Home")).toBeDefined();
        expect(screen.getAllByText(/Team Lead for Northstar Labs/i).length).toBeGreaterThan(0);

        fireEvent.click(screen.getByRole("button", { name: "Retry AI Engine change" }));

        expect(await screen.findByText("Current profile: Balanced.")).toBeDefined();
        expect(screen.queryByText("Unable to update AI Engine Settings")).toBeNull();
    }, 15000);
});
