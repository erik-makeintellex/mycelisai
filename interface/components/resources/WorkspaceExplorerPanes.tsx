import { ChevronUp, File, Folder, FolderPlus, RefreshCw, Save } from "lucide-react";
import type { WorkspaceEntry, WorkspacePane } from "./WorkspaceExplorer";

const WORKSPACE_PANES: Array<{ id: WorkspacePane; label: string; summary: string }> = [
    { id: "browse", label: "Find outputs", summary: "Open folders and choose retained files." },
    { id: "preview", label: "Preview", summary: "Read the selected generated file." },
    { id: "create", label: "Create", summary: "Add small handoff files or folders." },
];

export function WorkspacePaneTabs({
    activePane,
    onSelect,
}: {
    activePane: WorkspacePane;
    onSelect: (pane: WorkspacePane) => void;
}) {
    return (
        <div
            className="grid grid-cols-1 gap-2 sm:grid-cols-3"
            role="tablist"
            aria-label="Workspace output panes"
        >
            {WORKSPACE_PANES.map((pane) => {
                const selected = activePane === pane.id;
                return (
                    <button
                        key={pane.id}
                        type="button"
                        role="tab"
                        aria-selected={selected}
                        aria-controls={`workspace-${pane.id}-panel`}
                        id={`workspace-${pane.id}-tab`}
                        onClick={() => onSelect(pane.id)}
                        className={`rounded border px-3 py-2 text-left transition-colors ${
                            selected
                                ? "border-cortex-primary/50 bg-cortex-primary/10 text-cortex-text-main"
                                : "border-cortex-border bg-cortex-surface/70 text-cortex-text-muted hover:bg-cortex-surface"
                        }`}
                    >
                        <span className="block text-xs font-semibold">{pane.label}</span>
                        <span className="mt-0.5 block text-[11px] leading-4">{pane.summary}</span>
                    </button>
                );
            })}
        </div>
    );
}

export function WorkspaceBrowsePane({
    currentPath,
    entries,
    busy,
    isFetchingMCPServers,
    onUp,
    onRefresh,
    onOpenFolder,
    onOpenFile,
}: {
    currentPath: string;
    entries: WorkspaceEntry[];
    busy: boolean;
    isFetchingMCPServers: boolean;
    onUp: () => void;
    onRefresh: () => void;
    onOpenFolder: (path: string) => void;
    onOpenFile: (path: string) => void;
}) {
    return (
        <div
            role="tabpanel"
            id="workspace-browse-panel"
            aria-labelledby="workspace-browse-tab"
            className="flex h-full min-h-0 flex-col"
        >
            <div className="mb-3 flex flex-wrap gap-2">
                <button
                    type="button"
                    onClick={onUp}
                    className="inline-flex items-center gap-1 rounded border border-cortex-border px-2 py-1 text-xs font-mono text-cortex-text-main hover:bg-cortex-border"
                >
                    <ChevronUp className="h-3 w-3" />
                    Up
                </button>
                <button
                    type="button"
                    onClick={onRefresh}
                    disabled={busy || isFetchingMCPServers}
                    className="inline-flex items-center gap-1 rounded border border-cortex-border px-2 py-1 text-xs font-mono text-cortex-text-main hover:bg-cortex-border disabled:opacity-60"
                >
                    <RefreshCw
                        className={`h-3 w-3 ${(busy || isFetchingMCPServers) ? "animate-spin" : ""}`}
                    />
                    Refresh
                </button>
            </div>

            <div className="min-h-0 flex-1 overflow-y-auto rounded border border-cortex-border/60 bg-cortex-bg">
                {entries.length === 0 ? (
                    <div className="p-3 text-xs font-mono text-cortex-text-muted">
                        No retained outputs in this folder
                    </div>
                ) : (
                    entries.map((entry, index) => (
                        <button
                            type="button"
                            key={`${entry.type}:${entry.path}:${index}`}
                            aria-label={
                                entry.type === "dir"
                                    ? `Open folder ${entry.name}`
                                    : `Preview output file ${entry.name}`
                            }
                            onClick={() =>
                                entry.type === "dir" ? onOpenFolder(entry.path) : onOpenFile(entry.path)
                            }
                            className="flex w-full items-center gap-2 border-b border-cortex-border/40 px-3 py-2 text-left transition-colors last:border-b-0 hover:bg-cortex-surface/70"
                        >
                            {entry.type === "dir" ? (
                                <Folder className="h-3.5 w-3.5 text-cortex-warning" />
                            ) : (
                                <File className="h-3.5 w-3.5 text-cortex-primary" />
                            )}
                            <span className="truncate text-xs font-mono text-cortex-text-main">
                                {entry.name}
                            </span>
                            {entry.type === "file" ? (
                                <span className="ml-auto text-[10px] uppercase tracking-normal text-cortex-text-muted">
                                    Preview
                                </span>
                            ) : null}
                        </button>
                    ))
                )}
            </div>
        </div>
    );
}

