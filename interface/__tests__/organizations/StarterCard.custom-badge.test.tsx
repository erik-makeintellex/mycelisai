import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { StarterCard } from "@/components/organizations/createOrganizationEntryUi";
import type { OrganizationStarterTemplate } from "@/lib/organizations";

function starter(overrides: Partial<OrganizationStarterTemplate>): OrganizationStarterTemplate {
    return {
        id: overrides.id || "starter-x",
        name: overrides.name || "Starter",
        description: "A starter organization.",
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

describe("StarterCard Custom badge", () => {
    it("shows a Custom badge for an operator-sourced (cloned) template", () => {
        render(
            <StarterCard
                template={starter({ id: "custom-acme-studio", name: "Acme Studio", source: "operator" })}
                active={false}
                onSelect={() => {}}
            />,
        );
        expect(screen.getByTestId("starter-card-custom-acme-studio-custom-badge").textContent).toMatch(/Custom/);
    });

    it("does not show a Custom badge for a file or bundle starter", () => {
        render(
            <StarterCard
                template={starter({ id: "starter-client-delivery-studio", source: "starter" })}
                active={false}
                onSelect={() => {}}
            />,
        );
        expect(screen.queryByTestId("starter-card-starter-client-delivery-studio-custom-badge")).toBeNull();
    });
});
