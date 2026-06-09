import { describe, expect, it } from "vitest";
import { outputWorkbenchItems } from "@/components/soma/OutputWorkbench";
import type { ExecutionSummaryData } from "@/store/useCortexStore";

describe("OutputWorkbench path projection", () => {
  it("treats file-like output ids as retained workspace paths", () => {
    const summary: ExecutionSummaryData = {
      outputs: [
        {
          kind: "file",
          id: "generated/workbench-review/operator-note.md",
          proof_artifact_id: "proof-operator-note",
          proof: {
            path_boundary_status: "verified",
            readback_status: "verified",
            checksum_algorithm: "sha256",
            checksum: "b94d27b9934d3e08a52e52d7da7dabfadebf1fde558f6ad0845e1274f7f9cde9",
          },
        },
      ],
    };

    expect(outputWorkbenchItems(summary)).toEqual([
      {
        text: "generated/workbench-review/operator-note.md",
        url: "/api/v1/workspace/files/view?path=generated%2Fworkbench-review%2Foperator-note.md",
        storagePath: "generated/workbench-review/operator-note.md",
        proofArtifactId: "proof-operator-note",
        proof: summary.outputs?.[0] && typeof summary.outputs[0] !== "string" ? summary.outputs[0].proof : undefined,
      },
    ]);
  });
});
