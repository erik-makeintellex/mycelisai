"use client";

import { useCallback, useEffect, useState } from "react";
import {
  compositeDeploymentId,
  getData,
  type DeploymentKindFilter,
  type DeploymentList,
  type DeploymentRecord,
  type DeploymentSourceStatus,
  type DeploymentStateFilter,
} from "./deploymentTypes";

/**
 * Fetches and filters the Deployments aggregate (GET /api/v1/deployments).
 * Filters are applied server-side via query params so the list, counts, and
 * source notices always reflect what Core actually returned.
 */
export function useDeployments(organizationId?: string | null) {
  const [deployments, setDeployments] = useState<DeploymentRecord[]>([]);
  const [sources, setSources] = useState<DeploymentSourceStatus[]>([]);
  const [kindFilter, setKindFilter] = useState<DeploymentKindFilter>("all");
  const [stateFilter, setStateFilter] = useState<DeploymentStateFilter>("all");
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [refreshing, setRefreshing] = useState(false);

  const load = useCallback(async () => {
    setRefreshing(true);
    try {
      const params = new URLSearchParams();
      if (kindFilter !== "all") params.set("kind", kindFilter);
      if (stateFilter !== "all") params.set("state", stateFilter);
      if (organizationId) params.set("organization_id", organizationId);
      const qs = params.toString();
      const res = await fetch(`/api/v1/deployments${qs ? `?${qs}` : ""}`, {
        cache: "no-store",
      });
      if (!res.ok) {
        throw new Error(`Deployments request failed (HTTP ${res.status}).`);
      }
      const data = await getData<DeploymentList>(res);
      const nextDeployments = Array.isArray(data?.deployments) ? data.deployments : [];
      setDeployments(nextDeployments);
      setSources(Array.isArray(data?.sources) ? data.sources : []);
      setSelectedId((current) => {
        if (current && nextDeployments.some((d) => compositeDeploymentId(d) === current)) {
          return current;
        }
        const requestedTeam =
          typeof window !== "undefined"
            ? new URLSearchParams(window.location.search).get("team")?.trim()
            : "";
        if (requestedTeam) {
          const matched = nextDeployments.find(
            (d) => d.id === requestedTeam || compositeDeploymentId(d) === requestedTeam,
          );
          if (matched) return compositeDeploymentId(matched);
        }
        return nextDeployments[0] ? compositeDeploymentId(nextDeployments[0]) : null;
      });
      setLoadError(null);
    } catch (failure) {
      setLoadError(
        failure instanceof Error ? failure.message : "Could not load deployments.",
      );
    } finally {
      setRefreshing(false);
    }
  }, [kindFilter, stateFilter, organizationId]);

  useEffect(() => {
    void load();
  }, [load]);

  const selected = deployments.find((d) => compositeDeploymentId(d) === selectedId) ?? null;

  return {
    deployments,
    sources,
    loadError,
    refreshing,
    kindFilter,
    stateFilter,
    selectedId,
    selected,
    setKindFilter,
    setStateFilter,
    setSelectedId,
    reload: () => void load(),
  };
}
