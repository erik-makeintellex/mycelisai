import type { TargetRef, TeamOutputRef } from "@/store/useCortexStore";

export function normalizeTargetRef(raw: unknown): TargetRef | undefined {
  const record = objectValue<Record<string, unknown>>(raw);
  const type = stringValue(record?.type);
  const id = stringValue(record?.id);
  if (!type || !id) return undefined;
  return {
    type,
    id,
    run_id: stringValue(record?.run_id) ?? undefined,
    team_id: stringValue(record?.team_id) ?? undefined,
    work_item_id: stringValue(record?.work_item_id) ?? undefined,
    project_id: stringValue(record?.project_id) ?? undefined,
    output_id: stringValue(record?.output_id) ?? undefined,
    label: stringValue(record?.label) ?? undefined,
  };
}

export function targetRefReference(targetRef?: TargetRef | null): string | null {
  if (!targetRef?.type || !targetRef.id) return null;
  return `${targetRef.type}:${targetRef.id}`;
}

export function targetRefHref(targetRef?: TargetRef | null): string | null {
  if (!targetRef) return null;
  if (targetRef.type === "run") {
    const runId = targetRef.run_id || targetRef.id;
    return runId ? `/runs/${encodeURIComponent(runId)}` : null;
  }
  if (targetRef.type === "work" || targetRef.type === "recovery") {
    const workItemId = targetRef.work_item_id || targetRef.id;
    return workItemId ? `/teams?view=work&work_item_id=${encodeURIComponent(workItemId)}` : null;
  }
  if (targetRef.type === "outcome_project") {
    return "/resources?tab=workspace";
  }
  if (targetRef.work_item_id) {
    return `/teams?view=work&work_item_id=${encodeURIComponent(targetRef.work_item_id)}`;
  }
  if (targetRef.run_id) {
    return `/runs/${encodeURIComponent(targetRef.run_id)}`;
  }
  return null;
}

export function outputRefArray(value: unknown, teamId: string, workItemId: string, fallbackCreatedAt?: string | null): TeamOutputRef[] {
  if (!Array.isArray(value)) return [];
  return value.filter(isRecord).map((item, index) => ({
    output_id: stringValue(item.output_id) ?? `${workItemId}-output-${index}`,
    team_id: stringValue(item.team_id) ?? teamId,
    work_item_id: stringValue(item.work_item_id) ?? workItemId,
    run_id: stringValue(item.run_id) ?? undefined,
    kind: stringValue(item.kind) ?? "file",
    label: stringValue(item.label) ?? "Team output",
    storage_ref: stringValue(item.storage_ref) ?? undefined,
    entrypoint: stringValue(item.entrypoint) ?? undefined,
    validation_ref: stringValue(item.validation_ref) ?? undefined,
    proof_ref: stringValue(item.proof_ref) ?? undefined,
    contract_id: stringValue(item.contract_id) ?? undefined,
    proof_id: stringValue(item.proof_id) ?? undefined,
    audit_refs: stringArray(item.audit_refs),
    created_at: stringValue(item.created_at) ?? stringValue(item.updated_at) ?? fallbackCreatedAt ?? undefined,
  }));
}

export function outputRefTime(output: TeamOutputRef) {
  const time = output.created_at ? Date.parse(output.created_at) : 0;
  return Number.isFinite(time) ? time : 0;
}

export function unique(values: string[]): string[] {
  return Array.from(new Set(values.map((value) => value.trim()).filter(Boolean)));
}

export function stringValue(value: unknown): string | null {
  return typeof value === "string" && value.trim() ? value.trim() : null;
}

export function stringArray(value: unknown): string[] {
  return Array.isArray(value)
    ? unique(value.map((item) => typeof item === "string" ? item : "").filter(Boolean))
    : [];
}

export function objectValue<T extends object>(value: unknown): T | null {
  return isRecord(value) ? value as T : null;
}

export function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}
