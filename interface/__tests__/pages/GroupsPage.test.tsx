import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";

const mocks = vi.hoisted(() => ({
    groupPanelMock: vi.fn(() => <div data-testid="group-management-panel">Groups Panel</div>),
    advancedMode: vi.fn(() => true),
    toggleAdvancedMode: vi.fn(),
}));

vi.mock("@/components/teams/GroupManagementPanel", () => ({
    __esModule: true,
    default: mocks.groupPanelMock,
}));

vi.mock("@/store/useCortexStore", () => ({
    useCortexStore: (selector: any) =>
        selector({
            advancedMode: mocks.advancedMode(),
            toggleAdvancedMode: mocks.toggleAdvancedMode,
        }),
}));

import GroupsPage from "@/app/(app)/groups/page";

describe("GroupsPage", () => {
    beforeEach(() => {
        mocks.groupPanelMock.mockClear();
        mocks.advancedMode.mockReturnValue(true);
        mocks.toggleAdvancedMode.mockReset();
    });

    it("shows the advanced gate when advanced mode is off", async () => {
        mocks.advancedMode.mockReturnValue(false);

        render(await GroupsPage({}));

        expect(screen.getByText("Groups are an Advanced coordination view")).toBeDefined();
        expect(screen.getByText(/standing teams, broadcasts, and cross-team coordination/i)).toBeDefined();
        expect(screen.queryByTestId("group-management-panel")).toBeNull();
    });

    it("renders the dedicated groups workspace surface", async () => {
        render(await GroupsPage({}));
        expect(screen.getByTestId("group-management-panel")).toBeDefined();
    });

    it("passes the requested group id into the groups workspace", async () => {
        render(await GroupsPage({ searchParams: Promise.resolve({ group_id: "group-temp" }) }));
        expect(mocks.groupPanelMock).toHaveBeenCalledWith(
            expect.objectContaining({ initialSelectedGroupId: "group-temp" }),
            undefined,
        );
    });
});
