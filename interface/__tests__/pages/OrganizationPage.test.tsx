import { describe, it, expect, vi, beforeEach } from "vitest";
import { act, fireEvent, render, screen } from "@testing-library/react";
import { mockFetch } from "../setup";

vi.mock("next/navigation", () => ({
    useRouter: () => ({ push: vi.fn(), replace: vi.fn(), back: vi.fn(), prefetch: vi.fn() }),
    usePathname: () => "/organizations/org-123",
}));

import OrganizationPage from "@/app/(app)/organizations/[id]/page";

const organizationHome = {
    id: "org-123",
    name: "Northstar Labs",
    purpose: "Ship a focused AI engineering organization for product delivery.",
    start_mode: "template",
    template_id: "engineering-starter",
    template_name: "Engineering Starter",
    team_lead_label: "Team Lead",
    advisor_count: 1,
    department_count: 1,
    specialist_count: 2,
    ai_engine_settings_summary: "Starter defaults included",
    memory_personality_summary: "Prepared for Adaptive Delivery work",
    status: "ready",
    description: "Guided AI Organization for engineering work",
};

function jsonResponse(body: unknown, status = 200) {
    return Promise.resolve(new Response(JSON.stringify(body), { status }));
}

function setupOrganizationFetch(options?: {
    homeHandler?: () => Promise<Response>;
    actionHandler?: (body: Record<string, unknown>) => Promise<Response>;
}) {
    mockFetch.mockImplementation((input, init) => {
        const url = typeof input === "string" ? input : input instanceof Request ? input.url : String(input);
        const method = init?.method ?? (input instanceof Request ? input.method : "GET");

        if (url.includes("/api/v1/organizations/org-123/home")) {
            return options?.homeHandler?.() ?? jsonResponse({ ok: true, data: organizationHome });
        }

        if (url.includes("/api/v1/organizations/org-123/workspace/actions") && method === "POST") {
            const body = init?.body ? JSON.parse(String(init.body)) as Record<string, unknown> : {};
            return options?.actionHandler?.(body) ?? jsonResponse({
                ok: true,
                data: {
                    action: "plan_next_steps",
                    request_label: "Plan next steps for this organization",
                    headline: "Team Lead plan for Northstar Labs",
                    summary: "Team Lead recommends a focused first delivery loop.",
                    priority_steps: [
                        "Align the first outcome with the AI Organization purpose.",
                        "Use the first Department as the routing layer for work.",
                    ],
                    suggested_follow_ups: [
                        "Review my organization setup",
                        "What should I focus on first?",
                    ],
                },
            });
        }

        throw new Error(`Unexpected fetch: ${method} ${url}`);
    });
}

