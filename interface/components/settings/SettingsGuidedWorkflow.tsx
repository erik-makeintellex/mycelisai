import Link from "next/link";
import type React from "react";
import {
  ArrowRight,
  Brain,
  KeyRound,
  Layers,
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
  id: SettingsTabId;
  title: string;
  summary: string;
  buttonLabel: string;
  icon: LucideIcon;
  href?: string;
};

const DEFAULT_WORKFLOW_CARDS: WorkflowCardDefinition[] = [
  {
    id: "profile",
    title: "Name Soma and set the workspace look",
    summary: "Start by setting your assistant identity and daily product theme.",
    buttonLabel: "Open Profile",
    icon: User,
  },
  {
    id: "profiles",
    title: "Shape reusable mission defaults",
    summary: "Keep workflow-ready mission profiles available before execution.",
    buttonLabel: "Open Mission Profiles",
    icon: Layers,
  },
  {
    id: "users",
    title: "Review people and access",
    summary: "Confirm who can work in this workspace.",
    buttonLabel: "Open People & Access",
    icon: Shield,
  },
];

const ADVANCED_WORKFLOW_CARDS: WorkflowCardDefinition[] = [
  {
    id: "engines",
    title: "Inspect AI engine posture",
    summary: "Review how Soma and the wider workspace are tuned.",
    buttonLabel: "Open AI Engines",
    icon: Brain,
  },
  {
    id: "tools",
    title: "Manage connected tools",
    summary: "Use Resources as the primary MCP, search, and tool readiness home.",
    buttonLabel: "Open Resources",
    icon: Wrench,
    href: "/resources?tab=tools",
  },
  {
    id: "auth",
    title: "Plan enterprise authentication",
    summary: "Review SSO, SAML, OIDC, Entra, Google Workspace, GitHub, and SCIM.",
    buttonLabel: "Open Auth Providers",
    icon: KeyRound,
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
          <h2 className="text-xl font-semibold text-cortex-text-main">
            Start with the controls most operators actually need.
          </h2>
          <p className="text-sm leading-6 text-cortex-text-muted">
            Settings moves from identity and workflow defaults into deeper admin
            controls without becoming the main connected-tools console.
          </p>
        </div>
        <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted lg:max-w-sm">
          <p className="font-medium text-cortex-text-main">
            Tools live in Resources
          </p>
          <p className="mt-1 leading-6">
            Open Resources for MCP servers, search readiness, and recent tool
            use. Settings keeps the admin setup path tidy.
          </p>
        </div>
      </div>
      <div className="mt-5 grid gap-3 lg:grid-cols-3">
        {DEFAULT_WORKFLOW_CARDS.map((card) => (
          <WorkflowCard
            key={card.id}
            {...card}
            active={activeTab === card.id}
            onSelect={() => onSelect(card.id)}
          />
        ))}
      </div>
      {advancedMode ? (
        <div className="mt-5 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4">
          <p className="text-sm font-semibold text-cortex-text-main">
            Advanced controls are open
          </p>
          <div className="mt-4 grid gap-3 lg:grid-cols-3">
            {ADVANCED_WORKFLOW_CARDS.map((card) => (
              <WorkflowCard
                key={card.id}
                {...card}
                active={activeTab === card.id}
                onSelect={() => onSelect(card.id)}
              />
            ))}
          </div>
        </div>
      ) : (
        <div className="mt-5 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4 text-sm text-cortex-text-muted">
          <p className="font-medium text-cortex-text-main">
            Advanced controls unlock when you need them
          </p>
          <p className="mt-1 leading-6">
            Turn on Advanced mode when you need engines, auth, or connected tool
            setup. Resources remains the main tool workspace.
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
