// Shared types and fetch helpers for the governed clone/apply config flows
// (RC-004G). Mirrors core/internal/server/config_clone.go and
// core/internal/server/config_apply*.go; see
// docs/API_REFERENCE.md#clone-and-apply-config for the wire contract.

export type ConfirmToken = {
  token: string;
  intent_proof_id: string;
  expires_at: string;
};

export type IntentProof = {
  id: string;
  resolved_intent: string;
  status: string;
};

export type ConfigDiff = {
  changed: string[];
  added: string[];
  removed: string[];
};

export type ConfigPreview = {
  kind: string;
  id: string;
  document_sha256: string;
  format?: string;
  document?: string;
  redactions?: string[];
  valid: boolean;
  validation_errors: string[];
  diff?: ConfigDiff;
  source_template_id?: string;
};

export type GovernedPreview = {
  requires_approval: boolean;
  blocked: boolean;
  confirm_token?: ConfirmToken;
  intent_proof: IntentProof;
  audit_event_id: string;
  preview: ConfigPreview;
};

/** Outcome of one governed POST: either a 202 preview or a terminal result. */
export type GovernedOutcome<T> =
  | { kind: "preview"; preview: GovernedPreview }
  | { kind: "success"; data: T }
  | { kind: "conflict" }
  | { kind: "write_back_violations"; violations: string[] }
  | { kind: "bad_request"; detail: string }
  | { kind: "forbidden" }
  | { kind: "not_found" }
  | { kind: "unavailable"; detail: string }
  | { kind: "error"; detail: string };

async function readJson(res: Response): Promise<unknown> {
  try {
    return await res.json();
  } catch {
    return null;
  }
}

function dataOf(payload: unknown): Record<string, unknown> | null {
  if (payload && typeof payload === "object") {
    const record = payload as Record<string, unknown>;
    if ("data" in record && record.data && typeof record.data === "object") {
      return record.data as Record<string, unknown>;
    }
    return record;
  }
  return null;
}

function errorOf(payload: unknown): string | undefined {
  if (payload && typeof payload === "object") {
    const error = (payload as Record<string, unknown>).error;
    if (typeof error === "string") return error;
  }
  return undefined;
}

/** Posts a governed clone/apply request and classifies the response. */
export async function postGoverned<T>(
  url: string,
  body: unknown,
): Promise<GovernedOutcome<T>> {
  let res: Response;
  try {
    res = await fetch(url, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
  } catch (failure) {
    return {
      kind: "error",
      detail: failure instanceof Error ? failure.message : "Network error.",
    };
  }
  const payload = await readJson(res);
  const data = dataOf(payload);

  if (res.status === 202) {
    return { kind: "preview", preview: data as unknown as GovernedPreview };
  }
  if (res.status === 200 || res.status === 201) {
    return { kind: "success", data: data as unknown as T };
  }
  if (res.status === 409) return { kind: "conflict" };
  if (res.status === 403) return { kind: "forbidden" };
  if (res.status === 404) return { kind: "not_found" };
  if (res.status === 503) {
    return {
      kind: "unavailable",
      detail: errorOf(payload) || "Config write is unavailable right now.",
    };
  }
  if (res.status === 400) {
    const violations = data?.write_back_violations;
    if (Array.isArray(violations)) {
      return { kind: "write_back_violations", violations: violations as string[] };
    }
    return { kind: "bad_request", detail: errorOf(payload) || "Request was rejected." };
  }
  return { kind: "error", detail: `Request failed (HTTP ${res.status}).` };
}

export type CloneTemplateResult = {
  template: {
    id: string;
    name: string;
    source: string;
    source_template_id?: string;
    document_sha256: string;
  };
  intent_proof_id: string;
  audit_event_id: string;
  run_id?: string;
  mission_event_id?: string;
};

export type ApplyConfigResult = {
  applied: boolean;
  diff: ConfigDiff;
  intent_proof_id: string;
  audit_event_id: string;
  run_id?: string;
  mission_event_id?: string;
  result?: unknown;
};

export function previewOrConfirmClone(
  templateId: string,
  body: { name: string; new_id?: string; overrides?: Record<string, unknown>; confirm_token?: string },
) {
  return postGoverned<CloneTemplateResult>(
    `/api/v1/config/templates/${encodeURIComponent(templateId)}/clone`,
    body,
  );
}

export function previewOrConfirmApply(
  kind: "templates" | "groups",
  id: string,
  body: { document: string; format: "yaml" | "json"; confirm_token?: string },
) {
  return postGoverned<ApplyConfigResult>(
    `/api/v1/config/${kind}/${encodeURIComponent(id)}/apply`,
    body,
  );
}
