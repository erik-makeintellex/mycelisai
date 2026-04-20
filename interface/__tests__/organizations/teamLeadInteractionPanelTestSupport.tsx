import { render } from "@testing-library/react";
import TeamLeadInteractionPanel from "@/components/organizations/TeamLeadInteractionPanel";

export const teamLeadInteractionPanelDefaults = {
    organizationId: "org-123",
    organizationName: "Northstar Labs",
    somaName: "Soma for Northstar Labs",
    teamLeadName: "Team Lead for Northstar Labs",
};

export function jsonResponse(body: unknown, status = 200) {
    return Promise.resolve(new Response(JSON.stringify(body), { status }));
}

export function renderTeamLeadInteractionPanel(
    overrides: Partial<typeof teamLeadInteractionPanelDefaults> & {
        promptSuggestions?: Array<{ label: string; prompt: string }>;
    } = {},
) {
    return render(
        <TeamLeadInteractionPanel
            {...teamLeadInteractionPanelDefaults}
            {...overrides}
        />,
    );
}
