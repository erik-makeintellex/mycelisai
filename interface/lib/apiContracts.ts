export interface ApiEnvelope<T = unknown> {
    ok?: boolean;
    data?: T;
    error?: string;
}

export interface ResourceCallRequest<TArgs = Record<string, unknown>> {
    arguments: TArgs;
}

export function extractApiData<T>(payload: unknown): T {
    if (payload && typeof payload === "object" && "data" in (payload as Record<string, unknown>)) {
        return (payload as ApiEnvelope<T>).data as T;
    }
    return payload as T;
}

export function extractApiError(payload: unknown): string | undefined {
    if (!payload || typeof payload !== "object") return undefined;
    const error = (payload as ApiEnvelope).error;
    return typeof error === "string" ? error : undefined;
}

export function formatMCPToolResult(payload: unknown): string {
    const data = extractApiData<unknown>(payload);
    if (data == null) return "(no output)";
    if (typeof data === "string") return data;

    if (typeof data === "object" && Array.isArray((data as { content?: unknown[] }).content)) {
        const texts = ((data as { content: unknown[] }).content ?? [])
            .map((c) => {
                if (!c || typeof c !== "object") return null;
                const rec = c as Record<string, unknown>;
                if (typeof rec.text === "string") return rec.text;
                return null;
            })
            .filter((v): v is string => Boolean(v));
        if (texts.length > 0) return texts.join("\n");
    }

    try {
        return JSON.stringify(data, null, 2);
    } catch {
        return String(data);
    }
}

