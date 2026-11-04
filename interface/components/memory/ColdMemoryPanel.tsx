"use client";

import React, { useState, useEffect } from "react";
import { Snowflake, Brain, AlertTriangle } from "lucide-react";
import {
  memoryResultScore,
  formatMemoryLayerLabel,
  formatVerificationStatusLabel,
  type SearchResult,
  type ArbitrationSummary,
} from "./memorySelection";

type LayerFilterValue = "" | "GOVERNED_DOCTRINE" | "PROJECT_MEMORY" | "AGENT_MEMORY" | "SOMA_MEMORY" | "REFLECTION_MEMORY";

const LAYER_FILTER_OPTIONS: Array<{ value: LayerFilterValue; label: string }> = [
  { value: "", label: "All" },
  { value: "GOVERNED_DOCTRINE", label: "Governed Doctrine" },
  { value: "PROJECT_MEMORY", label: "Project" },
  { value: "AGENT_MEMORY", label: "Team / Agent" },
  { value: "SOMA_MEMORY", label: "Soma" },
  { value: "REFLECTION_MEMORY", label: "Reflection" },
];

// ── Timestamp Formatting ─────────────────────────────────────

function formatTimestamp(ts: string): string {
  try {
    const d = new Date(ts);
    if (isNaN(d.getTime())) return ts;
    return d.toLocaleDateString("en-GB", { day: "2-digit", month: "short", hour: "2-digit", minute: "2-digit" });
  } catch {
    return ts;
  }
}

// ── Result Card ──────────────────────────────────────────────

function ResultCard({ result, onSelect }: { result: SearchResult; onSelect: (result: SearchResult) => void }) {
  const pct = Math.round(memoryResultScore(result) * 100);
  const preview = result.content.length > 200 ? result.content.slice(0, 200) + "..." : result.content;
  const layerLabel = formatMemoryLayerLabel(result.memory_layer);
  const statusLabel = formatVerificationStatusLabel(result.verification_status);

  return (
    <button
      type="button"
      onClick={() => onSelect(result)}
      className="w-full bg-cortex-surface border border-cortex-border rounded-xl p-4 space-y-2 text-left hover:border-cortex-primary/40 hover:bg-cortex-primary/5 focus:outline-none focus:ring-2 focus:ring-cortex-primary/40"
    >
      {/* Relevance bar */}
      <div className="flex items-center gap-2">
        <div className="flex-1 h-1 rounded-full bg-cortex-border/50 overflow-hidden">
          <div
            className="h-1 rounded-full bg-cortex-primary transition-all"
            style={{ width: `${pct}%` }}
          />
        </div>
        <span className="text-[10px] font-mono text-cortex-primary flex-shrink-0">
          {pct}%
        </span>
      </div>

      {/* Content preview */}
      <p className="text-sm text-cortex-text-main leading-relaxed">{preview}</p>

      {/* Source + Layer + Verification + Timestamp */}
      <div className="flex flex-wrap items-center gap-1.5">
        <span className="text-[9px] font-mono uppercase px-1.5 py-0.5 rounded bg-cortex-info/15 text-cortex-info">
          {result.source}
        </span>
        {layerLabel ? (
          <span className="text-[9px] font-mono px-1.5 py-0.5 rounded bg-cortex-primary/15 text-cortex-primary">
            {layerLabel}
          </span>
        ) : null}
        {statusLabel ? (
          <span className="text-[9px] font-mono px-1.5 py-0.5 rounded bg-cortex-border/60 text-cortex-text-main">
            {statusLabel}
          </span>
        ) : null}
        <span className="text-[9px] font-mono text-cortex-text-muted ml-auto">
          {formatTimestamp(result.created_at)}
        </span>
      </div>
    </button>
  );
}

// ── ColdMemoryPanel ──────────────────────────────────────────

interface ColdMemoryPanelProps {
  searchQuery?: string;
  onSelectResult?: (result: SearchResult) => void;
}

