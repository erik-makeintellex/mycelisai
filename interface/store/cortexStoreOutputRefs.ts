import type {
    ChatArtifactRef,
    DeliveryPackageSummary,
    ExecutionSummary,
    OutputProofEnvelope,
    OutputRef,
} from '@/store/cortexStoreTypes';

function isProofPath(path?: string, title?: string): boolean {
    const lowerPath = path?.toLowerCase() ?? '';
    const lowerTitle = title?.toLowerCase() ?? '';
    return lowerPath.endsWith('proof.md') || lowerTitle.includes('proof');
}

function refKind(contentType?: string, artifactType?: string): string {
    const lower = contentType?.toLowerCase() ?? '';
    if (lower.includes('markdown') || lower.startsWith('text/')) return 'document';
    if (lower.includes('json')) return 'data';
    return artifactType || 'file';
}

// Legacy package messages (older backend responses, or messages restored from
// local persistence before proof envelopes existed) never carried a `proof`
// field. Rather than silently treating those outputs as verified, project a
// degraded envelope from whatever legacy metadata is available so the UI can
// be explicit that readback was never confirmed.
function inferDegradedProof(ref: OutputRef): OutputProofEnvelope | undefined {
    if (!ref.storage_ref && !ref.sha256 && !ref.artifact_id && !ref.proof_artifact_id) {
        return undefined;
    }
    return {
        proof_id: ref.id,
        output_ref_id: ref.id,
        artifact_id: ref.artifact_id ?? ref.proof_artifact_id,
        storage_ref: ref.storage_ref,
        source_run_id: ref.source_run_id,
        source_contract_id: ref.source_contract_id,
        checksum: ref.sha256,
        checksum_algorithm: ref.sha256 ? 'sha256' : undefined,
        bytes: ref.bytes,
        content_type: ref.content_type,
        readback_status: 'degraded',
        recovery_hint: 'Backend did not report a proof envelope for this output; readback is unconfirmed.',
    };
}

function normalizeRef(ref: OutputRef): OutputRef {
    return {
        ...ref,
        status: ref.status || 'final',
        proof: ref.proof ?? inferDegradedProof(ref),
    };
}

export function normalizeOutputRefs(input: {
    artifacts?: ChatArtifactRef[];
    deliveryPackage?: DeliveryPackageSummary;
    executionSummary?: ExecutionSummary;
    runId?: string | null;
}): OutputRef[] {
    const explicit = input.executionSummary?.output_refs
        ?? input.deliveryPackage?.items
        ?? input.deliveryPackage?.inspect_refs;
    if (explicit?.length) {
        return explicit.map(normalizeRef);
    }

    const packageID = input.deliveryPackage?.package_id
        ?? input.deliveryPackage?.package_slug
        ?? input.deliveryPackage?.base_path;
    if (input.deliveryPackage?.files?.length) {
        return input.deliveryPackage.files.map((file, index) => {
            const proof = isProofPath(file.path, file.title);
            return normalizeRef({
                id: packageID ? `${packageID}:${file.path.split('/').pop() ?? index}` : undefined,
                kind: refKind(file.content_type),
                label: file.title || file.path,
                storage_ref: file.path,
                source_run_id: input.runId ?? input.deliveryPackage?.run_id,
                source_contract_id: input.deliveryPackage?.proof_id,
                proof_artifact_id: proof ? undefined : input.deliveryPackage?.proof_path,
                package_id: packageID,
                role: proof ? 'proof' : index === 0 ? 'primary' : 'supporting',
                content_type: file.content_type,
                bytes: file.bytes,
                sha256: file.sha256,
            });
        });
    }

    return (input.artifacts ?? []).map((artifact, index) => {
        const storageRef = artifact.saved_path || artifact.id || artifact.url;
        const proof = isProofPath(artifact.saved_path, artifact.title);
        return normalizeRef({
            id: artifact.id || storageRef || `artifact-${index}`,
            kind: refKind(artifact.content_type, artifact.type),
            label: artifact.title,
            storage_ref: storageRef,
            open_url: artifact.url || (artifact.id ? `/api/v1/artifacts/${encodeURIComponent(artifact.id)}/download` : undefined),
            source_run_id: input.runId ?? input.deliveryPackage?.run_id,
            source_contract_id: input.deliveryPackage?.proof_id,
            proof_artifact_id: proof ? undefined : input.deliveryPackage?.proof_path,
            package_id: packageID,
            role: proof ? 'proof' : index === 0 ? 'primary' : 'supporting',
            content_type: artifact.content_type,
        });
    });
}