describe("OrganizationPage (/organizations/[id])", () => {
    beforeEach(() => {
        mockFetch.mockReset();
    });

    it("renders a Team Lead-first organization workspace with guided actions and no generic or dev wording", async () => {
        setupOrganizationFetch();

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByText("AI Organization Home")).toBeDefined();
        expect(screen.getByText("Team Lead ready")).toBeDefined();
        expect(screen.getAllByText("Team Lead for Northstar Labs").length).toBeGreaterThan(0);
        expect(screen.getByText("What I can help with")).toBeDefined();
        expect(screen.getByText("Work with the Team Lead")).toBeDefined();
        expect(screen.getByRole("heading", { name: "Advisors" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "Departments" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "AI Engine Settings" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "Memory & Personality" })).toBeDefined();
        expect(screen.getByText("Advisor support")).toBeDefined();
        expect(screen.getByText("Department view")).toBeDefined();
        expect(screen.getByText("Planning review")).toBeDefined();
        expect(screen.getByText("Started from Engineering Starter")).toBeDefined();
        expect(screen.getAllByText("What this affects").length).toBeGreaterThan(0);
        expect(screen.getByText("Response style")).toBeDefined();
        expect(screen.getByText("Planning depth")).toBeDefined();
        expect(screen.getByText("Working tone")).toBeDefined();
        expect(screen.getByText("Context continuity")).toBeDefined();
        expect(screen.getByRole("button", { name: /Plan next steps for this organization/i })).toBeDefined();
        expect(screen.getByRole("button", { name: /What should I focus on first\?/i })).toBeDefined();
        expect(screen.getByRole("button", { name: /Review my organization setup/i })).toBeDefined();
        expect(screen.queryByText(/context shell/i)).toBeNull();
        expect(screen.queryByText(/bounded slice/i)).toBeNull();
        expect(screen.queryByText(/implementation slice/i)).toBeNull();
        expect(screen.queryByText(/contract/i)).toBeNull();
        expect(screen.queryByText(/New Chat/i)).toBeNull();
        expect(screen.queryByText(/generic chat/i)).toBeNull();
    });

    it("keeps the organization frame visible while moving from home into the Team Lead interaction flow", async () => {
        setupOrganizationFetch({
            actionHandler: (body) => jsonResponse({
                ok: true,
                data: {
                    action: body.action,
                    request_label: "Review my organization setup",
                    headline: "Organization setup review for Northstar Labs",
                    summary: "Team Lead is reviewing the current AI Organization shape.",
                    priority_steps: [
                        "Advisors: 1 advisor ready.",
                        "Departments: 1 department ready.",
                    ],
                    suggested_follow_ups: [
                        "Plan next steps for this organization",
                        "What should I focus on first?",
                    ],
                },
            }),
        });

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByText("Work with the Team Lead")).toBeDefined();
        fireEvent.click(screen.getByRole("button", { name: /Review my organization setup/i }));
        expect(await screen.findByText("Organization setup review for Northstar Labs")).toBeDefined();
        expect(screen.getByText("Priority steps")).toBeDefined();
        expect(screen.getByText("Northstar Labs")).toBeDefined();
        expect(screen.getByText("AI Organization Home")).toBeDefined();
        expect(screen.getAllByText("Team Lead for Northstar Labs").length).toBeGreaterThan(0);
        expect(screen.getByRole("heading", { name: "Advisors" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "Departments" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "AI Engine Settings" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "Memory & Personality" })).toBeDefined();
    });

    it("preserves the organization context when a Team Lead action fails and then succeeds on retry", async () => {
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
                        request_label: "Plan next steps for this organization",
                        headline: "Team Lead plan for Northstar Labs",
                        summary: "Team Lead recommends a focused first delivery loop.",
                        priority_steps: [
                            "Align the first outcome with the AI Organization purpose.",
                            "Use the first Department as the routing layer for work.",
                        ],
                        suggested_follow_ups: [
                            "Review my organization setup",
                            "What should I focus on first?",
                        ],
                    },
                });
            },
        });

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByText("Work with the Team Lead")).toBeDefined();
        fireEvent.click(screen.getByRole("button", { name: /Plan next steps for this organization/i }));

        expect(await screen.findByText("Team Lead guidance is unavailable")).toBeDefined();
        expect(screen.getByText("AI Organization Home")).toBeDefined();
        expect(screen.getByText("Team Lead ready")).toBeDefined();
        expect(screen.getAllByText("Team Lead for Northstar Labs").length).toBeGreaterThan(0);

        fireEvent.click(screen.getByRole("button", { name: "Retry Team Lead action" }));

        expect(await screen.findByText("Team Lead plan for Northstar Labs")).toBeDefined();
        expect(screen.getByText("Priority steps")).toBeDefined();
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
                    },
                }),
        });

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByRole("heading", { name: "Advisors" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "Departments" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "AI Engine Settings" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "Memory & Personality" })).toBeDefined();
        expect(screen.getAllByText("Inspect only").length).toBeGreaterThan(0);
        expect(screen.getByText("Advisor roles appear here once they are added")).toBeDefined();
        expect(screen.getByText("Add the first Department when ready")).toBeDefined();
        expect(screen.getByText("Started from Empty")).toBeDefined();
        expect(screen.getByText("The current AI Engine Settings keep the organization on a simple starter profile until deeper tuning is needed.")).toBeDefined();
        expect(screen.getByText("Memory & Personality stay on a simple starter posture so the Team Lead keeps a consistent tone and working style.")).toBeDefined();
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
        expect(await screen.findByText("Team Lead ready")).toBeDefined();
    });
});

