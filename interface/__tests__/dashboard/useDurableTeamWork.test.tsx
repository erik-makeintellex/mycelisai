import { describe, expect, it } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { useDurableTeamWork } from "@/components/soma/useDurableTeamWork";
import { mockFetch } from "../setup";
import type { TeamDetailEntry } from "@/store/useCortexStore";

const team: TeamDetailEntry = {
  id: "team-alpha",
  name: "Alpha Squad",
  role: "delivery",
  type: "mission",
  mission_id: "mission-1",
  mission_intent: "Build a package",
  inputs: [],
  deliveries: [],
  agents: [],
};

describe("useDurableTeamWork", () => {
  it("loads durable team work and retained output refs", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        data: [
          {
            work_item_id: "work-1",
            team_id: "team-alpha",
            objective: "Build a launch package",
            execution_shape: "deliverable",
            state: "output_ready",
            output_refs: [
              {
                output_id: "out-1",
                team_id: "team-alpha",
                work_item_id: "work-1",
                kind: "file",
                label: "Launch brief",
                storage_ref: "generated/launch/brief.md",
              },
            ],
            updated_at: "2026-05-17T18:00:00Z",
          },
        ],
      }),
    });

    const { result } = renderHook(() => useDurableTeamWork({ teams: [team] }));

    await waitFor(() => {
      expect(result.current.status).toBe("durable");
    });
    expect(result.current.items[0]).toMatchObject({
      title: "Build a launch package",
      source: "durable",
      outputCount: 1,
    });
    expect(result.current.outputRefs).toHaveLength(1);
  });

  it("falls back to degraded roster projection when durable state is unavailable", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
    });

    const { result } = renderHook(() => useDurableTeamWork({ teams: [team] }));

    await waitFor(() => {
      expect(result.current.status).toBe("degraded");
    });
    expect(result.current.items[0]).toMatchObject({
      title: "Alpha Squad",
      state: "degraded",
      source: "projection",
      scopeLabel: "Projection fallback",
    });
    expect(result.current.degradedMessage).toContain("durable TeamWorkItem API was unavailable");
  });
});
