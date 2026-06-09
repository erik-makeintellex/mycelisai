import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { useTeamWorkActionHandler } from "@/components/teams/useTeamWorkActionHandler";
import type { TeamWorkItem } from "@/store/useCortexStore";
import { mockFetch } from "../setup";

const sourceItem: TeamWorkItem = {
  id: "work-source",
  title: "Build playable prototype",
  state: "running",
  ownerLabel: "Game lead",
  scopeLabel: "Delegated work",
  teamIds: ["team-game"],
  interactions: [],
  source: "durable",
};

describe("useTeamWorkActionHandler", () => {
  it("keeps returned team ask output and proof visible before the next durable poll", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        data: {
          accepted: true,
          dispatch_state: "accepted",
          work_item: {
            work_item_id: "work-game-output",
            team_id: "team-game",
            run_id: "run-game-1",
            objective: "Generate the playable prototype package",
            state: "output_ready",
            output_refs: [{
              output_id: "out-game-1",
              team_id: "team-game",
              work_item_id: "work-game-output",
              kind: "project_package",
              label: "Playable prototype",
              storage_ref: "groups/team-game/generated/prototype",
              entrypoint: "index.html",
              proof_id: "proof-output-1",
              audit_refs: ["audit-output-1"],
            }],
            proof_refs: ["proof-work-1"],
            audit_refs: ["audit-work-1"],
          },
        },
      }),
    });

    const selectTeam = vi.fn();
    const { result } = renderHook(() => useTeamWorkActionHandler(selectTeam));

    await act(async () => {
      await result.current.handleTeamAsk(sourceItem, "Create a playable code-only game.");
    });

    await waitFor(() => {
      expect(result.current.submittedTeamWorkItems).toHaveLength(1);
    });
    expect(result.current.activeWorkActionNotice).toBe(
      "Team response is ready and retained in Active Work.",
    );
    expect(result.current.submittedTeamWorkItems[0]).toMatchObject({
      id: "work-game-output",
      title: "Generate the playable prototype package",
      state: "output_ready",
      runId: "run-game-1",
      outputCount: 1,
      proofRefs: ["proof-work-1"],
      auditRefs: ["audit-work-1"],
      nextAction: "Review retained output and proof.",
    });
    expect(result.current.submittedTeamWorkItems[0]?.outputRefs?.[0]).toMatchObject({
      output_id: "out-game-1",
      label: "Playable prototype",
      entrypoint: "index.html",
      proof_id: "proof-output-1",
      audit_refs: ["audit-output-1"],
    });
    expect(mockFetch).toHaveBeenCalledWith(
      "/api/v1/teams/team-game/work/ask",
      expect.objectContaining({
        method: "POST",
      }),
    );
  });
});
