"use client";

import { type ReactNode, useEffect, useState } from "react";
import { GitBranch, MessageSquareMore, PackageSearch, RefreshCw } from "lucide-react";

type ExchangeChannel = {
    id: string;
    name: string;
    type: string;
    schema_id: string;
    visibility: string;
    sensitivity_class?: string;
    owner: string;
};

type ExchangeThread = {
    id: string;
    channel_name?: string;
    thread_type: string;
    title: string;
    status: string;
    participants: string[];
    allowed_reviewers?: string[];
};

type ExchangeItem = {
    id: string;
    channel_name?: string;
    schema_id: string;
    summary: string;
    created_by: string;
    created_at: string;
    sensitivity_class?: string;
    trust_class?: string;
    capability_id?: string;
    review_required?: boolean;
};

export default function ExchangeInspector() {
    const [channels, setChannels] = useState<ExchangeChannel[]>([]);
    const [threads, setThreads] = useState<ExchangeThread[]>([]);
    const [items, setItems] = useState<ExchangeItem[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const load = async () => {
        setLoading(true);
        try {
            const [channelRes, threadRes, itemRes] = await Promise.all([
                fetch("/api/v1/exchange/channels"),
                fetch("/api/v1/exchange/threads?limit=12"),
                fetch("/api/v1/exchange/items?limit=12"),
            ]);
            if (!channelRes.ok || !threadRes.ok || !itemRes.ok) {
                throw new Error("Managed exchange is unavailable.");
            }
            const [channelData, threadData, itemData] = await Promise.all([
                channelRes.json(),
                threadRes.json(),
                itemRes.json(),
            ]);
            setChannels(Array.isArray(channelData) ? channelData : []);
            setThreads(Array.isArray(threadData) ? threadData : []);
            setItems(Array.isArray(itemData) ? itemData : []);
            setError(null);
        } catch (err) {
            setError(err instanceof Error ? err.message : "Managed exchange failed to load.");
            setChannels([]);
            setThreads([]);
            setItems([]);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        load();
    }, []);

    return (
        <div className="h-full flex flex-col bg-cortex-bg">
            <div className="h-12 border-b border-cortex-border bg-cortex-surface/50 backdrop-blur-sm flex items-center justify-between px-6 flex-shrink-0">
                <div>
                    <h2 className="text-xs font-mono font-bold uppercase tracking-wider text-cortex-text-muted">
                        Managed Exchange
                    </h2>
                    <p className="text-[11px] text-cortex-text-muted mt-1">
                        Inspect governed channels, active threads, and recent outputs crossing teams, automations, and tools.
                    </p>
                </div>
                <button
                    onClick={load}
                    className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-cortex-primary/10 border border-cortex-primary/30 text-xs font-mono font-bold text-cortex-primary hover:bg-cortex-primary/20 transition-colors"
                >
                    <RefreshCw className={`w-3.5 h-3.5 ${loading ? "animate-spin" : ""}`} />
                    REFRESH
                </button>
            </div>

            <div className="flex-1 overflow-y-auto">
                {error ? (
                    <div className="m-6 rounded-xl border border-cortex-danger/30 bg-cortex-danger/5 p-4">
                        <p className="text-sm text-cortex-danger">Managed exchange unavailable</p>
                        <p className="text-xs text-cortex-text-muted mt-1">{error}</p>
                    </div>
                ) : (
                    <div className="grid grid-cols-1 xl:grid-cols-3 gap-4 p-6 max-w-7xl mx-auto">
                        <Section
                            icon={<GitBranch className="w-4 h-4 text-cortex-primary" />}
                            title="Channels"
                            subtitle="Named work, review, learning, and tool lanes"
                            empty="No channels registered yet."
                            loading={loading}
                        >
                            {channels.map((channel) => (
                                <div key={channel.id} className="rounded-xl border border-cortex-border bg-cortex-surface p-3 space-y-1">
                                    <div className="flex items-center justify-between gap-2">
                                        <span className="text-xs font-semibold text-cortex-text-main break-all">{channel.name}</span>
                                        <span className="text-[10px] font-mono uppercase px-1.5 py-0.5 rounded bg-cortex-primary/15 text-cortex-primary">
                                            {channel.type}
                                        </span>
                                    </div>
                                    <p className="text-[11px] text-cortex-text-muted">
                                        {channel.schema_id} · owner {channel.owner} · {channel.visibility}
                                    </p>
                                    {channel.sensitivity_class ? (
                                        <p className="text-[10px] text-cortex-text-muted">
                                            sensitivity {channel.sensitivity_class}
                                        </p>
                                    ) : null}
                                </div>
                            ))}
                        </Section>

                        <Section
                            icon={<MessageSquareMore className="w-4 h-4 text-cortex-info" />}
                            title="Threads"
                            subtitle="Ordered planning, work, review, escalation, and learning traces"
                            empty="No threads created yet."
                            loading={loading}
                        >
                            {threads.map((thread) => (
                                <div key={thread.id} className="rounded-xl border border-cortex-border bg-cortex-surface p-3 space-y-1">
                                    <div className="flex items-center justify-between gap-2">
                                        <span className="text-xs font-semibold text-cortex-text-main">{thread.title}</span>
                                        <span className="text-[10px] font-mono uppercase px-1.5 py-0.5 rounded bg-cortex-info/15 text-cortex-info">
                                            {thread.status}
                                        </span>
                                    </div>
                                    <p className="text-[11px] text-cortex-text-muted">
                                        {thread.channel_name ?? "unbound"} · {thread.thread_type}
                                    </p>
                                    {thread.participants?.length > 0 && (
                                        <p className="text-[10px] text-cortex-text-muted">
                                            {thread.participants.join(", ")}
                                        </p>
                                    )}
                                    {thread.allowed_reviewers?.length ? (
                                        <p className="text-[10px] text-cortex-text-muted">
                                            reviewers {thread.allowed_reviewers.join(", ")}
                                        </p>
                                    ) : null}
                                </div>
                            ))}
                        </Section>

                        <Section
                            icon={<PackageSearch className="w-4 h-4 text-cortex-warning" />}
                            title="Recent Outputs"
                            subtitle="Latest normalized exchange items available to Soma and advanced review"
                            empty="No managed outputs have been published yet."
                            loading={loading}
                        >
                            {items.map((item) => (
                                <div key={item.id} className="rounded-xl border border-cortex-border bg-cortex-surface p-3 space-y-1">
                                    <div className="flex items-center justify-between gap-2">
                                        <span className="text-xs font-semibold text-cortex-text-main">
                                            {item.summary || "Untitled exchange item"}
                                        </span>
                                        <span className="text-[10px] font-mono uppercase px-1.5 py-0.5 rounded bg-cortex-warning/15 text-cortex-warning">
                                            {item.schema_id}
                                        </span>
                                    </div>
                                    <p className="text-[11px] text-cortex-text-muted">
                                        {item.channel_name ?? "unbound"} · {item.created_by}
                                    </p>
                                    <p className="text-[10px] text-cortex-text-muted">
                                        {item.sensitivity_class ?? "role_scoped"} · {item.trust_class ?? "trusted_internal"}
                                        {item.capability_id ? ` · ${item.capability_id}` : ""}
                                        {item.review_required ? " · review required" : ""}
                                    </p>
                                    <p className="text-[10px] text-cortex-text-muted">
                                        {formatTimestamp(item.created_at)}
                                    </p>
                                </div>
                            ))}
                        </Section>
                    </div>
                )}
            </div>
        </div>
    );
}

function Section({
    icon,
    title,
    subtitle,
    empty,
    loading,
    children,
}: {
    icon: ReactNode;
    title: string;
    subtitle: string;
    empty: string;
    loading: boolean;
    children: ReactNode;
}) {
    const count = Array.isArray(children) ? children.length : 0;

    return (
        <div className="rounded-2xl border border-cortex-border bg-cortex-bg/60 min-h-[320px]">
            <div className="border-b border-cortex-border px-4 py-3">
                <div className="flex items-center gap-2">
                    {icon}
                    <h3 className="text-sm font-semibold text-cortex-text-main">{title}</h3>
                </div>
                <p className="text-[11px] text-cortex-text-muted mt-1">{subtitle}</p>
            </div>
            <div className="p-4 space-y-3">
                {loading ? (
                    <p className="text-xs font-mono text-cortex-text-muted animate-pulse">Loading…</p>
                ) : count === 0 ? (
                    <p className="text-xs text-cortex-text-muted">{empty}</p>
                ) : children}
            </div>
        </div>
    );
}

function formatTimestamp(value: string): string {
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
        return value;
    }
    return date.toLocaleString();
}
