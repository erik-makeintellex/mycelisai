import { X } from "lucide-react";
import type { MissionChatContinuationContext } from "@/store/cortexStoreTypes";

export function MissionControlContinuationChip({
    context,
    onClear,
}: {
    context: MissionChatContinuationContext;
    onClear: () => void;
}) {
    return (
        <div
            aria-live="polite"
            className="mb-2 inline-flex max-w-full items-center gap-2 rounded-full border border-cortex-primary/25 bg-cortex-primary/10 px-3 py-1 text-xs text-cortex-text-muted"
        >
            <span className="min-w-0 truncate">
                Continuing from <span className="font-semibold text-cortex-text-main">{context.title}</span>
            </span>
            <button
                type="button"
                onClick={onClear}
                aria-label={`Clear continuation from ${context.title}`}
                className="rounded-full p-0.5 text-cortex-text-muted transition-colors hover:bg-cortex-primary/15 hover:text-cortex-text-main focus:outline-none focus:ring-2 focus:ring-cortex-primary/40"
            >
                <X className="h-3.5 w-3.5" />
            </button>
        </div>
    );
}
