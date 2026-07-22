import { describe, expect, it } from "vitest";
import {
  aggregateOutcomeHealth,
  outcomeHealthFromRunStatus,
  outcomeHealthLabel,
} from "@/lib/outcomeHealth";

describe("outcome health", () => {
  it("uses the canonical user-facing vocabulary", () => {
    expect(outcomeHealthLabel("healthy")).toBe("Healthy");
    expect(outcomeHealthLabel("completed")).toBe("Completed");
    expect(outcomeHealthLabel("blocked")).toBe("Blocked");
  });

  it("normalizes runtime status without leaking source-specific labels", () => {
    expect(outcomeHealthFromRunStatus(" output_ready ")).toBe("completed");
    expect(outcomeHealthFromRunStatus("FAILED")).toBe("blocked");
    expect(outcomeHealthFromRunStatus("needs_attention")).toBe("degraded");
  });

  it("prioritizes actionable states and archives only fully archived outcomes", () => {
    expect(aggregateOutcomeHealth(["completed", "blocked"])).toBe("blocked");
    expect(aggregateOutcomeHealth(["archived", "completed"])).toBe("completed");
    expect(aggregateOutcomeHealth(["archived", "archived"])).toBe("archived");
  });
});
