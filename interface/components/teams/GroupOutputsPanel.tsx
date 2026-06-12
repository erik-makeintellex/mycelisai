import { Download, ShieldCheck } from "lucide-react";
import OutputAccessActions from "@/components/soma/OutputAccessActions";
import type { Artifact } from "@/store/cortexStoreTypesPlanning";
import { relativeTime, type OutputSummary } from "./groupWorkspaceTypes";
import { Badge } from "./GroupDetailPane";

export function OutputsPanel({
  archived,
  outputs,
  outputSummary,
}: {
  archived?: boolean;
  outputs: Artifact[];
  outputSummary: OutputSummary;
}) {
  return (
    <section className="rounded-xl border border-cortex-border bg-cortex-bg p-4">
      <h3 className="text-sm font-semibold text-cortex-text-main">
        {archived ? "Retained outputs" : "Recent outputs"}
      </h3>
      <div
        className="mt-2 flex flex-wrap gap-2"
        data-testid="groups-output-summary"
      >
        <Badge muted>
          {outputSummary.artifactCount} output
          {outputSummary.artifactCount === 1 ? "" : "s"}
        </Badge>
        <Badge muted>
          {outputSummary.agentCount} contributing lead
          {outputSummary.agentCount === 1 ? "" : "s"}
        </Badge>
      </div>
      {archived ? (
        <p
          className="mt-2 text-sm leading-6 text-cortex-text-muted"
          data-testid="groups-retained-outputs-note"
        >
          Review the outputs this archived temporary group already produced.
          Downloads remain available so the work can still be inspected after
          the collaboration window closes.
        </p>
      ) : null}
      <div className="mt-3 max-h-[calc(100vh-24rem)] min-h-[16rem] space-y-3 overflow-y-auto pr-1">
        {outputs.length === 0 ? (
          <p className="text-sm text-cortex-text-muted">
            No recent team outputs found for this group yet.
          </p>
        ) : (
          outputs
            .slice(0, 8)
            .map((artifact) => (
              <ArtifactRow key={artifact.id} artifact={artifact} />
            ))
        )}
      </div>
    </section>
  );
}

function ArtifactRow({ artifact }: { artifact: Artifact }) {
  const projectPackage = artifact.artifact_type === "project_package";
  const entrypoint = stringMetadata(artifact.metadata, "entrypoint");
  const folder = stringMetadata(artifact.metadata, "folder");
  const files = stringArrayMetadata(artifact.metadata, "files");
  const validation = stringMetadata(artifact.metadata, "validation");
  const outputPath = projectPackage
    ? entrypoint || artifact.file_path || ""
    : outputPathForArtifact(artifact);
  const packageHref = projectPackage
    ? workspaceFileHref(outputPath)
    : null;
  const outputHref = !projectPackage ? workspaceFileHref(outputPath) : null;
  const htmlOutput = isHtmlArtifact(artifact);
  const readable =
    artifact.artifact_type === "code" ||
    artifact.artifact_type === "document" ||
    artifact.artifact_type === "data" ||
    artifact.artifact_type === "chart";
  return (
    <div className="rounded-xl border border-cortex-border bg-cortex-surface px-3 py-3">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="truncate text-sm font-semibold text-cortex-text-main">
            {artifact.title}
          </p>
          <p className="mt-1 text-[11px] font-mono uppercase tracking-[0.12em] text-cortex-text-muted">
            {artifact.artifact_type} | {artifact.agent_id} |{" "}
            {relativeTime(artifact.created_at)}
          </p>
        </div>
        <div className="flex shrink-0 flex-wrap justify-end gap-2">
          {!projectPackage && (outputHref || outputPath) ? (
            <OutputAccessActions
              label={artifact.title}
              url={outputHref}
              storagePath={outputPath}
              openLabel={htmlOutput ? "Open output" : "Open file"}
            />
          ) : null}
          <a
            href={`/api/v1/artifacts/${encodeURIComponent(artifact.id)}/download`}
            download
            className="inline-flex items-center gap-1 rounded-xl border border-cortex-primary/30 px-3 py-2 text-xs font-semibold text-cortex-primary hover:bg-cortex-primary/10"
          >
            <Download className="h-3.5 w-3.5" />
            Download
          </a>
        </div>
      </div>
      {projectPackage ? (
        <ProjectPackage
          artifact={artifact}
          entrypoint={entrypoint}
          folder={folder}
          files={files}
          validation={validation}
          packageHref={packageHref}
        />
      ) : htmlOutput && outputHref ? (
        <div className="mt-3 rounded-xl border border-cortex-border bg-cortex-bg p-3 text-sm leading-6 text-cortex-text-muted">
          HTML output is saved as a browser-viewable file. Use{" "}
          <span className="font-semibold text-cortex-text-main">
            Open output
          </span>{" "}
          to inspect the rendered result.
        </div>
      ) : readable && artifact.content ? (
        <pre className="mt-3 max-h-48 overflow-auto rounded-xl border border-cortex-border bg-cortex-bg p-3 text-xs leading-6 text-cortex-text-muted">
          {artifact.content}
        </pre>
      ) : artifact.file_path ? (
        <p className="mt-3 text-sm text-cortex-text-muted">
          Saved path: {artifact.file_path}
        </p>
      ) : null}
    </div>
  );
}

