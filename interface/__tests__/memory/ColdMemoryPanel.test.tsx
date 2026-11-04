import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { mockFetch } from "../setup";
import ColdMemoryPanel from "@/components/memory/ColdMemoryPanel";
import MemoryDetailPanel from "@/components/memory/MemoryDetailPanel";

describe("ColdMemoryPanel & Trusted-Memory UI", () => {
  beforeEach(() => {
    mockFetch.mockReset();
  });

  it("renders search input and layer filter bar", () => {
    render(<ColdMemoryPanel />);

    expect(screen.getByLabelText("Search semantic memory")).toBeDefined();
    expect(screen.getByLabelText("Filter memory layer")).toBeDefined();
    expect(screen.getByRole("button", { name: "All" })).toBeDefined();
    expect(screen.getByRole("button", { name: "Governed Doctrine" })).toBeDefined();
    expect(screen.getByRole("button", { name: "Project" })).toBeDefined();
    expect(screen.getByRole("button", { name: "Team / Agent" })).toBeDefined();
    expect(screen.getByRole("button", { name: "Soma" })).toBeDefined();
    expect(screen.getByRole("button", { name: "Reflection" })).toBeDefined();
  });

  it("appends layer_filter parameter when a memory layer filter is selected", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      json: async () => ({
        results: [
          {
            id: "mem-1",
            content: "Doctrine rule on governed telemetry",
            score: 0.94,
            source: "doctrine",
            created_at: "2026-04-01T12:00:00Z",
            memory_layer: "GOVERNED_DOCTRINE",
            verification_status: "verified",
          },
        ],
        arbitration: {
          total_candidates: 1,
          returned_count: 1,
          verified_count: 1,
          has_contradictions: false,
        },
      }),
    });

    render(<ColdMemoryPanel searchQuery="telemetry" />);

    await waitFor(
      () => {
        expect(mockFetch).toHaveBeenCalledWith(
          "/api/v1/memory/search?q=telemetry",
          expect.any(Object),
        );
      },
      { timeout: 2000 },
    );

    fireEvent.click(screen.getByRole("button", { name: "Governed Doctrine" }));

    await waitFor(
      () => {
        expect(mockFetch).toHaveBeenCalledWith(
          "/api/v1/memory/search?q=telemetry&layer_filter=GOVERNED_DOCTRINE",
          expect.any(Object),
        );
      },
      { timeout: 2000 },
    );
  });

  it("renders memory_layer and verification_status badges on ResultCard and contradiction banner when arbitration.has_contradictions is true", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      json: async () => ({
        results: [
          {
            id: "mem-contradicted",
            content: "Conflicting project deployment target",
            score: 0.82,
            source: "architecture_doc",
            created_at: "2026-04-02T15:30:00Z",
            memory_layer: "PROJECT_MEMORY",
            verification_status: "contradicted",
            contradiction_halt: {
              halted: true,
              reason: "Contradicts higher-trust governed doctrine entry doc-1",
              conflicting_id: "doc-1",
              conflicting_layer: "GOVERNED_DOCTRINE",
            },
          },
        ],
        arbitration: {
          total_candidates: 2,
          returned_count: 1,
          contradicted_count: 1,
          has_contradictions: true,
        },
      }),
    });

    const onSelect = vi.fn();
    render(<ColdMemoryPanel searchQuery="deployment" onSelectResult={onSelect} />);

    await waitFor(
      () => {
        expect(screen.getByTestId("memory-contradiction-banner")).toBeDefined();
      },
      { timeout: 2000 },
    );

    expect(
      screen.getByText("Conflicting project deployment target"),
    ).toBeDefined();
    expect(screen.getByText("Contradicted")).toBeDefined();

    fireEvent.click(screen.getByText("Conflicting project deployment target"));
    expect(onSelect).toHaveBeenCalledTimes(1);
  });

  it("renders Layer, Verification, and contradiction_halt warning in MemoryDetailPanel", () => {
    render(
      <MemoryDetailPanel
        selection={{
          kind: "search",
          result: {
            id: "mem-42",
            content: "Memory record body",
            score: 0.88,
            source: "reflection_log",
            created_at: "2026-04-03T10:00:00Z",
            memory_layer: "REFLECTION_MEMORY",
            verification_status: "candidate",
            contradiction_halt: {
              halted: true,
              reason: "Halted due to conflicting governed doctrine",
            },
          },
        }}
      />,
    );

    expect(screen.getByText("Reflection")).toBeDefined();
    expect(screen.getByText("Candidate")).toBeDefined();
    expect(
      screen.getByText("Halted due to conflicting governed doctrine"),
    ).toBeDefined();
  });
});
