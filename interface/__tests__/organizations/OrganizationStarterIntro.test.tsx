import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { OrganizationStarterIntro } from "@/components/organizations/OrganizationStarterIntro";
import type { OrganizationDeliveryDefaults } from "@/lib/organizations";

function deliveryDefaults(overrides: Partial<OrganizationDeliveryDefaults>): OrganizationDeliveryDefaults {
    return {
        work_mode: "execute_with_approval",
        approval_policy_ref: "default",
        approval_posture: "approve_changes",
        audit: "standard",
        outputs: { retain: true, proof_required: false },
        max_team_members: 6,
        ...overrides,
    };
}

describe("OrganizationStarterIntro", () => {
    it("renders the soma intro and posture line for a regulated starter", () => {
        render(
            <OrganizationStarterIntro
                somaIntro="I keep every decision reviewable before it ships."
                deliveryDefaults={deliveryDefaults({
                    approval_posture: "approve_all_mutations",
                    audit: "full",
                    outputs: { retain: true, proof_required: true },
                    max_team_members: 4,
                })}
            />,
        );

        expect(screen.getByText(/I keep every decision reviewable before it ships\./)).toBeDefined();
        expect(screen.getByText(/every mutation needs your approval/)).toBeDefined();
        expect(screen.getByText(/outputs are kept with proof/)).toBeDefined();
        expect(screen.getByText(/teams top out at 4 members/)).toBeDefined();
        expect(screen.getByText(/Departments are planned/)).toBeDefined();
    });

    it("renders the soma intro and posture line for a studio starter", () => {
        render(
            <OrganizationStarterIntro
                somaIntro="I keep client work moving with a light approval touch."
                deliveryDefaults={deliveryDefaults({
                    approval_posture: "approve_external_sends",
                    audit: "light",
                    outputs: { retain: true, proof_required: false },
                    max_team_members: 8,
                })}
            />,
        );

        expect(screen.getByText(/I keep client work moving with a light approval touch\./)).toBeDefined();
        expect(screen.getByText(/nothing leaves the organization without your approval/)).toBeDefined();
        expect(screen.getByText(/outputs are kept/)).toBeDefined();
        expect(screen.getByText(/teams top out at 8 members/)).toBeDefined();
    });

    it("renders nothing when neither soma_intro nor delivery_defaults are present", () => {
        const { container } = render(<OrganizationStarterIntro />);
        expect(container.innerHTML).toBe("");
    });
});
