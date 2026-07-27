"use client";

import React, { useState, useCallback } from "react";
import {
  Brain,
  ChevronDown,
  ChevronRight,
  FileText,
  Radio,
  Search,
} from "lucide-react";
import HotMemoryPanel from "./HotMemoryPanel";
import WarmMemoryPanel from "./WarmMemoryPanel";
import ColdMemoryPanel from "./ColdMemoryPanel";
import MemoryDetailPanel from "./MemoryDetailPanel";
import type { MemorySelection } from "./memorySelection";
import { useCortexStore } from "@/store/useCortexStore";
import type { Artifact } from "@/store/cortexStoreTypesPlanning";

type MemoryView = "work" | "search" | "details";

const MEMORY_VIEWS: Array<{
  id: MemoryView;
  label: string;
  compactLabel: string;
  description: string;
  icon: React.ComponentType<{ className?: string }>;
}> = [
  {
    id: "work",
    label: "Recent Work",
    compactLabel: "Recent",
    description: "Warm records",
    icon: FileText,
  },
  {
    id: "search",
    label: "Search Memory",
    compactLabel: "Search",
    description: "Cold recall",
    icon: Search,
  },
  {
    id: "details",
    label: "Details",
    compactLabel: "Details",
    description: "Inspect record",
    icon: Brain,
  },
];

// ── MemoryExplorer ────────────────────────────────────────────

export default function MemoryExplorer() {
  const advancedMode = useCortexStore((s) => s.advancedMode);
  const [coldSearchQuery, setColdSearchQuery] = useState<string | undefined>(
    undefined,
  );
  const [activeView, setActiveView] = useState<MemoryView>(readMemoryView);
  const [selection, setSelection] = useState<MemorySelection | null>(null);
  const [signalExpanded, setSignalExpanded] = useState(false);
  const getArtifactDetail = useCortexStore((s) => s.getArtifactDetail);
  const selectedArtifactDetail = useCortexStore((s) => s.selectedArtifactDetail);

  const selectView = useCallback((view: MemoryView, mode: RouteMode = "push") => {
    setActiveView(view);
    updateMemoryRoute(view, mode);
  }, []);

  const handleSearchRelated = useCallback((query: string) => {
    setColdSearchQuery(query);
    selectView("search");
  }, [selectView]);

  const handleSelectArtifact = useCallback(
    (artifact: Artifact) => {
      setSelection({ kind: "artifact", artifact });
      selectView("details");
      void getArtifactDetail(artifact.id);
    },
    [getArtifactDetail, selectView],
  );

  React.useEffect(() => {
    const restoreView = () => setActiveView(readMemoryView());
    window.addEventListener("popstate", restoreView);
    return () => window.removeEventListener("popstate", restoreView);
  }, []);

  React.useEffect(() => {
    if (!selectedArtifactDetail || selection?.kind !== "artifact") return;
    if (selectedArtifactDetail.id !== selection.artifact.id) return;
    setSelection({ kind: "artifact", artifact: selectedArtifactDetail });
  }, [selectedArtifactDetail, selection]);

  return (
    <div className="h-full flex flex-col bg-cortex-bg text-cortex-text-main">
      {/* Header */}
      <header className="h-12 border-b border-cortex-border flex items-center justify-between px-4 bg-cortex-surface/50 backdrop-blur-sm flex-shrink-0">
        <div className="flex items-center gap-2.5">
          <Brain className="w-4 h-4 text-cortex-success" />
          <div>
            <h1 className="font-mono font-bold text-sm text-cortex-text-main">
              Memory
            </h1>
            <span className="hidden md:inline text-[10px] font-mono text-cortex-text-muted ml-3">
              Retained knowledge and conversations from your swarm
            </span>
          </div>
        </div>
      </header>

      {/* Main layout: focused tabs instead of compressed columns */}
      <div className="flex min-h-0 flex-1 flex-col overflow-hidden">
        <nav
          aria-label="Memory views"
          className="flex flex-nowrap gap-1 overflow-x-auto border-b border-cortex-border bg-cortex-surface/30 px-2 py-2 [scrollbar-width:none] sm:px-3 [&::-webkit-scrollbar]:hidden"
          role="tablist"
          onKeyDown={(event) => handleMemoryTabKeyDown(event, activeView, selectView)}
        >
          {MEMORY_VIEWS.map((view) => {
            const Icon = view.icon;
            const isActive = activeView === view.id;
            return (
              <button
                key={view.id}
                type="button"
                role="tab"
                aria-label={view.label}
                aria-selected={isActive}
                aria-controls={`memory-${view.id}-panel`}
                id={`memory-${view.id}-tab`}
                tabIndex={isActive ? 0 : -1}
                onClick={() => selectView(view.id)}
                className={`flex h-11 min-w-max flex-none items-center gap-2 rounded-lg border px-3 text-left transition-colors sm:min-w-[150px] sm:flex-1 lg:max-w-[220px] ${
                  isActive
                    ? "border-cortex-primary/40 bg-cortex-primary/15 text-cortex-text-main"
                    : "border-cortex-border bg-cortex-bg/60 text-cortex-text-muted hover:border-cortex-primary/30 hover:text-cortex-text-main"
                }`}
              >
                <Icon className="h-4 w-4 text-cortex-primary" />
                <span>
                  {view.compactLabel === view.label ? (
                    <span className="block text-sm font-bold">{view.label}</span>
                  ) : (
                    <>
                      <span className="block text-sm font-bold sm:hidden">
                        {view.compactLabel}
                      </span>
                      <span className="hidden text-sm font-bold sm:block">
                        {view.label}
                      </span>
                    </>
                  )}
                  <span className="hidden text-[10px] text-cortex-text-muted sm:block">
                    {view.description}
                  </span>
                </span>
              </button>
            );
          })}
        </nav>

        <div
          className="min-h-0 flex-1 overflow-hidden"
          role="tabpanel"
          id={`memory-${activeView}-panel`}
          aria-labelledby={`memory-${activeView}-tab`}
        >
          {activeView === "work" ? (
            <WarmMemoryPanel
              onSearchRelated={handleSearchRelated}
              onSelectArtifact={(artifact) => handleSelectArtifact(artifact)}
            />
          ) : activeView === "search" ? (
            <ColdMemoryPanel
              searchQuery={coldSearchQuery}
              onSelectResult={(result) => {
                setSelection({ kind: "search", result });
                selectView("details");
              }}
            />
          ) : (
            <MemoryDetailPanel selection={selection} embedded />
          )}
        </div>
      </div>

      {/* Advanced Mode only: Signal Stream (collapsible) */}
      {advancedMode && (
        <div className="flex-shrink-0 border-t border-cortex-border bg-cortex-surface/30">
          <button
            onClick={() => setSignalExpanded(!signalExpanded)}
            className="w-full h-8 flex items-center gap-2 px-3 hover:bg-cortex-bg/40 transition-colors"
          >
            {signalExpanded ? (
              <ChevronDown className="w-3.5 h-3.5 text-cortex-text-muted" />
            ) : (
              <ChevronRight className="w-3.5 h-3.5 text-cortex-text-muted" />
            )}
            <Radio className="w-3 h-3 text-cortex-text-muted" />
            <span className="text-[9px] font-mono uppercase tracking-widest text-cortex-text-muted font-bold">
              Signal Stream
            </span>
          </button>
          {signalExpanded && (
            <div
              style={{ height: 220 }}
              className="border-t border-cortex-border/50 overflow-hidden"
            >
              <HotMemoryPanel />
            </div>
          )}
        </div>
      )}
    </div>
  );
}

