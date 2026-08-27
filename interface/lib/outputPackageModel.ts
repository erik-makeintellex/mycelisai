import { OUTPUT_PACKAGE_ACTION_LABELS } from "@/lib/deliveryRuntimeLanguage";

export interface OutputPackagePathInput {
  folder?: string | null;
  entrypoint?: string | null;
  filePath?: string | null;
}

export const OUTPUT_PACKAGE_OPEN_LABEL = OUTPUT_PACKAGE_ACTION_LABELS.openFile;
export const OUTPUT_PACKAGE_FOLDER_LABEL = OUTPUT_PACKAGE_ACTION_LABELS.openFolder;
export const OUTPUT_PACKAGE_RESOURCES_LABEL = OUTPUT_PACKAGE_ACTION_LABELS.openInResources;
export const OUTPUT_CANVAS_PATH = "/outputs/view";

export interface OutputCanvasHrefInput {
  label: string;
  url?: string | null;
  storagePath?: string | null;
  returnTo?: string | null;
  proofArtifactId?: string | null;
  teamId?: string | null;
  runId?: string | null;
  workItemId?: string | null;
  outputId?: string | null;
  contentDigest?: string | null;
}

export function normalizeWorkspacePath(path?: string | null) {
  const normalized = path?.trim().replace(/\\/g, "/").replace(/^\/+/, "");
  return normalized || null;
}

export function parentWorkspacePath(path?: string | null) {
  const normalized = normalizeWorkspacePath(path);
  if (!normalized?.includes("/")) return null;
  return normalized.slice(0, normalized.lastIndexOf("/"));
}

export function workspaceBrowserPath(path?: string | null) {
  const normalized = normalizeWorkspacePath(path);
  if (!normalized) return null;
  if (normalized.startsWith("workspace/") || normalized === "workspace") return normalized;
  return /^(groups|generated|outputs|reports|logs|saved-media)(\/|$)/i.test(normalized)
    ? `workspace/${normalized}`
    : normalized;
}

export function joinWorkspacePath(folder?: string | null, entrypoint?: string | null) {
  const normalizedEntry = normalizeWorkspacePath(entrypoint);
  if (!normalizedEntry) return null;
  const normalizedFolder = normalizeWorkspacePath(folder);
  if (isWorkspaceRootedPath(normalizedEntry) || !normalizedFolder) return normalizedEntry;
  return normalizedFolder ? `${normalizedFolder}/${normalizedEntry}` : normalizedEntry;
}

function isWorkspaceRootedPath(path: string) {
  return /^(workspace|groups|generated|outputs|reports|logs|saved-media)\//i.test(path);
}

export function projectPackageOpenPath(input: OutputPackagePathInput) {
  return joinWorkspacePath(input.folder, input.entrypoint)
    ?? normalizeWorkspacePath(input.filePath);
}

export function projectPackageRevealPath(input: OutputPackagePathInput) {
  return normalizeWorkspacePath(input.folder)
    ?? parentWorkspacePath(projectPackageOpenPath(input))
    ?? normalizeWorkspacePath(input.filePath);
}

export function workspaceFileHref(path?: string | null) {
  const normalized = normalizeWorkspacePath(path);
  return normalized ? `/api/v1/workspace/files/view?path=${encodeURIComponent(normalized)}` : null;
}

export function resourcesWorkspaceHref(path?: string | null) {
  const normalized = workspaceBrowserPath(path);
  return normalized ? `/resources?tab=workspace&path=${encodeURIComponent(normalized)}` : null;
}

export function projectPackageResourcesHref(input: OutputPackagePathInput) {
  return resourcesWorkspaceHref(projectPackageRevealPath(input));
}

export function outputCanvasSourceHref(url?: string | null, storagePath?: string | null) {
  const source = url?.trim() || workspaceFileHref(storagePath);
  if (!source) return null;
  try {
    const parsed = new URL(source, "http://mycelis.local");
    if (parsed.origin !== "http://mycelis.local" || parsed.pathname !== "/api/v1/workspace/files/view") return null;
    return `${parsed.pathname}${parsed.search}${parsed.hash}`;
  } catch {
    return null;
  }
}

export function somaReturnHref(returnTo?: string | null) {
  if (!returnTo) return "/dashboard";
  try {
    const parsed = new URL(returnTo, "http://mycelis.local");
    if (parsed.origin !== "http://mycelis.local" || parsed.pathname !== "/dashboard") return "/dashboard";
    return `${parsed.pathname}${parsed.search}${parsed.hash}`;
  } catch {
    return "/dashboard";
  }
}

export function outputCanvasHref(input: OutputCanvasHrefInput) {
  const { label, url, storagePath, returnTo, proofArtifactId } = input;
  const source = outputCanvasSourceHref(url, storagePath);
  if (!source) return null;
  const params = new URLSearchParams({ source, label: label.trim() || "Retained output" });
  const normalizedPath = normalizeWorkspacePath(storagePath);
  if (normalizedPath) params.set("path", normalizedPath);
  if (proofArtifactId?.trim()) params.set("proof", proofArtifactId.trim());
  const trustedParams = {
    team_id: input.teamId,
    run_id: input.runId,
    work_item_id: input.workItemId,
    output_id: input.outputId,
    digest: input.contentDigest,
  };
  for (const [key, value] of Object.entries(trustedParams)) {
    if (value?.trim()) params.set(key, value.trim());
  }
  params.set("return_to", somaReturnHref(returnTo));
  return `${OUTPUT_CANVAS_PATH}?${params.toString()}`;
}
