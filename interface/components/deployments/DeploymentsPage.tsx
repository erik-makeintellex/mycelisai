"use client";

import Link from "next/link";
import LoadFailureState from "@/components/shared/LoadFailureState";
import { useCortexStore } from "@/store/useCortexStore";
import { DeploymentDetailPane } from "./DeploymentDetailPane";
import { DeploymentFilters } from "./DeploymentFilters";
import { DeploymentList } from "./DeploymentList";
import { SourceNotices } from "./SourceNotices";
import { useDeployments } from "./useDeployments";

/**
 * Read-only "what is running" surface (RC-004E). Lists standing teams, mission
 * teams, groups, and planned departments in one place; selecting a row opens a
 * detail pane. Advanced mode adds origin, scope, provider origin, and "View config".
 */
export default function DeploymentsPage() {
  const advancedMode = useCortexStore((s) => s.advancedMode);
  const {
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
    reload,
  } = useDeployments();

  if (loadError && deployments.length === 0) {
    return (
      <div className="p-4">
        <LoadFailureState
          title="Could not load deployments"
          detail={loadError}
          onRetry={reload}
          retrying={refreshing}
          testId="deployments-load-failure"
        />
      </div>
    );
  }

  return (
    <div className="flex min-h-0 flex-col gap-3 p-4">
      <header className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-xl font-semibold text-cortex-text-main">Deployments</h1>
          <p className="mt-1 text-sm text-cortex-text-muted">
            What is running, where, with which models and tools, and what it has produced.
          </p>
        </div>
        <Link
          href="/dashboard?team_action=start_team"
          data-testid="deployments-start-new-team"
          className="inline-flex items-center justify-center rounded-xl border border-cortex-primary/50 bg-cortex-primary/15 px-3.5 py-2 text-sm font-semibold text-cortex-primary hover:bg-cortex-primary/25"
        >
          + Start a New Team in Soma
        </Link>
      </header>
      <SourceNotices sources={sources} />
      <div className="grid min-h-0 flex-1 gap-3 lg:grid-cols-[minmax(280px,340px)_minmax(0,1fr)]">
        <aside className="flex min-h-0 flex-col gap-3 rounded-2xl border border-cortex-border bg-cortex-surface p-3">
          <DeploymentFilters
            kindFilter={kindFilter}
            stateFilter={stateFilter}
            onKindChange={setKindFilter}
            onStateChange={setStateFilter}
          />
          <DeploymentList
            deployments={deployments}
            selectedId={selectedId}
            onSelect={setSelectedId}
          />
        </aside>
        <DeploymentDetailPane record={selected} advancedMode={advancedMode} />
      </div>
    </div>
  );
}
