import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { StarterCard } from "@/components/organizations/createOrganizationEntryUi";
import { useCortexStore } from "@/store/useCortexStore";
import { mockFetch } from "../setup";
import type { OrganizationStarterTemplate } from "@/lib/organizations";

// RC-004G-2: "View config" on organization setup starter cards, gated by
// Advanced mode, opening the governed ConfigViewer for the starter template
// without selecting the card underneath it.

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

function jsonResponse(status: number, body: unknown) {
    return { ok: status >= 200 && status < 300, status, json: async () => body };
}

describe("StarterCard View config", () => {
    beforeEach(() => {
        mockFetch.mockReset();
        useCortexStore.setState({ advancedMode: false });
    });

    afterEach(() => {
        useCortexStore.setState({ advancedMode: false });
    });

    it("is hidden when Advanced mode is off", () => {
        render(<StarterCard template={starter({ id: "starter-x" })} active={false} onSelect={() => {}} />);
        expect(screen.queryByTestId("starter-card-starter-x-view-config")).toBeNull();
    });

    it("is visible when Advanced mode is on", () => {
        useCortexStore.setState({ advancedMode: true });
        render(<StarterCard template={starter({ id: "starter-x" })} active={false} onSelect={() => {}} />);
        expect(screen.getByTestId("starter-card-starter-x-view-config")).toBeTruthy();
    });

    it("is visible for a Custom (operator) starter too", () => {
        useCortexStore.setState({ advancedMode: true });
        render(
            <StarterCard
                template={starter({ id: "custom-acme", source: "operator" })}
                active={false}
                onSelect={() => {}}
            />,
        );
        expect(screen.getByTestId("starter-card-custom-acme-view-config")).toBeTruthy();
    });

    it("opens the ConfigViewer for the starter template path without selecting the card", async () => {
        useCortexStore.setState({ advancedMode: true });
        mockFetch.mockResolvedValue(
            jsonResponse(200, {
                ok: true,
                data: {
                    kind: "template",
                    id: "starter-x",
                    format: "yaml",
                    document: "id: starter-x\nname: Starter\n",
                    source: "file",
                    redactions: [],
                    generated_at: "2026-09-25T12:00:00Z",
                },
            }),
        );
        let selected = false;
        render(
            <StarterCard
                template={starter({ id: "starter-x" })}
                active={false}
                onSelect={() => {
                    selected = true;
                }}
            />,
        );

        fireEvent.click(screen.getByTestId("starter-card-starter-x-view-config"));

        expect(await screen.findByTestId("config-viewer")).toBeTruthy();
        expect(mockFetch).toHaveBeenCalledWith(
            "/api/v1/config/templates/starter-x?format=yaml",
            expect.any(Object),
        );
        expect(selected).toBe(false);
        expect(await screen.findByTestId("config-viewer-clone-template")).toBeTruthy();
    });

    it("clicking the card body still selects it", () => {
        useCortexStore.setState({ advancedMode: true });
        let selected = false;
        render(
            <StarterCard
                template={starter({ id: "starter-x" })}
                active={false}
                onSelect={() => {
                    selected = true;
                }}
            />,
        );

        fireEvent.click(screen.getByTestId("starter-card-starter-x"));
        expect(selected).toBe(true);
    });

    it("does not select the card when the starter is unavailable", () => {
        useCortexStore.setState({ advancedMode: true });
        let selected = false;
        render(
            <StarterCard
                template={starter({ id: "starter-x", available: false, unavailable_reason: "blocked" })}
                active={false}
                onSelect={() => {
                    selected = true;
                }}
            />,
        );

        fireEvent.click(screen.getByTestId("starter-card-starter-x"));
        expect(selected).toBe(false);
    });
});
