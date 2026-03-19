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
    ai_engine_profile_id: "starter_defaults",
    ai_engine_settings_summary: "Starter defaults included",
    memory_personality_summary: "Prepared for Adaptive Delivery work",
    status: "ready",
    description: "Guided AI Organization for engineering work",
    departments: [
        {
            id: "platform",
            name: "Platform Department",
            specialist_count: 2,
            ai_engine_effective_profile_id: "starter_defaults",
            ai_engine_effective_summary: "Starter defaults included",
            inherits_organization_ai_engine: true,
        },
    ],
};

function jsonResponse(body: unknown, status = 200) {
    return Promise.resolve(new Response(JSON.stringify(body), { status }));
}

function applyOrganizationAIEngineToDepartments(
    home: typeof organizationHome,
    profileId: string,
    summary: string,
) {
    return {
        ...home,
        ai_engine_profile_id: profileId,
        ai_engine_settings_summary: summary,
        departments: home.departments.map((department) =>
            department.inherits_organization_ai_engine
                ? {
                      ...department,
                      ai_engine_effective_profile_id: profileId,
                      ai_engine_effective_summary: summary,
                  }
                : department,
        ),
    };
}

function setupOrganizationFetch(options?: {
    homeHandler?: () => Promise<Response>;
    actionHandler?: (body: Record<string, unknown>) => Promise<Response>;
    aiEngineUpdateHandler?: (body: Record<string, unknown>) => Promise<Response>;
    departmentAIEngineUpdateHandler?: (body: Record<string, unknown>) => Promise<Response>;
}) {
    let currentOrganizationHome = structuredClone(organizationHome);

    mockFetch.mockImplementation((input, init) => {
        const url = typeof input === "string" ? input : input instanceof Request ? input.url : String(input);
        const method = init?.method ?? (input instanceof Request ? input.method : "GET");

        if (url.includes("/api/v1/organizations/org-123/home")) {
            return options?.homeHandler?.() ?? jsonResponse({ ok: true, data: currentOrganizationHome });
        }

        if (url.includes("/api/v1/organizations/org-123/ai-engine") && method === "PATCH") {
            const body = init?.body ? JSON.parse(String(init.body)) as Record<string, unknown> : {};
            if (options?.aiEngineUpdateHandler) {
                return options.aiEngineUpdateHandler(body);
            }

            const summaries: Record<string, string> = {
                starter_defaults: "Starter Defaults",
                balanced: "Balanced",
                high_reasoning: "High Reasoning",
                fast_lightweight: "Fast & Lightweight",
                deep_planning: "Deep Planning",
            };
            const profileId = String(body.profile_id ?? "");
            currentOrganizationHome = applyOrganizationAIEngineToDepartments(
                currentOrganizationHome,
                profileId,
                summaries[profileId] ?? currentOrganizationHome.ai_engine_settings_summary,
            );

            return jsonResponse({ ok: true, data: currentOrganizationHome });
        }

        if (url.includes("/api/v1/organizations/org-123/departments/platform/ai-engine") && method === "PATCH") {
            const body = init?.body ? JSON.parse(String(init.body)) as Record<string, unknown> : {};
            if (options?.departmentAIEngineUpdateHandler) {
                return options.departmentAIEngineUpdateHandler(body);
            }

            if (body.revert_to_organization_default) {
                currentOrganizationHome = {
                    ...currentOrganizationHome,
                    departments: currentOrganizationHome.departments.map((department) =>
                        department.id === "platform"
                            ? {
                                  ...department,
                                  ai_engine_override_profile_id: undefined,
                                  ai_engine_override_summary: undefined,
                                  ai_engine_effective_profile_id: currentOrganizationHome.ai_engine_profile_id,
                                  ai_engine_effective_summary: currentOrganizationHome.ai_engine_settings_summary,
                                  inherits_organization_ai_engine: true,
                              }
                            : department,
                    ),
                };
                return jsonResponse({ ok: true, data: currentOrganizationHome });
            }

            const summaries: Record<string, string> = {
                starter_defaults: "Starter Defaults",
                balanced: "Balanced",
                high_reasoning: "High Reasoning",
                fast_lightweight: "Fast & Lightweight",
                deep_planning: "Deep Planning",
            };
            const profileId = String(body.profile_id ?? "");
            currentOrganizationHome = {
                ...currentOrganizationHome,
                departments: currentOrganizationHome.departments.map((department) =>
                    department.id === "platform"
                        ? {
                              ...department,
                              ai_engine_override_profile_id: profileId,
                              ai_engine_override_summary: summaries[profileId],
                              ai_engine_effective_profile_id: profileId,
                              ai_engine_effective_summary: summaries[profileId] ?? department.ai_engine_effective_summary,
                              inherits_organization_ai_engine: false,
                          }
                        : department,
                ),
            };

            return jsonResponse({ ok: true, data: currentOrganizationHome });
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
        expect(screen.getAllByRole("button", { name: "Review Advisors" }).length).toBeGreaterThan(0);
        expect(screen.getAllByRole("button", { name: "Open Departments" }).length).toBeGreaterThan(0);
        expect(screen.getAllByRole("button", { name: "Review AI Engine Settings" }).length).toBeGreaterThan(0);
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

    it("opens Advisor details from the Team Lead action and keeps the Team Lead workspace visible", async () => {
        setupOrganizationFetch();

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByText("Work with the Team Lead")).toBeDefined();
        fireEvent.click(screen.getAllByRole("button", { name: "Review Advisors" })[0]);

        expect(await screen.findByRole("heading", { name: "Advisor details" })).toBeDefined();
        expect(screen.getByText("Planning Advisor")).toBeDefined();
        expect(screen.getByText("Decision support")).toBeDefined();
        expect(screen.getByText("AI Organization Home")).toBeDefined();
        expect(screen.getAllByText("Team Lead for Northstar Labs").length).toBeGreaterThan(0);
        expect(screen.getByText("Work with the Team Lead")).toBeDefined();
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
        expect(screen.getByText("AI Organization Home")).toBeDefined();
        expect(screen.getAllByText("Team Lead for Northstar Labs").length).toBeGreaterThan(0);

        fireEvent.click(screen.getByRole("button", { name: "Back to Team Lead" }));
        expect(screen.queryByRole("heading", { name: "Department details" })).toBeNull();
        expect(screen.getByText("Work with the Team Lead")).toBeDefined();
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
        expect(screen.getByText("Work with the Team Lead")).toBeDefined();
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
        expect(screen.getByText("Work with the Team Lead")).toBeDefined();
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

        fireEvent.click(screen.getByRole("button", { name: "Back to Team Lead" }));
        fireEvent.click(screen.getAllByRole("button", { name: "Review AI Engine Settings" })[0]);
        fireEvent.click(screen.getByRole("button", { name: "Change AI Engine" }));
        fireEvent.click(screen.getByRole("button", { name: /Fast & Lightweight/i }));
        fireEvent.click(screen.getByRole("button", { name: "Use selected AI Engine" }));
        expect(await screen.findByText("Current profile: Fast & Lightweight.")).toBeDefined();

        fireEvent.click(screen.getByRole("button", { name: "Back to Team Lead" }));
        fireEvent.click(screen.getAllByRole("button", { name: "Open Departments" })[0]);
        expect(await screen.findByText("Overridden: High Reasoning")).toBeDefined();

        fireEvent.click(screen.getByRole("button", { name: "Revert to Organization Default" }));
        expect(await screen.findByText("Using Organization Default: Fast & Lightweight")).toBeDefined();
    });

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
        expect(screen.getAllByText("Team Lead for Northstar Labs").length).toBeGreaterThan(0);

        fireEvent.click(screen.getByRole("button", { name: "Retry AI Engine change" }));

        expect(await screen.findByText("Current profile: Balanced.")).toBeDefined();
        expect(screen.queryByText("Unable to update AI Engine Settings")).toBeNull();
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

