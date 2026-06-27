import type { SomaPinnedAction } from "./SomaActionShelf";

const QUICK_ACTION_TAG = "quick_action";

export function isSavedAction(value: unknown): value is SomaPinnedAction {
  if (!value || typeof value !== "object") return false;
  const candidate = value as Partial<SomaPinnedAction>;
  return typeof candidate.label === "string" && typeof candidate.prompt === "string";
}

type APITemplate = {
  id?: string;
  name?: string;
  template_body?: string;
  governance_tags?: string[];
  output_contract?: Record<string, unknown>;
};

type APIEnvelope<T> = {
  ok?: boolean;
  data?: T;
};

export async function loadBackendActions(): Promise<SomaPinnedAction[]> {
  const response = await fetch("/api/v1/conversation-templates?scope=soma&status=active&limit=50", { cache: "no-store" });
  if (!response.ok) return [];
  const envelope = await response.json() as APIEnvelope<APITemplate[]>;
  const templates = Array.isArray(envelope.data) ? envelope.data : [];
  return templates
    .filter((template) => template.governance_tags?.includes(QUICK_ACTION_TAG))
    .map(actionFromTemplate)
    .filter(isSavedAction)
    .filter(uniqueActionLabel())
    .slice(0, 2);
}

export async function saveBackendAction(action: SomaPinnedAction): Promise<SomaPinnedAction | null> {
  const response = await fetch("/api/v1/conversation-templates", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      name: action.label,
      description: "Pinned Soma quick action from Button Studio.",
      scope: "soma",
      creator_kind: "user",
      status: "active",
      template_body: action.prompt,
      output_contract: {
        output_format: action.outputFormat || "",
        approval_behavior: action.approvalBehavior || "Ask before running",
      },
      governance_tags: [QUICK_ACTION_TAG, "button_studio"],
    }),
  });
  if (!response.ok) return null;
  const envelope = await response.json() as APIEnvelope<APITemplate>;
  return envelope.data ? actionFromTemplate(envelope.data) : null;
}

export async function instantiateBackendAction(actionId: string): Promise<string> {
  const response = await fetch(`/api/v1/conversation-templates/${encodeURIComponent(actionId)}/instantiate`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ variables: {} }),
  });
  if (!response.ok) return "";
  const envelope = await response.json() as APIEnvelope<{ rendered_prompt?: string }>;
  return envelope.data?.rendered_prompt || "";
}

export function persistLocalActions(actions: SomaPinnedAction[]) {
  window.localStorage.setItem("mycelis-soma-saved-actions", JSON.stringify(actions));
}

function actionFromTemplate(template: APITemplate): SomaPinnedAction {
  const outputContract = template.output_contract || {};
  return {
    id: template.id,
    label: template.name || "",
    prompt: template.template_body || "",
    outputFormat: String(outputContract.output_format || ""),
    approvalBehavior: String(outputContract.approval_behavior || ""),
    userSaved: true,
  };
}

function uniqueActionLabel() {
  const seen = new Set<string>();
  return (action: SomaPinnedAction) => {
    const key = action.label.trim().toLowerCase();
    if (!key || seen.has(key)) return false;
    seen.add(key);
    return true;
  };
}
