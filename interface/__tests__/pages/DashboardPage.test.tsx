import { describe, it, expect, vi, beforeEach } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { mockFetch } from "../setup";

const push = vi.fn();

vi.mock("next/navigation", () => ({
    useRouter: () => ({ push, replace: vi.fn(), back: vi.fn(), prefetch: vi.fn() }),
}));

import DashboardPage from "@/app/(app)/dashboard/page";

const starterTemplate = {
    id: "engineering-starter",
    name: "Engineering Starter",
    description: "Guided AI Organization for engineering work",
    organization_type: "AI Organization starter",
    team_lead_label: "Team Lead",
    advisor_count: 1,
    department_count: 1,
    specialist_count: 2,
    ai_engine_settings_summary: "Starter defaults included",
    memory_personality_summary: "Prepared for Adaptive Delivery work",
};

const organizationSummary = {
    id: "org-42",
    name: "Atlas",
    purpose: "Resume me later",
    start_mode: "empty",
    team_lead_label: "Team Lead",
    advisor_count: 0,
    department_count: 0,
    specialist_count: 0,
    ai_engine_settings_summary: "Set up later in Advanced mode",
    memory_personality_summary: "Set up later in Advanced mode",
    status: "ready",
};

const diagnosticOrganizationSummary = {
    id: "org-qa-1",
    name: "QA Scenario A 171717",
    purpose: "Live governance verification",
    start_mode: "empty",
    team_lead_label: "Team Lead",
    advisor_count: 0,
    department_count: 0,
    specialist_count: 0,
    ai_engine_settings_summary: "Set up later in Advanced mode",
    memory_personality_summary: "Set up later in Advanced mode",
    status: "ready",
};

const diagnosticOrganizationSummaryTwo = {
    ...diagnosticOrganizationSummary,
    id: "org-qa-2",
    name: "Testing Setup B 171718",
    purpose: "UI verification lane",
};

const diagnosticOrganizationSummaryThree = {
    ...diagnosticOrganizationSummary,
    id: "org-qa-3",
    name: "QA Scenario C 171719",
};

function jsonResponse(body: unknown, status = 200) {
    return Promise.resolve(new Response(JSON.stringify(body), { status }));
}

function setupEntryFlowFetch(options?: {
    templateHandler?: (attempt: number) => Promise<Response>;
    organizationsHandler?: (attempt: number) => Promise<Response>;
    createHandler?: () => Promise<Response>;
}) {
    let templateAttempts = 0;
    let organizationAttempts = 0;

    mockFetch.mockImplementation((input, init) => {
        const url = typeof input === "string" ? input : input instanceof Request ? input.url : String(input);
        const method = init?.method ?? (input instanceof Request ? input.method : "GET");

        if (url.includes("/api/v1/templates?view=organization-starters")) {
            templateAttempts += 1;
            return options?.templateHandler?.(templateAttempts) ?? jsonResponse({ ok: true, data: [starterTemplate] });
        }

        if (url.includes("/api/v1/organizations?view=summary")) {
            organizationAttempts += 1;
            return options?.organizationsHandler?.(organizationAttempts) ?? jsonResponse({ ok: true, data: [] });
        }

        if (url.endsWith("/api/v1/organizations") && method === "POST") {
            return options?.createHandler?.() ?? jsonResponse({
                ok: true,
                data: {
                    id: "org-123",
                    name: "Northstar Labs",
                    purpose: "Ship a focused AI engineering organization for product delivery.",
                    start_mode: "empty",
                    team_lead_label: "Team Lead",
                    advisor_count: 0,
                    department_count: 0,
                    specialist_count: 0,
                    ai_engine_settings_summary: "Set up later in Advanced mode",
                    memory_personality_summary: "Set up later in Advanced mode",
                    status: "ready",
                },
            }, 201);
        }

        throw new Error(`Unexpected fetch: ${method} ${url}`);
    });
}

