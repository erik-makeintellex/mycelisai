export type SomaSuggestion = {
  label: string;
  detail: string;
  prompt: string;
};

export const DEFAULT_SOMA_SUGGESTIONS: SomaSuggestion[] = [
  {
    label: "Plan something",
    detail: "Shape a goal into next steps.",
    prompt: "Help me plan the next useful step and show what you understood.",
  },
  {
    label: "Research something",
    detail: "Search or review sources, then summarize.",
    prompt: "Research this, cite sources, and tell me what changed.",
  },
  {
    label: "Create something",
    detail: "Draft an output and store it visibly.",
    prompt: "Create a first version and tell me where the output was stored.",
  },
  {
    label: "Review something",
    detail: "Check work, risks, and approvals.",
    prompt: "Review this, identify the risks, and ask before taking action.",
  },
  {
    label: "Configure tools",
    detail: "Check tools and guide setup.",
    prompt: "Check available tools and walk me through enabling what is missing.",
  },
];

export function SomaSuggestionBar({
  suggestions = DEFAULT_SOMA_SUGGESTIONS,
}: {
  suggestions?: readonly SomaSuggestion[];
}) {
  const visibleSuggestions = suggestions.slice(0, 3);

  return (
    <div className="w-full max-w-2xl px-1 text-center">
      <p className="text-xs leading-5 text-cortex-text-muted">
        Ask in plain language. Soma will clarify, propose, or start only when it is safe to run.
      </p>
      <div className="mt-2 flex flex-wrap justify-center gap-x-3 gap-y-1 text-[11px] leading-5 text-cortex-text-main">
        {visibleSuggestions.map((item) => (
          <figure key={item.label} className="max-w-[16rem] min-w-0">
            <blockquote className="truncate">
              &ldquo;{item.prompt}&rdquo;
            </blockquote>
          </figure>
        ))}
      </div>
    </div>
  );
}
