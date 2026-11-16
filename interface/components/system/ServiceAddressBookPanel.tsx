"use client";

import React, { useEffect, useState } from "react";
import { AlertTriangle, CheckCircle2, Globe, Network } from "lucide-react";

export interface ServiceAddressEntry {
    name: string;
    address: string;
    protocol: string;
    source: string;
    loopback_warning?: string;
}

export interface ServiceAddressBookPayload {
    cluster_mode: boolean;
    entries: ServiceAddressEntry[];
    warnings?: string[];
}

const ADDRESS_LABELS: Record<string, string> = {
    nats: "NATS Message Bus",
    postgres: "PostgreSQL + pgvector",
    text_ai: "Text AI Engine",
    media_ai: "Media AI Engine",
    search: "Governed Search Provider",
    mcp_gateway: "MCP Gateway",
    cors_origin: "Interface CORS Origin",
};

export default function ServiceAddressBookPanel() {
    const [book, setBook] = useState<ServiceAddressBookPayload | null>(null);

    useEffect(() => {
        let cancelled = false;
        async function loadAddresses() {
            try {
                const res = await fetch("/api/v1/services/addresses");
                if (!res.ok) return;
                const payload = await res.json();
                const data = payload?.data ?? payload;
                if (!cancelled && data && Array.isArray(data.entries)) {
                    setBook(data as ServiceAddressBookPayload);
                }
            } catch {
                // Keep panel silent when endpoint is not mocked or offline
            }
        }
        loadAddresses();
        return () => {
            cancelled = true;
        };
    }, []);

    if (!book || book.entries.length === 0) return null;

    return (
        <section
            data-testid="service-address-book-panel"
            className="rounded-lg border border-cortex-border bg-cortex-surface/50 p-4 space-y-3"
        >
            <div className="flex flex-wrap items-center justify-between gap-2">
                <div className="flex items-center gap-2">
                    <Network className="w-4 h-4 text-cortex-primary" />
                    <h4 className="text-xs font-mono font-bold uppercase tracking-wider text-cortex-text-main">
                        Backend Service Address Book
                    </h4>
                </div>
                <span
                    data-testid="service-address-cluster-badge"
                    className="inline-flex items-center gap-1 rounded-full border border-cortex-primary/30 bg-cortex-primary/10 px-2.5 py-0.5 text-[10px] font-mono text-cortex-primary"
                >
                    <Globe className="w-3 h-3" />
                    {book.cluster_mode ? "Kubernetes Cluster Mode" : "Local / Host Mode"}
                </span>
            </div>

            {book.warnings && book.warnings.length > 0 && (
                <div
                    data-testid="service-address-warnings"
                    className="rounded border border-cortex-warning/35 bg-cortex-warning/10 p-2.5 text-xs text-cortex-warning space-y-1"
                >
                    {book.warnings.map((w) => (
                        <div key={w} className="flex items-start gap-1.5 font-mono text-[11px]">
                            <AlertTriangle className="w-3.5 h-3.5 flex-shrink-0 mt-0.5" />
                            <span>{w}</span>
                        </div>
                    ))}
                </div>
            )}

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                {book.entries.map((entry) => (
                    <div
                        key={entry.name}
                        data-testid={`address-book-entry-${entry.name}`}
                        className="rounded border border-cortex-border bg-cortex-bg px-3 py-2 space-y-1"
                    >
                        <div className="flex items-center justify-between gap-2">
                            <span className="text-xs font-semibold text-cortex-text-main">
                                {ADDRESS_LABELS[entry.name] ?? entry.name}
                            </span>
                            <span className="inline-flex items-center gap-1 text-[10px] font-mono text-cortex-text-muted">
                                <CheckCircle2 className="w-3 h-3 text-cortex-success" />
                                {entry.source}
                            </span>
                        </div>
                        <p className="text-[11px] font-mono text-cortex-primary truncate" title={entry.address || "unconfigured"}>
                            {entry.address || "(optional / unconfigured)"}
                        </p>
                    </div>
                ))}
            </div>
        </section>
    );
}