describe("Dashboard Page (V8 AI Organization entry flow)", () => {
    beforeEach(() => {
        push.mockReset();
        mockFetch.mockReset();
        localStorage.clear();
    });

    it("renders Central Soma home above the AI Organization entry flow without architecture copy leaks", async () => {
        setupEntryFlowFetch();

        render(<DashboardPage />);

        expect(await screen.findByRole("heading", { name: "Work with one Soma across every AI Organization." })).toBeDefined();
        expect(screen.getByText("Central Soma")).toBeDefined();
        expect(screen.getByText(/Soma and Council stay persistent across organizations/i)).toBeDefined();
        expect(screen.getByRole("link", { name: /Review the Soma context model/i })).toBeDefined();
        expect(await screen.findByRole("heading", { name: "Create AI Organization" })).toBeDefined();
        expect(screen.getByText("AI Organization Setup")).toBeDefined();
        expect(screen.getByRole("button", { name: "Explore Templates" })).toBeDefined();
        expect(screen.getAllByText("Create AI Organization").length).toBeGreaterThan(0);
        expect(screen.getByRole("button", { name: "Start Empty" })).toBeDefined();
        expect(screen.getByText("Memory & Continuity")).toBeDefined();
        expect(screen.queryByText("Memory & Personality")).toBeNull();
        expect(screen.queryByText("V8 Entry Flow")).toBeNull();
        expect(screen.queryByText(/contract/i)).toBeNull();
        expect(screen.queryByText(/context shell/i)).toBeNull();
        expect(screen.queryByText(/implementation slice/i)).toBeNull();
        expect(screen.queryByText(/bounded slice/i)).toBeNull();
        expect(screen.queryByText(/raw architecture controls/i)).toBeNull();
    });

    it("preserves recent organizations when starter templates fail and lets the operator retry or start empty", async () => {
        setupEntryFlowFetch({
            templateHandler: (attempt) =>
                attempt === 1
                    ? jsonResponse({ ok: false, error: "Starter templates are unavailable right now." }, 500)
                    : jsonResponse({ ok: true, data: [starterTemplate] }),
            organizationsHandler: () => jsonResponse({ ok: true, data: [organizationSummary] }),
        });

        render(<DashboardPage />);

        expect(await screen.findByText("Starter templates are unavailable right now.")).toBeDefined();
        expect(screen.getByText("Atlas")).toBeDefined();
        expect(screen.getByRole("button", { name: "Retry starters" })).toBeDefined();
        expect(screen.getByRole("button", { name: "Start empty instead" })).toBeDefined();

        fireEvent.click(screen.getByRole("button", { name: "Start empty instead" }));
        expect(screen.getByText(/creates the AI Organization first/i)).toBeDefined();

        fireEvent.click(screen.getByRole("button", { name: "Retry starters" }));
        fireEvent.click(await screen.findByRole("button", { name: /Start from template/i }));
        expect(await screen.findByText("Engineering Starter")).toBeDefined();
    });

    it("preserves template selection when recent organizations fail and offers a retry path", async () => {
        setupEntryFlowFetch({
            templateHandler: () => jsonResponse({ ok: true, data: [starterTemplate] }),
            organizationsHandler: () => jsonResponse({ ok: false, error: "Recent AI Organizations are unavailable right now." }, 500),
        });

        render(<DashboardPage />);

        expect(await screen.findByText("Recent AI Organizations are unavailable right now.")).toBeDefined();
        expect(screen.getByRole("button", { name: "Retry recent AI Organizations" })).toBeDefined();

        fireEvent.click(screen.getByRole("button", { name: /Start from template/i }));
        expect(await screen.findByText("Engineering Starter")).toBeDefined();
        expect(screen.getByText("AI Organization starter")).toBeDefined();
    });

    it("retries recent organizations successfully after an initial failure", async () => {
        setupEntryFlowFetch({
            organizationsHandler: (attempt) =>
                attempt === 1
                    ? jsonResponse({ ok: false, error: "Recent AI Organizations are unavailable right now." }, 500)
                    : jsonResponse({ ok: true, data: [organizationSummary] }),
        });

        render(<DashboardPage />);

        expect(await screen.findByText("Recent AI Organizations are unavailable right now.")).toBeDefined();
        fireEvent.click(screen.getByRole("button", { name: "Retry recent AI Organizations" }));
        expect(await screen.findByText("Atlas")).toBeDefined();
    });

    it("opens a recent AI Organization from the entry screen", async () => {
        setupEntryFlowFetch({
            organizationsHandler: () => jsonResponse({ ok: true, data: [organizationSummary] }),
        });

        render(<DashboardPage />);

        expect(await screen.findByText("Atlas")).toBeDefined();
        fireEvent.click(screen.getByRole("button", { name: /Open AI Organization/i }));

        await waitFor(() => {
            expect(push).toHaveBeenCalledWith("/organizations/org-42");
        });
    });

    it("shows a direct return path for the last opened AI Organization", async () => {
        localStorage.setItem("mycelis-last-organization-id", "org-42");
        localStorage.setItem("mycelis-last-organization-name", "Atlas");
        setupEntryFlowFetch();

        render(<DashboardPage />);

        expect(await screen.findByRole("link", { name: /Return to Organization/i })).toBeDefined();
        expect(screen.getByText("Atlas")).toBeDefined();
        expect(screen.getByRole("link", { name: /Return to Organization/i }).getAttribute("href")).toBe("/organizations/org-42");
    });

    it("hides diagnostic QA organizations by default and lets the operator reveal them intentionally", async () => {
        setupEntryFlowFetch({
            organizationsHandler: () => jsonResponse({ ok: true, data: [organizationSummary, diagnosticOrganizationSummary] }),
        });

        render(<DashboardPage />);

        expect(await screen.findByText("Atlas")).toBeDefined();
        expect(screen.queryByText("QA Scenario A 171717")).toBeNull();
        expect(screen.getByRole("button", { name: /Show 1 testing setup/i })).toBeDefined();

        fireEvent.click(screen.getByRole("button", { name: /Show 1 testing setup/i }));
        expect(await screen.findByText("QA Scenario A 171717")).toBeDefined();
    });

    it("caps revealed testing setups to two and explains when older fixtures stay hidden", async () => {
        setupEntryFlowFetch({
            organizationsHandler: () => jsonResponse({
                ok: true,
                data: [organizationSummary, diagnosticOrganizationSummary, diagnosticOrganizationSummaryTwo, diagnosticOrganizationSummaryThree],
            }),
        });

        render(<DashboardPage />);

        expect(await screen.findByText("Atlas")).toBeDefined();
        expect(screen.queryByText("QA Scenario A 171717")).toBeNull();
        expect(screen.queryByText("Testing Setup B 171718")).toBeNull();
        expect(screen.queryByText("QA Scenario C 171719")).toBeNull();

        fireEvent.click(screen.getByRole("button", { name: /Show 2 testing setups/i }));

        expect(await screen.findByText("QA Scenario A 171717")).toBeDefined();
        expect(screen.getByText("Testing Setup B 171718")).toBeDefined();
        expect(screen.queryByText("QA Scenario C 171719")).toBeNull();
        expect(screen.getByText(/Keeping 1 older testing setup out of the default product view/i)).toBeDefined();
    });

    it("submits an empty-start organization and routes into the landing screen", async () => {
        setupEntryFlowFetch();

        render(<DashboardPage />);

        expect(await screen.findByRole("heading", { name: "Create AI Organization" })).toBeDefined();
        fireEvent.click(screen.getByRole("button", { name: /Start empty begin with a clean ai organization/i }));
        fireEvent.change(screen.getByLabelText("AI Organization name"), { target: { value: "Northstar Labs" } });
        fireEvent.change(screen.getByLabelText("Purpose"), { target: { value: "Ship a focused AI engineering organization for product delivery." } });
        fireEvent.click(screen.getByRole("button", { name: "Create AI Organization" }));

        await waitFor(() => {
            expect(push).toHaveBeenCalledWith("/organizations/org-123");
        });
    });
});