type RouteMode = "push" | "replace";

const MEMORY_VIEW_ORDER: MemoryView[] = ["work", "search", "details"];

function readMemoryView(): MemoryView {
  if (typeof window === "undefined") return "work";
  const view = new URL(window.location.href).searchParams.get("view");
  return view === "search" || view === "details" ? view : "work";
}

function updateMemoryRoute(view: MemoryView, mode: RouteMode) {
  if (typeof window === "undefined") return;
  const nextUrl = new URL(window.location.href);
  if (view === "work") nextUrl.searchParams.delete("view");
  else nextUrl.searchParams.set("view", view);
  const updateHistory =
    mode === "push" ? window.history.pushState : window.history.replaceState;
  updateHistory.call(
    window.history,
    window.history.state,
    "",
    `${nextUrl.pathname}${nextUrl.search}${nextUrl.hash}`,
  );
}

function handleMemoryTabKeyDown(
  event: React.KeyboardEvent<HTMLElement>,
  activeView: MemoryView,
  selectView: (view: MemoryView, mode?: RouteMode) => void,
) {
  const offsets: Record<string, number | undefined> = {
    ArrowRight: 1,
    ArrowDown: 1,
    ArrowLeft: -1,
    ArrowUp: -1,
  };
  let nextIndex = MEMORY_VIEW_ORDER.indexOf(activeView);
  if (event.key === "Home") nextIndex = 0;
  else if (event.key === "End") nextIndex = MEMORY_VIEW_ORDER.length - 1;
  else if (offsets[event.key]) {
    nextIndex =
      (nextIndex + offsets[event.key]! + MEMORY_VIEW_ORDER.length) %
      MEMORY_VIEW_ORDER.length;
  } else return;

  event.preventDefault();
  const nextView = MEMORY_VIEW_ORDER[nextIndex];
  selectView(nextView);
  requestAnimationFrame(() => {
    document.getElementById(`memory-${nextView}-tab`)?.focus();
  });
}
