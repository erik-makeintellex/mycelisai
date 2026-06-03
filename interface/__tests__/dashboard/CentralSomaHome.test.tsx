import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { useCortexStore } from "@/store/useCortexStore";
import { mockFetch } from "../setup";

vi.mock("@/lib/lastOrganization", () => ({
    readLastOrganization: () => ({ id: "org-1", name: "Northstar Labs" }),
    subscribeLastOrganizationChange: () => () => undefined,
}));

vi.mock("@/components/soma/SomaOperatingSurface", () => ({
    __esModule: true,
    SomaOperatingSurface: ({ organizationName }: { organizationName?: string | null }) => (
        <div data-testid="soma-operating-surface">
            <h2>What do you want Soma to do?</h2>
            <p>Ready for your first request</p>
            <p>Evidence of Soma's work</p>
            <p>{organizationName ?? "No organization"}</p>
        </div>
    ),
}));

import CentralSomaHome from "@/components/dashboard/CentralSomaHome";

describe("CentralSomaHome", () => {
    beforeEach(() => {
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({
                ok: true,
                data: {
                    authenticated: true,
                    user: {
                        email: "erik@mycelis.link",
                        name: "Erik",
                        role: "admin",
                        provider: "google",
                        hd: "mycelis.link",
                    },
                },
            }),
        });
        useCortexStore.setState({
            assistantName: "Soma",
            teamsDetail: [],
            streamLogs: [],
            fetchTeamsDetail: vi.fn().mockResolvedValue(undefined),
            selectTeam: vi.fn(),
            selectSignalDetail: vi.fn(),
        });
    });

    it("renders the central Soma chat and signed-in operating environment on the front page", async () => {
        const { container } = render(<CentralSomaHome />);

        expect(screen.getByTestId("soma-operating-surface")).toBeDefined();
        expect(screen.getByText("What do you want Soma to do?")).toBeDefined();
        expect(screen.getByText("Ready for your first request")).toBeDefined();
        expect(screen.getByTestId("soma-environment-entry")).toBeDefined();
        expect(await screen.findByText("erik@mycelis.link")).toBeDefined();
        expect(screen.getByText("Google Workspace")).toBeDefined();
        expect(screen.getByText("mycelis.link")).toBeDefined();
        expect(screen.queryByText("Start here")).toBeNull();
        expect(screen.queryByRole("button", { name: "Set up an AI Organization" })).toBeNull();
        expect(screen.queryByText("Soma just did this")).toBeNull();
        expect(screen.getByText("Evidence of Soma's work")).toBeDefined();
        expect(screen.queryByText("Live team interaction stream")).toBeNull();
        expect(screen.queryByRole("link", { name: /Return to Northstar Labs/i })).toBeNull();
        expect(screen.queryByRole("link", { name: /Review Soma context model/i })).toBeNull();
        const somaSurface = screen.getByTestId("soma-operating-surface");
        const environment = screen.getByTestId("soma-environment-entry");
        expect([...container.querySelectorAll("*")].indexOf(somaSurface)).toBeLessThan(
            [...container.querySelectorAll("*")].indexOf(environment),
        );
    }, 15000);
});
