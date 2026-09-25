import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { mockFetch } from "../setup";

vi.mock("reactflow", async () => {
    const mock = await import("../mocks/reactflow");
    return mock;
});

import StatusDrawer from "@/components/dashboard/StatusDrawer";
import { buildMissionChatFailure } from "@/lib/missionChatFailure";
import { useCortexStore } from "@/store/useCortexStore";

function resetStore() {
    useCortexStore.setState({
        isStatusDrawerOpen: false,
        missionChat: [],
        missionChatError: null,
        councilTarget: "admin",
        councilMembers: [],
        isStreamConnected: true,
        streamConnectionState: "online",
        activeBrain: null,
        governanceMode: "passive",
        missions: [],
        fetchCouncilMembers: vi.fn().mockResolvedValue(undefined),
        fetchMissions: vi.fn().mockResolvedValue(undefined),
        missionChatFailure: null,
        servicesStatus: [],
        servicesStatusFailure: null,
        servicesStatusUpdatedAt: null,
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
            missionChatFailure: buildMissionChatFailure({
                assistantName: "Soma",
                targetId: "council-sentry",
                message: "Council timeout",
            }),
            councilTarget: "council-sentry",
            councilMembers: [{ id: "council-sentry", role: "sentry", team: "council" }],
            isStreamConnected: false,
            streamConnectionState: "connecting",
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
            expect(screen.getByText("Last workspace failure: council-sentry (timeout)")).toBeDefined();
            expect(screen.getByText("SSE Stream")).toBeDefined();
            expect(screen.getByText("connecting")).toBeDefined();
            expect(screen.getByText("Active Missions")).toBeDefined();
        });
    });

    it("closes when close button is clicked", () => {
        useCortexStore.setState({ isStatusDrawerOpen: true });
        render(<StatusDrawer />);

        fireEvent.click(screen.getByLabelText("Close status drawer"));
        expect(useCortexStore.getState().isStatusDrawerOpen).toBe(false);
    });

    it("leads with the Core API root cause and marks dependents as not checked when Core is unreachable", async () => {
        useCortexStore.setState({ isStatusDrawerOpen: true });
        mockFetch.mockRejectedValue(new TypeError("Failed to fetch"));

        render(<StatusDrawer />);

        const rootCause = await screen.findByTestId("status-drawer-root-cause");
        expect(rootCause.textContent).toMatch(/Core API unreachable/i);
        expect(rootCause.textContent).toMatch(/uv run inv lifecycle\.up --frontend/);
        expect(screen.getAllByText("not checked (Core unreachable)").length).toBeGreaterThanOrEqual(5);
        expect(screen.queryByText("unknown")).toBeNull();
        expect(screen.queryByText("0 members")).toBeNull();

        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({ data: [{ name: "nats", status: "online" }] }),
        });
        fireEvent.click(screen.getByRole("button", { name: /Retry/i }));
        await waitFor(() => {
            expect(screen.queryByTestId("status-drawer-root-cause")).toBeNull();
        });
        expect(screen.getByText("online")).toBeDefined();
    });
});
