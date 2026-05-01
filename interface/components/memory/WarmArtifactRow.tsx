import type { Artifact } from "@/store/useCortexStore";

export default function WarmArtifactRow({
  artifact,
  onSelect,
}: {
  artifact: Artifact;
  onSelect: (artifact: Artifact) => void;
}) {
  return (
    <button
      type="button"
      onClick={() => onSelect(artifact)}
      className="flex w-full items-center gap-2 border-b border-cortex-border/30 px-3 py-2 text-left transition-colors hover:bg-cortex-surface/40 focus:bg-cortex-primary/10 focus:outline-none"
    >
      <span
        className={`flex-shrink-0 rounded px-1.5 py-0.5 text-[9px] font-mono uppercase ${artifactTypeBadge(artifact.artifact_type)}`}
      >
        {artifact.artifact_type}
      </span>
      <span className="flex-1 truncate text-[11px] text-cortex-text-main">
        {artifact.title}
      </span>
      <span
        className={`flex-shrink-0 rounded px-1.5 py-0.5 text-[9px] font-mono uppercase ${statusBadge(artifact.status)}`}
      >
        {artifact.status}
      </span>
      <span className="max-w-[80px] flex-shrink-0 truncate text-[9px] font-mono text-cortex-text-muted">
        {artifact.agent_id}
      </span>
    </button>
  );
}

function artifactTypeBadge(type: string): string {
  switch (type) {
    case "code":
      return "bg-cortex-primary/20 text-cortex-primary";
    case "document":
      return "bg-cortex-info/20 text-cortex-info";
    case "image":
      return "bg-cortex-warning/20 text-cortex-warning";
    case "data":
    case "chart":
      return "bg-cortex-success/20 text-cortex-success";
    default:
      return "bg-cortex-text-muted/20 text-cortex-text-muted";
  }
}

function statusBadge(status: string): string {
  switch (status) {
    case "approved":
      return "bg-cortex-success/20 text-cortex-success";
    case "rejected":
      return "bg-cortex-danger/20 text-cortex-danger";
    case "pending":
      return "bg-cortex-warning/20 text-cortex-warning";
    case "archived":
      return "bg-cortex-text-muted/20 text-cortex-text-muted";
    default:
      return "bg-cortex-text-muted/20 text-cortex-text-muted";
  }
}
