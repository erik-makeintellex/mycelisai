import type { ServiceHealthStatus } from '@/store/cortexStoreTypes';

/**
 * Single operator-facing signal for "can the interface reach Core right now?".
 * Derived from the services-status poll (ShellLayout) plus the SSE stream state.
 */
export type CoreReachability = 'online' | 'degraded' | 'unreachable' | 'unknown';

export interface ServicesStatusFailure {
    /** unreachable = network error or 5xx from the proxy; http_error = Core answered with a 4xx. */
    kind: 'unreachable' | 'http_error';
    statusCode?: number;
    at: string;
}

export interface CoreReachabilityInputs {
    servicesStatus?: ServiceHealthStatus[];
    servicesStatusFailure?: ServicesStatusFailure | null;
    servicesStatusUpdatedAt?: string | null;
    streamConnectionState?: 'idle' | 'connecting' | 'online' | 'offline';
}

const CORE_CRITICAL_SERVICES = ['nats', 'postgres', 'cognitive'];

export function classifyServicesStatusFailure(statusCode?: number): ServicesStatusFailure {
    const unreachable = statusCode == null || statusCode >= 500;
    return {
        kind: unreachable ? 'unreachable' : 'http_error',
        statusCode,
        at: new Date().toISOString(),
    };
}

export function deriveCoreReachability(state: CoreReachabilityInputs): CoreReachability {
    const failure = state.servicesStatusFailure;
    if (failure?.kind === 'unreachable') return 'unreachable';
    if (failure) return 'degraded';
    if (!state.servicesStatusUpdatedAt) return 'unknown';
    const services = state.servicesStatus ?? [];
    const serviceDown = services.some(
        (svc) => CORE_CRITICAL_SERVICES.includes(svc.name) && svc.status !== 'online',
    );
    if (serviceDown || state.streamConnectionState === 'offline') return 'degraded';
    return 'online';
}

/** Zustand selector: `useCortexStore(selectCoreReachability)`. Returns a primitive, so it is render-stable. */
export const selectCoreReachability = (state: CoreReachabilityInputs): CoreReachability =>
    deriveCoreReachability(state);

// ── Shared operator copy ─────────────────────────────────────────────────────
export const CORE_UNREACHABLE_TITLE = 'Core API unreachable';
export const CORE_UNREACHABLE_SUMMARY =
    'The interface cannot reach the Mycelis Core API, so Soma, groups, runs, and service health are unavailable.';
export const CORE_START_COMMAND = 'uv run inv lifecycle.up --frontend';
export const CORE_STATUS_COMMAND = 'uv run inv lifecycle.status';
export const CORE_RECOVERY_HINT = `Start the runtime with \`${CORE_START_COMMAND}\`, then confirm with \`${CORE_STATUS_COMMAND}\`.`;
export const NOT_CHECKED_CORE_UNREACHABLE = 'not checked (Core unreachable)';
