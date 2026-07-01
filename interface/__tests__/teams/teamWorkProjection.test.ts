import { describe, expect, it } from "vitest";
import {
  mapDurableTeamWorkItem,
  parseTeamWorkAPIItems,
  projectTeamWorkItem,
  teamOutputRefsFromItems,
} from "@/components/teams/teamWorkProjection";

describe("teamWorkProjection", () => {
  it("maps durable TeamWorkItem API records into active-work rows with output refs", () => {
    const payload = {
      data: [
        {
          work_item_id: "work-1",
          team_id: "team-alpha",
          run_id: "run-1",
          objective: "Build the launch package",
          owner: "Alpha lead",
          execution_shape: "deliverable",
          state: "output_ready",
          expected_outputs: ["reviewable package"],
          expected_proof: ["smoke proof"],
          last_event: {
            headline: "Package retained",
            details: "Open the output workbench to review it.",
            next_action: "Review package",
          },
          output_refs: [
            {
              output_id: "output-1",
              team_id: "team-alpha",
              work_item_id: "work-1",
              kind: "project_package",
              label: "Launch package",
              storage_ref: "generated/launch",
              entrypoint: "index.html",
              proof_ref: "proof-1",
              proof: {
                proof_id: "proof-envelope-1",
                path_boundary_status: "verified",
                readback_status: "verified",
                checksum: "b94d27b9934d3e08a52e52d7da7dabfadebf1fde558f6ad0845e1274f7f9cde9",
                checksum_algorithm: "sha256",
              },
            },
          ],
          proof_refs: ["proof-1"],
          audit_refs: ["audit-1"],
          target_ref: {
            type: "recovery",
            id: "recover-work-1",
            team_id: "team-alpha",
            work_item_id: "work-1",
            label: "Recovery item",
          },
          updated_at: "2026-05-17T18:00:00Z",
          version: "v1",
        },
      ],
    };

    const [raw] = parseTeamWorkAPIItems(payload);
    const item = mapDurableTeamWorkItem(raw);

    expect(item).toMatchObject({
      id: "work-1",
      title: "Build the launch package",
      state: "output_ready",
      source: "durable",
      sourceLabel: "Durable team work",
      outputCount: 1,
      nextAction: "Review package",
      targetRef: {
        type: "recovery",
        id: "recover-work-1",
        team_id: "team-alpha",
        work_item_id: "work-1",
        label: "Recovery item",
      },
    });
    expect(item?.interactions.find((action) => action.action === "inspect")?.label).toBe("Open run");
    expect(item?.interactions.find((action) => action.action === "archive")?.label).toBe("Clear from review");
    expect(item?.advanced?.expectedOutputs).toEqual(["reviewable package"]);
    expect(item?.outputRefs?.[0]?.proof).toMatchObject({
      proof_id: "proof-envelope-1",
      path_boundary_status: "verified",
      readback_status: "verified",
    });
    expect(teamOutputRefsFromItems(item ? [item] : [])).toHaveLength(1);
  });

  it("falls back to the last event target ref when the work item does not carry one", () => {
    const item = mapDurableTeamWorkItem({
      work_item_id: "work-2",
      team_id: "team-alpha",
      objective: "Review retained proof",
      execution_shape: "delegated_work",
      state: "degraded",
      last_event: {
        headline: "Proof needs review",
        target_ref: {
          type: "run",
          id: "run-from-event",
          run_id: "run-from-event",
          work_item_id: "work-2",
          label: "Run receipt",
        },
      },
    });

    expect(item?.targetRef).toMatchObject({
      type: "run",
      id: "run-from-event",
      run_id: "run-from-event",
      work_item_id: "work-2",
    });
  });

  it("marks roster-only projection as degraded and inspectable", () => {
    const item = projectTeamWorkItem({
      id: "team-bravo",
      name: "Bravo Ops",
      role: "delivery",
      type: "mission",
      mission_id: "mission-1",
      mission_intent: "Create a demo",
      inputs: ["internal.command"],
      deliveries: ["signal.result"],
      agents: [],
    });

    expect(item.state).toBe("degraded");
    expect(item.source).toBe("projection");
    expect(item.scopeLabel).toBe("Projection fallback");
    expect(item.fallbackReason).toContain("Durable TeamWorkItem records were unavailable");
    expect(item.interactions.find((action) => action.action === "inspect")?.href).toBe("/teams?view=work");
    expect(item.interactions.find((action) => action.action === "archive")?.label).toBe("Clear from review");
  });
});
