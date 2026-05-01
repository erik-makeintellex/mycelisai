import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import ActivityPage from "@/app/(app)/activity/page";
import { useCortexStore } from "@/store/useCortexStore";

describe("ActivityPage", () => {
  beforeEach(() => {
    useCortexStore.setState({
      recentRuns: [
        {
          id: "run-active",
          mission_id: "mission-1",
          tenant_id: "tenant-1",
          status: "running",
          run_depth: 0,
          started_at: new Date(Date.now() - 30_000).toISOString(),
        },
        {
          id: "run-complete",
          mission_id: "mission-2",
          tenant_id: "tenant-1",
          status: "completed",
          run_depth: 0,
          started_at: new Date(Date.now() - 90_000).toISOString(),
        },
      ],
      isFetchingRuns: false,
      fetchRecentRuns: vi.fn().mockResolvedValue(undefined),
      runTimeline: [
        {
          id: "event-1",
          run_id: "run-active",
          tenant_id: "tenant-1",
          event_type: "run.started",
          severity: "info",
          source_agent: "Soma",
          payload: { summary: "Started workflow review" },
          emitted_at: new Date().toISOString(),
        },
      ],
      isFetchingTimeline: false,
      fetchRunTimeline: vi.fn().mockResolvedValue(undefined),
      streamLogs: [
        {
          type: "tool.completed",
          source: "Soma",
          message: "Generated delivery summary",
          topic: "swarm.team.soma.signal.result",
          timestamp: new Date().toISOString(),
        },
        {
          type: "approval.required",
          source_kind: "workspace_ui",
          message: "High-risk action needs review",
          timestamp: new Date().toISOString(),
        },
      ],
      streamConnectionState: "online",
      initializeStream: vi.fn(),
      servicesStatus: [
        { name: "nats", status: "online" },
        { name: "groups_bus", status: "online" },
      ],
      fetchServicesStatus: vi.fn().mockResolvedValue([]),
    });
  });

  it("renders a clean admin workflow and bus overview", async () => {
    render(<ActivityPage />);

    expect(
      screen.getByRole("heading", { name: "Workflow and bus review" }),
    ).toBeDefined();
    expect(screen.getByText("Active workflows")).toBeDefined();
    expect(screen.getByText("Recent runs")).toBeDefined();
    expect(screen.getByText("Live signals")).toBeDefined();
    expect(screen.getByText("Bus state")).toBeDefined();
    expect(
      screen
        .getByRole("link", { name: "Active workflows: 1" })
        .getAttribute("href"),
    ).toBe("/runs?status=running");
    expect(
      screen.getByRole("link", { name: "Recent runs: 2" }).getAttribute("href"),
    ).toBe("/runs");
    expect(
      screen
        .getByRole("link", { name: "Live signals: 2" })
        .getAttribute("href"),
    ).toBe("/activity#message-bus");
    expect(
      screen
        .getByRole("link", { name: "Bus state: online" })
        .getAttribute("href"),
    ).toBe("/system?tab=nats");
    expect(screen.getAllByText("run-active").length).toBeGreaterThan(0);
    expect(screen.getByText("Run events")).toBeDefined();
    await waitFor(() => expect(screen.getByText("run.started")).toBeDefined());
    expect(screen.getByText(/Started workflow review/)).toBeDefined();
    expect(screen.getByText("Generated delivery summary")).toBeDefined();
    expect(screen.getByText(/NATS online/)).toBeDefined();
    expect(
      screen.getByRole("link", { name: /System/i }).getAttribute("href"),
    ).toBe("/system?tab=nats");
  });
});
