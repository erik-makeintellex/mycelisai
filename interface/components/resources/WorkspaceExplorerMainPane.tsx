"use client";

import React from "react";
import {
    WorkspaceBrowsePane,
    WorkspaceCreatePane,
    WorkspacePaneTabs,
    WorkspacePreviewPane,
} from "./WorkspaceExplorerPanes";
import { WORKSPACE_ROOT_PATH } from "./WorkspaceExplorerSupport";
import { normalizePath } from "./WorkspaceExplorerUtils";
import type { WorkspaceEntry, WorkspacePane } from "./WorkspaceExplorer";

export default function WorkspaceExplorerMainPane({
    activePane,
    busy,
    currentPath,
    entries,
    isFetchingMCPServers,
    newDir,
    newFile,
    newFileContent,
    preview,
    selectedFile,
    status,
    onActivePaneChange,
    onCreateDirectory,
    onCreateFile,
    onCurrentPathChange,
    onNewDirChange,
    onNewFileChange,
    onNewFileContentChange,
    onOpenFile,
    onPreviewChange,
    onRefresh,
}: {
    activePane: WorkspacePane;
    busy: boolean;
    currentPath: string;
    entries: WorkspaceEntry[];
    isFetchingMCPServers: boolean;
    newDir: string;
    newFile: string;
    newFileContent: string;
    preview: string;
    selectedFile: string | null;
    status: string;
    onActivePaneChange: (pane: WorkspacePane) => void;
    onCreateDirectory: () => void;
    onCreateFile: () => void;
    onCurrentPathChange: (path: string) => void;
    onNewDirChange: (value: string) => void;
    onNewFileChange: (value: string) => void;
    onNewFileContentChange: (value: string) => void;
    onOpenFile: (path: string) => void;
    onPreviewChange: (value: string) => void;
    onRefresh: () => void;
}) {
    return (
        <section className="flex min-h-0 min-w-0 flex-col overflow-hidden rounded-lg border border-cortex-border bg-cortex-surface">
            <div className="border-b border-cortex-border bg-cortex-bg/70 px-3 py-3">
                <div className="mb-3 flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
                    <div className="min-w-0">
                        <p className="text-xs font-semibold text-cortex-text-main">Workspace output access</p>
                        <p className="mt-1 text-[11px] leading-5 text-cortex-text-muted">
                            Start with retained files, then open a preview or create a small handoff artifact when needed.
                        </p>
                    </div>
                    <div className="flex min-w-0 items-center gap-2">
                        <span className="text-[11px] font-mono text-cortex-text-muted">Path</span>
                        <code className="max-w-full truncate rounded border border-cortex-border bg-cortex-surface px-2 py-1 text-xs font-mono text-cortex-primary md:max-w-[22rem]">
                            {currentPath}
                        </code>
                    </div>
                </div>

                <WorkspacePaneTabs activePane={activePane} onSelect={onActivePaneChange} />
            </div>

            <div className="min-h-0 flex-1 overflow-y-auto p-3">
                {activePane === "browse" && (
                    <WorkspaceBrowsePane
                        currentPath={currentPath}
                        entries={entries}
                        busy={busy}
                        isFetchingMCPServers={isFetchingMCPServers}
                        onUp={() =>
                            onCurrentPathChange(
                                currentPath === WORKSPACE_ROOT_PATH
                                    ? WORKSPACE_ROOT_PATH
                                    : normalizePath(`${currentPath}/..`),
                            )
                        }
                        onRefresh={onRefresh}
                        onOpenFolder={onCurrentPathChange}
                        onOpenFile={onOpenFile}
                    />
                )}

                {activePane === "preview" && (
                    <WorkspacePreviewPane
                        selectedFile={selectedFile}
                        preview={preview}
                        onPreviewChange={onPreviewChange}
                    />
                )}

                {activePane === "create" && (
                    <WorkspaceCreatePane
                        newDir={newDir}
                        newFile={newFile}
                        newFileContent={newFileContent}
                        onNewDirChange={onNewDirChange}
                        onNewFileChange={onNewFileChange}
                        onNewFileContentChange={onNewFileContentChange}
                        onCreateDirectory={onCreateDirectory}
                        onCreateFile={onCreateFile}
                    />
                )}
            </div>

            <div className="border-t border-cortex-border bg-cortex-bg px-3 py-2 text-[11px] font-mono text-cortex-text-muted">
                <span className="block truncate">{status}</span>
            </div>
        </section>
    );
}
