import { describe, it, expect, vi, beforeEach } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
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
        expect(screen.getByText("Start here")).toBeDefined();
        expect(screen.getByText(/Ask Soma to plan, review, create, or execute governed work/i)).toBeDefined();
        expect(screen.getByRole("button", { name: "Set up an AI Organization" })).toBeDefined();
        expect(screen.queryByText("Soma just did this")).toBeNull();
        expect(screen.getByText("Evidence of Soma's work")).toBeDefined();
        expect(screen.getByText("Live team interaction stream")).toBeDefined();
        expect(screen.getByRole("link", { name: /Return to Northstar Labs/i }).getAttribute("href")).toBe("/organizations/org-1");
        expect(screen.queryByRole("link", { name: /Review Soma context model/i })).toBeNull();
        const somaSurface = screen.getByTestId("soma-operating-surface");
        const startHere = screen.getByText("Start here");
        expect([...container.querySelectorAll("*")].indexOf(somaSurface)).toBeLessThan(
            [...container.querySelectorAll("*")].indexOf(startHere),
        );
    }, 15000);

    it("opens the AI Organization setup details from the quick action", async () => {
        const details = document.createElement("details");
        details.id = "dashboard-organization-setup";
        document.body.appendChild(details);

        try {
            render(<CentralSomaHome />);
            await screen.findByText("erik@mycelis.link");

            const trigger = screen.getByRole("button", { name: "Set up an AI Organization" });
            fireEvent.click(trigger);

            expect(details.open).toBe(true);
            expect(window.location.hash).toBe("#dashboard-organization-setup");
        } finally {
            details.remove();
            window.history.replaceState(null, "", window.location.pathname);
        }
    });
});
