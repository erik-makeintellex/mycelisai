// labels.ts — Human-facing label translation layer
// Internal names preserved everywhere; this module maps them to natural language at render time.

// ---------------------------------------------------------------------------
// Tool labels: internal name → human-facing
// ---------------------------------------------------------------------------
export const TOOL_LABELS: Record<string, string> = {
  consult_council: "Ask Specialist",
  delegate_task: "Assign to Team",
  search_memory: "Search Memory",
  list_teams: "View Teams",
  list_missions: "View Missions",
  get_system_status: "System Status",
  list_available_tools: "Available Tools",
  generate_blueprint: "Design Mission",
  list_catalogue: "Agent Catalogue",
  remember: "Remember",
  recall: "Recall",
  store_artifact: "Save Output",
  publish_signal: "Send Signal",
  broadcast: "Broadcast",
  read_signals: "Listen to Bus",
  read_file: "Read File",
  write_file: "Write File",
  generate_image: "Generate Image",
  research_for_blueprint: "Research Mission",
  summarize_conversation: "Save Context",
};

export function toolLabel(name: string): string {
  return TOOL_LABELS[name] ?? name;
}

// ---------------------------------------------------------------------------
// Council labels: ID → name + subtitle
// ---------------------------------------------------------------------------
export interface CouncilLabel {
  name: string;
  subtitle: string;
}

export const COUNCIL_LABELS: Record<string, CouncilLabel> = {
  admin: { name: "Soma", subtitle: "Executive Cortex" },
  "council-architect": { name: "Architect", subtitle: "Systems Design" },
  "council-coder": { name: "Coder", subtitle: "Implementation" },
  "council-creative": { name: "Creative", subtitle: "Design & Ideation" },
  "council-sentry": { name: "Sentry", subtitle: "Risk & Security" },
};

export function councilLabel(id: string): CouncilLabel {
  return COUNCIL_LABELS[id] ?? { name: id, subtitle: "" };
}

export function councilOptionLabel(id: string, role: string): string {
  const label = COUNCIL_LABELS[id];
  if (label) return `${label.name} — ${label.subtitle}`;
  // Fallback: capitalize role
  return role.charAt(0).toUpperCase() + role.slice(1);
}

export function sourceNodeLabel(sourceNode: string): string {
  const label = COUNCIL_LABELS[sourceNode];
  if (label) return label.name;
  // Strip "council-" prefix and capitalize
  const cleaned = sourceNode.replace("council-", "");
  return cleaned.charAt(0).toUpperCase() + cleaned.slice(1);
}

// ---------------------------------------------------------------------------
// Governance labels
// ---------------------------------------------------------------------------
export const GOV_LABELS = {
  governanceReview: "Review Required",
  agentOutput: "Agent Output",
  proofOfWork: "Verification Evidence",
  verificationMethod: "Verification Method",
  rubricScore: "Quality Score",
  noProof: "No verification evidence provided",
  noProofSub: "This agent did not submit verification evidence",
};

// ---------------------------------------------------------------------------
// Trust labels
// ---------------------------------------------------------------------------
export function trustBadge(score: number): string {
  return `C:${score.toFixed(1)}`;
}

export function trustTooltip(score: number): string {
  const level =
    score >= 0.9
      ? "High confidence"
      : score >= 0.6
        ? "Moderate confidence"
        : "Low confidence";
  return `Confidence: ${score.toFixed(2)} — ${level}`;
}

// ---------------------------------------------------------------------------
// Workspace labels
// ---------------------------------------------------------------------------
export const WORKSPACE_LABELS = {
  spectrum: "Spectrum",
  squadRoom: "Squad Room",
  internalDebate: "Internal Deliberation",
  metaArchitect: "Mission Architect",
  toolRegistry: "Capabilities",
  internalTools: "Core Capabilities",
};

// ---------------------------------------------------------------------------
// CE-1: Orchestration Template labels
// ---------------------------------------------------------------------------

export const MODE_LABELS: Record<string, { label: string; color: string }> = {
  answer: { label: "ANSWER", color: "text-cortex-primary" },
  proposal: { label: "PROPOSAL", color: "text-amber-400" },
  broadcast: { label: "BROADCAST", color: "text-cortex-warning" },
  execute: { label: "EXECUTE", color: "text-cortex-success" },
};

