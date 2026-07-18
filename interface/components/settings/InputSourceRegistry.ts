"use client";

import { useCallback, useState } from "react";
import { extractApiData } from "@/lib/apiContracts";

export type InputSourceStatus = "available" | "paused" | "error" | string;

export type InputSource = {
    id: string;
    name: string;
    source_type: string;
    adapter_kind: string;
    scope_kind: string;
    scope_ref?: string;
    target_outcome_id?: string;
    target_group_id?: string;
    target_host_id?: string;
    auth_scheme: string;
    secret_ref?: string;
    allowed_ingress_subject: string;
    payload_schema_ref?: string;
    buffer_mode: string;
    sensitivity_class: string;
    trust_class: string;
    status: InputSourceStatus;
    recovery?: string;
};

type BufferEvent = {
    event_id: string;
    channel_key: string;
    payload?: unknown;
    received_at?: string;
    source_kind?: string;
    payload_kind?: string;
};

type LatestValue = {
    event_id?: string;
    channel_key: string;
    payload?: unknown;
    received_at?: string;
};

type WindowSummary = {
    channel_key: string;
    window_key: string;
    summary?: string;
    count?: number;
    started_at?: string;
    ended_at?: string;
};

export type InputSourceBufferView = {
    mode: string;
    source?: InputSource;
    events?: BufferEvent[];
    latest?: LatestValue[];
    windows?: WindowSummary[];
};

export function useInputSourceRegistry() {
    const [sources, setSources] = useState<InputSource[]>([]);
    const [isFetchingSources, setIsFetchingSources] = useState(false);
    const [sourcesError, setSourcesError] = useState<string | null>(null);
    const [registrySupported, setRegistrySupported] = useState(false);
    const [selectedSourceId, setSelectedSourceId] = useState<string | null>(null);
    const [bufferView, setBufferView] = useState<InputSourceBufferView | null>(null);
    const [isFetchingBuffer, setIsFetchingBuffer] = useState(false);
    const [bufferError, setBufferError] = useState<string | null>(null);

    const fetchInputSources = useCallback(async () => {
        let touchedLoadingState = false;
        try {
            const res = await fetch("/api/v1/input-sources");
            if (!res || res.status === 404 || res.status === 405) return;
            setRegistrySupported(true);
            setIsFetchingSources(true);
            touchedLoadingState = true;
            setSourcesError(null);
            if (res.ok) {
                const payload = await res.json();
                const normalized = normalizeInputSources(extractApiData<unknown>(payload));
                setSources(normalized);
                setSelectedSourceId((current) => current ?? normalized[0]?.id ?? null);
                return;
            }
            setSources([]);
            setSourcesError(`Live input registry unreachable (HTTP ${res.status})`);
        } catch {
            setRegistrySupported(false);
            setSources([]);
        } finally {
            if (touchedLoadingState) setIsFetchingSources(false);
        }
    }, []);

    const fetchSourceBuffer = useCallback(async (sourceId: string, mode?: string) => {
        setSelectedSourceId(sourceId);
        setIsFetchingBuffer(true);
        setBufferError(null);
        try {
            const query = new URLSearchParams({ limit: "5" });
            if (mode) query.set("mode", mode);
            const res = await fetch(`/api/v1/input-sources/${encodeURIComponent(sourceId)}/buffer?${query.toString()}`);
            if (res.ok) {
                const payload = await res.json();
                setBufferView(normalizeBufferView(extractApiData<unknown>(payload)));
                return;
            }
            setBufferView(null);
            setBufferError(`Buffer preview unavailable (HTTP ${res.status})`);
        } catch (err) {
            const message = err instanceof Error ? err.message : "network error";
            setBufferView(null);
            setBufferError(`Buffer preview unavailable (${message})`);
        } finally {
            setIsFetchingBuffer(false);
        }
    }, []);

    return {
        bufferError,
        bufferView,
        fetchInputSources,
        fetchSourceBuffer,
        isFetchingBuffer,
        isFetchingSources,
        registrySupported,
        selectedSourceId,
        sources,
        sourcesError,
    };
}