export default function ColdMemoryPanel({
  searchQuery,
  onSelectResult,
}: ColdMemoryPanelProps) {
  const [query, setQuery] = useState(searchQuery ?? "");
  const [layerFilter, setLayerFilter] = useState<LayerFilterValue>("");
  const [results, setResults] = useState<SearchResult[]>([]);
  const [arbitration, setArbitration] = useState<ArbitrationSummary | null>(null);
  const [isSearching, setIsSearching] = useState(false);
  const [hasSearched, setHasSearched] = useState(false);

  // Sync prop changes into local state
  useEffect(() => {
    if (searchQuery !== undefined && searchQuery !== query) {
      setQuery(searchQuery);
    }
    // Only react to prop changes, not local query changes
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchQuery]);

  // Debounced search
  useEffect(() => {
    if (!query.trim()) {
      setResults([]);
      setArbitration(null);
      setHasSearched(false);
      setIsSearching(false);
      return;
    }
    let controller: AbortController | null = null;
    const timer = setTimeout(async () => {
      controller = new AbortController();
      setIsSearching(true);
      setHasSearched(true);
      const filterParam = layerFilter
        ? `&layer_filter=${encodeURIComponent(layerFilter)}`
        : "";
      try {
        const res = await fetch(
          `/api/v1/memory/search?q=${encodeURIComponent(query)}${filterParam}`,
          { signal: controller.signal },
        );
        if (controller.signal.aborted) return;
        if (res.ok) {
          const data = await res.json();
          if (controller.signal.aborted) return;
          setResults(data.results ?? []);
          setArbitration(data.arbitration ?? null);
        } else {
          setResults([]);
          setArbitration(null);
        }
      } catch (err) {
        if (err instanceof DOMException && err.name === "AbortError") return;
        setResults([]);
        setArbitration(null);
      }
      if (!controller.signal.aborted) setIsSearching(false);
    }, 500);
    return () => {
      clearTimeout(timer);
      controller?.abort();
    };
  }, [query, layerFilter]);

  return (
    <div className="h-full flex flex-col">
      {/* Sub-header */}
      <div className="h-10 flex items-center px-3 border-b border-cortex-border bg-cortex-surface/50 flex-shrink-0">
        <div className="flex items-center gap-2">
          <Snowflake className="w-3.5 h-3.5 text-cortex-info" />
          <span className="text-[10px] font-bold uppercase tracking-widest text-cortex-text-muted">
            Cold
          </span>
        </div>
      </div>

      {/* Search input + Layer filter bar */}
      <div className="px-3 py-2 border-b border-cortex-border/50 flex-shrink-0 space-y-2">
        <div className="relative">
          <Brain className="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-cortex-text-muted pointer-events-none" />
          <input
            type="text"
            aria-label="Search semantic memory"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search memory..."
            className="w-full bg-cortex-bg border border-cortex-border rounded-lg pl-8 pr-3 py-2 text-sm font-mono text-cortex-text-main placeholder:text-cortex-text-muted/50 focus:outline-none focus:border-cortex-primary/50 transition-colors"
          />
        </div>
        <div
          role="group"
          aria-label="Filter memory layer"
          className="flex flex-wrap items-center gap-1"
        >
          {LAYER_FILTER_OPTIONS.map((opt) => (
            <button
              key={opt.value || "all"}
              type="button"
              aria-pressed={layerFilter === opt.value}
              onClick={() => setLayerFilter(opt.value)}
              className={`text-[9px] font-mono px-2 py-0.5 rounded border transition-colors ${
                layerFilter === opt.value
                  ? "bg-cortex-primary/20 border-cortex-primary text-cortex-primary"
                  : "bg-cortex-bg border-cortex-border text-cortex-text-muted hover:text-cortex-text-main"
              }`}
            >
              {opt.label}
            </button>
          ))}
        </div>
      </div>

      {/* Contradiction Warning Banner */}
      {arbitration?.has_contradictions ? (
        <div
          data-testid="memory-contradiction-banner"
          className="mx-3 mt-2 flex items-center gap-2 rounded-lg border border-amber-500/40 bg-amber-500/10 px-3 py-2 text-xs text-amber-300 flex-shrink-0"
        >
          <AlertTriangle className="h-3.5 w-3.5 flex-shrink-0" />
          <span>
            Conflicting memory candidates were detected during retrieval arbitration.
          </span>
        </div>
      ) : null}

      {/* Results */}
      <div className="flex-1 overflow-y-auto scrollbar-thin scrollbar-thumb-cortex-border min-h-0">
        {isSearching ? (
          <div className="flex items-center justify-center h-full">
            <span className="text-[10px] font-mono text-cortex-text-muted animate-pulse">
              Searching vectors...
            </span>
          </div>
        ) : hasSearched && results.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full gap-2">
            <Snowflake className="w-8 h-8 text-cortex-text-muted opacity-20" />
            <span className="text-[10px] font-mono text-cortex-text-muted">
              No results found.
            </span>
          </div>
        ) : !hasSearched ? (
          <div className="flex flex-col items-center justify-center h-full gap-2">
            <Brain className="w-8 h-8 text-cortex-text-muted opacity-20" />
            <span className="text-[10px] font-mono text-cortex-text-muted">
              Enter a query to search semantic memory.
            </span>
          </div>
        ) : (
          <div className="p-3 space-y-2">
            {results.map((result) => (
              <ResultCard
                key={result.id}
                result={result}
                onSelect={(next) => onSelectResult?.(next)}
              />
            ))}
          </div>
        )}
      </div>

      {/* Footer */}
      <div className="h-7 flex items-center justify-center border-t border-cortex-border/50 bg-cortex-surface/30 flex-shrink-0">
        <span className="text-[9px] font-mono text-cortex-text-muted">
          Vectors: {results.length}
        </span>
      </div>
    </div>
  );
}
