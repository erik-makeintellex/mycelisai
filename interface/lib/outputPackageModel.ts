import { OUTPUT_PACKAGE_ACTION_LABELS } from "@/lib/deliveryRuntimeLanguage";

export interface OutputPackagePathInput {
  folder?: string | null;
  entrypoint?: string | null;
  filePath?: string | null;
}

export const OUTPUT_PACKAGE_OPEN_LABEL = OUTPUT_PACKAGE_ACTION_LABELS.openFile;
export const OUTPUT_PACKAGE_FOLDER_LABEL = OUTPUT_PACKAGE_ACTION_LABELS.openFolder;
export const OUTPUT_PACKAGE_RESOURCES_LABEL = OUTPUT_PACKAGE_ACTION_LABELS.openInResources;

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
