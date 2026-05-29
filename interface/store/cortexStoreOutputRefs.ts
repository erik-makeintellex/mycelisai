import type {
    ChatArtifactRef,
    DeliveryPackageSummary,
    ExecutionSummary,
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

function normalizeRef(ref: OutputRef): OutputRef {
    return {
        ...ref,
        status: ref.status || 'final',
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
