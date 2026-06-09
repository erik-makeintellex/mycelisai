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
    if (!degradation || typeof degradation !== "object") return null;
    const joined = [
        compactText(degradation.what_failed),
        compactText(degradation.trusted_state),
        compactText(degradation.invalidated_proof),
        compactText(degradation.safe_continuation),
    ].filter(Boolean).join(" ").toLowerCase();
    if (
        !joined.includes("comfyui")
        && !joined.includes("media engine")
        && !joined.includes("media capability")
        && !joined.includes("local/private")
    ) {
        return null;
    }
    return {
        failed: "Local media generation is not reachable, so Soma could not create the image output.",
        trusted: "The approval, request, failed run record, and audit trail remain available for review.",
        invalid: "No completed image output or execution proof should be trusted for this attempt.",
        recovery: "Start or reconnect the configured ComfyUI upstream, then retry. If you only need text/files, ask Soma to rerun without image generation.",
    };
}
