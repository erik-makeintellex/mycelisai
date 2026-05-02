"use client";

import Link from "next/link";
import { useEffect } from "react";
import type React from "react";
import { Globe, Wrench, ShieldCheck } from "lucide-react";
import { useCortexStore } from "@/store/useCortexStore";

export function SomaReadinessStrip() {
  const mcpServers = useCortexStore((s) => s.mcpServers);
  const mcpError = useCortexStore((s) => s.mcpServersError);
  const search = useCortexStore((s) => s.searchCapability);
  const searchError = useCortexStore((s) => s.searchCapabilityError);
  const fetchMCPServers = useCortexStore((s) => s.fetchMCPServers);
  const fetchSearchCapability = useCortexStore((s) => s.fetchSearchCapability);

  useEffect(() => {
    void fetchMCPServers();
    void fetchSearchCapability();
  }, [fetchMCPServers, fetchSearchCapability]);

  const installed = mcpServers.length;
  const searchReady = Boolean(search?.enabled && search.configured);
  return (
    <section className="grid gap-2 md:grid-cols-3">
      <ReadinessCard
        icon={<Globe className="h-4 w-4" />}
        label="Web search"
        value={searchReady ? search?.provider ?? "configured" : "needs setup"}
        tone={searchReady ? "ready" : "attention"}
        detail={searchError || search?.blocker?.message || "Soma can use configured search when requests need current information."}
      />
      <ReadinessCard
        icon={<Wrench className="h-4 w-4" />}
        label="Connected tools"
        value={mcpError ? "unreachable" : `${installed} installed`}
        tone={!mcpError && installed > 0 ? "ready" : "attention"}
        detail={mcpError || "Manage MCP servers, tool bindings, and recent tool activity from Resources."}
        href="/resources?tab=tools"
      />
      <ReadinessCard
        icon={<ShieldCheck className="h-4 w-4" />}
        label="Protected actions"
        value="confirmation first"
        tone="ready"
        detail="Team creation, private data, private services, recurring behavior, and tool changes should become explicit confirmation steps."
      />
    </section>
  );
}

function ReadinessCard({
  icon,
  label,
  value,
  detail,
  tone,
  href,
}: {
  icon: React.ReactNode;
  label: string;
  value: string;
  detail: string;
  tone: "ready" | "attention";
  href?: string;
}) {
  const body = (
    <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-3 py-3">
      <div className="flex items-center gap-2 text-cortex-primary">
        {icon}
        <p className="text-[10px] font-mono uppercase tracking-[0.16em]">{label}</p>
      </div>
      <p className={`mt-2 text-sm font-semibold ${tone === "ready" ? "text-cortex-success" : "text-amber-300"}`}>
        {value}
      </p>
      <p className="mt-1 line-clamp-2 text-xs leading-5 text-cortex-text-muted">
        {detail}
      </p>
    </div>
  );
  return href ? <Link href={href}>{body}</Link> : body;
}
