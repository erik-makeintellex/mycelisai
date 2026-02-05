export type EnvelopeType = 'thought' | 'metric' | 'artifact' | 'governance';

export interface Envelope<T = any> {
    type: EnvelopeType;
    source: string;
    timestamp: string; // ISO8601
    content: T;
}

export interface ThoughtContent {
    summary: string;
    detail: string;
    model: string;
}

export interface MetricContent {
    label: string;
    value: number | string;
    unit: string;
    status: 'nominal' | 'warning' | 'critical';
}

export interface ArtifactContent {
    id: string;
    mime_type: string;
    title: string;
    preview: string; // Summary or snippet
    uri: string;
}

export interface GovernanceContent {
    request_id: string;
    agent_id: string;
    description: string;
    action: string;
    status: 'pending' | 'approved' | 'denied';
}
