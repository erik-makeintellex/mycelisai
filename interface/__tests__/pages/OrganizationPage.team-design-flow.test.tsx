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

describe("OrganizationPage team design slices", () => {
    beforeEach(() => {
        mockFetch.mockReset();
        resetOrganizationPageStoreState();
    });

    afterEach(() => {
        vi.restoreAllMocks();
        vi.useRealTimers();
    });

    it("keeps the organization frame visible while moving from home into the Soma interaction flow", async () => {
        setupOrganizationFetch({
            actionHandler: (body) => jsonResponse({
                ok: true,
                data: {
                    action: body.action,
                    request_label: "Review your organization setup",
                    headline: "Organization setup review for Northstar Labs",
                    summary: "Soma is reviewing the current AI Organization shape.",
                    priority_steps: [
                        "Advisors: 1 advisor ready.",
                        "Departments: 1 department ready.",
                    ],
                    suggested_follow_ups: [
                        "Run a quick strategy check",
                        "Choose the first priority",
                    ],
                },
            }),
        });

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByText("Create teams with Soma")).toBeDefined();
        fireEvent.click(screen.getByRole("button", { name: "Create teams with Soma" }));
        fireEvent.click(screen.getByRole("button", { name: /Review your organization setup/i }));
        expect(await screen.findByText("Organization setup review for Northstar Labs")).toBeDefined();
        expect(screen.getByText("Priority steps")).toBeDefined();
        expect(screen.getByRole("heading", { name: /^Northstar Labs$/ })).toBeDefined();
        expect(screen.getByText("AI Organization Home")).toBeDefined();
        expect(screen.getAllByText(/Team Lead for Northstar Labs/i).length).toBeGreaterThan(0);
        expect(screen.getByRole("heading", { name: "Advisors" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "Departments" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "AI Engine Settings" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "Response Style" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "Memory & Continuity" })).toBeDefined();
    }, 15000);

    it("shows native team execution guidance for image-oriented team design requests", async () => {
        setupOrganizationFetch({
            actionHandler: (body) => {
                expect(body).toEqual({
                    action: "plan_next_steps",
                    request_context: "Create a creative team to generate a launch hero image.",
                });

                return jsonResponse({
                    ok: true,
                    data: {
                        action: "plan_next_steps",
                        request_label: "Plan next steps for this organization",
                        headline: "Team Lead plan for Northstar Labs",
                        summary: "Soma is ready to shape a creative delivery path for Northstar Labs.",
                        priority_steps: [
                            "Align the first outcome with the AI Organization purpose.",
                            "Use the first Department as the routing layer for work.",
                        ],
                        suggested_follow_ups: [
                            "Review your organization setup",
                            "Choose the first priority",
                        ],
                        execution_contract: {
                            execution_mode: "native_team",
                            owner_label: "Native Mycelis team",
                            team_name: "Creative Delivery Team",
                            summary: "Use a bounded creative team inside Northstar Labs so Soma can shape the work and return the generated image as a managed artifact.",
                            target_outputs: [
                                "Reviewable image artifact",
                                "Short concept note",
                            ],
                        },
                    },
                });
            },
        });

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByText("Create teams with Soma")).toBeDefined();
        fireEvent.click(screen.getByRole("button", { name: "Create teams with Soma" }));
        fireEvent.change(screen.getByLabelText("Tell Soma what team or delivery lane you want to create"), {
            target: { value: "Create a creative team to generate a launch hero image." },
        });
        fireEvent.click(screen.getByRole("button", { name: "Start team design" }));

        expect(await screen.findByText("Execution path")).toBeDefined();
        expect(screen.getAllByText("Native Mycelis team").length).toBeGreaterThan(0);
        expect(screen.getByText("Creative Delivery Team")).toBeDefined();
        expect(screen.getByText("Reviewable image artifact")).toBeDefined();
        expect(screen.getByRole("heading", { name: /^Northstar Labs$/ })).toBeDefined();
        expect(screen.getByText("AI Organization Home")).toBeDefined();
    });

    it("preserves the organization context when a Soma action fails and then succeeds on retry", async () => {
        let attempt = 0;
        setupOrganizationFetch({
            actionHandler: (body) => {
                attempt += 1;
                if (attempt === 1) {
                    return jsonResponse({ ok: false, error: "Team Lead guidance is unavailable right now." }, 500);
                }
                return jsonResponse({
                    ok: true,
                    data: {
                        action: body.action,
                        request_label: "Run a quick strategy check",
                        headline: "Team Lead plan for Northstar Labs",
                        summary: "Team Lead recommends a clear next move for Northstar Labs.",
                        priority_steps: [
                            "Align the first outcome with the AI Organization purpose.",
                            "Use the first Department as the routing layer for work.",
                        ],
                        suggested_follow_ups: [
                            "Review your organization setup",
                            "Choose the first priority",
                        ],
                    },
                });
            },
        });

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByText("Create teams with Soma")).toBeDefined();
        fireEvent.click(screen.getByRole("button", { name: "Create teams with Soma" }));
        fireEvent.click(screen.getByRole("button", { name: /Run a quick strategy check/i }));

        expect(await screen.findByText("Soma guidance is unavailable")).toBeDefined();
        expect(screen.getByText("AI Organization Home")).toBeDefined();
        expect(screen.getByText("Soma ready")).toBeDefined();
        expect(screen.getAllByText(/Team Lead for Northstar Labs/i).length).toBeGreaterThan(0);

        fireEvent.click(screen.getByRole("button", { name: "Retry Soma action" }));

        expect(await screen.findByText("Soma plan for Northstar Labs")).toBeDefined();
        expect(screen.getByText("Priority steps")).toBeDefined();
    }, 20000);

    it("opens the create-team flow from the primary Soma workspace without leaving the organization page", async () => {
        setupOrganizationFetch();

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByRole("heading", { name: "Talk with Soma" })).toBeDefined();
        fireEvent.click(screen.getByRole("button", { name: "Create teams with Soma" }));
        fireEvent.click(await screen.findByRole("button", { name: "Open crew launcher" }));

        expect(await screen.findByRole("heading", { name: "Launch a Crew" })).toBeDefined();
        expect(screen.getByText(/must return a real outcome, not a planning stub/i)).toBeDefined();
        expect(screen.getByRole("heading", { name: "Soma for Northstar Labs" })).toBeDefined();
    });
});
