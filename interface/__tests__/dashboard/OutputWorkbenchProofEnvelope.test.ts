import { describe, expect, it } from "vitest";
import {
  teamOutputProjectPackages,
  teamOutputWorkbenchItems,
} from "@/components/soma/OutputWorkbench";
import type { TeamOutputRef } from "@/store/useCortexStore";

describe("OutputWorkbench proof envelopes", () => {
  it("preserves proof envelopes on retained file and package output refs", () => {
    const teamOutputs: TeamOutputRef[] = [
      {
        output_id: "out-1",
        team_id: "team-alpha",
        work_item_id: "work-1",
        kind: "file",
        label: "Launch brief",
        storage_ref: "generated/launch/brief.md",
        proof: {
          proof_id: "proof-file-1",
          path_boundary_status: "verified",
          readback_status: "verified",
        },
        created_at: "2026-05-17T18:00:00Z",
      },
      {
        output_id: "out-2",
        team_id: "team-alpha",
        work_item_id: "work-1",
        kind: "project_package",
        label: "Launch package",
        storage_ref: "generated/launch",
        entrypoint: "index.html",
        proof_ref: "proof-1",
        proof_id: "proof-package-1",
        proof: {
          proof_id: "proof-package-1",
          checksum: "b94d27b9934d3e08a52e52d7da7dabfadebf1fde558f6ad0845e1274f7f9cde9",
          checksum_algorithm: "sha256",
        },
        created_at: "2026-05-17T18:01:00Z",
      },
    ];

    expect(teamOutputWorkbenchItems(teamOutputs)).toEqual([
      expect.objectContaining({
        text: "Launch brief",
        proof: expect.objectContaining({
          proof_id: "proof-file-1",
          path_boundary_status: "verified",
          readback_status: "verified",
        }),
      }),
    ]);
    expect(teamOutputProjectPackages(teamOutputs)).toEqual([
      expect.objectContaining({
        title: "Launch package",
        proof_artifact_id: "proof-package-1",
        proof: expect.objectContaining({
          proof_id: "proof-package-1",
          checksum_algorithm: "sha256",
        }),
      }),
    ]);
  });
});
