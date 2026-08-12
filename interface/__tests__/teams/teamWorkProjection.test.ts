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
		  execution_mode: "team_async",
		  work_intent: {
			kind: "service",
			lifecycle: {
			  stop_action: "stop_service",
			  retry_action: "restart_service",
			  recovery_action: "inspect_and_restart",
			},
		  },
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
      outcomeHealth: "completed",
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
	expect(item?.advanced?.executionMode).toEqual(["team_async"]);
	expect(item?.advanced?.lifecycleControls).toEqual(["stop_service", "restart_service", "inspect_and_restart"]);
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

  it("requires Soma verification before an uncertain external mutation can retry", () => {
    const item = mapDurableTeamWorkItem({
      work_item_id: "work-external-1",
      team_id: "team-alpha",
      objective: "Update the customer system",
      execution_shape: "delegated_work",
      state: "degraded",
      needs_operator: true,
      degradation_state: "external_mutation_outcome_unknown",
      recovery_options: [
        "Ask Soma to verify the external system before deciding whether any retry is safe.",
      ],
    });

    const recover = item?.interactions.find((action) => action.action === "recover");
    const steer = item?.interactions.find((action) => action.action === "steer");
    expect(recover).toMatchObject({
      label: "Retry unavailable",
      disabled: true,
      disabledReason: "Ask Soma to verify the external outcome before considering a retry.",
    });
    expect(steer).toMatchObject({
      label: "Tell Soma what you found",
      disabled: false,
    });
  });

  it("uses last event evidence when durable work refs have not caught up yet", () => {
    const item = mapDurableTeamWorkItem({
      work_item_id: "work-3",
      team_id: "team-charlie",
      objective: "Package the game build",
      execution_shape: "deliverable",
      state: "output_ready",
      updated_at: "2026-05-17T19:00:00Z",
      last_event: {
        headline: "Game package retained",
        details: "Proof and output refs arrived with the latest bus event.",
        expected_outputs: ["Playable browser package"],
        expected_proof: ["Launchable index.html smoke proof"],
        output_refs: [{
          output_id: "output-event-1",
          kind: "project_package",
          label: "Playable game",
          storage_ref: "groups/game-team/generated/playable-game",
          entrypoint: "index.html",
          proof_ref: "proof-event-1",
          proof: {
            proof_id: "proof-event-1",
            path_boundary_status: "verified",
            readback_status: "verified",
          },
        }],
        proof_refs: ["proof-event-1"],
        audit_refs: ["audit-event-1"],
      },
    });

    expect(item).toMatchObject({
      id: "work-3",
      state: "output_ready",
      outputCount: 1,
      proofRefs: ["proof-event-1"],
      auditRefs: ["audit-event-1"],
    });
    expect(item?.advanced?.expectedOutputs).toEqual(["Playable browser package"]);
    expect(item?.advanced?.expectedProof).toEqual(["Launchable index.html smoke proof"]);
    expect(item?.outputRefs?.[0]).toMatchObject({
      output_id: "output-event-1",
      label: "Playable game",
      proof_ref: "proof-event-1",
      proof: {
        proof_id: "proof-event-1",
        path_boundary_status: "verified",
      },
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
