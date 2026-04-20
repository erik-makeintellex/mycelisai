"use client";

import { use, type ReactNode } from "react";
import { ArrowLeft, GitBranch, List, MessageSquare, Zap } from "lucide-react";
import ViewChain from "@/components/runs/ViewChain";

type TabKey = "conversation" | "events" | "chain";

export default function RunChainPage({ params }: { params: Promise<{ id: string }> }) {
    const { id } = use(params);

    const tabs: Array<{ key: TabKey; label: string; href: string; icon: ReactNode }> = [
        { key: "conversation", label: "Conversation", href: `/runs/${id}`, icon: <MessageSquare className="w-3.5 h-3.5" /> },
        { key: "events", label: "Events", href: `/runs/${id}?tab=events`, icon: <List className="w-3.5 h-3.5" /> },
        { key: "chain", label: "Chain", href: `/runs/${id}/chain`, icon: <GitBranch className="w-3.5 h-3.5" /> },
    ];

    return (
        <div className="min-h-screen bg-cortex-bg text-cortex-text-main">
            <div className="sticky top-0 z-10 bg-cortex-bg border-b border-cortex-border px-4 py-3 flex items-center gap-3">
                <a
                    href="/dashboard"
                    className="flex items-center gap-1.5 text-cortex-text-muted hover:text-cortex-primary transition-colors text-[11px] font-mono"
                >
                    <ArrowLeft className="w-3.5 h-3.5" />
                    Workspace
                </a>

                <div className="w-px h-4 bg-cortex-border" />

                <div className="flex items-center gap-2">
                    <GitBranch className="w-3.5 h-3.5 text-cortex-primary" />
                    <span className="text-[11px] font-mono text-cortex-text-main font-bold">
                        Run Chain: {id.slice(0, 8)}...
                    </span>
                </div>

                <div className="ml-auto flex items-center gap-0.5 bg-cortex-surface rounded-md border border-cortex-border p-0.5">
                    {tabs.map((tab) => (
                        <a
                            key={tab.key}
                            href={tab.href}
                            className={`flex items-center gap-1.5 px-3 py-1 rounded text-[10px] font-mono font-bold transition-colors ${
                                tab.key === "chain"
                                    ? "bg-cortex-primary/15 text-cortex-primary"
                                    : "text-cortex-text-muted hover:text-cortex-text-main"
                            }`}
                        >
                            {tab.icon}
                            {tab.label}
                        </a>
                    ))}
                </div>

                <a
                    href={`/runs/${id}`}
                    className="flex items-center gap-1.5 rounded-md border border-cortex-border bg-cortex-surface px-3 py-1 text-[10px] font-mono font-bold text-cortex-text-muted hover:text-cortex-primary transition-colors"
                >
                    <Zap className="w-3.5 h-3.5" />
                    Run
                </a>
            </div>

            <div className="max-w-2xl mx-auto px-4 py-6">
                <ViewChain runId={id} />
            </div>
        </div>
    );
}
