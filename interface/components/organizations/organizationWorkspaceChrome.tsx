import { ArrowRight, Sparkles } from "lucide-react";

export type GuidedWorkspaceCardDefinition = {
    eyebrow: string;
    title: string;
    summary: string;
    buttonLabel: string;
    onClick: () => void;
};

export function Metric({ label, value }: { label: string; value: string }) {
    return (
        <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3">
            <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">{label}</p>
            <p className="mt-1 text-sm font-medium text-cortex-text-main">{value}</p>
        </div>
    );
}

export function HelpPill({ label }: { label: string }) {
    return (
        <div className="inline-flex items-center gap-2 rounded-full border border-cortex-border bg-cortex-bg px-3 py-2 text-sm text-cortex-text-main">
            <Sparkles className="h-4 w-4 text-cortex-primary" />
            <span>{label}</span>
        </div>
    );
}

export function ActionPill({
    label,
    isActive,
    onClick,
}: {
    label: string;
    isActive?: boolean;
    onClick: () => void;
}) {
    return (
        <button
            type="button"
            onClick={onClick}
            className={`inline-flex items-center gap-2 rounded-full border px-3 py-2 text-sm transition-colors ${
                isActive
                    ? "border-cortex-primary/40 bg-cortex-primary/10 text-cortex-text-main"
                    : "border-cortex-border bg-cortex-bg text-cortex-text-main hover:border-cortex-primary/20"
            }`}
        >
            <Sparkles className="h-4 w-4 text-cortex-primary" />
            <span>{label}</span>
        </button>
    );
}

export function GuidedWorkspaceCard({
    eyebrow,
    title,
    summary,
    buttonLabel,
    onClick,
}: GuidedWorkspaceCardDefinition) {
    return (
        <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4">
            <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-primary">{eyebrow}</p>
            <p className="mt-2 text-sm font-semibold text-cortex-text-main">{title}</p>
            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{summary}</p>
            <button
                type="button"
                onClick={onClick}
                className="mt-4 inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-surface px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/20 hover:text-cortex-primary"
            >
                {buttonLabel}
                <ArrowRight className="h-4 w-4" />
            </button>
        </div>
    );
}
