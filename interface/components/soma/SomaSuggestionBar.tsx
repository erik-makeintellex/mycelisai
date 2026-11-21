import { useState } from "react";
import type { LucideIcon } from "lucide-react";
import { BarChart3, BrainCircuit, ChevronDown, ChevronUp, ClipboardCheck, Compass, FileText, Gamepad2, GitBranch, LayoutTemplate, Megaphone, Search, Settings2, ShieldAlert, Sparkles, Terminal, Wrench } from "lucide-react";
import { useCortexStore } from "@/store/useCortexStore";

export type SomaSuggestion = {
  id?: string;
  label: string;
  detail?: string;
  summary?: string;
  badge?: string;
  hint?: string;
  prompt: string;
  icon?: LucideIcon;
};

export const DEFAULT_SOMA_SUGGESTIONS: SomaSuggestion[] = [
  {
    label: "Plan something",
    detail: "Break a project idea into simple steps.",
    prompt: "Help me break my project idea into simple, practical steps.",
    icon: Compass,
  },
  {
    label: "Research something",
    detail: "Look up options and summarize clearly.",
    prompt: "Research this topic and give me a simple summary of the best options.",
    icon: Search,
  },
  {
    label: "Create something",
    detail: "Build an app, game, or project files.",
    prompt: "Create a clean task tracker web app project with an interactive HTML UI and quickstart README.",
    icon: Sparkles,
  },
  {
    label: "Review something",
    detail: "Check work and highlight what matters.",
    prompt: "Review this project and give me a simple summary of what looks good and what to fix next.",
    icon: FileText,
  },
  {
    label: "Configure tools",
    detail: "See active AI engines and tools.",
    prompt: "What is the current runtime and AI engine status of Mycelis?",
    icon: Settings2,
  },
];

export const POPULAR_DELIVERABLE_STARTERS: SomaSuggestion[] = [
  {
    label: "Marketing Launch Kit",
    detail: "Launch brief, email copy, and FAQ.",
    prompt: "Create marketing documentation for our product launch including a positioning brief, launch email, and customer FAQ.",
    icon: Megaphone,
  },
  {
    label: "Playable Puzzle Game",
    detail: "Interactive browser game + rules.",
    prompt: "Create a simple browser sliding-tile puzzle game with playable HTML, game rules, and controls.",
    icon: Gamepad2,
  },
  {
    label: "Engineering Task Pack",
    detail: "Simple project steps and test checklist.",
    prompt: "Create a Codex task package for MYC-RC-005 with task steps and validation checklist.",
    icon: Terminal,
  },
  {
    label: "Customer Onboarding Kit",
    detail: "Welcome guide, 30-day checklist, and FAQ.",
    prompt: "Create a customer onboarding kit with a welcome guide, first 30 days checklist, and escalation FAQ.",
    icon: ClipboardCheck,
  },
  {
    label: "Interactive KPI Dashboard",
    detail: "Live HTML5 metrics view and definitions.",
    prompt: "Create an interactive browser KPI dashboard with playable HTML, metrics definitions, and governance notes.",
    icon: BarChart3,
  },
  {
    id: "product-wireframe",
    label: "Product Spec & Wireframe",
    detail: "Interactive HTML5 clickable UI prototype and user flows.",
    summary: "Interactive HTML5 clickable UI prototype and user flows.",
    prompt: "Create an interactive product wireframe prototype and product spec in the workspace with clickable screens and user flows.",
    icon: LayoutTemplate,
  },
];

export const ORCHESTRATION_PLAYBOOK_STARTERS: SomaSuggestion[] = [
  {
    label: "Council Architecture & Security Review",
    detail: "Consult Architect + Sentry on system design and risk posture.",
    prompt: "Consult the Architect and Sentry council specialists to review our current architecture tradeoffs, security boundaries, and recommended next steps.",
    icon: ShieldAlert,
  },
  {
    label: "Multi-Team Mission Blueprint",
    detail: "Design a plan -> develop -> verify -> deliver team workflow.",
    prompt: "Design a multi-agent mission blueprint with explicit plan, develop, verify, and deliver team lanes and governance checkpoints.",
    icon: GitBranch,
  },
  {
    label: "Trusted-Memory & Doctrine Audit",
    detail: "Inspect governed doctrine, project memory, and continuity.",
    prompt: "Audit our current trusted memory layers, highlight any candidate lessons ready for promotion to Governed Doctrine, and check for contradictions.",
    icon: BrainCircuit,
  },
  {
    label: "Service & Tool Capability Plan",
    detail: "Verify backend service addresses and MCP tool readiness.",
    prompt: "Check our current backend service status and installed MCP tools, and recommend the next capability or workflow to enable.",
    icon: Wrench,
  },
  {
    label: "Chain Multi-Team & MCP Channels",
    detail: "Links 2+ teams via NATS listen/speak channels and registered MCP outputs",
    prompt: "Chain two team asks together: have an upstream MCP discovery team publish registered MCP outputs on its result channel, and have a downstream execution team listen on that channel to produce the final deliverable package.",
    badge: "Channel Chain",
    hint: "Links 2+ teams via NATS listen/speak channels and registered MCP outputs",
    icon: GitBranch,
  },
];

