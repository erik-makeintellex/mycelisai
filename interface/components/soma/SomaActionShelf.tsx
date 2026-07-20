"use client";

import { useEffect, useMemo, useState, useSyncExternalStore } from "react";
import { Bolt, Plus, X } from "lucide-react";
import {
  instantiateBackendAction,
  isSavedAction,
  loadBackendActions,
  persistLocalActions,
  saveBackendAction,
} from "./somaActionPersistence";
import { useClientReady } from "@/lib/browserLocation";

export type SomaPinnedAction = {
  id?: string;
  label: string;
  prompt: string;
  outputFormat?: string;
  approvalBehavior?: string;
  userSaved?: boolean;
};

export const DEFAULT_PINNED_ACTIONS: SomaPinnedAction[] = [
  {
    label: "Plan next step",
    prompt: "Help me turn this goal into a clear next step. Keep it lightweight, show what you understood, and ask before running work.",
  },
  {
    label: "Create output",
    prompt: "Create a useful first version, tell me where it will be saved, and keep proof visible if work is run.",
  },
  {
    label: "Review work",
    prompt: "Review this work, summarize what matters, identify risks, and ask before taking action.",
  },
];
const SAVED_ACTIONS_KEY = "mycelis-soma-saved-actions";

function subscribeSavedActions(onChange: () => void) {
  window.addEventListener("storage", onChange);
  return () => window.removeEventListener("storage", onChange);
}

function readSavedActionsSnapshot() {
  return window.localStorage.getItem(SAVED_ACTIONS_KEY) ?? "[]";
}

function parseSavedActions(raw: string): SomaPinnedAction[] {
  try {
    const parsed: unknown = JSON.parse(raw);
    return Array.isArray(parsed) ? parsed.filter(isSavedAction).slice(0, 2) : [];
  } catch {
    return [];
  }
}

export function SomaActionShelf({
  actions = DEFAULT_PINNED_ACTIONS,
  onRunAction,
}: {
  actions?: readonly SomaPinnedAction[];
  onRunAction: (prompt: string) => void;
}) {
  const localSnapshot = useSyncExternalStore(subscribeSavedActions, readSavedActionsSnapshot, () => "[]");
  const localActions = useMemo(() => parseSavedActions(localSnapshot), [localSnapshot]);
  const [loadedActions, setLoadedActions] = useState<SomaPinnedAction[] | null>(null);
  const savedActions = loadedActions ?? localActions;
  const isClientReady = useClientReady();
  const [studioOpen, setStudioOpen] = useState(false);
  const visibleActions = useMemo(() => [...savedActions, ...actions].slice(0, 3), [actions, savedActions]);

  useEffect(() => {
    let cancelled = false;
    loadBackendActions()
      .then((actions) => {
        if (!cancelled && actions.length > 0) {
          setLoadedActions(actions);
          persistLocalActions(actions);
        }
      })
      .catch(() => {
        // Local saved actions remain available if the runtime is unavailable.
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const saveAction = async (action: SomaPinnedAction) => {
    const next = [action, ...savedActions.filter((item) => item.label !== action.label)].slice(0, 2);
    setLoadedActions(next);
    setStudioOpen(false);
    persistLocalActions(next);
    try {
      const saved = await saveBackendAction(action);
      if (saved) {
        const backendNext = [saved, ...savedActions.filter((item) => item.label !== saved.label)].slice(0, 2);
        setLoadedActions(backendNext);
        persistLocalActions(backendNext);
      }
    } catch {
      // Local fallback preserves the user's action when Core is unavailable.
    }
  };

  const runAction = async (action: SomaPinnedAction) => {
    if (!action.id) {
      onRunAction(action.prompt);
      return;
    }
    try {
      const rendered = await instantiateBackendAction(action.id);
      onRunAction(rendered || action.prompt);
    } catch {
      onRunAction(action.prompt);
    }
  };

  return (
    <section
      className="border-b border-cortex-border bg-cortex-surface/95 px-3 py-1.5 lg:px-4"
      aria-label="Pinned Soma actions"
      data-testid="soma-action-shelf"
      data-hydrated={isClientReady ? "true" : "false"}
    >
      <div className="flex min-w-0 flex-wrap items-center gap-1.5">
        <div className="mr-0.5 shrink-0 text-[9px] font-bold uppercase tracking-[0.12em] text-cortex-text-muted">
          Start with
        </div>
        {visibleActions.map((action) => (
          <button
            key={`${action.id || (action.userSaved ? "saved" : "default")}:${action.label}`}
            type="button"
            onClick={() => runAction(action)}
            className="inline-flex h-7 min-w-0 max-w-full flex-1 basis-[9rem] items-center justify-center gap-1.5 rounded-md border border-cortex-border bg-cortex-bg px-2 text-[10px] font-semibold text-cortex-text-main shadow-sm transition hover:border-cortex-primary/50 hover:bg-cortex-primary/10 focus:outline-none focus:ring-2 focus:ring-cortex-primary/30 sm:max-w-[13rem] sm:flex-none"
          >
            <Bolt className="h-3 w-3 shrink-0 text-cortex-warning" />
            <span className="truncate">{action.label}</span>
          </button>
        ))}
        <button
          type="button"
          onClick={() => setStudioOpen(true)}
          className="inline-flex h-7 min-w-0 max-w-full flex-1 basis-[7rem] items-center justify-center gap-1.5 rounded-md border border-dashed border-cortex-border bg-cortex-bg px-2 text-[10px] font-semibold text-cortex-text-muted transition hover:border-cortex-primary/50 hover:text-cortex-text-main sm:max-w-[10rem] sm:flex-none"
          aria-label="Create new quick action"
        >
          <Plus className="h-3 w-3 shrink-0" />
          <span className="truncate">Create ask</span>
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
      outputFormat: format.trim(),
      approvalBehavior: approval,
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

export { SAVED_ACTIONS_KEY };
