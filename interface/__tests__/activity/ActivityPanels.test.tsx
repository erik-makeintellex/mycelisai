import { describe, expect, it } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";

import { BusActivityPanel } from "@/components/activity/ActivityPanels";
import { RunEventInspector } from "@/components/activity/RunEventInspector";
import type { MissionEvent, MissionRun, StreamSignal } from "@/store/useCortexStore";

const run: MissionRun = {
  id: "run-1234567890",
  mission_id: "mission-1",
  tenant_id: "tenant-1",
  status: "completed",
  run_depth: 0,
  started_at: new Date().toISOString(),
};

describe("activity compression", () => {
  it("shows run event summaries first and keeps raw payloads behind inspect", () => {
    const events: MissionEvent[] = [
      {
        id: "event-1",
        run_id: run.id,
        tenant_id: "tenant-1",
        event_type: "proof.created",
        severity: "info",
        emitted_at: new Date().toISOString(),
        payload: {
          operator_summary: "Package created and proof linked.",
          internal_tool_trace: "raw diagnostic detail",
        },
      },
    ];

    render(<RunEventInspector run={run} events={events} isFetching={false} />);

    expect(screen.getByText("Package created and proof linked.")).toBeDefined();
    expect(screen.queryByText(/raw diagnostic detail/i)).toBeNull();

    fireEvent.click(screen.getByRole("button", { name: /inspect payload/i }));

    expect(screen.getByText(/raw diagnostic detail/i)).toBeDefined();
  });

  it("caps live signals and hides raw topics until inspect", () => {
    const signals: StreamSignal[] = Array.from({ length: 13 }, (_, index) => ({
      type: "status",
      source: "system",
      message: `Visible summary ${index + 1}`,
      timestamp: new Date().toISOString(),
      topic: index === 0 ? "swarm.raw.secret" : `swarm.raw.${index}`,
    }));

    render(
      <BusActivityPanel
        signals={signals}
        summary={{ status: 13, output: 0, tools: 0, governance: 0, error: 0 }}
        natsStatus="online"
        groupsStatus="online"
      />,
    );

    expect(screen.getByText("Operational activity")).toBeDefined();
    expect(screen.getByText("Showing latest 12 of 13")).toBeDefined();
    expect(screen.queryByText("Visible summary 13")).toBeNull();
    expect(screen.queryByText("swarm.raw.secret")).toBeNull();

    fireEvent.click(screen.getAllByText("Inspect signal")[0]);

    expect(screen.getByText("swarm.raw.secret")).toBeDefined();
  });
});