function ProjectPackage({
  artifact,
  entrypoint,
  folder,
  files,
  validation,
  packageHref,
}: {
  artifact: Artifact;
  entrypoint: string;
  folder: string;
  files: string[];
  validation: string;
  packageHref: string | null;
}) {
  return (
    <div className="mt-3 space-y-3 rounded-lg border border-cortex-border bg-cortex-bg p-3">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <p className="text-xs font-semibold uppercase tracking-[0.12em] text-cortex-text-muted">
          Project package
        </p>
        <OutputAccessActions
          label={artifact.title}
          url={packageHref}
          storagePath={folder || entrypoint || artifact.file_path}
          openLabel="Open Game"
        />
      </div>
      {entrypoint ? <MetaLine label="Entrypoint" value={entrypoint} /> : null}
      {folder ? <MetaLine label="Storage" value={folder} /> : null}
      {files.length > 0 ? (
        <div className="flex flex-wrap gap-1.5">
          {files.slice(0, 8).map((file) => (
            <span
              key={file}
              className="rounded border border-cortex-border bg-cortex-surface px-2 py-1 font-mono text-[10px] text-cortex-text-muted"
            >
              {file}
            </span>
          ))}
        </div>
      ) : null}
      {validation ? (
        <p className="flex gap-2 text-xs leading-5 text-cortex-text-muted">
          <ShieldCheck className="mt-0.5 h-3.5 w-3.5 shrink-0 text-cortex-success" />
          <span>{validation}</span>
        </p>
      ) : null}
    </div>
  );
}

function MetaLine({ label, value }: { label: string; value: string }) {
  return (
    <p className="break-all text-xs leading-5 text-cortex-text-muted">
      {label}: <span className="font-mono text-cortex-text-main">{value}</span>
    </p>
  );
}

function stringMetadata(metadata: Record<string, any> | undefined, key: string) {
  const value = metadata?.[key];
  return typeof value === "string" ? value.trim() : "";
}

function stringArrayMetadata(
  metadata: Record<string, any> | undefined,
  key: string,
) {
  const value = metadata?.[key];
  if (!Array.isArray(value)) return [];
  return value
    .map((item) => (typeof item === "string" ? item.trim() : ""))
    .filter(Boolean);
}

function workspaceFileHref(path: string) {
  const trimmed = path.trim();
  if (!trimmed) return null;
  return `/api/v1/workspace/files/view?path=${encodeURIComponent(trimmed)}`;
}

function outputPathForArtifact(artifact: Artifact) {
  const metadataPath =
    stringMetadata(artifact.metadata, "entrypoint") ||
    stringMetadata(artifact.metadata, "path") ||
    stringMetadata(artifact.metadata, "file_path");
  if (metadataPath) return metadataPath;
  if (artifact.file_path?.trim()) return artifact.file_path.trim();
  return workspaceLikePath(artifact.title);
}

function workspaceLikePath(value: string) {
  const trimmed = value.trim();
  if (!trimmed) return "";
  if (/^(workspace|groups|generated|outputs|reports|logs)\//i.test(trimmed)) {
    return trimmed;
  }
  return "";
}

function isHtmlArtifact(artifact: Artifact) {
  const contentType = artifact.content_type?.toLowerCase() ?? "";
  const title = artifact.title.toLowerCase();
  const path = outputPathForArtifact(artifact).toLowerCase();
  const content = artifact.content?.trimStart().toLowerCase() ?? "";
  return (
    contentType.includes("html") ||
    title.endsWith(".html") ||
    path.endsWith(".html") ||
    content.startsWith("<!doctype html") ||
    content.startsWith("<html")
  );
}
