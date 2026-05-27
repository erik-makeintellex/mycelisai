import type {
    ChatArtifactRef,
    ExecutionSummaryCapabilityUse,
    ExecutionSummaryData,
    ExecutionSummaryItem,
    ExecutionSummaryLink,
} from "@/store/useCortexStore";

type SummaryValue = string | ExecutionSummaryItem;

export function compactText(value?: string | null) {
    const trimmed = value?.trim();
    return trimmed || null;
}

export function executionShapeLabel(value?: string | null) {
    const shape = compactText(value);
    if (!shape) return null;
    const labels: Record<string, string> = {
        direct_answer: "Direct answer",
        directed_execution: "Directed execution",
        governed_artifact: "Governed artifact",
        proposal: "Governed proposal",
        guided_proposal: "Governed proposal",
        team_execution: "Directed execution",
        native_team: "Native team execution",
        external_workflow: "External workflow",
        external_workflow_contract: "External workflow contract",
        tool_assisted_work: "Tool-assisted work",
    };
    return labels[shape] ?? shape.replace(/[_-]+/g, " ");
}

export function itemText(item: SummaryValue): string | null {
    if (typeof item === "string") return compactText(item);
    return compactText(item.label)
        ?? compactText(item.title)
        ?? compactText(item.name)
        ?? compactText(item.summary)
        ?? compactText(item.value)
        ?? compactText(item.id)
        ?? null;
}

export function itemUrl(item: SummaryValue): string | null {
    if (typeof item === "string") return null;
    return normalizeWorkspaceOutputUrl(compactText(item.url) ?? compactText(item.href) ?? compactText(item.path));
}

