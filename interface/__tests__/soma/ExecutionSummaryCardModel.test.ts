import { describe, expect, it } from "vitest";
import { auditText, degradationLines, executionSummaryHeading, trustVerdict } from "@/components/soma/ExecutionSummaryCardModel";
import type { ExecutionSummaryData } from "@/store/useCortexStore";

describe("ExecutionSummaryCardModel", () => {
  it("translates local ComfyUI outages into operator recovery language", () => {
    const summary: ExecutionSummaryData = {
      execution: {
        status: "failed",
        summary: "Soma could not complete the approved proposal.",
      },
      audit_recovery: {
        recovery_state: "failed",
        degradation: {
          what_failed: 'media engine error (HTTP 503): {"detail":"local/private ComfyUI engine unreachable at configured upstream"}',
          trusted_state: "The approval, intent proof, failed run record, and audit event remain trusted.",
          invalidated_proof: "No completed execution proof or retained output should be trusted for this attempt.",
          safe_continuation: "Review the failed run, adjust the request or runtime dependency, then retry the proposal.",
          requires_attention: true,
        },
      },
    };

    expect(auditText(summary.audit_recovery)).toContain("Start or reconnect the configured ComfyUI upstream");
    expect(degradationLines(summary.audit_recovery)).toEqual([
      "Local media generation is not reachable, so Soma could not create the image output.",
      "Still available: The approval, request, failed run record, and audit trail remain available for review.",
      "Not reliable: No completed image output or execution proof should be trusted for this attempt.",
      "Safe next: Start or reconnect the configured ComfyUI upstream, then retry. If you only need text/files, ask Soma to rerun without image generation.",
    ]);
    expect(trustVerdict(summary)).toMatchObject({
      label: "Needs review",
      detail: "Local media generation is not reachable, so Soma could not create the image output.",
      tone: "attention",
    });
  });

  it("labels proposed and failed summaries without calling them ready results", () => {
    expect(executionSummaryHeading({ execution: { status: "proposed" } })).toBe("Proposal ready");
    expect(executionSummaryHeading({ execution: { status: "failed" } })).toBe("Could not run");
    expect(executionSummaryHeading({ execution: { status: "completed" } }, 1)).toBe("Output ready");
  });
});
