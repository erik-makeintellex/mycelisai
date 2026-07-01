"use client";

import { useCallback, useMemo, useState } from "react";
import { extractApiData } from "@/lib/apiContracts";
import { normalizeSearchSourcesPayload } from "@/store/cortexStoreMcpCapabilities";
import type { SearchCapabilitySource, SearchCapabilityStatus } from "@/store/useCortexStore";
import type { SearchSourceDraft } from "./SearchSourceForm";

export function useSearchSourceRegistry(
    searchSources: SearchCapabilityStatus["sources"] | undefined,
    fetchSearchCapability: () => Promise<void> | void,
) {
    const [registrySearchSources, setRegistrySearchSources] = useState<SearchCapabilitySource[]>([]);
    const [isFetchingSearchSources, setIsFetchingSearchSources] = useState(false);
    const [searchSourceRegistrySupported, setSearchSourceRegistrySupported] = useState(false);
    const [searchSourcesError, setSearchSourcesError] = useState<string | null>(null);
    const [searchSourceNotice, setSearchSourceNotice] = useState<string | null>(null);
    const [isAddingSearchSource, setIsAddingSearchSource] = useState(false);

    const fetchOptionalSearchSources = useCallback(async () => {
        let touchedLoadingState = false;
        try {
            const res = await fetch("/api/v1/search/sources");
            if (!res || res.status === 404 || res.status === 405) return;
            setIsFetchingSearchSources(true);
            touchedLoadingState = true;
            setSearchSourcesError(null);
            setSearchSourceRegistrySupported(true);
            if (res.ok) {
                const payload = await res.json();
                setRegistrySearchSources(normalizeSearchSourcesPayload(extractApiData<unknown>(payload)));
                return;
            }
            setRegistrySearchSources([]);
            setSearchSourcesError(`Search source registry unreachable (HTTP ${res.status})`);
        } catch {
            setSearchSourceRegistrySupported(false);
            setRegistrySearchSources([]);
        } finally {
            if (touchedLoadingState) setIsFetchingSearchSources(false);
        }
    }, []);

    const addSearchSource = useCallback(async (input: SearchSourceDraft): Promise<boolean> => {
        setIsAddingSearchSource(true);
        setSearchSourcesError(null);
        setSearchSourceNotice(null);
        try {
            const res = await fetch("/api/v1/search/sources", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(input),
            });
            if (res.status === 404 || res.status === 405) {
                setSearchSourceRegistrySupported(false);
                return false;
            }
            if (res.ok) {
                setSearchSourceNotice(`Added ${input.name}.`);
                await fetchOptionalSearchSources();
                await fetchSearchCapability();
                return true;
            }
            setSearchSourcesError(await responseText(res) || `Search source was rejected (HTTP ${res.status})`);
        } catch (err) {
            const message = err instanceof Error ? err.message : "network error";
            setSearchSourcesError(`Search source could not be saved (${message})`);
        } finally {
            setIsAddingSearchSource(false);
        }
        return false;
    }, [fetchOptionalSearchSources, fetchSearchCapability]);

    const updateSearchSource = useCallback(async (sourceId: string, input: SearchSourceDraft): Promise<boolean> => {
        setIsAddingSearchSource(true);
        setSearchSourcesError(null);
        setSearchSourceNotice(null);
        try {
            const res = await fetch(`/api/v1/search/sources/${encodeURIComponent(sourceId)}`, {
                method: "PATCH",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(input),
            });
            if (res.ok) {
                setSearchSourceNotice(`Updated ${input.name}.`);
                await fetchOptionalSearchSources();
                await fetchSearchCapability();
                return true;
            }
            setSearchSourcesError(await responseText(res) || `Search source was not updated (HTTP ${res.status})`);
        } catch (err) {
            const message = err instanceof Error ? err.message : "network error";
            setSearchSourcesError(`Search source could not be updated (${message})`);
        } finally {
            setIsAddingSearchSource(false);
        }
        return false;
    }, [fetchOptionalSearchSources, fetchSearchCapability]);

    const deleteSearchSource = useCallback(async (sourceId: string, sourceName: string): Promise<boolean> => {
        setIsAddingSearchSource(true);
        setSearchSourcesError(null);
        setSearchSourceNotice(null);
        try {
            const res = await fetch(`/api/v1/search/sources/${encodeURIComponent(sourceId)}`, {
                method: "DELETE",
            });
            if (res.ok) {
                setSearchSourceNotice(`Removed ${sourceName}.`);
                await fetchOptionalSearchSources();
                await fetchSearchCapability();
                return true;
            }
            setSearchSourcesError(await responseText(res) || `Search source was not removed (HTTP ${res.status})`);
        } catch (err) {
            const message = err instanceof Error ? err.message : "network error";
            setSearchSourcesError(`Search source could not be removed (${message})`);
        } finally {
            setIsAddingSearchSource(false);
        }
        return false;
    }, [fetchOptionalSearchSources, fetchSearchCapability]);

    const visibleSearchSources = useMemo(
        () => mergeSearchSources(searchSources ?? [], registrySearchSources),
        [registrySearchSources, searchSources],
    );

    return {
        addSearchSource,
        deleteSearchSource,
        fetchOptionalSearchSources,
        isAddingSearchSource,
        isFetchingSearchSources,
        searchSourceNotice,
        searchSourceRegistrySupported,
        searchSourcesError,
        updateSearchSource,
        visibleSearchSources,
    };
}

function mergeSearchSources(primary: SearchCapabilitySource[], secondary: SearchCapabilitySource[]) {
    return Array.from(new Map([...primary, ...secondary].map((source) => [source.id, source])).values());
}

async function responseText(res: Response): Promise<string> {
    try {
        return await res.text();
    } catch {
        return "";
    }
}