function normalizeInputSources(value: unknown): InputSource[] {
    const raw = Array.isArray(value)
        ? value
        : Array.isArray((value as { sources?: unknown[] } | null)?.sources)
            ? (value as { sources: unknown[] }).sources
            : [];
    return raw.map(normalizeInputSource).filter((source): source is InputSource => Boolean(source));
}

function normalizeInputSource(value: unknown): InputSource | null {
    if (!value || typeof value !== "object") return null;
    const rec = value as Record<string, unknown>;
    const id = stringValue(rec.id);
    const name = stringValue(rec.name);
    if (!id || !name) return null;
    return {
        id,
        name,
        source_type: stringValue(rec.source_type) || "event",
        adapter_kind: stringValue(rec.adapter_kind) || "api",
        scope_kind: stringValue(rec.scope_kind) || "all",
        scope_ref: stringValue(rec.scope_ref),
        target_outcome_id: stringValue(rec.target_outcome_id),
        target_group_id: stringValue(rec.target_group_id),
        target_host_id: stringValue(rec.target_host_id),
        auth_scheme: stringValue(rec.auth_scheme) || "none",
        secret_ref: stringValue(rec.secret_ref),
        allowed_ingress_subject: stringValue(rec.allowed_ingress_subject),
        payload_schema_ref: stringValue(rec.payload_schema_ref),
        buffer_mode: stringValue(rec.buffer_mode) || "append_log",
        sensitivity_class: stringValue(rec.sensitivity_class) || "governed",
        trust_class: stringValue(rec.trust_class) || "bounded_input",
        status: stringValue(rec.status) || "available",
        recovery: stringValue(rec.recovery),
    };
}

function normalizeBufferView(value: unknown): InputSourceBufferView {
    if (!value || typeof value !== "object") return { mode: "append_log" };
    const rec = value as Record<string, unknown>;
    return {
        mode: stringValue(rec.mode) || "append_log",
        source: normalizeInputSource(rec.source) ?? undefined,
        events: Array.isArray(rec.events) ? rec.events.map(normalizeEvent).filter(Boolean) as BufferEvent[] : [],
        latest: Array.isArray(rec.latest) ? rec.latest.map(normalizeLatest).filter(Boolean) as LatestValue[] : [],
        windows: Array.isArray(rec.windows) ? rec.windows.map(normalizeWindow).filter(Boolean) as WindowSummary[] : [],
    };
}

function normalizeEvent(value: unknown): BufferEvent | null {
    if (!value || typeof value !== "object") return null;
    const rec = value as Record<string, unknown>;
    return {
        event_id: stringValue(rec.event_id) || stringValue(rec.id),
        channel_key: stringValue(rec.channel_key) || "default",
        payload: rec.payload,
        received_at: stringValue(rec.received_at),
        source_kind: stringValue(rec.source_kind),
        payload_kind: stringValue(rec.payload_kind),
    };
}

function normalizeLatest(value: unknown): LatestValue | null {
    if (!value || typeof value !== "object") return null;
    const rec = value as Record<string, unknown>;
    return {
        event_id: stringValue(rec.event_id),
        channel_key: stringValue(rec.channel_key) || "default",
        payload: rec.payload,
        received_at: stringValue(rec.received_at),
    };
}

function normalizeWindow(value: unknown): WindowSummary | null {
    if (!value || typeof value !== "object") return null;
    const rec = value as Record<string, unknown>;
    return {
        channel_key: stringValue(rec.channel_key) || "default",
        window_key: stringValue(rec.window_key) || "window",
        summary: stringValue(rec.summary),
        count: typeof rec.count === "number" ? rec.count : undefined,
        started_at: stringValue(rec.started_at),
        ended_at: stringValue(rec.ended_at),
    };
}

function stringValue(value: unknown): string {
    return typeof value === "string" ? value : "";
}
