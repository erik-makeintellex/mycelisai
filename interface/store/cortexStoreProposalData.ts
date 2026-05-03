import type { ModuleBindingData, ProposalData, TeamExpressionData } from '@/store/cortexStoreTypes';

function uniqueStrings(values: string[]): string[] {
    const seen = new Set<string>();
    const out: string[] = [];
    for (const value of values) {
        const v = value.trim();
        if (!v || seen.has(v)) continue;
        seen.add(v);
        out.push(v);
    }
    return out;
}

function normalizeModuleBindings(raw: unknown): ModuleBindingData[] {
    if (!Array.isArray(raw)) return [];
    const bindings: ModuleBindingData[] = [];
    for (const item of raw) {
        if (!item || typeof item !== 'object') continue;
        const rec = item as Record<string, unknown>;
        const moduleID = typeof rec.module_id === 'string' ? rec.module_id : typeof rec.moduleId === 'string' ? rec.moduleId : '';
        if (!moduleID.trim()) continue;
        bindings.push({
            binding_id: typeof rec.binding_id === 'string' ? rec.binding_id : typeof rec.bindingId === 'string' ? rec.bindingId : undefined,
            module_id: moduleID.trim(),
            adapter_kind: typeof rec.adapter_kind === 'string' ? rec.adapter_kind : typeof rec.adapterKind === 'string' ? rec.adapterKind : undefined,
            operation: typeof rec.operation === 'string' ? rec.operation : undefined,
        });
    }
    return bindings;
}

function normalizeTeamExpressions(raw: unknown): TeamExpressionData[] {
    if (!Array.isArray(raw)) return [];
    const expressions: TeamExpressionData[] = [];
    for (const item of raw) {
        if (!item || typeof item !== 'object') continue;
        const rec = item as Record<string, unknown>;
        const objective = typeof rec.objective === 'string' ? rec.objective.trim() : '';
        const rolePlanRaw = Array.isArray(rec.role_plan) ? rec.role_plan : Array.isArray(rec.rolePlan) ? rec.rolePlan : [];
        const rolePlan = rolePlanRaw.filter((v): v is string => typeof v === 'string').map((v) => v.trim()).filter(Boolean);
        const moduleBindings = normalizeModuleBindings(rec.module_bindings ?? rec.moduleBindings);
        if (!objective && moduleBindings.length === 0) continue;
        expressions.push({
            expression_id: typeof rec.expression_id === 'string' ? rec.expression_id : typeof rec.expressionId === 'string' ? rec.expressionId : undefined,
            team_id: typeof rec.team_id === 'string' ? rec.team_id : typeof rec.teamId === 'string' ? rec.teamId : undefined,
            objective: objective || `Execute ${moduleBindings[0]?.module_id ?? 'operation'}`,
            role_plan: rolePlan,
            module_bindings: moduleBindings,
        });
    }
    return expressions;
}

function normalizeStringArray(raw: unknown): string[] | undefined {
    if (!Array.isArray(raw)) return undefined;
    const values = uniqueStrings(raw.filter((value): value is string => typeof value === 'string'));
    return values.length > 0 ? values : undefined;
}

function pickString(rec: Record<string, unknown>, snake: string, camel: string): string | undefined {
    const value = typeof rec[snake] === 'string' ? rec[snake] : typeof rec[camel] === 'string' ? rec[camel] : '';
    const trimmed = value.trim();
    return trimmed || undefined;
}

export function normalizeProposalData(raw: unknown): ProposalData | undefined {
    if (!raw || typeof raw !== 'object') return undefined;
    const rec = raw as Record<string, unknown>;
    const approval = rec.approval as Record<string, unknown> | undefined;
    const teamExpressions = normalizeTeamExpressions(rec.team_expressions ?? rec.teamExpressions);
    const derivedTools = uniqueStrings(teamExpressions.flatMap((expr) => (expr.module_bindings ?? []).map((b) => b.module_id)));
    const derivedTeams = uniqueStrings(teamExpressions.map((expr) => expr.team_id ?? '')).length;
    const derivedAgents = teamExpressions.reduce((sum, expr) => sum + (expr.role_plan?.length ?? 0), 0);
    const rawTools = Array.isArray(rec.tools) ? rec.tools.filter((v): v is string => typeof v === 'string') : [];

    return {
        intent: typeof rec.intent === 'string' && rec.intent.trim() ? rec.intent : 'chat-action',
        operator_summary: typeof rec.operator_summary === 'string' ? rec.operator_summary.trim() || undefined : undefined,
        expected_result: typeof rec.expected_result === 'string' ? rec.expected_result.trim() || undefined : undefined,
        affected_resources: Array.isArray(rec.affected_resources)
            ? rec.affected_resources.filter((value): value is string => typeof value === 'string').map((value) => value.trim()).filter(Boolean)
            : undefined,
        teams: typeof rec.teams === 'number' ? rec.teams : derivedTeams,
        agents: typeof rec.agents === 'number' ? rec.agents : derivedAgents,
        tools: uniqueStrings(rawTools.length > 0 ? rawTools : derivedTools),
        risk_level: typeof rec.risk_level === 'string' && rec.risk_level.trim() ? rec.risk_level : 'medium',
        confirm_token: typeof rec.confirm_token === 'string' ? rec.confirm_token : '',
        intent_proof_id: typeof rec.intent_proof_id === 'string' ? rec.intent_proof_id : '',
        approval_required: typeof approval?.approval_required === 'boolean' ? Boolean(approval.approval_required) : undefined,
        approval_reason: typeof approval?.approval_reason === 'string' ? String(approval.approval_reason) : undefined,
        approval_mode: typeof approval?.approval_mode === 'string' ? String(approval.approval_mode) : undefined,
        capability_risk: typeof approval?.capability_risk === 'string' ? String(approval.capability_risk) : undefined,
        capability_ids: Array.isArray(approval?.capability_ids) ? (approval.capability_ids as string[]) : undefined,
        external_data_use: typeof approval?.external_data_use === 'boolean' ? Boolean(approval.external_data_use) : undefined,
        estimated_cost: typeof approval?.estimated_cost === 'number' ? Number(approval.estimated_cost) : undefined,
        team_expressions: teamExpressions.length > 0 ? teamExpressions : undefined,
        task_cadence: pickString(rec, 'task_cadence', 'taskCadence') as ProposalData['task_cadence'],
        schedule_summary: pickString(rec, 'schedule_summary', 'scheduleSummary'),
        runtime_posture: pickString(rec, 'runtime_posture', 'runtimePosture'),
        bus_scope: pickString(rec, 'bus_scope', 'busScope') as ProposalData['bus_scope'],
        nats_subjects: normalizeStringArray(rec.nats_subjects ?? rec.natsSubjects),
    };
}
