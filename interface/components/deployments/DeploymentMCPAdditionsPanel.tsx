"use client";

import { useState } from "react";
import Link from "next/link";
import { useCortexStore } from "@/store/useCortexStore";
import type { DeploymentRecord } from "./deploymentTypes";

export interface DeploymentMCPAdditionOption {
  id: string;
  title: string;
  group: "standard" | "custom";
  transport: string;
  category: string;
  summary: string;
  toolCue: string;
}

export const STANDARD_DEPLOYMENT_MCPS: DeploymentMCPAdditionOption[] = [
  { id: "filesystem", title: "Filesystem", group: "standard", transport: "stdio", category: "Workspace & Files", summary: "Read, write, and manage workspace files and deliverable folders.", toolCue: "file" },
  { id: "github", title: "GitHub", group: "standard", transport: "stdio", category: "Code & PRs", summary: "Repositories, pull requests, issues, and code search.", toolCue: "github" },
  { id: "git", title: "Git Repository", group: "standard", transport: "stdio", category: "Version Control", summary: "Local Git status, branch diffs, commits, and log inspection.", toolCue: "git" },
  { id: "fetch", title: "Fetch Web & Docs", group: "standard", transport: "stdio", category: "Web & Docs", summary: "Fetch live web pages and convert documentation to markdown.", toolCue: "fetch" },
  { id: "postgres", title: "PostgreSQL", group: "standard", transport: "stdio", category: "Database", summary: "Query and inspect relational PostgreSQL schemas and tables.", toolCue: "postgres" },
  { id: "sqlite", title: "SQLite", group: "standard", transport: "stdio", category: "Local DB", summary: "Embedded SQLite database querying and state management.", toolCue: "sqlite" },
  { id: "docker-kubernetes", title: "Docker & Kubernetes", group: "standard", transport: "stdio", category: "Cluster & Containers", summary: "Inspect container runtimes, K8s pods, logs, and deployment health.", toolCue: "kubernetes" },
  { id: "memory", title: "Memory Graph", group: "standard", transport: "stdio", category: "Knowledge", summary: "Persistent entity and relationship memory across runs.", toolCue: "memory" },
  { id: "puppeteer", title: "Browser Automation", group: "standard", transport: "stdio", category: "UI Verification", summary: "Headless browser navigation, screenshots, and E2E UI checks.", toolCue: "puppeteer" },
  { id: "sequential-thinking", title: "Sequential Thinking", group: "standard", transport: "stdio", category: "Reasoning", summary: "Structured multi-step problem decomposition and planning.", toolCue: "sequential" },
  { id: "brave-search", title: "Brave Web Search", group: "standard", transport: "stdio", category: "Search", summary: "Governed external web search for technical and market research.", toolCue: "brave" },
  { id: "slack", title: "Slack Workspace", group: "standard", transport: "stdio", category: "Messaging", summary: "Post deployment updates and read team channels.", toolCue: "slack" },
];

export const CUSTOM_DEVELOPED_MCPS: DeploymentMCPAdditionOption[] = [
  { id: "custom-python-fastmcp", title: "Custom Python FastMCP Server", group: "custom", transport: "stdio (python)", category: "Python FastMCP", summary: "Custom-developed Python FastMCP server for domain tools and internal data adapters.", toolCue: "custom_python" },
  { id: "custom-ts-mcp-sdk", title: "Custom TypeScript MCP SDK Server", group: "custom", transport: "stdio (node)", category: "TypeScript SDK", summary: "Custom-developed Node/TypeScript @modelcontextprotocol/sdk server for workspace automations.", toolCue: "custom_ts" },
  { id: "custom-http-sse-mcp", title: "Custom HTTP / SSE Microservice MCP", group: "custom", transport: "sse (http)", category: "HTTP / SSE Endpoint", summary: "In-cluster or remote custom MCP service exposed over HTTP Server-Sent Events.", toolCue: "custom_http" },
  { id: "custom-nats-service-bridge", title: "Custom Mycelis NATS & Service Bridge", group: "custom", transport: "stdio / nats", category: "NATS Bridge", summary: "Exposes team NATS request/result channels and internal REST APIs as governed MCP tools.", toolCue: "custom_nats" },
];

function isMCPAlreadyBound(record: DeploymentRecord, option: DeploymentMCPAdditionOption): boolean {
  const hay = record.members
    .flatMap((m) => m.tools.map((t) => t.name.toLowerCase()))
    .join(" ");
  return hay.includes(option.id.toLowerCase()) || hay.includes(option.toolCue.toLowerCase());
}

