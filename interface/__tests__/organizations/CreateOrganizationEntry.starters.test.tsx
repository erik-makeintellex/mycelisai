import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { act, fireEvent, render, screen, within } from "@testing-library/react";
import { mockFetch } from "../setup";
import CreateOrganizationEntry from "@/components/organizations/CreateOrganizationEntry";
import type { OrganizationStarterTemplate } from "@/lib/organizations";

vi.mock("next/navigation", () => ({
    useRouter: () => ({ push: vi.fn(), replace: vi.fn(), back: vi.fn(), prefetch: vi.fn() }),
}));

function jsonResponse(body: unknown, status = 200) {
    return Promise.resolve({
        ok: status >= 200 && status < 300,
        status,
        json: async () => body,
    } as Response);
}

function starter(overrides: Partial<OrganizationStarterTemplate>): OrganizationStarterTemplate {
    return {
        id: overrides.id || "starter-x",
        name: overrides.name || "Starter",
        description: overrides.description || "A starter organization.",
        organization_type: "AI Organization starter",
        team_lead_label: "Team Lead",
        advisor_count: 1,
        department_count: 1,
        specialist_count: 2,
        ai_engine_settings_summary: "Local engines",
        response_contract_summary: "Balanced",
        memory_personality_summary: "Starter defaults",
        source: "starter",
        available: true,
        ...overrides,
    };
}

const studio = starter({
    id: "starter-client-delivery-studio",
    name: "Client delivery studio",
    starter: { label: "Client delivery studio", persona: "studio", order: 10, sample_name: "Brightline Studio", sample_purpose: "Draft client work and hold it for approval." },
    delivery_defaults: { work_mode: "propose_only", approval_policy_ref: "delivery-approve-external", approval_posture: "approve_external_sends", audit: "standard", outputs: { retain: true, proof_required: true }, max_team_members: 3 },
    delivery_summary: { work_mode_label: "Propose only", approval_posture: "Nothing leaves the organization until you approve it.", output_proof: "Outputs are retained with proof.", team_shape: "Up to 3 members per team, plus Soma." },
});

const product = starter({
    id: "starter-product-delivery-team",
    name: "Product delivery team",
    starter: { label: "Product delivery team", persona: "saas_product_team", order: 20, sample_name: "Northstar Labs", sample_purpose: "Plan priorities, propose changes, and ship them with proof." },
    delivery_defaults: { work_mode: "execute_with_approval", approval_policy_ref: "delivery-approve-changes", approval_posture: "approve_changes", audit: "standard", outputs: { retain: true, proof_required: true }, max_team_members: 3 },
    delivery_summary: { work_mode_label: "Execute with approval", approval_posture: "Deploy and write actions wait for your approval.", output_proof: "Outputs are retained with proof.", team_shape: "Up to 3 members per team, plus Soma." },
});

const operations = starter({
    id: "starter-operations-desk",
    name: "Operations desk",
    starter: { label: "Operations desk", persona: "operations", order: 30, sample_name: "Skylight Works", sample_purpose: "Watch what you connect and summarize daily." },
    delivery_defaults: { work_mode: "execute_bounded", approval_policy_ref: "delivery-approve-state-change", approval_posture: "approve_high_risk", audit: "standard", outputs: { retain: true, proof_required: true }, max_team_members: 3 },
    delivery_summary: { work_mode_label: "Bounded execution", approval_posture: "Routine checks run within bounds; high-risk actions wait for your approval.", output_proof: "Outputs are retained with proof.", team_shape: "Up to 3 members per team, plus Soma." },
});

const regulated = starter({
    id: "starter-governed-enterprise-program",
    name: "Governed enterprise program",
    starter: { label: "Governed enterprise program", persona: "regulated_enterprise", order: 40, sample_name: "Meridian Compliance", sample_purpose: "Every mutation is a proposal with proof." },
    available: false,
    unavailable_reason: "approval policy \"delivery-approve-all\" is not a group in policy.yaml",
    delivery_summary: undefined,
});

function mockTemplatesAndOrganizations(templates: OrganizationStarterTemplate[] | null, templatesOk = true) {
    mockFetch.mockImplementation((url: string) => {
        if (url.includes("/api/v1/templates")) {
            return templatesOk
                ? jsonResponse({ ok: true, data: templates })
                : jsonResponse({ ok: false, error: "Starter templates are unavailable right now." }, 500);
        }
        if (url.includes("/api/v1/organizations")) {
            return jsonResponse({ ok: true, data: [] });
        }
        return jsonResponse({ ok: true, data: null });
    });
}