export function normalizeWorkspaceOutputUrl(value?: string | null): string | null {
    const raw = compactText(value);
    if (!raw) return null;
    if (/^(https?:)?\/\//i.test(raw) || raw.startsWith("/")) return raw;
    const normalized = raw.replace(/\\/g, "/");
    if (normalized.startsWith("workspace/") || normalized.includes("/") || /\.[a-z0-9]{1,8}$/i.test(normalized)) {
        return `/api/v1/workspace/files/view?path=${encodeURIComponent(normalized)}`;
    }
    return raw;
}

export function intentLines(intent: ExecutionSummaryData["intent"]): string[] {
    if (!intent) return [];
    if (typeof intent === "string") return compactText(intent) ? [intent] : [];
    return [
        compactText(intent.original),
        compactText(intent.resolved) ? `Resolved: ${intent.resolved}` : null,
    ].filter(Boolean) as string[];
}

export function understandingLines(understanding: ExecutionSummaryData["understanding"]): string[] {
    if (!understanding) return [];
    if (typeof understanding === "string") return compactText(understanding) ? [understanding] : [];
    return [
        compactText(understanding.summary),
        ...(understanding.assumptions ?? []).map((item) => `Assumption: ${item}`),
    ].filter(Boolean) as string[];
}

export function asItems(value: ExecutionSummaryData["outputs"]): SummaryValue[] {
    if (!value) return [];
    if (typeof value === "string") return [value];
    return value;
}

export function linkLabel(link: string | ExecutionSummaryLink): string | null {
    if (typeof link === "string") return compactText(link);
    return compactText(link.label)
        ?? compactText(link.title)
        ?? (compactText(link.run_id) ? `Run ${link.run_id}` : null)
        ?? (compactText(link.audit_event_id) ? `Audit ${link.audit_event_id}` : null)
        ?? (compactText(link.intent_proof_id) ? `Proof ${link.intent_proof_id}` : null)
        ?? compactText(link.id)
        ?? compactText(link.path)
        ?? compactText(link.url)
        ?? compactText(link.href)
        ?? null;
}

export function linkHref(link: string | ExecutionSummaryLink): string | null {
    if (typeof link === "string") return link.startsWith("/") || link.startsWith("http") ? link : null;
    return compactText(link.url)
        ?? compactText(link.href)
        ?? compactText(link.path)
        ?? (compactText(link.run_id) ? `/runs/${link.run_id}` : null)
        ?? null;
}

export function proofLinks(proof: ExecutionSummaryData["proof"]): Array<string | ExecutionSummaryLink> {
    if (!proof) return [];
    return Array.isArray(proof) ? proof : [proof];
}

export function linkRunId(link: string | ExecutionSummaryLink): string | null {
    if (typeof link === "string") return null;
    return compactText(link.run_id);
}

export function capabilityGroups(capabilityUse: ExecutionSummaryData["capability_use"]) {
    if (!capabilityUse) return [];
    if (Array.isArray(capabilityUse)) {
        const values = capabilityUse.map(itemText).filter(Boolean) as string[];
        return values.length ? [{ label: "Used", values }] : [];
    }

    const groups: Array<{ label: string; values: string[] }> = [];
    const source = capabilityUse as ExecutionSummaryCapabilityUse;
    const candidates: Array<[keyof ExecutionSummaryCapabilityUse, string]> = [
        ["capabilities", "Capabilities"],
        ["teams", "Teams"],
        ["agents", "Agents"],
        ["tools", "Tools"],
        ["used", "Used"],
    ];

    for (const [key, label] of candidates) {
        const values = source[key]?.map(itemText).filter(Boolean) as string[] | undefined;
        if (values?.length) groups.push({ label, values });
    }
    return groups;
}

export function searchSourceLines(capabilityUse: ExecutionSummaryData["capability_use"]): string[] {
    if (!Array.isArray(capabilityUse)) return [];
    return capabilityUse
        .filter((item): item is ExecutionSummaryItem => typeof item !== "string")
        .filter((item) => compactText(item.id) === "web_search" || compactText(item.label) === "web_search")
        .map((item) => compactText(item.reason))
        .filter((reason): reason is string => Boolean(reason?.startsWith("Search source:")));
}

export function auditText(value: ExecutionSummaryData["audit_recovery"]) {
    if (!value) return null;
    if (typeof value === "string") return compactText(value);
    const status = compactText(value.status) ?? compactText(value.approval_status);
    const recovery = compactText(value.recovery_state);
    const summary = compactText(value.summary) ?? compactText(value.value) ?? compactText(value.label);
    const blocker = compactText(value.blocker);
    return [status, recovery, summary, blocker].filter(Boolean).join(": ") || null;
}

export function degradationLines(value: ExecutionSummaryData["audit_recovery"]): string[] {
    if (!value || typeof value === "string" || !value.degradation) return [];
    const degradation = value.degradation;
    return [
        compactText(degradation.what_failed) ? `Failed: ${degradation.what_failed}` : null,
        compactText(degradation.trusted_state) ? `Still trusted: ${degradation.trusted_state}` : null,
        compactText(degradation.invalidated_proof) ? `Invalid proof: ${degradation.invalidated_proof}` : null,
        compactText(degradation.safe_continuation) ? `Safe next: ${degradation.safe_continuation}` : null,
    ].filter(Boolean) as string[];
}

export function nextStepText(value: ExecutionSummaryData["next_step"]) {
    if (!value) return null;
    if (typeof value === "string") return compactText(value);
    return compactText(value.label)
        ?? compactText(value.title)
        ?? compactText(value.action)
        ?? compactText(value.href)
        ?? compactText(value.url)
        ?? null;
}

export function artifactOutputItems(artifacts?: ChatArtifactRef[]) {
    return artifacts?.map((artifact) => ({
        text: artifact.title || artifact.type || artifact.id || "Artifact",
        url: artifact.url ?? null,
    })) ?? [];
}

export type TrustVerdictTone = "trusted" | "review" | "attention";

export interface TrustVerdict {
    label: string;
    detail: string;
    tone: TrustVerdictTone;
}

function proofObjects(proof: ExecutionSummaryData["proof"]) {
    return proofLinks(proof).filter((item): item is ExecutionSummaryLink => typeof item !== "string");
}

function auditObject(value: ExecutionSummaryData["audit_recovery"]) {
    return value && typeof value !== "string" ? value : null;
}

export function trustVerdict(summary: ExecutionSummaryData, runId?: string, artifacts?: ChatArtifactRef[]): TrustVerdict {
    const status = (compactText(summary.execution?.status) ?? compactText(summary.execution_status) ?? "").toLowerCase();
    const audit = auditObject(summary.audit_recovery);
    const degradation = audit?.degradation;
    const proofs = proofObjects(summary.proof);
    const proofClass = proofs.map((proof) => compactText(proof.proof_class)).find(Boolean);
    const verified = proofs.some((proof) => proof.verified === true);
    const hasRun = Boolean(compactText(runId) ?? proofs.map((proof) => compactText(proof.run_id)).find(Boolean));
    const retainedOutput = asItems(summary.outputs).some((item) => typeof item !== "string" && (item.retained === true || item.kind === "code" || item.kind === "file"))
        || Boolean(artifacts?.some((artifact) => artifact.id || artifact.cached || artifact.saved_path || artifact.url));

    if (degradation?.requires_attention || ["failed", "blocked", "cancelled"].includes(status) || compactText(audit?.blocker)) {
        return {
            label: "Needs operator attention",
            detail: compactText(degradation?.what_failed)
                ?? "Part of the work is blocked or failed. Review recovery before trusting the result.",
            tone: "attention",
        };
    }
    if (status === "proposed" || audit?.approval_status === "approval_required" || audit?.recovery_state === "awaiting_confirmation") {
        return {
            label: "Awaiting approval",
            detail: "Soma has intent proof, but execution trust is not established until approval runs.",
            tone: "review",
        };
    }
    if (hasRun && retainedOutput) {
        return {
            label: "Run proof + retained output",
            detail: "A run is linked and the produced output is available for review.",
            tone: "trusted",
        };
    }
    if (hasRun || verified) {
        return {
            label: "Verified execution proof",
            detail: "Run or audit proof is linked for this result.",
            tone: "trusted",
        };
    }
    if (proofClass === "audit_only") {
        return {
            label: "Audit-only proof",
            detail: "No execution run was needed; the audit record is the trust anchor.",
            tone: "review",
        };
    }
    if (proofClass === "intent_proof") {
        return {
            label: "Intent proof only",
            detail: "Soma captured the governed intent; execution proof will appear after approval.",
            tone: "review",
        };
    }
    return {
        label: "Proof needs review",
        detail: "Review the available output and proof before relying on this result.",
        tone: "review",
    };
}
