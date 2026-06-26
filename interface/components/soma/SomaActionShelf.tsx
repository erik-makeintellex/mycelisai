"use client";

import { useEffect, useMemo, useState } from "react";
import { Bolt, Plus, X } from "lucide-react";

export type SomaPinnedAction = {
  label: string;
  prompt: string;
  userSaved?: boolean;
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
const SAVED_ACTIONS_KEY = "mycelis-soma-saved-actions";

export function SomaActionShelf({
  actions = DEFAULT_PINNED_ACTIONS,
  onRunAction,
}: {
  actions?: readonly SomaPinnedAction[];
  onRunAction: (prompt: string) => void;
}) {
  const [savedActions, setSavedActions] = useState<SomaPinnedAction[]>([]);
  const [isClientReady, setIsClientReady] = useState(false);
  const [studioOpen, setStudioOpen] = useState(false);
  const visibleActions = useMemo(() => [...actions, ...savedActions].slice(0, 5), [actions, savedActions]);

  useEffect(() => {
    try {
      const raw = window.localStorage.getItem(SAVED_ACTIONS_KEY);
      const parsed = raw ? JSON.parse(raw) : [];
      if (Array.isArray(parsed)) {
        setSavedActions(parsed.filter(isSavedAction).slice(0, 2));
      }
    } catch {
      setSavedActions([]);
    }
    setIsClientReady(true);
  }, []);

  const saveAction = (action: SomaPinnedAction) => {
    const next = [action, ...savedActions.filter((item) => item.label !== action.label)].slice(0, 2);
    setSavedActions(next);
    window.localStorage.setItem(SAVED_ACTIONS_KEY, JSON.stringify(next));
    setStudioOpen(false);
  };

  return (
    <section
      className="flex flex-col gap-2 border-b border-cortex-border bg-cortex-surface px-4 py-2.5 md:flex-row md:items-center"
      aria-label="Pinned Soma actions"
      data-testid="soma-action-shelf"
      data-hydrated={isClientReady ? "true" : "false"}
    >
      <div className="shrink-0 text-[11px] font-bold uppercase tracking-[0.14em] text-cortex-text-muted">
        Quick actions:
      </div>
      <div className="flex w-full min-w-0 max-w-full flex-1 gap-2 overflow-x-auto [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
        {visibleActions.map((action) => (
          <button
            key={action.label}
            type="button"
            onClick={() => onRunAction(action.prompt)}
            className="inline-flex min-h-9 shrink-0 items-center gap-2 rounded-lg border border-cortex-border bg-cortex-bg px-3 py-1.5 text-xs font-semibold text-cortex-text-main shadow-sm transition hover:border-cortex-primary/50 hover:bg-cortex-primary/10 focus:outline-none focus:ring-2 focus:ring-cortex-primary/30 sm:px-3.5"
          >
            <Bolt className="h-3.5 w-3.5 text-cortex-warning" />
            {action.label}
          </button>
        ))}
        <button
          type="button"
          onClick={() => setStudioOpen(true)}
          className="inline-flex min-h-9 shrink-0 items-center gap-2 rounded-lg border border-dashed border-cortex-border bg-cortex-bg px-3 py-1.5 text-xs font-semibold text-cortex-text-muted transition hover:border-cortex-primary/50 hover:text-cortex-text-main sm:px-3.5"
          aria-label="Create new quick action"
        >
          <Plus className="h-3.5 w-3.5" />
          Create action
        </button>
      </div>
      {studioOpen ? (
        <ButtonStudio
          onClose={() => setStudioOpen(false)}
          onSave={saveAction}
        />
      ) : null}
    </section>
  );
}

function ButtonStudio({
  onClose,
  onSave,
}: {
  onClose: () => void;
  onSave: (action: SomaPinnedAction) => void;
}) {
  const [label, setLabel] = useState("");
  const [outcome, setOutcome] = useState("");
  const [format, setFormat] = useState("");
  const [approval, setApproval] = useState("Ask before running");
  const canSave = label.trim() && outcome.trim();

  const save = () => {
    if (!canSave) return;
    onSave({
      label: label.trim(),
      userSaved: true,
      prompt: [
        `Quick action: ${label.trim()}.`,
        `Outcome: ${outcome.trim()}.`,
        format.trim() ? `Output format: ${format.trim()}.` : "Output format: let Soma recommend the best format.",
        `Approval behavior: ${approval}.`,
        "Shape the request conversationally first if anything is unclear; keep outputs, proof, and recovery visible.",
      ].join(" "),
    });
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/45 p-4" role="presentation">
      <div
        className="w-full max-w-lg rounded-2xl border border-cortex-border bg-cortex-surface p-5 shadow-2xl"
        role="dialog"
        aria-modal="true"
        aria-labelledby="button-studio-title"
      >
        <div className="flex items-start justify-between gap-3">
          <div>
            <h2 id="button-studio-title" className="text-lg font-semibold text-cortex-text-main">Save quick action</h2>
            <p className="mt-1 text-sm leading-5 text-cortex-text-muted">
              Turn a repeated request into a pinned Soma ask.
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            aria-label="Close quick action studio"
            className="inline-flex h-9 w-9 items-center justify-center rounded-full border border-cortex-border text-cortex-text-muted hover:text-cortex-text-main"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
        <div className="mt-4 grid gap-3">
          <TextField label="Button label" value={label} onChange={setLabel} placeholder="Weekly client brief" />
          <TextField label="Outcome" value={outcome} onChange={setOutcome} placeholder="Create a retained brief with sources and next steps" />
          <TextField label="Output format" value={format} onChange={setFormat} placeholder="Markdown, slide outline, checklist, package..." />
          <label className="grid gap-1 text-sm font-semibold text-cortex-text-main">
            Approval
            <select
              value={approval}
              onChange={(event) => setApproval(event.target.value)}
              className="h-11 rounded-lg border border-cortex-border bg-cortex-bg px-3 text-sm text-cortex-text-main"
            >
              <option>Ask before running</option>
              <option>Run only if low risk</option>
              <option>Always propose first</option>
            </select>
          </label>
        </div>
        <div className="mt-5 flex justify-end gap-2">
          <button
            type="button"
            onClick={onClose}
            className="rounded-lg border border-cortex-border px-4 py-2 text-sm font-semibold text-cortex-text-muted hover:text-cortex-text-main"
          >
            Cancel
          </button>
          <button
            type="button"
            disabled={!canSave}
            onClick={save}
            className="rounded-lg border border-cortex-primary/40 bg-cortex-primary px-4 py-2 text-sm font-semibold text-cortex-bg disabled:cursor-not-allowed disabled:opacity-50"
          >
            Save action
          </button>
        </div>
      </div>
    </div>
  );
}

function TextField({
  label,
  value,
  onChange,
  placeholder,
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  placeholder: string;
}) {
  return (
    <label className="grid gap-1 text-sm font-semibold text-cortex-text-main">
      {label}
      <input
        value={value}
        onChange={(event) => onChange(event.target.value)}
        placeholder={placeholder}
        className="h-11 rounded-lg border border-cortex-border bg-cortex-bg px-3 text-sm text-cortex-text-main placeholder:text-cortex-text-muted"
      />
    </label>
  );
}

function isSavedAction(value: unknown): value is SomaPinnedAction {
  if (!value || typeof value !== "object") return false;
  const candidate = value as Partial<SomaPinnedAction>;
  return typeof candidate.label === "string" && typeof candidate.prompt === "string";
}
