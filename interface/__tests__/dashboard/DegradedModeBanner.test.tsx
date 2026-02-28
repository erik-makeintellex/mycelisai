import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { mockFetch } from "../setup";

vi.mock("reactflow", async () => {
    const mock = await import("../mocks/reactflow");
    return mock;
});

import DegradedModeBanner from "@/components/dashboard/DegradedModeBanner";
import { useCortexStore } from "@/store/useCortexStore";

function resetStore() {
    useCortexStore.setState({
        missionChatError: null,
        isStreamConnected: true,
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

    it("renders in degraded mode and supports actions", async () => {
        useCortexStore.setState({
            missionChatError: "Council timeout",
            isStreamConnected: false,
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
});
