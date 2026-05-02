import { describe, it, expect, vi, beforeEach } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import { useCortexStore } from "@/store/useCortexStore";

vi.mock("@/lib/lastOrganization", () => ({
    readLastOrganization: () => ({ id: "org-1", name: "Northstar Labs" }),
    subscribeLastOrganizationChange: () => () => undefined,
}));

vi.mock("@/components/dashboard/MissionControlChat", () => ({
    __esModule: true,
    default: () => <div data-testid="mission-control-chat">Mission Chat</div>,
}));

vi.mock("@/components/dashboard/SomaReadinessStrip", () => ({
    SomaReadinessStrip: () => <div data-testid="soma-readiness-strip">Readiness</div>,
}));

import CentralSomaHome from "@/components/dashboard/CentralSomaHome";

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

    it("renders the central Soma chat and the live interaction stream on the front page", () => {
        render(<CentralSomaHome />);

        expect(screen.getByText("Work directly with Soma from the admin home.")).toBeDefined();
        expect(screen.getByText("Say this to Soma")).toBeDefined();
        expect(screen.getAllByText("Web search").length).toBeGreaterThan(0);
        expect(screen.getByText("Create a team")).toBeDefined();
        expect(screen.getByText("Private data")).toBeDefined();
        expect(screen.getByRole("link", { name: /Manage tools/i }).getAttribute("href")).toBe("/resources?tab=tools");
        expect(screen.getByTestId("mission-control-chat")).toBeDefined();
        expect(screen.getByTestId("central-soma-chat-frame").className).toContain("h-[72vh]");
        expect(screen.getByTestId("central-soma-chat-frame").className).toContain("overflow-hidden");
        expect(screen.getByText("Live team interaction stream")).toBeDefined();
        expect(screen.getByRole("link", { name: /Open groups workspace/i }).getAttribute("href")).toBe("/groups");
        expect(screen.getByRole("link", { name: /Review workflow activity/i }).getAttribute("href")).toBe("/activity");
        expect(screen.getByRole("link", { name: /Approval queue/i }).getAttribute("href")).toBe("/approvals");
        expect(screen.getByRole("link", { name: /Tool readiness/i }).getAttribute("href")).toBe("/resources?tab=tools");
        expect(screen.getByRole("link", { name: /Return to Northstar Labs/i }).getAttribute("href")).toBe("/organizations/org-1");
        expect(screen.queryByRole("link", { name: /Review Soma context model/i })).toBeNull();
    }, 15000);

    it("opens the AI Organization setup details from the quick action", () => {
        const details = document.createElement("details");
        details.id = "dashboard-organization-setup";
        document.body.appendChild(details);

        try {
            render(<CentralSomaHome />);

            const trigger = screen.getByRole("button", { name: "Create or open AI Organizations" });
            fireEvent.click(trigger);

            expect(details.open).toBe(true);
            expect(window.location.hash).toBe("#dashboard-organization-setup");
        } finally {
            details.remove();
            window.history.replaceState(null, "", window.location.pathname);
        }
    });
});
