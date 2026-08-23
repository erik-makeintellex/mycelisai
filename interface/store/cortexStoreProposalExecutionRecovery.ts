import { trimToNonEmpty } from '@/store/cortexStoreChatWorkflow';
import type { ChatMessage } from '@/store/cortexStoreTypes';
import type { ExecutionSummaryData } from '@/store/cortexStoreTypesExecutionSummary';

export type ConfirmFailureBody = {
    error?: string;
    data?: {
        run_id?: string;
        execution_summary?: ExecutionSummaryData;
    };
};

export function recoveryTextFromExecutionSummary(summary: ExecutionSummaryData | undefined) {
    const auditRecovery = summary?.audit_recovery && typeof summary.audit_recovery === 'object'
        ? summary.audit_recovery
        : undefined;
    const degradation = auditRecovery?.degradation;
    const code = trimToNonEmpty(degradation?.code);
    const whatFailed = trimToNonEmpty(degradation?.what_failed)
        ?? trimToNonEmpty(auditRecovery?.blocker);
    const safeContinuation = trimToNonEmpty(degradation?.safe_continuation);
    const diagnostics = [
        code,
        whatFailed,
        trimToNonEmpty(degradation?.trusted_state),
        trimToNonEmpty(degradation?.invalidated_proof),
        safeContinuation,
    ].filter(Boolean).join(' | ');
    return { code, whatFailed, safeContinuation, diagnostics: trimToNonEmpty(diagnostics) };
}

export function failureThreadEvent({
    runId,
    recovery,
}: {
    runId?: string | null;
    recovery: ReturnType<typeof recoveryTextFromExecutionSummary>;
}): NonNullable<ChatMessage['thread_events']>[number] {
    const code = recovery.code ?? 'approved_execution_failed';
    const contractUnsatisfied = code === 'result_contract_unsatisfied';
    return {
        kind: 'attention_required',
        label: contractUnsatisfied ? 'Output is not playable yet' : 'Soma needs your direction',
        detail: contractUnsatisfied
            ? 'The team stopped because the required runnable output was not validated. No playable output should be trusted for this attempt.'
            : recovery.whatFailed ?? 'This work stopped before a usable result was produced.',
        tone: 'warning',
        status: code,
        run_id: runId ?? undefined,
        source_kind: 'web_api',
        source_channel: 'api.intent.confirm-action',
        payload_kind: 'soma_thread_event',
        target_reference: code,
        timestamp: new Date().toISOString(),
    };
}

export function isMediaDependencyFailure(message?: string | null) {
    const lower = (message ?? '').toLowerCase();
    return lower.includes('comfyui')
        || lower.includes('forge')
        || lower.includes('media engine')
        || lower.includes('media capability')
        || lower.includes('local/private');
}

export function mediaDependencyRecoveryCopy(diagnostics: string) {
    const forgeAPIDisabled = diagnostics.toLowerCase().includes('forge')
        && (diagnostics.toLowerCase().includes('api mode') || diagnostics.toLowerCase().includes('api access is off'));
    return {
        summary: forgeAPIDisabled
            ? 'Forge is open, but Soma cannot use image generation until its API mode is enabled.'
            : 'The configured image generator is not ready, so Soma could not create the requested image output.',
        recommendedAction: forgeAPIDisabled
            ? "Enable API mode in Forge's Pinokio launch settings, restart Forge, then tell Soma to try the image again."
            : 'Start or reconnect the configured image generator, then tell Soma to try the image again.',
        diagnostics,
    };
}
