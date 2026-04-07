import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";

vi.mock("@/components/teams/GroupManagementPanel", () => ({
    __esModule: true,
    default: () => <div data-testid="group-management-panel">Groups Panel</div>,
}));

import GroupsPage from "@/app/(app)/groups/page";

describe("GroupsPage", () => {
    it("renders the dedicated groups workspace surface", () => {
        render(<GroupsPage />);
        expect(screen.getByTestId("group-management-panel")).toBeDefined();
    });
});
