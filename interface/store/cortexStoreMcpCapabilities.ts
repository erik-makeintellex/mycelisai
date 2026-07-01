import type { CapabilityManifest, SearchCapabilitySource, SearchCapabilityStatus } from '@/store/cortexStoreTypes';

export function isSearchCapabilityStatus(value: unknown): value is SearchCapabilityStatus {
    return typeof value === 'object'
        && value !== null
        && typeof (value as SearchCapabilityStatus).provider === 'string'
        && typeof (value as SearchCapabilityStatus).enabled === 'boolean'
        && typeof (value as SearchCapabilityStatus).configured === 'boolean';
}

export function normalizeSearchCapabilityStatus(value: unknown): SearchCapabilityStatus | null {
    if (!isSearchCapabilityStatus(value)) return null;
    return {
        ...value,
        sources: normalizeSearchSourcesPayload(value.sources),
    };
}

export function normalizeSearchSourcesPayload(value: unknown): SearchCapabilitySource[] {
    const candidate = Array.isArray(value)
        ? value
        : typeof value === 'object' && value !== null && Array.isArray((value as { sources?: unknown }).sources)
        ? (value as { sources: unknown[] }).sources
        : typeof value === 'object' && value !== null && Array.isArray((value as { search_sources?: unknown }).search_sources)
        ? (value as { search_sources: unknown[] }).search_sources
        : [];
    return candidate.map(normalizeSearchSource).filter((source): source is SearchCapabilitySource => source !== null);
}

export function normalizeCapabilitiesPayload(value: unknown): CapabilityManifest[] {
    const candidate = Array.isArray(value)
        ? value
        : typeof value === 'object' && value !== null && Array.isArray((value as { capabilities?: unknown }).capabilities)
        ? (value as { capabilities: unknown[] }).capabilities
        : typeof value === 'object' && value !== null && Array.isArray((value as { manifests?: unknown }).manifests)
        ? (value as { manifests: unknown[] }).manifests
        : [];
    return candidate.map(normalizeCapabilityManifest).filter(isCapabilityManifest);
}

function normalizeSearchSource(value: unknown): SearchCapabilitySource | null {
    if (typeof value !== 'object' || value === null) return null;
    const raw = value as Record<string, unknown>;
    const id = stringField(raw.id) ?? stringField(raw.name);
    const name = stringField(raw.name) ?? stringField(raw.display_name) ?? id;
    if (!id || !name) return null;
    return {
        id,
        name,
        provider: stringField(raw.provider),
        source_type: stringField(raw.source_type) ?? stringField(raw.type) ?? 'unknown',
        endpoint: stringField(raw.endpoint),
        base_url: stringField(raw.base_url),
        scope_kind: stringField(raw.scope_kind) ?? 'all',
        scope_ref: stringField(raw.scope_ref),
        boundary: stringField(raw.boundary) ?? stringField(raw.description) ?? 'Configured search boundary',
        auth_scheme: stringField(raw.auth_scheme) ?? stringField(raw.auth) ?? 'none',
        secret_ref: stringField(raw.secret_ref),
        mode: stringField(raw.mode) ?? 'live',
        sensitivity_class: stringField(raw.sensitivity_class) ?? 'unknown',
        trust_class: stringField(raw.trust_class) ?? 'unknown',
        status: stringField(raw.status) ?? stringField(raw.availability_status) ?? 'available',
        recovery: stringField(raw.recovery),
    };
}

function isCapabilityManifest(value: unknown): value is CapabilityManifest {
    return typeof value === 'object'
        && value !== null
        && typeof (value as CapabilityManifest).id === 'string'
        && typeof (value as CapabilityManifest).name === 'string'
        && typeof (value as CapabilityManifest).source === 'string'
        && typeof (value as CapabilityManifest).category === 'string'
        && typeof (value as CapabilityManifest).risk === 'string'
        && typeof (value as CapabilityManifest).approval === 'string';
}

function normalizeCapabilityManifest(value: unknown): CapabilityManifest | null {
    if (isCapabilityManifest(value)) return value;
    if (typeof value !== 'object' || value === null) return null;
    const raw = value as Record<string, unknown>;
    const id = stringField(raw.id);
    const source = stringField(raw.source);
    const category = stringField(raw.category) ?? stringField(raw.kind);
    const risk = stringField(raw.risk) ?? stringField(raw.risk_class);
    if (!id || !source || !category || !risk) return null;

    const metadata = raw.metadata && typeof raw.metadata === 'object'
        ? raw.metadata as Record<string, unknown>
        : {};
    const approvalRequired = booleanField(raw.approval_required);
    const auditRequired = booleanField(raw.audit_required);
    const toolRefs = stringArrayField(raw.tool_refs);

    return {
        id,
        name: stringField(raw.name) ?? stringField(raw.display_name) ?? id,
        description: stringField(raw.description),
        source,
        category,
        risk,
        approval: stringField(raw.approval) ?? (approvalRequired ? 'required' : 'none'),
        inputs: stringArrayField(raw.inputs),
        outputs: operatorOutputLabels(raw),
        writes: stringArrayField(raw.writes),
        allowed_roles: stringArrayField(raw.allowed_roles) ?? stringArrayField(raw.default_allowed_roles),
        audit: stringField(raw.audit) ?? (auditRequired ? 'required' : 'none'),
        health_check: booleanField(raw.health_check),
        availability_status: stringField(raw.availability_status) ?? stringField(raw.status),
        fallback_behavior: stringField(raw.fallback_behavior),
        retention_policy: stringField(raw.retention_policy),
        review_required: booleanField(raw.review_required),
        server_or_package: stringField(raw.server_or_package)
            ?? stringField(metadata.server_name)
            ?? stringField(metadata.repository)
            ?? stringField(metadata.provider),
        config_refs: stringArrayField(raw.config_refs),
        secret_refs: stringArrayField(raw.secret_refs) ?? stringArrayField(metadata.required_env),
        provider: stringField(raw.provider) ?? stringField(metadata.provider),
        bound_server_id: stringField(raw.bound_server_id) ?? stringField(metadata.server_id),
        bound_server_name: stringField(raw.bound_server_name) ?? stringField(metadata.server_name),
        bound_tool_id: stringField(raw.bound_tool_id) ?? stringField(metadata.tool_id),
        bound_tool_name: stringField(raw.bound_tool_name) ?? firstToolName(toolRefs),
    };
}

function stringField(value: unknown): string | undefined {
    return typeof value === 'string' && value.trim() ? value : undefined;
}

function booleanField(value: unknown): boolean | undefined {
    return typeof value === 'boolean' ? value : undefined;
}

function stringArrayField(value: unknown): string[] | undefined {
    if (!Array.isArray(value)) return undefined;
    const values = value.filter((item): item is string => typeof item === 'string' && item.trim().length > 0);
    return values.length > 0 ? values : undefined;
}

function operatorOutputLabels(raw: Record<string, unknown>): string[] | undefined {
    const direct = stringArrayField(raw.outputs);
    if (direct) return direct;
    const schema = stringField(raw.output_schema_ref);
    if (!schema) return undefined;
    const label = schema.split(/[/.#]/).filter(Boolean).pop();
    return [label ? `${label} output` : 'Structured output'];
}

function firstToolName(toolRefs?: string[]): string | undefined {
    const ref = toolRefs?.[0];
    if (!ref) return undefined;
    const last = ref.split('/').pop()?.trim();
    return last && last !== '*' ? last : undefined;
}
