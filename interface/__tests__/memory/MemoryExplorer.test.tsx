import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import type { Artifact } from "@/store/cortexStoreTypesPlanning";
import type { SearchResult } from "@/components/memory/memorySelection";

type MemoryExplorerStore = {
  advancedMode: boolean;
  getArtifactDetail: ReturnType<typeof vi.fn>;
  selectedArtifactDetail: Artifact | null;
};

// Mock Zustand store
const mockAdvancedMode = vi.fn(() => false);
vi.mock("@/store/useCortexStore", () => ({
  useCortexStore: (selector: (state: MemoryExplorerStore) => unknown) => {
    const state = {
      advancedMode: mockAdvancedMode(),
      getArtifactDetail: vi.fn().mockResolvedValue(undefined),
      selectedArtifactDetail: null,
    };
    return selector(state);
  },
}));

// Mock the three sub-panels to isolate MemoryExplorer layout tests
vi.mock("@/components/memory/HotMemoryPanel", () => ({
  __esModule: true,
  default: () => <div data-testid="hot-panel">Hot Panel</div>,
}));

vi.mock("@/components/memory/WarmMemoryPanel", () => ({
  __esModule: true,
  default: ({
    onSearchRelated,
    onSelectArtifact,
  }: {
    onSearchRelated: (query: string) => void;
    onSelectArtifact: (artifact: Artifact) => void;
  }) => (
    <div data-testid="warm-panel">
      Warm Panel
      <button
        data-testid="warm-search-btn"
        onClick={() => onSearchRelated("test query")}
      >
        Search Related
      </button>
      <button
        data-testid="artifact-select-btn"
        onClick={() =>
          onSelectArtifact({
            id: "artifact-1",
            agent_id: "Soma",
            artifact_type: "document",
            title: "Launch Plan",
            content_type: "text/markdown",
            content: "# Launch Plan",
            metadata: { audience: "admins" },
            status: "approved",
            created_at: "2026-01-01T00:00:00.000Z",
          })
        }
      >
        Select Artifact
      </button>
    </div>
  ),
}));

vi.mock("@/components/memory/ColdMemoryPanel", () => ({
  __esModule: true,
  default: ({
    searchQuery,
    onSelectResult,
  }: {
    searchQuery?: string;
    onSelectResult: (result: SearchResult) => void;
  }) => (
    <div data-testid="cold-panel">
      Cold Panel
      {searchQuery && <span data-testid="cold-query">{searchQuery}</span>}
      <button
        data-testid="search-result-select-btn"
        onClick={() =>
          onSelectResult({
            id: "memory-1",
            content: "Full memory content for inspection.",
            similarity: 0.91,
            source: "conversation",
            created_at: "2026-01-02T00:00:00.000Z",
          })
        }
      >
        Select Search Result
      </button>
    </div>
  ),
}));

import MemoryExplorer from "@/components/memory/MemoryExplorer";

describe("MemoryExplorer", () => {
  beforeEach(() => {
    mockAdvancedMode.mockReturnValue(false);
    window.history.replaceState(null, "", "/memory");
  });

  it("renders the Memory header and focused view tabs", () => {
    render(<MemoryExplorer />);

    // Header displays "Memory"
    expect(screen.getByText("Memory")).toBeDefined();

    // Focused tab layout avoids compressed columns.
    expect(screen.getByText("Recent Work")).toBeDefined();
    expect(screen.getByText("Search Memory")).toBeDefined();
    expect(screen.getByText("Details")).toBeDefined();

    // Warm panel is the default view.
    expect(screen.getByTestId("warm-panel")).toBeDefined();
    expect(screen.queryByTestId("cold-panel")).toBeNull();
  });

  it("does not show Signal Stream when advancedMode is off", () => {
    mockAdvancedMode.mockReturnValue(false);
    render(<MemoryExplorer />);

    expect(screen.queryByText("Signal Stream")).toBeNull();
    expect(screen.queryByTestId("hot-panel")).toBeNull();
  });

  it("shows Signal Stream toggle when advancedMode is on", () => {
    mockAdvancedMode.mockReturnValue(true);
    render(<MemoryExplorer />);

    expect(screen.getByText("Signal Stream")).toBeDefined();
  });

  it("passes search query from WarmMemoryPanel to ColdMemoryPanel", async () => {
    const { fireEvent } = await import("@testing-library/react");

    render(<MemoryExplorer />);

    // Initially no search query passed to cold panel
    expect(screen.queryByTestId("cold-query")).toBeNull();
    expect(screen.queryByTestId("cold-panel")).toBeNull();

    // Trigger a "search related" action from the warm panel
    fireEvent.click(screen.getByTestId("warm-search-btn"));

    // The search view should open and receive the query.
    expect(screen.getByTestId("cold-panel")).toBeDefined();
    expect(screen.getByTestId("cold-query")).toBeDefined();
    expect(screen.getByText("test query")).toBeDefined();
  });

  it("opens full details for search results and artifacts", async () => {
    const { fireEvent } = await import("@testing-library/react");

    render(<MemoryExplorer />);

    fireEvent.click(screen.getByRole("tab", { name: /Details/i }));
    expect(
      screen.getByText(/Select a memory search result or artifact/i),
    ).toBeDefined();

    fireEvent.click(screen.getByRole("tab", { name: /Search memory/i }));
    fireEvent.click(screen.getByTestId("search-result-select-btn"));
    expect(
      screen.getByText("Full memory content for inspection."),
    ).toBeDefined();
    expect(screen.getByText("91% relevance")).toBeDefined();

    fireEvent.click(screen.getByRole("tab", { name: /Recent work/i }));
    fireEvent.click(screen.getByTestId("artifact-select-btn"));
    expect(screen.getByRole("heading", { name: "Launch Plan" })).toBeDefined();
    expect(screen.getByText("# Launch Plan")).toBeDefined();
    expect(screen.getByText(/Download artifact/i)).toBeDefined();
  });

  it("keeps focused views in browser history and supports arrow keys", async () => {
    const { fireEvent } = await import("@testing-library/react");
    render(<MemoryExplorer />);

    const tablist = screen.getByRole("tablist", { name: "Memory views" });
    fireEvent.keyDown(tablist, { key: "ArrowRight" });
    expect(
      screen
        .getByRole("tab", { name: "Search Memory" })
        .getAttribute("aria-selected"),
    ).toBe("true");
    expect(new URL(window.location.href).searchParams.get("view")).toBe("search");

    window.history.replaceState(null, "", "/memory");
    fireEvent.popState(window);
    expect(
      screen
        .getByRole("tab", { name: "Recent Work" })
        .getAttribute("aria-selected"),
    ).toBe("true");
  });
});
