"use client";

import { Brain, Copy, LockKeyhole, Trash2 } from "lucide-react";
import type { CatalogueAgent } from "@/store/useCortexStore";

interface AgentCardProps {
  agent: CatalogueAgent;
  onSelect: (agent: CatalogueAgent) => void;
  onDelete: (id: string) => void;
}

export default function AgentCard({ agent, onSelect, onDelete }: AgentCardProps) {
  const isBuiltIn = agent.source === "built_in" || agent.locked;
  const capabilities = agent.capability_refs?.length ? agent.capability_refs : agent.tools;
  const contextCount = agent.context_bindings?.length ?? 0;

  const handleDelete = (event: React.MouseEvent) => {
    event.stopPropagation();
    if (window.confirm(`Delete profile "${agent.name}"?`)) onDelete(agent.id);
  };

  return (
    <article
      onClick={() => onSelect(agent)}
      className="group cursor-pointer rounded-lg border border-cortex-border bg-cortex-surface p-4 transition-colors hover:border-cortex-primary/60"
    >
      <div className="mb-3 flex items-start justify-between gap-3">
        <div className="flex min-w-0 items-center gap-2">
          <div className="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-md bg-cortex-primary/10">
            <Brain className="h-4 w-4 text-cortex-primary" />
          </div>
          <div className="min-w-0">
            <h3 className="truncate text-sm font-semibold text-cortex-text-main">{agent.name}</h3>
            <p className="truncate text-xs text-cortex-text-muted">{agent.role}</p>
          </div>
        </div>
        {isBuiltIn ? (
          <LockKeyhole className="h-4 w-4 text-cortex-text-muted" aria-label="Built-in profile" />
        ) : (
          <button
            type="button"
            aria-label={`Delete ${agent.name}`}
            onClick={handleDelete}
            className="rounded p-1 text-cortex-text-muted opacity-0 transition group-hover:opacity-100 hover:bg-cortex-danger/15 hover:text-cortex-danger"
          >
            <Trash2 className="h-4 w-4" />
          </button>
        )}
      </div>
      <p className="mb-4 line-clamp-2 min-h-10 text-xs leading-5 text-cortex-text-muted">
        {agent.description || "Reusable teammate profile for governed team work."}
      </p>
      <div className="flex flex-wrap items-center gap-2 text-[11px] text-cortex-text-muted">
        <span>{capabilities.length} capabilities</span><span aria-hidden="true">·</span>
        <span>{contextCount} context sources</span>
        {isBuiltIn && <><span aria-hidden="true">·</span><span className="inline-flex items-center gap-1"><Copy className="h-3 w-3" /> Copy to customize</span></>}
      </div>
    </article>
  );
}
