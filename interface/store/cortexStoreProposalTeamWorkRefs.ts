import { trimToNonEmpty } from "@/store/cortexStoreChatWorkflow";

export type TeamWorkConfirmationRef = {
    id?: string;
    work_item_id?: string;
    work_id?: string;
    state?: string;
    status?: string;
    run_id?: string;
    runId?: string;
    output_count?: number;
    outputCount?: number;
    expected_outputs?: unknown[];
    expectedOutputs?: unknown[];
    expected_proof?: unknown[];
    expectedProof?: unknown[];
    output_refs?: unknown[];
    outputRefs?: unknown[];
    proof_refs?: unknown[];
    proofRefs?: unknown[];
    audit_refs?: unknown[];
    auditRefs?: unknown[];
};

export function extractTeamWorkRefs(raw: unknown): TeamWorkConfirmationRef[] {
    if (!raw || typeof raw !== "object") return [];
    const record = raw as Record<string, unknown>;
    const data = record.data && typeof record.data === "object"
        ? record.data as Record<string, unknown>
        : record;
    const candidate = Array.isArray(data.team_work_refs)
        ? data.team_work_refs
        : Array.isArray(data.team_work_items)
            ? data.team_work_items
            : [];
    return candidate.filter((item): item is TeamWorkConfirmationRef => !!item && typeof item === "object");
}

function shortIdentifier(value: unknown): string | null {
    const id = trimToNonEmpty(value);
    return id ? id.slice(0, 8) : null;
}

function teamWorkStateLabel(refs: TeamWorkConfirmationRef[]): string {
    const states = refs
        .map((ref) => trimToNonEmpty(ref.state) ?? trimToNonEmpty(ref.status))
        .map((state) => state?.toLowerCase())
        .filter(Boolean);
    const hasOutput = refs.some((ref) => (
        (ref.output_count ?? 0) > 0
        || (ref.outputCount ?? 0) > 0
        || (Array.isArray(ref.output_refs) && ref.output_refs.length > 0)
        || (Array.isArray(ref.outputRefs) && ref.outputRefs.length > 0)
    ));
    if (hasOutput || states.some((state) => state === "output_ready" || state === "output-ready")) return "output-ready";
    if (states.some((state) => state === "degraded" || state === "needs_operator" || state === "needs-operator")) return "needs recovery";
    if (states.includes("running")) return "running";
    return "queued";
}

function firstExpectedOutput(refs: TeamWorkConfirmationRef[]): string | null {
    for (const ref of refs) {
        const values = Array.isArray(ref.expected_outputs)
            ? ref.expected_outputs
            : Array.isArray(ref.expectedOutputs)
                ? ref.expectedOutputs
                : [];
        for (const value of values) {
            const text = trimToNonEmpty(value);
            if (text) return text;
        }
    }
    return null;
}

function refArray(...values: unknown[]): unknown[] {
    for (const value of values) {
        if (Array.isArray(value) && value.length > 0) return value;
    }
    return [];
}

function hasProofOrReceipt(refs: TeamWorkConfirmationRef[]): boolean {
    return refs.some((ref) => {
        const outputRefs = refArray(ref.output_refs, ref.outputRefs);
        const proofRefs = refArray(ref.proof_refs, ref.proofRefs);
        const auditRefs = refArray(ref.audit_refs, ref.auditRefs);
        return Boolean(trimToNonEmpty(ref.run_id) ?? trimToNonEmpty(ref.runId))
            || proofRefs.length > 0
            || auditRefs.length > 0
            || outputRefs.some((output) => (
                !!output
                && typeof output === "object"
                && (
                    Boolean(trimToNonEmpty((output as Record<string, unknown>).proof_ref))
                    || Boolean(trimToNonEmpty((output as Record<string, unknown>).proofRef))
                    || Boolean(trimToNonEmpty((output as Record<string, unknown>).proof_id))
                    || Boolean(trimToNonEmpty((output as Record<string, unknown>).proofId))
                )
            ));
    });
}

export function teamWorkMessage(refs: TeamWorkConfirmationRef[]): string | null {
    if (refs.length === 0) return null;
    const identifiers = refs
        .map((ref) => shortIdentifier(ref.work_item_id) ?? shortIdentifier(ref.work_id) ?? shortIdentifier(ref.id))
        .filter(Boolean);
    const uniqueIdentifiers = Array.from(new Set(identifiers)).slice(0, 2);
    const workLabel = uniqueIdentifiers.length > 0 ? `Work ${uniqueIdentifiers.join(", ")}` : "Team work";
    const expected = firstExpectedOutput(refs);
    const outputHint = expected ? ` Expected output: ${expected}.` : "";
    const state = teamWorkStateLabel(refs);
    const destination = state === "needs recovery" ? "Review Current Work for recovery." : "Review Current Work and the latest output.";
    const stateCopy = state === "needs recovery" ? "needs recovery" : `is ${state}`;
    const proofHint = hasProofOrReceipt(refs) ? " Proof/receipt is available in Current Work." : "";
    return `${workLabel} ${stateCopy}.${outputHint} ${destination}${proofHint}`;
}
