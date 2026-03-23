import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { mockFetch } from "../setup";

vi.mock("reactflow", async () => {
    const mock = await import("../mocks/reactflow");
    return mock;
});

import DegradedModeBanner from "@/components/dashboard/DegradedModeBanner";
import { buildMissionChatFailure } from "@/lib/missionChatFailure";
import { useCortexStore } from "@/store/useCortexStore";

function resetStore() {
    useCortexStore.setState({
        missionChatError: null,
        isStreamConnected: true,
        streamConnectionState: "online",
        assistantName: "Soma",
        councilTarget: "council-sentry",
        isStatusDrawerOpen: false,
    });
}

describe("DegradedModeBanner", () => {
    beforeEach(() => {
        resetStore();
    });

    it("does not render when all health checks are healthy", async () => {
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({
                data: [
                    { name: "nats", status: "online" },
                    { name: "postgres", status: "online" },
                ],
            }),
        });

        render(<DegradedModeBanner />);

        await waitFor(() => {
            expect(screen.queryByText(/System in Degraded Mode/i)).toBeNull();
        });
    });

    it("does not treat a connecting stream as degraded on startup", async () => {
        useCortexStore.setState({
            isStreamConnected: false,
            streamConnectionState: "connecting",
            councilTarget: "admin",
        });

        render(<DegradedModeBanner />);

        await waitFor(() => {
            expect(screen.queryByText(/System in Degraded Mode/i)).toBeNull();
        });
    });

    it("renders in degraded mode and supports actions", async () => {
        useCortexStore.setState({
            missionChatError: "Council timeout",
            missionChatFailure: buildMissionChatFailure({
                assistantName: "Soma",
                targetId: "council-architect",
                message: "Council timeout",
            }),
            isStreamConnected: false,
            streamConnectionState: "offline",
            councilTarget: "council-architect",
            disconnectStream: vi.fn(),
            initializeStream: vi.fn(),
        });

        mockFetch
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({ data: [{ name: "nats", status: "offline" }] }),
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({ ok: true, data: [] }),
            })
            .mockResolvedValue({
                ok: true,
                json: async () => ({ data: [{ name: "nats", status: "offline" }] }),
            });

        render(<DegradedModeBanner />);

        await waitFor(() => {
            expect(screen.getByText(/System in Degraded Mode/i)).toBeDefined();
        });
        expect(screen.getByText(/Core functionality is still available/i)).toBeDefined();

        fireEvent.click(screen.getByText("Open Status"));
        expect(useCortexStore.getState().isStatusDrawerOpen).toBe(true);

        fireEvent.click(screen.getByText("Switch to Soma"));
        expect(useCortexStore.getState().councilTarget).toBe("admin");

        const callsBeforeRetry = mockFetch.mock.calls.length;
        fireEvent.click(screen.getByText("Retry"));
        await waitFor(() => {
            expect(mockFetch.mock.calls.length).toBeGreaterThan(callsBeforeRetry);
        });
        expect(useCortexStore.getState().disconnectStream).toHaveBeenCalled();
        expect(useCortexStore.getState().initializeStream).toHaveBeenCalledWith(true);
    });

    it("does not offer a redundant Soma switch when Soma is already the active route", async () => {
        useCortexStore.setState({
            isStreamConnected: false,
            streamConnectionState: "offline",
            councilTarget: "admin",
        });

        render(<DegradedModeBanner />);

        await waitFor(() => {
            expect(screen.getByText(/System in Degraded Mode/i)).toBeDefined();
        });
        expect(screen.queryByText("Switch to Soma")).toBeNull();
    });

    it("shows a specific workspace chat server-error reason", async () => {
        useCortexStore.setState({
            missionChatError: "Council agent unreachable (500)",
            missionChatFailure: buildMissionChatFailure({
                assistantName: "Soma",
                targetId: "admin",
                message: "Council agent unreachable (500)",
                statusCode: 500,
            }),
            isStreamConnected: true,
            streamConnectionState: "online",
        });

        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({
                data: [
                    { name: "nats", status: "online" },
                    { name: "postgres", status: "online" },
                ],
            }),
        });

        render(<DegradedModeBanner />);

        await waitFor(() => {
            expect(screen.getByText(/Workspace chat server error/i)).toBeDefined();
        });
    });
});
