"use client";

import { Layers2, X } from "lucide-react";

export function SomaContextFocusBar({
  teamName,
  teamId,
  onClear,
}: {
  teamName?: string | null;
  teamId?: string | null;
  onClear: () => void;
}) {
  const isFocused = Boolean(teamId);
  return (
    <div
      className="mb-3 flex flex-col gap-2 rounded-xl border border-cortex-border bg-cortex-bg px-3 py-2 text-sm text-cortex-text-muted lg:flex-row lg:items-center lg:justify-between"
      data-testid="soma-context-focus-bar"
    >
      <div className="flex min-w-0 items-start gap-2">
        <span className="mt-0.5 text-cortex-primary">
          <Layers2 className="h-4 w-4" />
        </span>
        <div className="min-w-0">
          <p className="font-mono text-[10px] uppercase tracking-[0.16em] text-cortex-primary">
            {isFocused ? "Focused team" : "Soma home"}
          </p>
          <p className="mt-1 leading-5">
            {isFocused
              ? `${teamName || teamId} is in focus. Soma will keep this team's chat, work, and outputs close by.`
              : "Select a team to focus Soma on that team's chat, work, and outputs."}
          </p>
        </div>
      </div>
      {isFocused ? (
        <button
          type="button"
          onClick={onClear}
          className="inline-flex shrink-0 items-center justify-center gap-2 rounded-lg border border-cortex-border px-2.5 py-1.5 text-xs font-semibold text-cortex-text-main hover:border-cortex-primary/30"
        >
          <X className="h-3.5 w-3.5" />
          Soma root
        </button>
      ) : null}
    </div>
  );
}
