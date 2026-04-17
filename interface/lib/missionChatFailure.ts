export type MissionChatFailureType =
    | 'setup_required'
    | 'timeout'
    | 'unreachable'
    | 'server_error'
    | 'permission_denied'
    | 'unknown';

export interface MissionChatAvailability {
    available?: boolean;
    code?: string;
    summary?: string;
    recommended_action?: string;
    profile?: string;
    provider_id?: string;
    model_id?: string;
    setup_required?: boolean;
    setup_path?: string;
    fallback_applied?: boolean;
}

export interface MissionChatFailure {
    routeKind: 'workspace' | 'council';
    targetId: string;
    targetLabel: string;
    type: MissionChatFailureType;
    title: string;
    bannerLabel: string;
    summary: string;
    recommendedAction: string;
    diagnostics: string;
    statusCode?: number;
    availability?: MissionChatAvailability;
    setupPath?: string;
}

function classifyMissionChatFailure(message: string, statusCode?: number, availability?: MissionChatAvailability): MissionChatFailureType {
    if (
        availability?.setup_required ||
        availability?.code === 'no_provider_available' ||
        availability?.code === 'profile_unbound' ||
        availability?.code === 'provider_missing' ||
        availability?.code === 'provider_disabled' ||
        availability?.code === 'provider_uninitialized' ||
        availability?.code === 'model_missing' ||
        availability?.code === 'router_unavailable'
    ) {
        return 'setup_required';
    }
    if (availability?.code === 'transport_timeout') {
        return 'timeout';
    }
    if (availability?.code === 'transport_unavailable' || availability?.code === 'runtime_unavailable') {
        return 'unreachable';
    }
    if (availability?.code === 'transport_backpressure') {
        return 'server_error';
    }
    const lower = message.toLowerCase();
    if (statusCode === 401 || statusCode === 403 || lower.includes('401') || lower.includes('403') || lower.includes('unauthorized') || lower.includes('forbidden')) {
        return 'permission_denied';
    }
    if (statusCode === 504 || lower.includes('timeout') || lower.includes('deadline exceeded')) {
        return 'timeout';
    }
    if (
        (statusCode != null && statusCode >= 500 && statusCode !== 502 && statusCode !== 503 && statusCode !== 504) ||
        lower.includes('500') ||
        lower.includes('internal server error') ||
        lower.includes('internal error') ||
        lower.includes('server error')
    ) {
        return 'server_error';
    }
    if (
        statusCode === 502 ||
        statusCode === 503 ||
        lower.includes('could not reach') ||
        lower.includes('unreachable') ||
        lower.includes('failed to fetch') ||
        lower.includes('bad gateway') ||
        lower.includes('connection refused') ||
        lower.includes('503') ||
        lower.includes('502') ||
        lower.includes('offline')
    ) {
        return 'unreachable';
    }
    return 'unknown';
}

export function buildMissionChatFailure({
    assistantName,
    targetId,
    message,
    statusCode,
    availability,
}: {
    assistantName: string;
    targetId: string;
    message: string;
    statusCode?: number;
    availability?: MissionChatAvailability;
}): MissionChatFailure {
    const routeKind = targetId === 'admin' ? 'workspace' : 'council';
    const targetLabel = routeKind === 'workspace' ? assistantName : targetId;
    const type = classifyMissionChatFailure(message, statusCode, availability);

    if (routeKind === 'workspace') {
        const bannerLabel = {
            setup_required: 'AI engine setup required',
            timeout: 'Workspace chat timeout',
            unreachable: 'Workspace chat unreachable',
            server_error: 'Workspace chat server error',
            permission_denied: 'Workspace chat auth failure',
            unknown: 'Workspace chat blocked',
        }[type];
        const summary = {
            setup_required: availability?.summary || `${assistantName} does not currently have an available AI engine for workspace chat.`,
            timeout: `${assistantName} did not return a response before the request deadline.`,
            unreachable: `${assistantName} or the local API proxy is currently unreachable from this client.`,
            server_error: `${assistantName} hit a server-side failure while handling the request.`,
            permission_denied: `${assistantName} rejected the request because the UI client is not authorized.`,
            unknown: `${assistantName} could not complete the request and returned an unclassified blocker.`,
        }[type];
        const recommendedAction = {
            setup_required: availability?.recommended_action || `Open Settings and verify that at least one AI Engine is enabled and reachable for ${assistantName}.`,
            timeout: 'Retry once, then open System Status if the timeout repeats.',
            unreachable: 'Open System Status and verify Core, NATS, and the local proxy are online.',
            server_error: 'Retry once. If the failure persists, inspect System Status and recent startup logs.',
            permission_denied: 'Refresh the session or re-check the local auth configuration before retrying.',
            unknown: 'Copy diagnostics and inspect System Status for the failing dependency.',
        }[type];
        return {
            routeKind,
            targetId,
            targetLabel,
            type,
            title: type === 'setup_required' ? `${assistantName} Setup Required` : `${assistantName} Chat Blocked`,
            bannerLabel,
            summary,
            recommendedAction,
            diagnostics: message,
            statusCode,
            availability,
            setupPath: availability?.setup_path || '/settings',
        };
    }

    const summary = {
        setup_required: availability?.summary || `${assistantName} cannot continue because the shared AI engine setup is incomplete.`,
        timeout: 'The council member did not respond before the request deadline.',
        unreachable: 'The council member service or proxy is currently unreachable from this client.',
        server_error: 'The council member service returned an internal error while handling the request.',
        permission_denied: 'The council member rejected the request because the UI client is not authorized.',
        unknown: 'The council request failed unexpectedly and did not return a classified blocker.',
    }[type];
    const recommendedAction = {
        setup_required: availability?.recommended_action || `Open Settings and verify that ${assistantName} has an available AI Engine before retrying council routing.`,
        timeout: `Retry once, then continue with ${assistantName} or open System Status if the blocker persists.`,
        unreachable: `Switch to ${assistantName} or open System Status to inspect runtime connectivity.`,
        server_error: `Retry once. If it fails again, continue with ${assistantName} and inspect System Status.`,
        permission_denied: `Continue with ${assistantName} or refresh the session before retrying this council member.`,
        unknown: `Copy diagnostics and continue with ${assistantName} while investigating the council runtime.`,
    }[type];

    return {
        routeKind,
        targetId,
        targetLabel,
        type,
        title: type === 'setup_required' ? `${assistantName} Setup Required` : 'Council Call Failed',
        bannerLabel: type === 'setup_required' ? 'AI engine setup required' : `Council member ${type.replace('_', ' ')}`,
        summary,
        recommendedAction,
        diagnostics: message,
        statusCode,
        availability,
        setupPath: availability?.setup_path || '/settings',
    };
}
