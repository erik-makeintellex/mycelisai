import { localMediaDependencyRecovery } from "@/lib/deliveryRuntimeLanguage";

export type DegradationShape = {
    what_failed?: string;
    trusted_state?: string;
    invalidated_proof?: string;
    safe_continuation?: string;
};

function compactText(value?: string | null) {
    const trimmed = value?.trim();
    return trimmed || null;
}

export function mediaDependencyRecovery(degradation?: DegradationShape | null) {
    return localMediaDependencyRecovery(degradation);
}
