export type MissionChatFailureType =
    | 'timeout'
    | 'unreachable'
    | 'server_error'
    | 'permission_denied'
    | 'unknown';

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
}

function classifyMissionChatFailure(message: string, statusCode?: number): MissionChatFailureType {
    const lower = message.toLowerCase();
    if (statusCode === 401 || statusCode === 403 || lower.includes('401') || lower.includes('403') || lower.includes('unauthorized') || lower.includes('forbidden')) {
        return 'permission_denied';
    }
    if (lower.includes('timeout') || lower.includes('deadline exceeded')) {
        return 'timeout';
    }
    if (
        (statusCode != null && statusCode >= 500) ||
        lower.includes('500') ||
        lower.includes('internal error') ||
        lower.includes('server error')
    ) {
        return 'server_error';
    }
    if (
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
}: {
    assistantName: string;
    targetId: string;
    message: string;
    statusCode?: number;
}): MissionChatFailure {
    const routeKind = targetId === 'admin' ? 'workspace' : 'council';
    const targetLabel = routeKind === 'workspace' ? assistantName : targetId;
    const type = classifyMissionChatFailure(message, statusCode);

    if (routeKind === 'workspace') {
        const bannerLabel = {
            timeout: 'Workspace chat timeout',
            unreachable: 'Workspace chat unreachable',
            server_error: 'Workspace chat server error',
            permission_denied: 'Workspace chat auth failure',
            unknown: 'Workspace chat blocked',
        }[type];
        const summary = {
            timeout: `${assistantName} did not return a response before the request deadline.`,
            unreachable: `${assistantName} or the local API proxy is currently unreachable from this client.`,
            server_error: `${assistantName} hit a server-side failure while handling the request.`,
            permission_denied: `${assistantName} rejected the request because the UI client is not authorized.`,
            unknown: `${assistantName} could not complete the request and returned an unclassified blocker.`,
        }[type];
        const recommendedAction = {
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
            title: `${assistantName} Chat Blocked`,
            bannerLabel,
            summary,
            recommendedAction,
            diagnostics: message,
            statusCode,
        };
    }

    const summary = {
        timeout: 'The council member did not respond before the request deadline.',
        unreachable: 'The council member service or proxy is currently unreachable from this client.',
        server_error: 'The council member service returned an internal error while handling the request.',
        permission_denied: 'The council member rejected the request because the UI client is not authorized.',
        unknown: 'The council request failed unexpectedly and did not return a classified blocker.',
    }[type];
    const recommendedAction = {
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
        title: 'Council Call Failed',
        bannerLabel: `Council member ${type.replace('_', ' ')}`,
        summary,
        recommendedAction,
        diagnostics: message,
        statusCode,
    };
}
