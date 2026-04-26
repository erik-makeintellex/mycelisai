import type {
    TeamLeadGuidanceResponse,
    TeamLeadGuidedAction,
} from "@/lib/organizations";

export type RequestState = "idle" | "loading" | "ready" | "error";

export type PersistedWorkspaceState = {
    draftPrompt?: string;
    selectedAction?: TeamLeadGuidedAction | null;
    requestState?: Extract<RequestState, "idle" | "ready">;
    requestContext?: string | null;
    guidance?: TeamLeadGuidanceResponse | null;
};

function storageKeyForOrganization(organizationId: string) {
    return `mycelis-soma-workspace:${organizationId}`;
}

export function readPersistedWorkspaceState(organizationId: string): PersistedWorkspaceState | null {
    if (typeof window === "undefined") {
        return null;
    }
    try {
        const raw = window.localStorage.getItem(storageKeyForOrganization(organizationId));
        if (!raw) {
            return null;
        }
        const parsed = JSON.parse(raw) as PersistedWorkspaceState;
        return typeof parsed === "object" && parsed !== null ? parsed : null;
    } catch {
        return null;
    }
}

export function persistWorkspaceState(organizationId: string, state: PersistedWorkspaceState) {
    if (typeof window === "undefined") {
        return;
    }
    try {
        window.localStorage.setItem(storageKeyForOrganization(organizationId), JSON.stringify(state));
    } catch {
        // best-effort continuity only
    }
}
