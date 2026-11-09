import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import MemoryDetailPanel from "../MemoryDetailPanel";
import { mockFetch } from "@/__tests__/setup";

describe("MemoryDetailPanel Governance & Conflict Resolution", () => {
  it("renders promote buttons for a candidate search result and promotes to Project Memory", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      json: async () => ({ ok: true }),
    });

    render(
      <MemoryDetailPanel
        selection={{
          kind: "search",
          result: {
            id: "mem-candidate-1",
            content: "Customer onboarding requires a 30-day milestone review.",
            score: 0.91,
            source: "onboarding_session",
            created_at: "2026-04-12T10:00:00Z",
            memory_layer: "REFLECTION_MEMORY",
            verification_status: "candidate",
          },
        }}
      />,
    );

    const promoteProjectBtn = screen.getByTestId("memory-promote-project-btn");
    const promoteDoctrineBtn = screen.getByTestId("memory-promote-doctrine-btn");
    expect(promoteProjectBtn).toBeDefined();
    expect(promoteDoctrineBtn).toBeDefined();

    fireEvent.click(promoteProjectBtn);

    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith(
        "/api/v1/memory/promote",
        expect.objectContaining({
          method: "POST",
          body: JSON.stringify({
            id: "mem-candidate-1",
            target_layer: "PROJECT_MEMORY",
            verification_status: "verified",
          }),
        }),
      );
      expect(screen.getByTestId("memory-governance-notice").textContent).toContain(
        "Promoted to Project Memory (Verified)",
      );
    });
  });

  it("renders conflict resolution buttons for a contradicted/halted search result and resolves conflict", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      json: async () => ({ ok: true }),
    });

    render(
      <MemoryDetailPanel
        selection={{
          kind: "search",
          result: {
            id: "mem-conflict-1",
            content: "Deploy directly to production without staging approval.",
            score: 0.84,
            source: "agent_reflection",
            created_at: "2026-04-12T11:00:00Z",
            memory_layer: "AGENT_MEMORY",
            verification_status: "contradicted",
            contradiction_halt: {
              halted: true,
              reason: "Contradicts Governed Doctrine: release gate required before production.",
            },
          },
        }}
      />,
    );

    expect(screen.getByTestId("memory-detail-contradiction-warning")).toBeDefined();
    const resolveKeepBtn = screen.getByTestId("memory-resolve-keep-btn");
    const resolveSupersedeBtn = screen.getByTestId("memory-resolve-supersede-btn");
    expect(resolveKeepBtn).toBeDefined();
    expect(resolveSupersedeBtn).toBeDefined();

    fireEvent.click(resolveKeepBtn);

    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith(
        "/api/v1/memory/resolve-conflict",
        expect.objectContaining({
          method: "POST",
          body: JSON.stringify({
            id: "mem-conflict-1",
            resolution: "keep",
          }),
        }),
      );
      expect(screen.getByTestId("memory-governance-notice").textContent).toContain(
        "Conflict resolved: record verified for active recall",
      );
    });
  });
});
