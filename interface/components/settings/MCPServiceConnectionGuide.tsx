"use client";

import type { ReactNode } from "react";

export function MCPServiceConnectionGuide() {
    return (
        <section className="rounded-xl border border-cortex-border bg-cortex-surface px-4 py-4">
            <div>
                <p className="text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-text-muted">
                    Service data connections
                </p>
                <p className="mt-1 text-sm font-semibold text-cortex-text-main">
                    Give Soma named access to databases and private services.
                </p>
                <p className="mt-1 text-xs leading-5 text-cortex-text-muted">
                    Service MCPs such as PostgreSQL should be configured as named connections, then scoped to Everyone, a Group, or a Host.
                </p>
            </div>
            <div className="mt-4 grid gap-3 md:grid-cols-3">
                <ConnectionStep title="1. Connect">
                    Add or configure a connector such as PostgreSQL with one secret-backed connection profile.
                </ConnectionStep>
                <ConnectionStep title="2. Name and scope">
                    Name the source in user language, then limit it to the workspace, a group, or a host.
                </ConnectionStep>
                <ConnectionStep title="3. Let Soma cite it">
                    Soma should say which named source it used. The system-owned Mycelis database stays reserved unless explicitly exposed.
                </ConnectionStep>
            </div>
        </section>
    );
}

function ConnectionStep({ children, title }: { children: ReactNode; title: string }) {
    return (
        <div className="rounded-lg border border-cortex-border bg-cortex-bg/60 px-3 py-3">
            <p className="text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-primary">{title}</p>
            <p className="mt-2 text-xs leading-5 text-cortex-text-main">{children}</p>
        </div>
    );
}
