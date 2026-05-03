"use client";

import Link from "next/link";
import { ClipboardCheck, Database, Globe, Users, Wrench } from "lucide-react";

const somaPromptExamples = [
  {
    icon: ClipboardCheck,
    label: "Confirm action",
    prompt:
      "Review my latest request, match it to related prior commands and tool metadata, tell me the action you infer, and ask me to confirm before you execute.",
  },
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
    label: "Private data",
    prompt:
      "Review the private data needed for this request, name the source and visibility boundary, and ask me to confirm before using or retaining outputs.",
  },
  {
    icon: Wrench,
    label: "Review tools",
    prompt:
      "Review the current MCP tool structure, tell me which tools are available to Soma and teams, recommend what should be connected next, and walk me through enabling missing MCP servers.",
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
            Use direct verbs and ask for visible outcomes. Teams, private data, private services, recurring behavior, and tool changes should become explicit confirmation or proposal steps.
          </p>
        </div>
        <Link
          href="/resources?tab=tools"
          className="rounded-xl border border-cortex-border bg-cortex-surface px-3 py-2 text-xs font-semibold text-cortex-text-main hover:border-cortex-primary/25"
        >
          Advanced tool setup
        </Link>
      </div>
      <div className="mt-3 grid gap-2 lg:grid-cols-3 xl:grid-cols-6">
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
