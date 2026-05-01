"use client";

import Link from "next/link";
import { BookOpen, Bot } from "lucide-react";
import type { CatalogueAgent } from "@/store/useCortexStore";

export function TeamMemberTemplatesPanel({
  highlightedTemplates,
  templateCoverage,
  isFetchingCatalogue,
  onNewTemplate,
  onEditTemplate,
}: {
  highlightedTemplates: CatalogueAgent[];
  templateCoverage: [string, CatalogueAgent[]][];
  isFetchingCatalogue: boolean;
  onNewTemplate: () => void;
  onEditTemplate: (agent: CatalogueAgent) => void;
}) {
  return (
    <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-4">
      <div className="flex items-start justify-between gap-3">
        <div className="flex items-start gap-3">
          <div className="rounded-full border border-cortex-success/20 bg-cortex-success/10 p-2 text-cortex-success">
            <BookOpen className="h-4 w-4" />
          </div>
          <div>
            <div className="flex flex-wrap items-center gap-2">
              <p className="text-sm font-semibold text-cortex-text-main">Soma team-member templates</p>
              {isFetchingCatalogue && <span className="text-[10px] font-mono text-cortex-text-muted animate-pulse">loading...</span>}
            </div>
            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
              By default, team members inherit the shared organization model policy. Give Soma reusable templates when certain work should prefer a specific specialist role, model, MCP/internal toolset, or output contract.
            </p>
          </div>
        </div>
        <button type="button" onClick={onNewTemplate} className="inline-flex items-center justify-center rounded-2xl border border-cortex-success/30 px-3 py-2 text-sm font-semibold text-cortex-success hover:bg-cortex-success/10">
          New template
        </button>
      </div>

      <div className="mt-4 space-y-3">
        {highlightedTemplates.length > 0 ? (
          highlightedTemplates.map((agent) => (
            <button key={agent.id} type="button" onClick={() => onEditTemplate(agent)} className="w-full rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-left transition-colors hover:border-cortex-primary/30 hover:bg-cortex-surface">
              <div className="flex items-start justify-between gap-3">
                <div>
                  <div className="flex flex-wrap items-center gap-2">
                    <p className="text-sm font-semibold text-cortex-text-main">{agent.name}</p>
                    <span className="rounded-full border border-cortex-border px-2 py-0.5 text-[10px] font-mono uppercase tracking-[0.16em] text-cortex-text-muted">{agent.role}</span>
                  </div>
                  <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{agent.system_prompt?.trim() || "Reusable specialist template for future team-member creation."}</p>
                </div>
                <Bot className="mt-0.5 h-4 w-4 text-cortex-text-muted" />
              </div>
              <div className="mt-3 flex flex-wrap gap-2 text-[10px] font-mono">
                <span className="rounded-full border border-cortex-primary/20 bg-cortex-primary/10 px-2 py-0.5 text-cortex-primary">{agent.model?.trim() || "inherits org model"}</span>
                <span className="rounded-full border border-cortex-border px-2 py-0.5 text-cortex-text-muted">{agent.tools.length} MCP/internal tool{agent.tools.length !== 1 ? "s" : ""}</span>
                <span className="rounded-full border border-cortex-border px-2 py-0.5 text-cortex-text-muted">
                  {agent.outputs.length > 0 ? `outputs: ${agent.outputs.slice(0, 2).join(", ")}${agent.outputs.length > 2 ? "..." : ""}` : "general lane support"}
                </span>
              </div>
            </button>
          ))
        ) : (
          <div className="rounded-2xl border border-dashed border-cortex-border bg-cortex-bg px-4 py-4">
            <p className="text-sm font-semibold text-cortex-text-main">No team-member templates yet</p>
            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">Create reusable agent templates here so Soma knows which kinds of specialists to reach for when a new team needs writers, coders, researchers, or reviewers.</p>
          </div>
        )}
      </div>

      <div className="mt-4 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4">
        <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <div>
            <p className="text-sm font-semibold text-cortex-text-main">Agent type, model, and MCP access</p>
            <p className="mt-1 text-sm leading-6 text-cortex-text-muted">Edit a template to choose its role and model. Manage MCP servers, direct web search, and reusable tool references from Resources.</p>
          </div>
          <Link href="/resources?tab=tools" className="inline-flex items-center justify-center rounded-2xl border border-cortex-primary/30 px-4 py-2 text-sm font-semibold text-cortex-primary hover:bg-cortex-primary/10">
            Manage MCP tools
          </Link>
        </div>
      </div>

      <div className="mt-4 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4">
        <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">What Soma can match by work type</p>
        <div className="mt-3 space-y-2">
          {templateCoverage.length > 0 ? (
            templateCoverage.map(([kind, agents]) => (
              <div key={kind} className="flex flex-col gap-1 rounded-2xl border border-cortex-border/80 bg-cortex-surface px-3 py-2 sm:flex-row sm:items-center sm:justify-between">
                <span className="text-sm font-medium text-cortex-text-main">{kind}</span>
                <span className="text-xs text-cortex-text-muted">{agents.slice(0, 2).map((agent) => agent.name).join(", ")}{agents.length > 2 ? ` +${agents.length - 2} more` : ""}</span>
              </div>
            ))
          ) : (
            <p className="text-sm leading-6 text-cortex-text-muted">Once you add templates, this page will show the work types Soma can map to specific team-member defaults.</p>
          )}
        </div>
        <div className="mt-4 flex flex-wrap gap-2">
          <Link href="/resources?tab=roles" className="inline-flex items-center justify-center rounded-2xl border border-cortex-border px-4 py-2 text-sm font-semibold text-cortex-text-main hover:bg-cortex-border">
            Open full role library
          </Link>
          <Link href="/dashboard" className="inline-flex items-center justify-center rounded-2xl border border-cortex-primary/30 px-4 py-2 text-sm font-semibold text-cortex-primary hover:bg-cortex-primary/10">
            Ask Soma to use these defaults
          </Link>
        </div>
      </div>
    </div>
  );
}