export const GOV_POSTURE_LABELS: Record<string, { label: string; color: string }> = {
  passive: { label: "PASSIVE", color: "text-cortex-text-muted" },
  active: { label: "ACTIVE", color: "text-amber-400" },
  strict: { label: "STRICT", color: "text-red-400" },
};

// Intent classification heuristic from tools_used array
export function deriveIntentClass(tools: string[]): string {
  if (!tools?.length) return "Direct Answer";
  if (tools.includes("generate_blueprint") || tools.includes("research_for_blueprint")) return "Mission Design";
  if (tools.includes("delegate_task")) return "Delegation";
  if (tools.includes("consult_council")) return "Consultation";
  if (tools.includes("search_memory") || tools.includes("recall")) return "Memory Recall";
  if (tools.includes("read_file") || tools.includes("write_file")) return "File Operation";
  return "Direct Answer";
}

// Governance-gated tools (require elevated scrutiny)
export const GOVERNANCE_TOOLS = new Set([
  "delegate_task",
  "generate_blueprint",
  "write_file",
  "publish_signal",
]);

export const isGovernanceTool = (name: string): boolean => GOVERNANCE_TOOLS.has(name);
export const isInternalTool = (name: string): boolean => name in TOOL_LABELS;

// Tool descriptions for tooltips
export const TOOL_DESCRIPTIONS: Record<string, string> = {
  consult_council: "Routes question to a council specialist",
  delegate_task: "Assigns work to an active team (governance-gated)",
  search_memory: "Semantic search over situation reports",
  generate_blueprint: "Creates mission blueprint via MetaArchitect (governance-gated)",
  write_file: "Writes file to sandboxed workspace (governance-gated)",
  publish_signal: "Publishes signal to NATS bus (governance-gated)",
  recall: "Retrieves knowledge by semantic similarity",
  research_for_blueprint: "Gathers context before blueprint generation",
  list_teams: "Lists active teams in the swarm",
  list_missions: "Lists active missions",
  get_system_status: "Retrieves system telemetry",
  list_available_tools: "Lists all available tools",
  list_catalogue: "Lists agent templates from catalogue",
  remember: "Stores fact to persistent memory",
  store_artifact: "Persists agent output to artifacts table",
  broadcast: "Sends message to all active teams",
  read_signals: "Subscribes to NATS topic pattern",
  read_file: "Reads file from sandboxed workspace",
  generate_image: "Generates image via cognitive pipeline",
  summarize_conversation: "Compresses conversation for long-term memory",
};

export function toolDescription(name: string): string {
  return TOOL_DESCRIPTIONS[name] ?? name;
}

// Memory tag labels
export const MEMORY_LABELS: Record<string, string> = {
  search_memory: "Semantic search",
  recall: "Context recall",
};

// ---------------------------------------------------------------------------
// Phase 19: Brain / Provider labels
// ---------------------------------------------------------------------------

export const PROVIDER_DISPLAY_NAMES: Record<string, string> = {
  ollama: "Ollama",
  vllm: "vLLM",
  lmstudio: "LM Studio",
  production_claude: "Claude",
  production_gpt4: "GPT-4",
  production_gemini: "Gemini",
  "emergency-ollama": "Ollama (Emergency)",
};

export function brainDisplayName(providerId: string): string {
  return PROVIDER_DISPLAY_NAMES[providerId] ?? providerId;
}

export function brainLocationLabel(location: string): string {
  if (location === "remote") return "Remote";
  return "Local";
}

export function brainBadge(providerId: string, location: string): string {
  const name = brainDisplayName(providerId);
  const loc = brainLocationLabel(location);
  return `${name} (${loc})`;
}

// Tool origin classification for enhanced badges
export const MCP_TOOL_PREFIX = "mcp.";
export const SANDBOXED_TOOLS = new Set(["read_file", "write_file"]);

export function toolOrigin(name: string): 'internal' | 'external' | 'sandboxed' {
  if (name.startsWith(MCP_TOOL_PREFIX)) return 'external';
  if (SANDBOXED_TOOLS.has(name)) return 'sandboxed';
  return 'internal';
}
