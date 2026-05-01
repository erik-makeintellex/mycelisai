"use client";

import Link from "next/link";
import { Database, Globe, Users, Wrench } from "lucide-react";

const somaPromptExamples = [
  {
    icon: Globe,
    label: "Web search",
    prompt:
      "Search the web for the latest changes in self-hosted AI agent products, summarize the top findings, and cite the sources you used.",
  },
  {
    icon: Users,
    label: "Create a team",
    prompt:
      "Create a small temporary team to review the release-readiness risks, assign the right specialists, and bring the retained output back here for approval.",
  },
  {
    icon: Users,
    label: "Talk with teams",
    prompt:
      "Ask the active delivery teams for current blockers, compare their answers, and tell me which workflow needs attention first.",
  },
  {
    icon: Database,
    label: "Use host data",
    prompt:
      "Use the host data under workspace/shared-sources to answer this question, and tell me which files or records shaped the answer.",
  },
  {
    icon: Wrench,
    label: "Review tools",
    prompt:
      "Review the current MCP tool structure, tell me which tools are available to Soma and teams, and recommend what should be connected next.",
  },
] as const;

export function SomaCapabilityGuide() {
  return (
    <div className="rounded-2xl border border-cortex-border bg-cortex-bg/70 p-3">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div>
          <p className="text-[11px] font-mono uppercase tracking-[0.18em] text-cortex-primary">
            Say this to Soma
          </p>
          <p className="mt-1 text-sm text-cortex-text-muted">
            Use direct verbs. Soma should search, create teams, ask teams, read host data, and review connected tools when those capabilities are configured.
          </p>
        </div>
        <Link
          href="/settings?tab=tools"
          className="rounded-xl border border-cortex-border bg-cortex-surface px-3 py-2 text-xs font-semibold text-cortex-text-main hover:border-cortex-primary/25"
        >
          Manage tools
        </Link>
      </div>
      <div className="mt-3 grid gap-2 lg:grid-cols-5">
        {somaPromptExamples.map((example) => {
          const Icon = example.icon;
          return (
            <button
              key={example.label}
              type="button"
              onClick={() => void navigator.clipboard?.writeText(example.prompt)}
              className="rounded-xl border border-cortex-border bg-cortex-surface px-3 py-2 text-left transition hover:border-cortex-primary/30 hover:bg-cortex-primary/5"
              title="Copy prompt"
            >
              <span className="flex items-center gap-2 text-xs font-semibold text-cortex-text-main">
                <Icon className="h-3.5 w-3.5 text-cortex-primary" />
                {example.label}
              </span>
              <span className="mt-2 line-clamp-3 block text-xs leading-5 text-cortex-text-muted">
                {example.prompt}
              </span>
            </button>
          );
        })}
      </div>
    </div>
  );
}
