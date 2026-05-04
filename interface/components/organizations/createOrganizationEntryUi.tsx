"use client";

import Link from "next/link";
import { ArrowRight, Loader2, RefreshCcw } from "lucide-react";

export function StartModeCard({
    active,
    title,
    description,
    icon,
    onClick,
}: {
    active: boolean;
    title: string;
    description: string;
    icon: React.ReactNode;
    onClick: () => void;
}) {
    return (
        <button onClick={onClick} className={`rounded-2xl border p-4 text-left transition-colors ${active ? "border-cortex-primary/40 bg-cortex-primary/10" : "border-cortex-border bg-cortex-bg hover:border-cortex-primary/20"}`}>
            <div className="flex items-start gap-3">
                <div className="rounded-xl border border-cortex-border bg-cortex-surface p-2 text-cortex-primary">{icon}</div>
                <div>
                    <p className="text-base font-semibold text-cortex-text-main">{title}</p>
                    <p className="mt-1 text-sm leading-6 text-cortex-text-muted">{description}</p>
                </div>
            </div>
        </button>
    );
}

export function Metric({ label, value }: { label: string; value: string }) {
    return (
        <div className="rounded-xl border border-cortex-border bg-cortex-surface/60 px-3 py-2">
            <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">{label}</p>
            <p className="mt-1 text-sm font-medium text-cortex-text-main">{value}</p>
        </div>
    );
}

export function HiddenLater({ label, icon }: { label: string; icon: React.ReactNode }) {
    return (
        <div className="flex items-center gap-2 rounded-xl border border-cortex-border px-3 py-2">
            <span className="text-cortex-primary">{icon}</span>
            <span>{label}</span>
        </div>
    );
}

export function LoadingState({ label }: { label: string }) {
    return (
        <div className="flex items-center gap-3 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4 text-sm text-cortex-text-muted">
            <Loader2 className="h-4 w-4 animate-spin" />
            <span>{label}</span>
        </div>
    );
}

export function ActionableState({
    title,
    message,
    guidance,
    actions,
}: {
    title: string;
    message: string;
    guidance: string;
    actions?: React.ReactNode;
}) {
    return (
        <div className="rounded-2xl border border-cortex-danger/30 bg-cortex-danger/10 px-4 py-4 text-sm text-cortex-text-main">
            <p className="font-semibold">{title}</p>
            <p className="mt-2 leading-6">{message}</p>
            <p className="mt-2 text-cortex-text-muted">{guidance}</p>
            {actions && <div className="mt-4 flex flex-wrap gap-3">{actions}</div>}
        </div>
    );
}

export function EmptyState({
    title,
    detail,
    actions,
}: {
    title: string;
    detail: string;
    actions?: React.ReactNode;
}) {
    return (
        <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4">
            <p className="text-sm font-semibold text-cortex-text-main">{title}</p>
            <p className="mt-1 text-sm leading-6 text-cortex-text-muted">{detail}</p>
            {actions && <div className="mt-4 flex flex-wrap gap-3">{actions}</div>}
        </div>
    );
}

export function ActionButton({
    children,
    onClick,
    kind,
    disabled = false,
    showRetryIcon = false,
}: {
    children: React.ReactNode;
    onClick: () => void;
    kind: "primary" | "secondary" | "ghost";
    disabled?: boolean;
    showRetryIcon?: boolean;
}) {
    const classes =
        kind === "primary"
            ? "border border-cortex-primary/35 bg-cortex-primary/10 text-cortex-primary hover:bg-cortex-primary/15"
            : kind === "secondary"
              ? "border border-cortex-border bg-cortex-bg text-cortex-text-main hover:border-cortex-primary/20"
              : "border border-transparent bg-transparent text-cortex-primary hover:bg-cortex-primary/10";

    return (
        <button onClick={onClick} disabled={disabled} className={`inline-flex items-center gap-2 rounded-xl px-3 py-2 text-sm font-medium transition-colors disabled:cursor-not-allowed disabled:opacity-50 ${classes}`}>
            {showRetryIcon && <RefreshCcw className="h-4 w-4" />}
            {children}
        </button>
    );
}

export function WhyStartHere() {
    return (
        <section className="rounded-3xl border border-cortex-border bg-cortex-surface p-6 text-sm text-cortex-text-muted">
            <p className="font-medium text-cortex-text-main">Why start here</p>
            <p className="mt-2 max-w-4xl leading-7">
                Create the AI Organization first so Mycelis opens with structure, not a one-off assistant session. The organization home keeps the Team Lead, Advisors, Departments, and Specialists visible from the beginning.
            </p>
            <div className="mt-4">
                <Link href="/docs?doc=v8-ui-api-operator-experience-contract" className="inline-flex items-center gap-2 text-cortex-primary hover:underline">
                    Learn about AI Organizations
                    <ArrowRight className="h-4 w-4" />
                </Link>
            </div>
        </section>
    );
}
