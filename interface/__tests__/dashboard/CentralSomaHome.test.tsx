import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { useCortexStore } from "@/store/useCortexStore";
vi.mock("@/lib/lastOrganization", () => ({
    readLastOrganization: () => ({ id: "org-1", name: "Northstar Labs" }),
    subscribeLastOrganizationChange: () => () => undefined,
}));

vi.mock("@/components/soma/SomaOperatingSurface", () => ({
    __esModule: true,
    SomaOperatingSurface: ({
        focusedTeamId,
        organizationId,
        organizationName,
    }: {
        focusedTeamId?: string | null;
        organizationId?: string | null;
        organizationName?: string | null;
    }) => (
        <div
            data-testid="soma-operating-surface"
            data-focused-team-id={focusedTeamId ?? ""}
            data-organization-id={organizationId ?? ""}
        >
            <h2>Talk to Soma</h2>
            <p>Ready for your first request</p>
            <p>Evidence of Soma's work</p>
            <p>{organizationName ?? "No organization"}</p>
        </div>
    ),
}));

import CentralSomaHome, { resolveDashboardRequestedTeamId } from "@/components/dashboard/CentralSomaHome";

describe("CentralSomaHome", () => {
    beforeEach(() => {
        useCortexStore.setState({
            assistantName: "Soma",
            teamsDetail: [],
            streamLogs: [],
            fetchTeamsDetail: vi.fn().mockResolvedValue(undefined),
            selectTeam: vi.fn(),
            selectSignalDetail: vi.fn(),
        });
    });

    it("renders the central Soma chat as the whole front-page operating surface", async () => {
        render(<CentralSomaHome />);

        expect(screen.getByTestId("soma-operating-surface")).toBeDefined();
        expect(screen.getByText("Talk to Soma")).toBeDefined();
        expect(screen.getByText("Ready for your first request")).toBeDefined();
        expect(screen.queryByTestId("soma-environment-entry")).toBeNull();
        expect(screen.queryByText("Start here")).toBeNull();
        expect(screen.queryByRole("button", { name: "Set up an AI Organization" })).toBeNull();
        expect(screen.queryByText("Soma just did this")).toBeNull();
        expect(screen.getByText("Evidence of Soma's work")).toBeDefined();
        expect(screen.queryByText("Live team interaction stream")).toBeNull();
        expect(screen.queryByRole("link", { name: /Return to Northstar Labs/i })).toBeNull();
        expect(screen.queryByRole("link", { name: /Review Soma context model/i })).toBeNull();
        const somaSurface = screen.getByTestId("soma-operating-surface");
        expect(somaSurface.getAttribute("data-organization-id")).toBe("org-1");
        await waitFor(() => {
            expect(useCortexStore.getState().selectTeam).toHaveBeenCalledWith(null);
        });
    }, 15000);

    it("resolves requested team focus only when the team exists", () => {
        expect(resolveDashboardRequestedTeamId("team-alpha", [{ id: "team-alpha" }])).toBe("team-alpha");
        expect(resolveDashboardRequestedTeamId("missing-team", [])).toBeNull();
        expect(resolveDashboardRequestedTeamId("", [])).toBeNull();
    });

    it("clears stale focused team state when no team is requested", async () => {
        useCortexStore.setState({
            selectedTeamId: "stale-team",
            teamsDetail: [],
        });

        render(<CentralSomaHome />);

        await waitFor(() => {
            expect(useCortexStore.getState().selectTeam).toHaveBeenCalledWith(null);
        });
        expect(screen.getByTestId("soma-operating-surface").getAttribute("data-focused-team-id")).toBe("");
    });
});
