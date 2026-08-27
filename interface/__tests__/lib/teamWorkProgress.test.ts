import { describe, expect, it } from "vitest";
import {
  progressForChatMessage,
  progressForDurableState,
  progressForThreadEvent,
} from "@/lib/teamWorkProgress";
import type { SomaThreadEvent } from "@/store/useCortexStore";

function event(
  kind: SomaThreadEvent["kind"],
  status?: string,
): SomaThreadEvent {
  return { kind, status, label: "Raw event", tone: "info" };
}

const proposal = {
  intent: "build-game",
  teams: 1,
  agents: 1,
  tools: [],
  risk_level: "low",
  confirm_token: "confirm-build-game",
  intent_proof_id: "proof-build-game",
};

describe("teamWorkProgress", () => {
  it("projects durable lifecycle truth without changing the source state", () => {
    expect(progressForDurableState("new").state).toBe("planning");
    expect(progressForDurableState("queued").state).toBe("queued");
    expect(progressForDurableState("running").state).toBe("building");
    expect(progressForDurableState("paused").state).toBe("building");
    expect(progressForDurableState("reviewing").state).toBe("validating");
    expect(progressForDurableState("output_ready").state).toBe("ready");
    expect(progressForDurableState("degraded").state).toBe("recovery");
    expect(progressForDurableState("needs_operator").state).toBe("recovery");
  });

  it("projects existing thread status with terminal kinds taking precedence", () => {
    expect(progressForThreadEvent(event("execution_update", "confirming")).state).toBe("queued");
    expect(progressForThreadEvent(event("execution_started", "running")).state).toBe("building");
    expect(progressForThreadEvent(event("execution_update", "reviewing")).state).toBe("validating");
    expect(progressForThreadEvent(event("result_ready", "running")).state).toBe("ready");
    expect(progressForThreadEvent(event("attention_required", "running")).state).toBe("recovery");
  });

  it("uses a proposal as planning only when no later thread truth exists", () => {
    expect(progressForChatMessage({
      role: "system",
      content: "Review this proposal",
      proposal,
    })?.state).toBe("planning");
    expect(progressForChatMessage({
      role: "system",
      content: "Approved",
      proposal,
      thread_event: event("execution_update", "queued"),
    })?.state).toBe("queued");
  });
});
