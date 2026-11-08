import type { LucideIcon } from "lucide-react";
import { Compass, FileText, Gamepad2, Megaphone, Search, Settings2, Sparkles, Terminal } from "lucide-react";

export type SomaSuggestion = {
  label: string;
  detail: string;
  prompt: string;
  icon?: LucideIcon;
};

export const DEFAULT_SOMA_SUGGESTIONS: SomaSuggestion[] = [
  {
    label: "Plan something",
    detail: "Shape a goal into next steps.",
    prompt: "Help me plan the next useful step and show what you understood.",
    icon: Compass,
  },
  {
    label: "Research something",
    detail: "Search or review sources, then summarize.",
    prompt: "Research this, cite sources, and tell me what changed.",
    icon: Search,
  },
  {
    label: "Create something",
    detail: "Draft an output and store it visibly.",
    prompt: "Create a first version and tell me where the output was stored.",
    icon: Sparkles,
  },
  {
    label: "Review something",
    detail: "Check work, risks, and approvals.",
    prompt: "Review this, identify the risks, and ask before taking action.",
    icon: FileText,
  },
  {
    label: "Configure tools",
    detail: "Check tools and guide setup.",
    prompt: "Check available tools and walk me through enabling what is missing.",
    icon: Settings2,
  },
];

export const POPULAR_DELIVERABLE_STARTERS: SomaSuggestion[] = [
  {
    label: "Marketing Launch Kit",
    detail: "Positioning brief, launch email, and FAQ.",
    prompt: "Create marketing documentation for our product launch including a positioning brief, launch email, and customer FAQ.",
    icon: Megaphone,
  },
  {
    label: "Playable Puzzle Game",
    detail: "Interactive browser sliding-tile game + rules.",
    prompt: "Create a simple browser sliding-tile puzzle game with playable HTML, game rules, and controls.",
    icon: Gamepad2,
  },
  {
    label: "Engineering Task Pack",
    detail: "Scoped implementation spec and checklist.",
    prompt: "Create a Codex task package for MYC-RC-005 with task steps and validation checklist.",
    icon: Terminal,
  },
];

export function SomaSuggestionBar({
  suggestions = DEFAULT_SOMA_SUGGESTIONS,
  starters = POPULAR_DELIVERABLE_STARTERS,
  onSelect,
}: {
  suggestions?: readonly SomaSuggestion[];
  starters?: readonly SomaSuggestion[];
  onSelect: (prompt: string) => void;
}) {
  return (
    <div className="rounded-2xl border border-cortex-border bg-cortex-bg/70 p-3 space-y-3">
      <div>
        <p className="text-[11px] font-mono uppercase tracking-[0.18em] text-cortex-primary">
          Start with Soma
        </p>
        <div className="mt-3 grid gap-2 md:grid-cols-5">
          {suggestions.map((item) => {
            const Icon = item.icon ?? Sparkles;
            return (
              <button
                key={item.label}
                type="button"
                onClick={() => onSelect(item.prompt)}
                className="rounded-xl border border-cortex-border bg-cortex-surface px-3 py-2 text-left transition hover:border-cortex-primary/30 hover:bg-cortex-primary/5"
              >
                <span className="flex items-center gap-2 text-xs font-semibold text-cortex-text-main">
                  <Icon className="h-3.5 w-3.5 text-cortex-primary" />
                  {item.label}
                </span>
                <span className="mt-2 block text-xs leading-5 text-cortex-text-muted">
                  {item.detail}
                </span>
              </button>
            );
          })}
        </div>
      </div>

      {starters.length > 0 && (
        <div className="border-t border-cortex-border/70 pt-2.5">
          <p className="text-[10px] font-mono uppercase tracking-[0.16em] text-cortex-text-muted">
            Popular deliverable starters
          </p>
          <div className="mt-2 grid gap-2 md:grid-cols-3">
            {starters.map((item, index) => {
              const Icon = item.icon ?? Sparkles;
              return (
                <button
                  key={item.label}
                  type="button"
                  data-testid={`deliverable-starter-${index}`}
                  onClick={() => onSelect(item.prompt)}
                  className="rounded-xl border border-cortex-primary/20 bg-cortex-primary/5 px-3 py-2 text-left transition hover:border-cortex-primary/40 hover:bg-cortex-primary/10"
                >
                  <span className="flex items-center gap-2 text-xs font-semibold text-cortex-text-main">
                    <Icon className="h-3.5 w-3.5 text-cortex-primary" />
                    {item.label}
                  </span>
                  <span className="mt-1 block text-xs leading-4 text-cortex-text-muted">
                    {item.detail}
                  </span>
                </button>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
}