export function WorkspacePreviewPane({
    selectedFile,
    preview,
    onPreviewChange,
}: {
    selectedFile: string | null;
    preview: string;
    onPreviewChange: (value: string) => void;
}) {
    return (
        <div
            role="tabpanel"
            id="workspace-preview-panel"
            aria-labelledby="workspace-preview-tab"
            className="flex h-full min-h-0 flex-col"
        >
            <div className="mb-2 min-w-0">
                <p className="text-xs font-semibold text-cortex-text-main">Selected output</p>
                <code className="mt-1 block truncate text-[11px] font-mono text-cortex-primary">
                    {selectedFile ?? "(no output selected)"}
                </code>
            </div>
            <textarea
                value={preview}
                onChange={(event) => onPreviewChange(event.target.value)}
                placeholder="Select a file from Find outputs to preview contents"
                className="min-h-0 flex-1 resize-none rounded border border-cortex-border bg-cortex-bg p-3 text-xs font-mono text-cortex-text-main"
            />
        </div>
    );
}

export function WorkspaceCreatePane({
    newDir,
    newFile,
    newFileContent,
    onNewDirChange,
    onNewFileChange,
    onNewFileContentChange,
    onCreateDirectory,
    onCreateFile,
}: {
    newDir: string;
    newFile: string;
    newFileContent: string;
    onNewDirChange: (value: string) => void;
    onNewFileChange: (value: string) => void;
    onNewFileContentChange: (value: string) => void;
    onCreateDirectory: () => void;
    onCreateFile: () => void;
}) {
    return (
        <div
            role="tabpanel"
            id="workspace-create-panel"
            aria-labelledby="workspace-create-tab"
            className="grid gap-3 lg:grid-cols-2"
        >
            <div className="rounded border border-cortex-border bg-cortex-bg p-3">
                <label
                    className="mb-2 block text-xs font-semibold text-cortex-text-main"
                    htmlFor="workspace-new-directory"
                >
                    New folder
                </label>
                <div className="flex flex-col gap-2 sm:flex-row">
                    <input
                        id="workspace-new-directory"
                        value={newDir}
                        onChange={(event) => onNewDirChange(event.target.value)}
                        placeholder="new directory name"
                        className="min-w-0 flex-1 rounded border border-cortex-border bg-cortex-surface px-2 py-1 text-xs font-mono text-cortex-text-main"
                    />
                    <button
                        type="button"
                        onClick={onCreateDirectory}
                        className="inline-flex items-center justify-center gap-1 rounded border border-cortex-border px-2 py-1 text-xs font-mono text-cortex-text-main hover:bg-cortex-border"
                    >
                        <FolderPlus className="h-3 w-3" />
                        Create Dir
                    </button>
                </div>
            </div>

            <div className="rounded border border-cortex-border bg-cortex-bg p-3">
                <label
                    className="mb-2 block text-xs font-semibold text-cortex-text-main"
                    htmlFor="workspace-new-file"
                >
                    New handoff file
                </label>
                <div className="flex flex-col gap-2 sm:flex-row">
                    <input
                        id="workspace-new-file"
                        value={newFile}
                        onChange={(event) => onNewFileChange(event.target.value)}
                        placeholder="new file name"
                        className="min-w-0 flex-1 rounded border border-cortex-border bg-cortex-surface px-2 py-1 text-xs font-mono text-cortex-text-main"
                    />
                    <button
                        type="button"
                        onClick={onCreateFile}
                        className="inline-flex items-center justify-center gap-1 rounded border border-cortex-primary/30 px-2 py-1 text-xs font-mono text-cortex-primary hover:bg-cortex-primary/10"
                    >
                        <Save className="h-3 w-3" />
                        Write File
                    </button>
                </div>
                <textarea
                    value={newFileContent}
                    onChange={(event) => onNewFileContentChange(event.target.value)}
                    placeholder="Optional content for new file"
                    className="mt-3 h-40 w-full resize-y rounded border border-cortex-border bg-cortex-surface p-2 text-xs font-mono text-cortex-text-main"
                />
            </div>
        </div>
    );
}
