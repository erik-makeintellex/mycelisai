"use client";

import React, { useState, useEffect, useCallback } from "react";
import {
  Database,
  FileText,
  Package,
  ChevronDown,
  ChevronRight,
  Search,
} from "lucide-react";
import { useCortexStore, type Artifact } from "@/store/useCortexStore";
import WarmArtifactRow from "./WarmArtifactRow";

interface SitRep {
  id: string;
  mission_id: string;
  summary: string;
  raw_count: number;
  created_at: string;
}

type TabId = "warm" | "sitreps" | "artifacts";

const WARM_TABS: Array<
  [TabId, string, string, React.ComponentType<{ className?: string }>]
> = [
  ["warm", "Warm", "Recent work", Database],
  ["sitreps", "SitReps", "Summaries", FileText],
  ["artifacts", "Artifacts", "Outputs", Package],
];

function formatTimestamp(ts: string): string {
  try {
    const d = new Date(ts);
    if (isNaN(d.getTime())) return ts;
    return d.toLocaleDateString("en-GB", {
      day: "2-digit",
      month: "short",
      hour: "2-digit",
      minute: "2-digit",
    });
  } catch {
    return ts;
  }
}

function SitRepCard({
  sitrep,
  onSearchRelated,
}: {
  sitrep: SitRep;
  onSearchRelated?: (query: string) => void;
}) {
  const [expanded, setExpanded] = useState(false);
  const summary = sitrep.summary ?? "";
  const needsTruncation = summary.length > 120;
  const displayText =
    expanded || !needsTruncation ? summary : summary.slice(0, 120) + "...";

  return (
    <div className="bg-cortex-surface border border-cortex-border rounded-xl p-4 space-y-2">
      {/* Summary */}
      <button
        onClick={() => setExpanded(!expanded)}
        className="text-left w-full flex items-start gap-2"
      >
        {needsTruncation &&
          (expanded ? (
            <ChevronDown className="w-3 h-3 text-cortex-text-muted flex-shrink-0 mt-0.5" />
          ) : (
            <ChevronRight className="w-3 h-3 text-cortex-text-muted flex-shrink-0 mt-0.5" />
          ))}
        <span className="text-xs text-cortex-text-main leading-relaxed">
          {displayText}
        </span>
      </button>

      {/* Meta row */}
      <div className="flex items-center gap-2 flex-wrap">
        <span className="text-[9px] font-mono uppercase px-1.5 py-0.5 rounded bg-cortex-primary/15 text-cortex-primary">
          {sitrep.mission_id}
        </span>
        <span className="text-[9px] font-mono uppercase px-1.5 py-0.5 rounded bg-cortex-info/15 text-cortex-info">
          {sitrep.raw_count} raw
        </span>
        <span className="text-[9px] font-mono text-cortex-text-muted ml-auto">
          {formatTimestamp(sitrep.created_at)}
        </span>
      </div>

      {/* Find Related */}
      {onSearchRelated && (
        <button
          onClick={() => onSearchRelated(summary.slice(0, 100))}
          className="flex items-center gap-1 text-[10px] font-mono text-cortex-info hover:text-cortex-primary transition-colors"
        >
          <Search className="w-3 h-3" />
          Find Related
        </button>
      )}
    </div>
  );
}

interface WarmMemoryPanelProps {
  onSearchRelated?: (query: string) => void;
  onSelectArtifact?: (artifact: Artifact) => void;
}

