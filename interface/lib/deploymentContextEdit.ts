// MEM W2: PATCH client for editing a saved deployment-context entry in
// place. Kept separate from DeploymentContextEntryActions.tsx's own
// archive/restore/delete fetch calls so this shape (changed-fields-only
// body, {ok, data:{changed}} response) has one place to unit test.
export type DeploymentContextEditPatch = {
    title?: string;
    content?: string;
    source_label?: string;
};

export type DeploymentContextEditResult =
    | { ok: true; changed: boolean }
    | { ok: false; code: string; httpStatus: number };

export async function editDeploymentContextEntry(
    artifactId: string,
    patch: DeploymentContextEditPatch,
): Promise<DeploymentContextEditResult> {
    try {
        const res = await fetch(`/api/v1/memory/deployment-context/${encodeURIComponent(artifactId)}`, {
            method: "PATCH",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(patch),
        });
        const payload = await res.json().catch(() => ({}));
        if (!res.ok || payload?.ok === false) {
            const code = typeof payload?.data?.code === "string" ? payload.data.code : "request_failed";
            return { ok: false, code, httpStatus: res.status };
        }
        return { ok: true, changed: payload?.data?.changed !== false };
    } catch {
        return { ok: false, code: "request_failed", httpStatus: 0 };
    }
}
