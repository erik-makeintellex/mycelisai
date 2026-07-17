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
  description: string;
  icon: React.ComponentType<{ className?: string }>;
}> = [
  {
    id: "work",
    label: "Recent Work",
    description: "Warm records",
    icon: FileText,
  },
  {
    id: "search",
    label: "Search Memory",
    description: "Cold recall",
    icon: Search,
  },
  {
    id: "details",
    label: "Details",
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
  const [activeView, setActiveView] = useState<MemoryView>("work");
  const [selection, setSelection] = useState<MemorySelection | null>(null);
  const [signalExpanded, setSignalExpanded] = useState(false);
  const getArtifactDetail = useCortexStore((s) => s.getArtifactDetail);
  const selectedArtifactDetail = useCortexStore((s) => s.selectedArtifactDetail);

  const handleSearchRelated = useCallback((query: string) => {
    setColdSearchQuery(query);
    setActiveView("search");
  }, []);

  const handleSelectArtifact = useCallback(
    (artifact: Artifact) => {
      setSelection({ kind: "artifact", artifact });
      setActiveView("details");
      void getArtifactDetail(artifact.id);
    },
    [getArtifactDetail],
  );

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
          className="flex flex-wrap gap-2 border-b border-cortex-border bg-cortex-surface/30 px-3 py-2"
          role="tablist"
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
                onClick={() => setActiveView(view.id)}
                className={`flex min-w-[160px] items-center gap-2 rounded-xl border px-3 py-2 text-left transition-colors ${
                  isActive
                    ? "border-cortex-primary/40 bg-cortex-primary/15 text-cortex-text-main"
                    : "border-cortex-border bg-cortex-bg/60 text-cortex-text-muted hover:border-cortex-primary/30 hover:text-cortex-text-main"
                }`}
              >
                <Icon className="h-4 w-4 text-cortex-primary" />
                <span>
                  <span className="block text-sm font-bold">{view.label}</span>
                  <span className="block text-[10px] text-cortex-text-muted">
                    {view.description}
                  </span>
                </span>
              </button>
            );
          })}
        </nav>

        <div className="min-h-0 flex-1 overflow-hidden">
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
                setActiveView("details");
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