describe("CreateOrganizationEntry starter cards", () => {
    beforeEach(() => {
        mockFetch.mockReset();
        window.localStorage.clear();
    });

    afterEach(() => {
        vi.restoreAllMocks();
    });

    it("renders one card per company type, in the order returned by the API (backend sorts by starter.order)", async () => {
        mockTemplatesAndOrganizations([studio, product, operations, regulated]);

        await act(async () => {
            render(<CreateOrganizationEntry />);
        });
        fireEvent.click(await screen.findByRole("button", { name: "Explore Templates" }));

        const cards = await screen.findAllByTestId(/^starter-card-/);
        expect(cards.map((card) => card.getAttribute("data-testid"))).toEqual([
            "starter-card-starter-client-delivery-studio",
            "starter-card-starter-product-delivery-team",
            "starter-card-starter-operations-desk",
            "starter-card-starter-governed-enterprise-program",
        ]);
    });

    it("selects a card's template and prefills sample name and purpose", async () => {
        mockTemplatesAndOrganizations([studio, product, operations, regulated]);

        await act(async () => {
            render(<CreateOrganizationEntry />);
        });
        fireEvent.click(await screen.findByRole("button", { name: "Explore Templates" }));

        const productCard = await screen.findByTestId("starter-card-starter-product-delivery-team");
        fireEvent.click(productCard);

        expect(screen.getByDisplayValue("Northstar Labs")).toBeDefined();
        expect(screen.getByDisplayValue("Plan priorities, propose changes, and ship them with proof.")).toBeDefined();

        let capturedBody: { template_id?: string } | null = null;
        mockFetch.mockImplementationOnce((url: string, init?: RequestInit) => {
            capturedBody = JSON.parse(String(init?.body));
            return jsonResponse({ ok: true, data: { id: "org-1", name: "Northstar Labs" } });
        });
        await act(async () => {
            fireEvent.click(screen.getByRole("button", { name: "Create AI Organization" }));
            await Promise.resolve();
            await Promise.resolve();
        });
        expect(capturedBody).toEqual(expect.objectContaining({ template_id: "starter-product-delivery-team" }));
    });

    it("disables an unavailable starter and shows its reason as plain text", async () => {
        mockTemplatesAndOrganizations([studio, product, operations, regulated]);

        await act(async () => {
            render(<CreateOrganizationEntry />);
        });
        fireEvent.click(await screen.findByRole("button", { name: "Explore Templates" }));

        const regulatedCard = await screen.findByTestId("starter-card-starter-governed-enterprise-program");
        expect(regulatedCard).toHaveProperty("disabled", true);
        expect(within(regulatedCard).getByText(regulated.unavailable_reason!)).toBeDefined();

        fireEvent.click(regulatedCard);
        expect(screen.queryByDisplayValue("Meridian Compliance")).toBeNull();
    });

    it("falls back to start-empty guidance when the starter API fails", async () => {
        mockTemplatesAndOrganizations(null, false);

        await act(async () => {
            render(<CreateOrganizationEntry />);
        });
        fireEvent.click(await screen.findByRole("button", { name: "Explore Templates" }));

        expect(await screen.findByText("Starter templates are unavailable")).toBeDefined();
        expect(screen.getByRole("button", { name: "Start empty instead" })).toBeDefined();
    });

    it("shows the plain-language delivery summary for each approval posture", async () => {
        mockTemplatesAndOrganizations([studio, product, operations]);

        await act(async () => {
            render(<CreateOrganizationEntry />);
        });
        fireEvent.click(await screen.findByRole("button", { name: "Explore Templates" }));

        fireEvent.click(await screen.findByTestId("starter-card-starter-client-delivery-studio"));
        expect(screen.getByText("Nothing leaves the organization until you approve it.")).toBeDefined();
        expect(screen.getByText("Outputs are retained with proof.")).toBeDefined();
        expect(screen.getByText("Up to 3 members per team, plus Soma.")).toBeDefined();
        expect(screen.getByText("Departments start as planned; Soma proposes launching teams and you approve.")).toBeDefined();

        fireEvent.click(await screen.findByTestId("starter-card-starter-product-delivery-team"));
        expect(screen.getByText("Deploy and write actions wait for your approval.")).toBeDefined();

        fireEvent.click(await screen.findByTestId("starter-card-starter-operations-desk"));
        expect(screen.getByText("Routine checks run within bounds; high-risk actions wait for your approval.")).toBeDefined();
    });
});