export function SomaSuggestionBar({
  suggestions = DEFAULT_SOMA_SUGGESTIONS,
  starters = POPULAR_DELIVERABLE_STARTERS,
  playbooks = ORCHESTRATION_PLAYBOOK_STARTERS,
  onSelect,
}: {
  suggestions?: readonly SomaSuggestion[];
  starters?: readonly SomaSuggestion[];
  playbooks?: readonly SomaSuggestion[];
  onSelect: (prompt: string) => void;
}) {
  const advancedMode = useCortexStore((s) => s.advancedMode);
  const [showAdvancedPlaybooks, setShowAdvancedPlaybooks] = useState(false);
  const displayPlaybooks = advancedMode || showAdvancedPlaybooks;

  return (
    <div className="rounded-2xl border border-cortex-border bg-cortex-bg/70 p-3 space-y-3">
      <div>
        <p className="text-[11px] font-mono uppercase tracking-[0.18em] text-cortex-primary">
          Start with Soma
        </p>
        <div className="mt-3 grid gap-2 md:grid-cols-5">
          {suggestions.map((item) => {
            const Icon = item.icon ?? Sparkles;
            return (
              <button
                key={item.label}
                type="button"
                onClick={() => onSelect(item.prompt)}
                className="rounded-xl border border-cortex-border bg-cortex-surface px-3 py-2 text-left transition hover:border-cortex-primary/30 hover:bg-cortex-primary/5"
              >
                <span className="flex items-center gap-2 text-xs font-semibold text-cortex-text-main">
                  <Icon className="h-3.5 w-3.5 text-cortex-primary" />
                  {item.label}
                </span>
                <span className="mt-2 block text-xs leading-5 text-cortex-text-muted">
                  {item.detail}
                </span>
              </button>
            );
          })}
        </div>
      </div>

      {starters.length > 0 && (
        <div className="border-t border-cortex-border/70 pt-2.5">
          <p className="text-[10px] font-mono uppercase tracking-[0.16em] text-cortex-text-muted">
            Popular deliverable starters
          </p>
          <div className="mt-2 grid gap-2 md:grid-cols-6">
            {starters.map((item, index) => {
              const Icon = item.icon ?? Sparkles;
              return (
                <button
                  key={item.label}
                  type="button"
                  data-testid={`deliverable-starter-${index}`}
                  onClick={() => onSelect(item.prompt)}
                  className="rounded-xl border border-cortex-primary/20 bg-cortex-primary/5 px-3 py-2 text-left transition hover:border-cortex-primary/40 hover:bg-cortex-primary/10"
                >
                  <span className="flex items-center gap-2 text-xs font-semibold text-cortex-text-main">
                    <Icon className="h-3.5 w-3.5 text-cortex-primary" />
                    {item.label}
                  </span>
                  <span className="mt-1 block text-xs leading-4 text-cortex-text-muted">
                    {item.detail || item.summary}
                  </span>
                </button>
              );
            })}
          </div>
        </div>
      )}

      {playbooks.length > 0 && (
        <div className="border-t border-cortex-border/70 pt-2.5">
          <div className="flex items-center justify-between gap-2">
            <p className="text-[10px] font-mono uppercase tracking-[0.16em] text-cortex-text-muted">
              Advanced orchestration & council playbooks
            </p>
            {!advancedMode && (
              <button
                type="button"
                data-testid="toggle-advanced-playbooks"
                onClick={() => setShowAdvancedPlaybooks((prev) => !prev)}
                className="inline-flex items-center gap-1 rounded border border-cortex-border bg-cortex-surface px-2 py-0.5 text-[10px] font-mono text-cortex-text-muted hover:border-cortex-primary/40 hover:text-cortex-primary"
              >
                {showAdvancedPlaybooks ? <ChevronUp className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />}
                {showAdvancedPlaybooks ? "Hide" : "Show"}
              </button>
            )}
          </div>
          {displayPlaybooks && (
            <div className="mt-2 grid gap-2 md:grid-cols-4">
              {playbooks.map((item, index) => {
                const Icon = item.icon ?? Sparkles;
                return (
                  <button
                    key={item.label}
                    type="button"
                    data-testid={`orchestration-playbook-starter-${index}`}
                    onClick={() => onSelect(item.prompt)}
                    className="rounded-xl border border-cortex-border bg-cortex-surface px-3 py-2 text-left transition hover:border-cortex-primary/35 hover:bg-cortex-primary/5"
                  >
                    <span className="flex items-center gap-2 text-xs font-semibold text-cortex-text-main">
                      <Icon className="h-3.5 w-3.5 text-cortex-primary" />
                      {item.label}
                    </span>
                    <span className="mt-1 block text-xs leading-4 text-cortex-text-muted">
                      {item.detail}
                    </span>
                  </button>
                );
              })}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
