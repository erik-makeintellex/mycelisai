"use client";

import Link from "next/link";
import { BookOpen, MessageSquareText } from "lucide-react";
import type { CatalogueAgent } from "@/store/useCortexStore";

export function TeamMemberTemplatesPanel({
  highlightedTemplates,
  isFetchingCatalogue,
  onNewTemplate,
  onEditTemplate,
}: {
  highlightedTemplates: CatalogueAgent[];
  isFetchingCatalogue: boolean;
  onNewTemplate: () => void;
  onEditTemplate: (agent: CatalogueAgent) => void;
}) {
  return (
    <section className="rounded-lg border border-cortex-border bg-cortex-surface p-4">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="flex min-w-0 items-start gap-3">
          <BookOpen className="mt-0.5 h-5 w-5 shrink-0 text-cortex-success" />
          <div>
            <div className="flex items-center gap-2">
              <h2 className="text-sm font-semibold text-cortex-text-main">Worker profiles</h2>
              {isFetchingCatalogue && <span className="text-xs text-cortex-text-muted">Loading</span>}
            </div>
            <p className="mt-1 text-sm text-cortex-text-muted">
              Name one when you ask Soma, or let Soma choose the smallest useful team.
            </p>
          </div>
        </div>
        <button type="button" onClick={onNewTemplate} className="inline-flex items-center gap-1.5 rounded-md border border-cortex-success/40 px-3 py-2 text-xs font-semibold text-cortex-success hover:bg-cortex-success/10">
          <MessageSquareText className="h-4 w-4" /> Create with Soma
        </button>
      </div>

      <div className="mt-4 divide-y divide-cortex-border overflow-hidden rounded-md border border-cortex-border">
        {highlightedTemplates.length > 0 ? highlightedTemplates.map((agent) => {
          const capabilities = agent.capability_refs?.length ?? agent.tools.length;
          const contexts = agent.context_bindings?.length ?? 0;
          const readyMade = agent.source === "built_in" || agent.locked;
          return (
            <button key={agent.id} type="button" onClick={() => onEditTemplate(agent)} className="flex w-full items-center justify-between gap-3 bg-cortex-bg px-3 py-3 text-left hover:bg-cortex-surface">
              <span className="min-w-0">
                <span className="block truncate text-sm font-semibold text-cortex-text-main">{agent.name}</span>
                <span className="mt-0.5 block truncate text-xs text-cortex-text-muted">{agent.description || agent.role.replaceAll("_", " ")}</span>
              </span>
              <span className="shrink-0 text-right text-xs text-cortex-text-muted">
                <span className="block font-semibold text-cortex-text-main">{readyMade ? "Ready-made" : "Your profile"}</span>
                {capabilities} access · {contexts} context
              </span>
            </button>
          );
        }) : (
          <p className="bg-cortex-bg px-3 py-4 text-sm text-cortex-text-muted">No worker profiles are available.</p>
        )}
      </div>

      <div className="mt-4 flex flex-wrap items-center justify-between gap-3 text-xs">
        <p className="text-cortex-text-muted">Example: “Use the Research Specialist and Quality Reviewer.”</p>
        <Link href="/resources?tab=roles" className="font-semibold text-cortex-primary hover:underline">Manage all profiles</Link>
      </div>
    </section>
  );
}
