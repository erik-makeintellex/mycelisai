import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { OutputWorkbench } from "@/components/soma/OutputWorkbench";
import {
  OUTPUT_CONTINUATION_EVENT,
  outputContinuationPrompt,
  type OutputContinuationDetail,
} from "@/components/soma/outputContinuation";

describe("OutputWorkbench continuation", () => {
  it("lets the operator reply to delivered outputs and packages in Soma", () => {
    const continuations: OutputContinuationDetail[] = [];
    const handleContinuation = (event: Event) => {
      continuations.push((event as CustomEvent<OutputContinuationDetail>).detail);
    };
    window.addEventListener(OUTPUT_CONTINUATION_EVENT, handleContinuation);

    render(
      <OutputWorkbench
        outputs={[{
          text: "Launch brief",
          url: "/api/v1/workspace/files/view?path=generated%2Flaunch%2Fbrief.md",
          storagePath: "generated/launch/brief.md",
          proofArtifactId: "proof-brief",
        }]}
        projectPackages={[{
          kind: "project_package",
          title: "Launch microsite",
          folder: "generated/launch",
          entrypoint: "dist/index.html",
          proof_artifact_id: "proof-package",
        }]}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: /Reply to Launch brief in Soma/i }));
    fireEvent.click(screen.getByRole("button", { name: /Reply to Launch microsite in Soma/i }));
    window.removeEventListener(OUTPUT_CONTINUATION_EVENT, handleContinuation);

    expect(continuations).toEqual([
      { title: "Launch brief", reference: "generated/launch/brief.md", proof: "proof-brief" },
      { title: "Launch microsite", reference: "generated/launch", proof: "proof-package" },
    ]);
    expect(outputContinuationPrompt(continuations[0])).toContain("update, alternate version, or follow-up generation");
  });
});
