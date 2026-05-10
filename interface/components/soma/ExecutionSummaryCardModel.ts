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
    return compactText(item.url) ?? compactText(item.href) ?? compactText(item.path) ?? null;
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

export function auditText(value: ExecutionSummaryData["audit_recovery"]) {
    if (!value) return null;
    if (typeof value === "string") return compactText(value);
    const status = compactText(value.status) ?? compactText(value.approval_status);
    const recovery = compactText(value.recovery_state);
    const summary = compactText(value.summary) ?? compactText(value.value) ?? compactText(value.label);
    const blocker = compactText(value.blocker);
    return [status, recovery, summary, blocker].filter(Boolean).join(": ") || null;
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
