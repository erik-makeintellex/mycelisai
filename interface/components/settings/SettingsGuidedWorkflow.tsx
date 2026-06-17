import Link from "next/link";
import type React from "react";
import {
  ArrowRight,
  Brain,
  FolderOpen,
  KeyRound,
  Layers,
  ServerCog,
  Shield,
  User,
  Wrench,
  type LucideIcon,
} from "lucide-react";

export type SettingsTabId =
  | "profile"
  | "profiles"
  | "users"
  | "engines"
  | "auth"
  | "tools";

type WorkflowCardDefinition = {
  id: string;
  selectTabId?: SettingsTabId;
  title: string;
  summary: string;
  buttonLabel: string;
  icon: LucideIcon;
  href?: string;
};

const DEFAULT_WORKFLOW_CARDS: WorkflowCardDefinition[] = [
  {
    id: "profile",
    selectTabId: "profile",
    title: "Name Soma and set the workspace look",
    summary: "Start by setting your assistant identity and daily product theme.",
    buttonLabel: "Open Profile",
    icon: User,
  },
  {
    id: "profiles",
    selectTabId: "profiles",
    title: "Shape reusable mission defaults",
    summary: "Keep workflow-ready mission profiles available before execution.",
    buttonLabel: "Open Mission Profiles",
    icon: Layers,
  },
  {
    id: "users",
    selectTabId: "users",
    title: "Review people and access",
    summary: "Confirm who can work in this workspace.",
    buttonLabel: "Open People & Access",
    icon: Shield,
  },
];

const ADVANCED_WORKFLOW_CARDS: WorkflowCardDefinition[] = [
  {
    id: "auth",
    selectTabId: "auth",
    title: "Set login and SSO",
    summary: "Review local owner login, Google Workspace, OIDC, SAML, Entra, GitHub, and SCIM posture.",
    buttonLabel: "Open Auth Providers",
    icon: KeyRound,
  },
  {
    id: "engines",
    selectTabId: "engines",
    title: "Choose the AI provider",
    summary: "Confirm which model provider Soma and the workspace will use before live work.",
    buttonLabel: "Open AI Engines",
    icon: Brain,
  },
  {
    id: "workspace-roots",
    title: "Check workspace and output roots",
    summary: "Confirm where generated files, packages, artifacts, and filesystem MCP writes land.",
    buttonLabel: "Open Deployments",
    icon: FolderOpen,
    href: "/system?tab=deployments",
  },
  {
    id: "mcp-tools",
    title: "Add MCP connected tools",
    summary: "Install filesystem, fetch, search, or other curated tool servers from Resources.",
    buttonLabel: "Add MCP",
    icon: Wrench,
    href: "/resources?tab=tools",
  },
  {
    id: "workspace-files",
    title: "Browse workspace files",
    summary: "Use the filesystem MCP boundary to verify the mounted workspace from the operator UI.",
    buttonLabel: "Open Workspace Files",
    icon: ServerCog,
    href: "/resources?tab=workspace",
  },
];

export function SettingsGuidedWorkflow({
  advancedMode,
  activeTab,
  onSelect,
}: {
  advancedMode: boolean;
  activeTab: SettingsTabId;
  onSelect: (tab: SettingsTabId) => void;
}) {
  return (
    <section className="rounded-3xl border border-cortex-border bg-cortex-surface px-5 py-5 shadow-sm">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div className="max-w-2xl space-y-2">
          <p className="text-[11px] font-mono uppercase tracking-[0.24em] text-cortex-primary">
            Guided setup path
          </p>
          <h2 className="text-xl font-semibold text-cortex-text-main">New admin setup checklist</h2>
          <p className="text-sm leading-6 text-cortex-text-muted">
            Make login, AI provider, storage roots, and connected tools obvious before handing
            the workspace to operators.
          </p>
        </div>
        <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted lg:max-w-sm">
          <p className="font-medium text-cortex-text-main">
            Setup has four concrete checks
          </p>
          <p className="mt-1 leading-6">
            Auth, AI provider, workspace/output roots, and Add MCP are linked directly
            from this panel.
          </p>
        </div>
      </div>
      <div className="mt-5 grid gap-3 lg:grid-cols-3">
        {DEFAULT_WORKFLOW_CARDS.map((card) => (
          <WorkflowCard
            key={card.id}
            {...card}
            active={activeTab === card.selectTabId}
            onSelect={() => card.selectTabId ? onSelect(card.selectTabId) : undefined}
          />
        ))}
      </div>
      {advancedMode ? (
        <div className="mt-5 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4">
          <p className="text-sm font-semibold text-cortex-text-main">
            Admin tools are open
          </p>
          <div className="mt-4 grid gap-3 lg:grid-cols-3">
            {ADVANCED_WORKFLOW_CARDS.map((card) => (
              <WorkflowCard
                key={`${card.id}-${card.title}`}
                {...card}
                active={activeTab === card.selectTabId}
                onSelect={() => card.selectTabId ? onSelect(card.selectTabId) : undefined}
              />
            ))}
          </div>
        </div>
      ) : (
        <div className="mt-5 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4 text-sm text-cortex-text-muted">
          <p className="font-medium text-cortex-text-main">
            Admin tools unlock the full setup checklist
          </p>
          <p className="mt-1 leading-6">
            Turn on Admin tools to configure SSO/auth, AI providers, workspace/output
            roots, and MCP connected tools.
          </p>
        </div>
      )}
    </section>
  );
}

function WorkflowCard({
  title,
  summary,
  buttonLabel,
  icon: Icon,
  active,
  href,
  onSelect,
}: WorkflowCardDefinition & {
  active: boolean;
  onSelect: () => void;
}) {
  const className = `mt-4 inline-flex items-center gap-2 rounded-xl border border-cortex-border bg-cortex-surface px-3 py-2 text-sm font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/25 hover:text-cortex-primary`;
  return (
    <div
      className={`rounded-2xl border px-4 py-4 transition-colors ${
        active ? "border-cortex-primary/40 bg-cortex-primary/10" : "border-cortex-border bg-cortex-bg"
      }`}
    >
      <div className="flex items-center gap-2 text-cortex-primary">
        <Icon className="h-4 w-4" />
        <p className="text-sm font-semibold text-cortex-text-main">{title}</p>
      </div>
      <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{summary}</p>
      {href ? (
        <Link href={href} className={className}>
          {buttonLabel}
          <ArrowRight className="h-4 w-4" />
        </Link>
      ) : (
        <button type="button" onClick={onSelect} className={className}>
          {buttonLabel}
          <ArrowRight className="h-4 w-4" />
        </button>
      )}
    </div>
  );
}
