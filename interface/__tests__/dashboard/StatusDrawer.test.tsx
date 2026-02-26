import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { mockFetch } from "../setup";

vi.mock("reactflow", async () => {
    const mock = await import("../mocks/reactflow");
    return mock;
});

import StatusDrawer from "@/components/dashboard/StatusDrawer";
import { useCortexStore } from "@/store/useCortexStore";

function resetStore() {
    useCortexStore.setState({
        isStatusDrawerOpen: false,
        missionChat: [],
        missionChatError: null,
        councilTarget: "admin",
        councilMembers: [],
        isStreamConnected: true,
        activeBrain: null,
        governanceMode: "passive",
        missions: [],
        fetchCouncilMembers: vi.fn().mockResolvedValue(undefined),
        fetchMissions: vi.fn().mockResolvedValue(undefined),
    });
}

describe("StatusDrawer", () => {
    beforeEach(() => {
        resetStore();
    });

    it("does not render when closed", () => {
        render(<StatusDrawer />);
        expect(screen.queryByRole("dialog", { name: "System status drawer" })).toBeNull();
    });

    it("renders live status details when opened", async () => {
        useCortexStore.setState({
            isStatusDrawerOpen: true,
            missionChatError: "Council timeout",
            councilTarget: "council-sentry",
            councilMembers: [{ id: "council-sentry", role: "sentry", team: "council" }],
            isStreamConnected: false,
            activeBrain: {
                provider_id: "ollama",
                provider_name: "Ollama",
                model_id: "qwen2.5",
                location: "local",
                data_boundary: "local_only",
            },
            governanceMode: "active",
            missions: [{ id: "m-1", intent: "x", status: "active", teams: 1, agents: 1 }],
        });

        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({
                data: [
                    { name: "nats", status: "online" },
                    { name: "postgres", status: "degraded" },
                ],
            }),
        });

        render(<StatusDrawer />);

        await waitFor(() => {
            expect(screen.getByRole("dialog", { name: "System status drawer" })).toBeDefined();
            expect(screen.getByText("System Status")).toBeDefined();
            expect(screen.getByText("Last council failure: council-sentry")).toBeDefined();
            expect(screen.getByText("SSE Stream")).toBeDefined();
            expect(screen.getByText("Active Missions")).toBeDefined();
        });
    });

    it("closes when close button is clicked", () => {
        useCortexStore.setState({ isStatusDrawerOpen: true });
        render(<StatusDrawer />);

        fireEvent.click(screen.getByLabelText("Close status drawer"));
        expect(useCortexStore.getState().isStatusDrawerOpen).toBe(false);
    });
});