export default function WarmMemoryPanel({
  onSearchRelated,
  onSelectArtifact,
}: WarmMemoryPanelProps) {
  const [activeTab, setActiveTab] = useState<TabId>("warm");
  const [sitreps, setSitreps] = useState<SitRep[]>([]);
  const [loading, setLoading] = useState(false);

  const artifacts = useCortexStore((s) => s.artifacts);
  const isFetchingArtifacts = useCortexStore((s) => s.isFetchingArtifacts);
  const fetchArtifacts = useCortexStore((s) => s.fetchArtifacts);

  const loadSitreps = useCallback(async () => {
    setLoading(true);
    try {
      const res = await fetch("/api/v1/memory/sitreps");
      if (res.ok) {
        const data = await res.json();
        setSitreps(Array.isArray(data) ? data : (data.sitreps ?? []));
      } else {
        setSitreps([]);
      }
    } catch {
      setSitreps([]);
    }
    setLoading(false);
  }, []);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      if (activeTab === "warm") {
        void loadSitreps();
        void fetchArtifacts({ limit: 50 });
      } else if (activeTab === "sitreps") {
        void loadSitreps();
      } else if (activeTab === "artifacts") {
        void fetchArtifacts({ limit: 50 });
      }
    }, 0);
    return () => window.clearTimeout(timer);
  }, [activeTab, loadSitreps, fetchArtifacts]);

  const isLoading = activeTab === "warm"
    ? loading || isFetchingArtifacts
    : activeTab === "sitreps" ? loading : isFetchingArtifacts;

  return (
    <div className="h-full flex flex-col">
      <div className="border-b border-cortex-border bg-cortex-surface/50 px-3 py-2 flex-shrink-0">
        <div className="flex flex-wrap gap-2" role="tablist" aria-label="Warm memory lanes">
          {WARM_TABS.map(([id, label, description, Icon]) => {
            const isActive = activeTab === id;
            return (
              <button
                key={id}
                type="button"
                role="tab"
                aria-label={label}
                aria-selected={isActive}
                onClick={() => setActiveTab(id)}
                className={`flex min-w-[120px] items-center gap-2 rounded-lg border px-3 py-2 text-left transition-colors ${
                  isActive
                    ? "border-cortex-warning/40 bg-cortex-warning/15 text-cortex-text-main"
                    : "border-cortex-border bg-cortex-bg/50 text-cortex-text-muted hover:border-cortex-warning/30 hover:text-cortex-text-main"
                }`}
              >
                <Icon className="w-3.5 h-3.5 text-cortex-warning" />
                <span>
                  <span className="block text-xs font-bold uppercase tracking-widest">
                    {label}
                  </span>
                  <span className="block text-[9px] text-cortex-text-muted">
                    {description}
                  </span>
                </span>
              </button>
            );
          })}
        </div>
      </div>

      <div className="flex-1 overflow-y-auto scrollbar-thin scrollbar-thumb-cortex-border min-h-0">
        {isLoading ? (
          <div className="flex items-center justify-center h-full">
            <span className="text-[10px] font-mono text-cortex-text-muted animate-pulse">
              Loading...
            </span>
          </div>
        ) : activeTab === "warm" ? (
          <WarmOverview
            sitreps={sitreps}
            artifacts={artifacts}
            onSearchRelated={onSearchRelated}
            onSelectArtifact={onSelectArtifact}
          />
        ) : activeTab === "sitreps" ? (
          <SitRepList sitreps={sitreps} onSearchRelated={onSearchRelated} />
        ) : artifacts.length === 0 ? (
          <EmptyWarmState
            icon={<Package className="w-8 h-8 text-cortex-text-muted opacity-20" />}
            label="No artifacts stored."
          />
        ) : (
          <div>
            {artifacts.map((artifact) => (
              <WarmArtifactRow
                key={artifact.id}
                artifact={artifact}
                onSelect={(next) => onSelectArtifact?.(next)}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

function WarmOverview({
  sitreps,
  artifacts,
  onSearchRelated,
  onSelectArtifact,
}: {
  sitreps: SitRep[];
  artifacts: Artifact[];
  onSearchRelated?: (query: string) => void;
  onSelectArtifact?: (artifact: Artifact) => void;
}) {
  if (sitreps.length === 0 && artifacts.length === 0) {
    return (
      <EmptyWarmState
        icon={<Database className="w-8 h-8 text-cortex-text-muted opacity-20" />}
        label="No recent warm records yet."
      />
    );
  }
  return (
    <div className="p-3 space-y-4">
      {sitreps.length > 0 ? (
        <section className="space-y-2">
          <h2 className="text-[10px] font-bold uppercase tracking-widest text-cortex-text-muted">
            Latest SitReps
          </h2>
          {sitreps.slice(0, 5).map((sitrep) => (
            <SitRepCard
              key={sitrep.id}
              sitrep={sitrep}
              onSearchRelated={onSearchRelated}
            />
          ))}
        </section>
      ) : null}
      {artifacts.length > 0 ? (
        <section>
          <h2 className="px-1 pb-2 text-[10px] font-bold uppercase tracking-widest text-cortex-text-muted">
            Recent Artifacts
          </h2>
          <div className="overflow-hidden rounded-xl border border-cortex-border">
            {artifacts.slice(0, 8).map((artifact) => (
              <WarmArtifactRow
                key={artifact.id}
                artifact={artifact}
                onSelect={(next) => onSelectArtifact?.(next)}
              />
            ))}
          </div>
        </section>
      ) : null}
    </div>
  );
}

function SitRepList({
  sitreps,
  onSearchRelated,
}: {
  sitreps: SitRep[];
  onSearchRelated?: (query: string) => void;
}) {
  if (sitreps.length === 0) {
    return (
      <EmptyWarmState
        icon={<FileText className="w-8 h-8 text-cortex-text-muted opacity-20" />}
        label="No sitreps recorded yet."
      />
    );
  }
  return (
    <div className="p-3 space-y-2">
      {sitreps.map((sitrep) => (
        <SitRepCard
          key={sitrep.id}
          sitrep={sitrep}
          onSearchRelated={onSearchRelated}
        />
      ))}
    </div>
  );
}

function EmptyWarmState({
  icon,
  label,
}: {
  icon: React.ReactNode;
  label: string;
}) {
  return (
    <div className="flex flex-col items-center justify-center h-full gap-2">
      {icon}
      <span className="text-[10px] font-mono text-cortex-text-muted">
        {label}
      </span>
    </div>
  );
}
