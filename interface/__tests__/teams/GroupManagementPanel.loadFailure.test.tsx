import { describe, it, expect, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import GroupManagementPanel from "@/components/teams/GroupManagementPanel";
import { useCortexStore } from "@/store/useCortexStore";
import { mockFetch } from "../setup";

function jsonResponse(status: number, body: unknown) {
    return { ok: status >= 200 && status < 300, status, json: async () => body };
}

describe("GroupManagementPanel load failure", () => {
    beforeEach(() => {
        mockFetch.mockReset();
        window.localStorage.clear();
        useCortexStore.setState({ servicesStatusFailure: null, servicesStatusUpdatedAt: null });
    });

    it("replaces the empty state with a visible error and Retry when groups fail to load", async () => {
        mockFetch.mockResolvedValue(jsonResponse(500, {}));

        render(<GroupManagementPanel />);

        const failure = await screen.findByTestId("groups-load-failure");
        expect(failure.textContent).toMatch(/Could not load groups/);
        expect(failure.textContent).toMatch(/HTTP 500/);
        expect(screen.queryByText(/Nothing here yet/i)).toBeNull();

        mockFetch.mockImplementation(async (url: string) =>
            String(url).includes("/monitor")
                ? jsonResponse(200, { ok: true, data: { status: "online" } })
                : jsonResponse(200, { ok: true, data: [] }),
        );
        fireEvent.click(screen.getByRole("button", { name: /Retry/i }));

        await waitFor(() => expect(screen.queryByTestId("groups-load-failure")).toBeNull());
    });

    it("names the Core API root cause when Core is unreachable", async () => {
        useCortexStore.setState({
            servicesStatusFailure: { kind: "unreachable", at: new Date().toISOString() },
        });
        mockFetch.mockRejectedValue(new TypeError("Failed to fetch"));

        render(<GroupManagementPanel />);

        const failure = await screen.findByTestId("groups-load-failure");
        expect(failure.textContent).toMatch(/Core API unreachable/);
        expect(failure.textContent).toMatch(/uv run inv lifecycle\.up --frontend/);
    });
});
