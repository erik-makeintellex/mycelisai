import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { useCortexStore } from "@/store/useCortexStore";

vi.mock("@/lib/lastOrganization", () => ({
    readLastOrganization: () => ({ id: "org-1", name: "Northstar Labs" }),
    subscribeLastOrganizationChange: () => () => undefined,
}));

vi.mock("@/components/dashboard/MissionControlChat", () => ({
    __esModule: true,
    default: () => <div data-testid="mission-control-chat">Mission Chat</div>,
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
        expect(screen.getByTestId("mission-control-chat")).toBeDefined();
        expect(screen.getByText("Live team interaction stream")).toBeDefined();
        expect(screen.getByRole("link", { name: /Open groups workspace/i }).getAttribute("href")).toBe("/groups");
    }, 15000);
});
