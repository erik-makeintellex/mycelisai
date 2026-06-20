"use client";

import { Bolt, Plus } from "lucide-react";

export type SomaPinnedAction = {
  label: string;
  prompt: string;
};

export const DEFAULT_PINNED_ACTIONS: SomaPinnedAction[] = [
  {
    label: "Run Expense Audit",
    prompt: "Run the expense audit workflow. Summarize exceptions, proof, and safe next actions before changing anything.",
  },
  {
    label: "Generate Client Brief",
    prompt: "Generate a client brief as a retained output with proof, source notes, and next-step recommendations.",
  },
  {
    label: "Weekly Media Pack",
    prompt: "Run the Weekly Media Pack. Gather assets, create a retained package, and show it in Outcomes and Vault.",
  },
];

export function SomaActionShelf({
  actions = DEFAULT_PINNED_ACTIONS,
  onRunAction,
}: {
  actions?: readonly SomaPinnedAction[];
  onRunAction: (prompt: string) => void;
}) {
  return (
    <section
      className="flex flex-col gap-3 border-b border-cortex-border bg-cortex-surface px-5 py-4 md:flex-row md:items-center"
      aria-label="Pinned Soma actions"
      data-testid="soma-action-shelf"
    >
      <div className="shrink-0 text-[12px] font-bold uppercase tracking-[0.14em] text-cortex-text-muted">
        Quick actions:
      </div>
      <div className="flex min-w-0 flex-1 gap-3 overflow-x-auto pb-1">
        {actions.map((action) => (
          <button
            key={action.label}
            type="button"
            onClick={() => onRunAction(action.prompt)}
            className="inline-flex min-h-11 shrink-0 items-center gap-2 rounded-lg border border-cortex-border bg-cortex-bg px-5 py-2.5 text-sm font-semibold text-cortex-text-main shadow-sm transition hover:border-cortex-primary/50 hover:bg-cortex-primary/10 focus:outline-none focus:ring-2 focus:ring-cortex-primary/30"
          >
            <Bolt className="h-4 w-4 text-cortex-warning" />
            {action.label}
          </button>
        ))}
        <button
          type="button"
          onClick={() => onRunAction("Help me create a reusable quick action. Ask what outcome, output format, source boundary, and approval behavior I want before saving anything.")}
          className="inline-flex min-h-11 shrink-0 items-center gap-2 rounded-lg border border-dashed border-cortex-border bg-cortex-bg px-5 py-2.5 text-sm font-semibold text-cortex-text-muted transition hover:border-cortex-primary/50 hover:text-cortex-text-main"
          aria-label="Create new quick action"
        >
          <Plus className="h-4 w-4" />
          Create action
        </button>
      </div>
    </section>
  );
}
