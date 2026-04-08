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

vi.mock("@/components/organizations/CreateOrganizationEntry", () => ({
    __esModule: true,
    default: () => <div data-testid="create-organization-entry">Create Organization Entry</div>,
}));

import DashboardPage from "@/app/(app)/dashboard/page";

describe("DashboardPage", () => {
    it("keeps the admin home centered on Soma and demotes organization setup", () => {
        render(<DashboardPage />);

        expect(screen.getByTestId("central-soma-home")).toBeDefined();
        expect(screen.getByTestId("central-soma-home").getAttribute("data-has-team-promise")).toBe("no");
        expect(screen.getByText("Create or open AI Organizations")).toBeDefined();
        expect(screen.getByText(/Keep the main admin home centered on Soma/i)).toBeDefined();
        expect(screen.getByTestId("create-organization-entry")).toBeDefined();
    });

    it("renders organization setup inside a secondary details section", () => {
        const { container } = render(<DashboardPage />);
        const details = container.querySelector("details#dashboard-organization-setup");
        expect(details).not.toBeNull();
        expect(details?.querySelector("[data-testid='create-organization-entry']")).not.toBeNull();
    });

    it("passes route team context through to the central Soma home", () => {
        render(<DashboardPage searchParams={Promise.resolve({ team_id: "team-alpha" })} />);
        expect(screen.getByTestId("central-soma-home").getAttribute("data-has-team-promise")).toBe("yes");
    });
});
