"use client";

import { ChevronRight, Search, X } from "lucide-react";
import type { DocEntry, DocSection } from "@/lib/docsManifest";

type DocsSidebarProps = {
  sections: DocSection[];
  activeSlug: string | null;
  query: string;
  onSelect: (entry: DocEntry) => void;
  onQueryChange: (query: string) => void;
};

export function DocsSidebar({
  sections,
  activeSlug,
  query,
  onSelect,
  onQueryChange,
}: DocsSidebarProps) {
  const normalizedQuery = query.toLowerCase();
  const filtered = query.trim()
    ? sections
        .map((section) => ({
          ...section,
          docs: section.docs.filter(
            (doc) =>
              doc.label.toLowerCase().includes(normalizedQuery) ||
              (doc.description ?? "").toLowerCase().includes(normalizedQuery),
          ),
        }))
        .filter((section) => section.docs.length > 0)
    : sections;

  return (
    <div className="w-56 flex-shrink-0 border-r border-cortex-border flex flex-col overflow-hidden">
      <div className="px-3 py-2.5 border-b border-cortex-border">
        <div className="flex items-center gap-2 bg-cortex-bg border border-cortex-border rounded px-2 py-1.5">
          <Search className="w-3 h-3 text-cortex-text-muted flex-shrink-0" />
          <input
            type="text"
            aria-label="Filter docs"
            value={query}
            onChange={(event) => onQueryChange(event.target.value)}
            placeholder="Filter docs..."
            className="flex-1 bg-transparent text-[11px] font-mono text-cortex-text-main placeholder:text-cortex-text-muted/50 outline-none"
          />
          {query ? (
            <button
              type="button"
              aria-label="Clear docs filter"
              onClick={() => onQueryChange("")}
              className="text-cortex-text-muted hover:text-cortex-text-main"
            >
              <X className="w-3 h-3" />
            </button>
          ) : null}
        </div>
      </div>

      <div className="flex-1 overflow-y-auto py-2">
        {filtered.map((section) => (
          <div key={section.section} className="mb-3">
            <div className="px-3 pb-1 text-[9px] font-mono font-bold uppercase tracking-widest text-cortex-text-muted/60">
              {section.section}
            </div>
            {section.docs.map((doc) => {
              const isActive = activeSlug === doc.slug;
              return (
                <button
                  key={doc.slug}
                  type="button"
                  onClick={() => onSelect(doc)}
                  title={doc.description}
                  className={`w-full text-left flex items-center gap-2 px-3 py-1.5 transition-colors ${
                    isActive
                      ? "bg-cortex-primary/10 text-cortex-primary border-r-2 border-cortex-primary"
                      : "text-cortex-text-muted hover:text-cortex-text-main hover:bg-cortex-bg/60"
                  }`}
                >
                  <ChevronRight
                    className={`w-3 h-3 flex-shrink-0 transition-transform ${
                      isActive ? "rotate-90 text-cortex-primary" : ""
                    }`}
                  />
                  <span className="text-[11px] font-mono truncate">
                    {doc.label}
                  </span>
                </button>
              );
            })}
          </div>
        ))}

        {filtered.length === 0 ? (
          <p className="px-3 py-4 text-[11px] font-mono text-cortex-text-muted/60 text-center">
            No docs match "{query}"
          </p>
        ) : null}
      </div>
    </div>
  );
}
