import type { LucideIcon } from "lucide-react";
import { Compass, FileText, Search, Settings2, Sparkles } from "lucide-react";

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

export function SomaSuggestionBar({
  suggestions = DEFAULT_SOMA_SUGGESTIONS,
  onSelect,
}: {
  suggestions?: readonly SomaSuggestion[];
  onSelect: (prompt: string) => void;
}) {
  return (
    <div className="w-full max-w-2xl rounded-xl border border-cortex-border bg-cortex-bg/60 px-3 py-2.5">
      <p className="text-[10px] font-mono uppercase tracking-[0.16em] text-cortex-primary">
        Try a starting phrase
      </p>
      <div className="mt-2 flex gap-2 overflow-x-auto pb-1 scrollbar-thin scrollbar-thumb-cortex-border">
        {suggestions.map((item) => {
          const Icon = item.icon ?? Sparkles;
          return (
            <button
              key={item.label}
              type="button"
              onClick={() => onSelect(item.prompt)}
              className="inline-flex min-h-9 shrink-0 items-center gap-2 rounded-full border border-cortex-border bg-cortex-surface px-3 py-1.5 text-left transition hover:border-cortex-primary/30 hover:bg-cortex-primary/5"
              title={item.detail}
            >
              <Icon className="h-3.5 w-3.5 shrink-0 text-cortex-primary" />
              <span className="whitespace-nowrap text-xs font-semibold text-cortex-text-main">
                {item.label}
              </span>
            </button>
          );
        })}
      </div>
    </div>
  );
}
