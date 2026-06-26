import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";

vi.mock("@/components/dashboard/CentralSomaHome", () => ({
    __esModule: true,
    default: ({ requestedTeamIdPromise }: { requestedTeamIdPromise?: Promise<{ team_id?: string | string[] }> }) => (
        <div data-testid="central-soma-home" data-has-team-promise={requestedTeamIdPromise ? "yes" : "no"}>
            Central Soma Home
        </div>
    ),
}));

import DashboardPage from "@/app/(app)/dashboard/page";

describe("DashboardPage", () => {
    it("keeps the admin home centered on Soma without secondary setup chrome", () => {
        render(<DashboardPage />);

        expect(screen.getByTestId("central-soma-home")).toBeDefined();
        expect(screen.getByTestId("central-soma-home").getAttribute("data-has-team-promise")).toBe("no");
        expect(screen.queryByText("Create or open AI Organizations")).toBeNull();
    });

    it("passes route team context through to the central Soma home", () => {
        render(<DashboardPage searchParams={Promise.resolve({ team_id: "team-alpha" })} />);
        expect(screen.getByTestId("central-soma-home").getAttribute("data-has-team-promise")).toBe("yes");
    });
});
