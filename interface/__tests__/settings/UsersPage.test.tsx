import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { mockFetch } from "../setup";
import UsersPage from "@/components/settings/UsersPage";

vi.mock("@/components/teams/GroupManagementPanel", () => ({
    __esModule: true,
    default: () => <div data-testid="group-management-panel">GroupManagementPanel</div>,
}));

describe("UsersPage", () => {
    beforeEach(() => {
        mockFetch.mockReset();
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({ ok: true, data: { id: "me-1", email: "me@local", role: "owner", name: "Current User" } }),
        });
    });

    it("renders user management and group management sections", async () => {
        render(<UsersPage />);

        await waitFor(() => {
            expect(screen.getByTestId("users-management-panel")).toBeDefined();
        });

        expect(screen.getByTestId("users-groups-section")).toBeDefined();
        expect(screen.getByTestId("group-management-panel")).toBeDefined();
    });

    it("adds a local user from the form", async () => {
        render(<UsersPage />);

        fireEvent.change(screen.getByLabelText("Name"), { target: { value: "Alex" } });
        fireEvent.change(screen.getByLabelText("Email"), { target: { value: "alex@example.com" } });
        fireEvent.click(screen.getByTestId("users-add-button"));

        await waitFor(() => {
            expect(screen.getByText("alex@example.com")).toBeDefined();
        });
    });
});

