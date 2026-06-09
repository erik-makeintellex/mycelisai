"use client";

import { FolderOpen } from "lucide-react";
import type { ExecutionSummaryItem } from "@/store/useCortexStore";
import { workspacePathFromOutputUrl } from "./OutputAccessActions";

export function itemWorkspacePath(item: ExecutionSummaryItem) {
  const id = item.id?.trim();
  return item.folder
    ?? item.entrypoint
    ?? workspacePathFromOutputUrl(item.path ?? null)
    ?? workspacePathFromOutputUrl(item.href ?? null)
    ?? workspacePathFromOutputUrl(item.url ?? null)
    ?? workspacePathFromOutputUrl(item.open_url ?? null)
    ?? (id && isWorkspacePathLike(id) ? workspacePathFromOutputUrl(id) : null);
}

function isWorkspacePathLike(value: string) {
  const normalized = value.replace(/\\/g, "/");
  return normalized.startsWith("workspace/")
    || normalized.includes("/")
    || /\.[a-z0-9]{1,8}$/i.test(normalized);
}

export function outputWorkspacePath(output: { url: string | null; storagePath?: string }) {
  return output.storagePath?.trim() || workspacePathFromOutputUrl(output.url);
}

export function OutputPathHint({ storagePath, url }: { storagePath?: string; url: string | null }) {
  const path = outputWorkspacePath({ storagePath, url });
  if (!path) return null;
  return (
    <div className="mt-1 flex min-w-0 items-center gap-1 text-[11px] leading-5 text-cortex-text-muted">
      <FolderOpen className="h-3 w-3 shrink-0 text-cortex-primary" />
      <span className="shrink-0">Workspace path</span>
      <code className="min-w-0 truncate rounded border border-cortex-border/60 bg-cortex-bg/70 px-1 py-0.5 font-mono text-[10px] text-cortex-text-main">
        {path}
      </code>
    </div>
  );
}