export function DeploymentMCPAdditionsPanel({ record }: { record: DeploymentRecord }) {
  const [activeGroup, setActiveGroup] = useState<"standard" | "custom">("standard");
  const [customSlug, setCustomSlug] = useState("");
  const [customTransport, setCustomTransport] = useState("stdio (python)");
  const mcpServers = useCortexStore((s) => s.mcpServers);

  const standardIds = new Set(STANDARD_DEPLOYMENT_MCPS.map((item) => item.id));
  const userRegisteredCustomServers: DeploymentMCPAdditionOption[] = mcpServers
    .filter((srv) => !standardIds.has(srv.name))
    .map((srv) => ({
      id: srv.name,
      title: `${srv.name} (Registered Custom MCP)`,
      group: "custom",
      transport: srv.transport || "stdio",
      category: "Registered Custom",
      summary: `User-registered custom MCP server (${srv.status || "configured"}).`,
      toolCue: srv.name,
    }));

  const customList = [...CUSTOM_DEVELOPED_MCPS, ...userRegisteredCustomServers];
  const normalizedCustomId = customSlug.trim().toLowerCase().replace(/[^a-z0-9-]+/g, "-") || "custom-team-domain-mcp";
  const customTitle = customSlug.trim() || "Custom Team Domain MCP";

  return (
    <section className="rounded-xl border border-cortex-border bg-cortex-bg p-4" data-testid="deployments-mcp-additions-panel">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div>
          <h3 className="text-sm font-semibold text-cortex-text-main">MCP Additions (Standard &amp; Custom Developed)</h3>
          <p className="mt-0.5 text-xs text-cortex-text-muted">
            Attach curated standard deployment MCPs or scaffold and bind custom-developed MCP servers to <span className="font-semibold text-cortex-text-main">{record.name}</span>.
          </p>
        </div>
        <Link
          href="/resources?tab=tools"
          data-testid="deployments-open-mcp-registry"
          className="rounded-lg border border-cortex-border bg-cortex-surface px-2.5 py-1 font-mono text-[11px] text-cortex-text-main hover:border-cortex-primary/35 hover:text-cortex-primary"
        >
          Manage MCP Registry &rarr;
        </Link>
      </div>

      <div className="mt-3 flex flex-wrap gap-2" role="tablist" aria-label="MCP Addition Groupings">
        <button
          type="button"
          onClick={() => setActiveGroup("standard")}
          data-testid="deployments-mcp-tab-standard"
          className={`rounded-lg border px-3 py-1.5 text-xs font-semibold transition-colors ${
            activeGroup === "standard"
              ? "border-cortex-primary/50 bg-cortex-primary/15 text-cortex-primary"
              : "border-cortex-border bg-cortex-surface text-cortex-text-muted hover:text-cortex-text-main"
          }`}
        >
          Standard Deployment MCPs ({STANDARD_DEPLOYMENT_MCPS.length})
        </button>
        <button
          type="button"
          onClick={() => setActiveGroup("custom")}
          data-testid="deployments-mcp-tab-custom"
          className={`rounded-lg border px-3 py-1.5 text-xs font-semibold transition-colors ${
            activeGroup === "custom"
              ? "border-cortex-primary/50 bg-cortex-primary/15 text-cortex-primary"
              : "border-cortex-border bg-cortex-surface text-cortex-text-muted hover:text-cortex-text-main"
          }`}
        >
          Custom Developed MCPs ({customList.length})
        </button>
      </div>

      {activeGroup === "standard" ? (
        <div className="mt-3 grid gap-2.5 sm:grid-cols-2" data-testid="deployments-standard-mcp-group">
          {STANDARD_DEPLOYMENT_MCPS.map((item) => {
            const bound = isMCPAlreadyBound(record, item);
            return (
              <div
                key={item.id}
                data-testid={`deployment-mcp-card-${item.id}`}
                className="flex flex-col justify-between rounded-lg border border-cortex-border bg-cortex-surface p-3"
              >
                <div>
                  <div className="flex items-center justify-between gap-2">
                    <span className="text-xs font-semibold text-cortex-text-main">{item.title}</span>
                    <span className="rounded border border-cortex-border bg-cortex-bg px-1.5 py-0.5 font-mono text-[10px] text-cortex-text-muted">
                      {item.category}
                    </span>
                  </div>
                  <p className="mt-1 text-[11px] leading-4 text-cortex-text-muted">{item.summary}</p>
                </div>
                <div className="mt-2.5 flex items-center justify-between gap-2 pt-1">
                  <span className="font-mono text-[10px] text-cortex-text-muted">
                    {item.id} · {item.transport}
                  </span>
                  <Link
                    href={`/dashboard?team_id=${encodeURIComponent(record.id)}&team_name=${encodeURIComponent(record.name)}&team_action=add_mcp&mcp_id=${encodeURIComponent(item.id)}&mcp_title=${encodeURIComponent(item.title)}&mcp_group=standard`}
                    data-testid={`deployment-add-standard-mcp-${item.id}`}
                    className="rounded border border-cortex-primary/40 bg-cortex-primary/15 px-2.5 py-1 font-mono text-[10px] font-semibold text-cortex-primary hover:bg-cortex-primary/25"
                  >
                    {bound ? "Re-sync Binding \u2192" : "+ Add to Deployment"}
                  </Link>
                </div>
              </div>
            );
          })}
        </div>
      ) : (
        <div className="mt-3 space-y-3" data-testid="deployments-custom-mcp-group">
          <div className="grid gap-2.5 sm:grid-cols-2">
            {customList.map((item) => (
              <div
                key={item.id}
                data-testid={`deployment-custom-mcp-card-${item.id}`}
                className="flex flex-col justify-between rounded-lg border border-cortex-border bg-cortex-surface p-3"
              >
                <div>
                  <div className="flex items-center justify-between gap-2">
                    <span className="text-xs font-semibold text-cortex-text-main">{item.title}</span>
                    <span className="rounded border border-cortex-primary/30 bg-cortex-primary/10 px-1.5 py-0.5 font-mono text-[10px] text-cortex-primary">
                      {item.category}
                    </span>
                  </div>
                  <p className="mt-1 text-[11px] leading-4 text-cortex-text-muted">{item.summary}</p>
                </div>
                <div className="mt-2.5 flex items-center justify-between gap-2 pt-1">
                  <span className="font-mono text-[10px] text-cortex-text-muted">
                    {item.id} · {item.transport}
                  </span>
                  <Link
                    href={`/dashboard?team_id=${encodeURIComponent(record.id)}&team_name=${encodeURIComponent(record.name)}&team_action=add_mcp&mcp_id=${encodeURIComponent(item.id)}&mcp_title=${encodeURIComponent(item.title)}&mcp_group=custom`}
                    data-testid={`deployment-add-custom-mcp-${item.id}`}
                    className="rounded border border-cortex-primary/40 bg-cortex-primary/15 px-2.5 py-1 font-mono text-[10px] font-semibold text-cortex-primary hover:bg-cortex-primary/25"
                  >
                    + Scaffold &amp; Bind Custom MCP
                  </Link>
                </div>
              </div>
            ))}
          </div>

          <div className="rounded-lg border border-cortex-primary/30 bg-cortex-surface p-3" data-testid="deployments-custom-mcp-builder">
            <div className="text-xs font-semibold text-cortex-text-main">
              Define Your Own Custom Developed MCP Addition
            </div>
            <p className="mt-0.5 text-[11px] text-cortex-text-muted">
              Name a bespoke MCP server for this deployment and Soma will scaffold the server template, binding manifest, and reviewer verification gate.
            </p>
            <div className="mt-2 flex flex-wrap items-center gap-2">
              <input
                type="text"
                value={customSlug}
                onChange={(e) => setCustomSlug(e.target.value)}
                placeholder="e.g. billing-ledger-mcp or internal-erp-bridge"
                data-testid="deployments-custom-mcp-name-input"
                className="min-w-[200px] flex-1 rounded-lg border border-cortex-border bg-cortex-bg px-2.5 py-1.5 font-mono text-xs text-cortex-text-main placeholder:text-cortex-text-muted/50 focus:outline-none"
              />
              <select
                value={customTransport}
                onChange={(e) => setCustomTransport(e.target.value)}
                data-testid="deployments-custom-mcp-transport-select"
                className="rounded-lg border border-cortex-border bg-cortex-bg px-2.5 py-1.5 font-mono text-xs text-cortex-text-main"
              >
                <option value="stdio (python)">Python FastMCP (stdio)</option>
                <option value="stdio (node)">TypeScript SDK (stdio)</option>
                <option value="sse (http)">HTTP / SSE Endpoint</option>
                <option value="stdio / nats">Mycelis NATS Bridge</option>
              </select>
              <Link
                href={`/dashboard?team_id=${encodeURIComponent(record.id)}&team_name=${encodeURIComponent(record.name)}&team_action=add_mcp&mcp_id=${encodeURIComponent(normalizedCustomId)}&mcp_title=${encodeURIComponent(`${customTitle} (${customTransport})`)}&mcp_group=custom`}
                data-testid="deployments-custom-mcp-launch"
                className="rounded-lg border border-cortex-primary/50 bg-cortex-primary/15 px-3 py-1.5 font-mono text-xs font-semibold text-cortex-primary hover:bg-cortex-primary/25"
              >
                Build &amp; Attach Custom MCP &rarr;
              </Link>
            </div>
          </div>
        </div>
      )}
    </section>
  );
}
