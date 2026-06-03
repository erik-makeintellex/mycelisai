import { trimToNonEmpty } from "@/store/cortexStoreChatWorkflow";

export type TeamWorkConfirmationRef = {
    id?: string;
    work_item_id?: string;
    work_id?: string;
    state?: string;
    status?: string;
    output_count?: number;
    outputCount?: number;
    output_refs?: unknown[];
    outputRefs?: unknown[];
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
    if (states.includes("running")) return "running";
    return "queued";
}

export function teamWorkMessage(refs: TeamWorkConfirmationRef[]): string | null {
    if (refs.length === 0) return null;
    const identifiers = refs
        .map((ref) => shortIdentifier(ref.work_item_id) ?? shortIdentifier(ref.work_id) ?? shortIdentifier(ref.id))
        .filter(Boolean);
    const uniqueIdentifiers = Array.from(new Set(identifiers)).slice(0, 2);
    const workLabel = uniqueIdentifiers.length > 0 ? `Work ${uniqueIdentifiers.join(", ")}` : "Team work";
    return `${workLabel} is ${teamWorkStateLabel(refs)}. Review Active Work and the latest output.`;
}
