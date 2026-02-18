"use client";

import React, { useState, useEffect } from "react";
import { Snowflake, Brain } from "lucide-react";

// ── Types ────────────────────────────────────────────────────

interface SearchResult {
    id: string;
    content: string;
    similarity: number;
    source: string;
    created_at: string;
}

// ── Timestamp Formatting ─────────────────────────────────────

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

// ── Result Card ──────────────────────────────────────────────

function ResultCard({ result }: { result: SearchResult }) {
    const pct = Math.round(result.similarity * 100);
    const preview = result.content.length > 200
        ? result.content.slice(0, 200) + "..."
        : result.content;

    return (
        <div className="bg-cortex-surface border border-cortex-border rounded-xl p-4 space-y-2">
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
            <p className="text-sm text-cortex-text-main leading-relaxed">
                {preview}
            </p>

            {/* Source + Timestamp */}
            <div className="flex items-center gap-2">
                <span className="text-[9px] font-mono uppercase px-1.5 py-0.5 rounded bg-cortex-info/15 text-cortex-info">
                    {result.source}
                </span>
                <span className="text-[9px] font-mono text-cortex-text-muted ml-auto">
                    {formatTimestamp(result.created_at)}
                </span>
            </div>
        </div>
    );
}

// ── ColdMemoryPanel ──────────────────────────────────────────

interface ColdMemoryPanelProps {
    searchQuery?: string;
}

export default function ColdMemoryPanel({ searchQuery }: ColdMemoryPanelProps) {
    const [query, setQuery] = useState(searchQuery ?? "");
    const [results, setResults] = useState<SearchResult[]>([]);
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
            setHasSearched(false);
            return;
        }
        const timer = setTimeout(async () => {
            setIsSearching(true);
            setHasSearched(true);
            try {
                const res = await fetch(`/api/v1/memory/search?q=${encodeURIComponent(query)}`);
                if (res.ok) {
                    const data = await res.json();
                    setResults(data.results ?? []);
                } else {
                    setResults([]);
                }
            } catch {
                setResults([]);
            }
            setIsSearching(false);
        }, 500);
        return () => clearTimeout(timer);
    }, [query]);

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

            {/* Search input */}
            <div className="px-3 py-2 border-b border-cortex-border/50 flex-shrink-0">
                <div className="relative">
                    <Brain className="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-cortex-text-muted pointer-events-none" />
                    <input
                        type="text"
                        value={query}
                        onChange={(e) => setQuery(e.target.value)}
                        placeholder="Search memory..."
                        className="w-full bg-cortex-bg border border-cortex-border rounded-lg pl-8 pr-3 py-2 text-sm font-mono text-cortex-text-main placeholder:text-cortex-text-muted/50 focus:outline-none focus:border-cortex-primary/50 transition-colors"
                    />
                </div>
            </div>

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
                            <ResultCard key={result.id} result={result} />
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
