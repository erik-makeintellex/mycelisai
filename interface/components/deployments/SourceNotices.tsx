import { AlertTriangle } from "lucide-react";
import { sourceNoticeText, type DeploymentSourceStatus } from "./deploymentTypes";

/** Small non-blocking notices for any source that is degraded, unavailable, or restricted. */
export function SourceNotices({ sources }: { sources: DeploymentSourceStatus[] }) {
  const notable = sources.filter((source) => source.status !== "ok");
  if (notable.length === 0) return null;
  return (
    <div className="flex flex-col gap-1.5" data-testid="deployments-source-notices">
      {notable.map((source) => (
        <div
          key={source.source}
          data-testid={`deployments-source-notice-${source.source}`}
          className="flex items-center gap-2 rounded-lg border border-cortex-warning/30 bg-cortex-warning/10 px-3 py-1.5 text-xs text-cortex-text-muted"
        >
          <AlertTriangle className="h-3.5 w-3.5 flex-shrink-0 text-cortex-warning" />
          <span>{sourceNoticeText(source)}</span>
        </div>
      ))}
    </div>
  );
}
