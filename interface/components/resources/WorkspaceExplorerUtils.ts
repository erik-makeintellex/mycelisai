import type { WorkspaceEntry } from "./WorkspaceExplorer";

export function normalizePath(path: string): string {
    const raw = (path || ".").replaceAll("\\", "/").trim();
    const parts = raw.split("/").filter(Boolean);
    const stack: string[] = [];
    for (const part of parts) {
        if (part === ".") continue;
        if (part === "..") {
            stack.pop();
            continue;
        }
        stack.push(part);
    }
    return stack.length === 0 ? "." : stack.join("/");
}

export function joinPath(base: string, child: string): string {
    return normalizePath(base === "." ? child : `${base}/${child}`);
}

export function parseListOutput(raw: string, currentPath: string): WorkspaceEntry[] {
    const text = raw.trim();
    if (!text) return [];

    const parsedJson = parseJsonList(text, currentPath);
    if (parsedJson) return parsedJson;

    return text
        .split(/\r?\n/)
        .map((line) => line.trim())
        .filter(Boolean)
        .map((line) => parseLineEntry(line, currentPath))
        .filter((entry) => entry.name.length > 0);
}

function parseJsonList(text: string, currentPath: string): WorkspaceEntry[] | null {
    try {
        const parsed = JSON.parse(text) as unknown;
        if (Array.isArray(parsed)) {
            return parsed
                .filter((value) => typeof value === "string")
                .map((name) => ({
                    name,
                    path: joinPath(currentPath, name),
                    type: name.endsWith("/") ? "dir" : "file",
                }));
        }
        if (!isEntriesObject(parsed)) return null;
        return parsed.entries
            .filter((value) => value && typeof value === "object")
            .map((value) => {
                const entry = value as Record<string, unknown>;
                const name = String(entry.name ?? "");
                const isDir =
                    entry.type === "directory" ||
                    entry.type === "dir" ||
                    Boolean(entry.isDirectory);
                return {
                    name,
                    path: joinPath(currentPath, name),
                    type: isDir ? "dir" : "file",
                } satisfies WorkspaceEntry;
            })
            .filter((entry) => entry.name.length > 0);
    } catch {
        return null;
    }
}

function isEntriesObject(value: unknown): value is { entries: unknown[] } {
    return Boolean(
        value &&
            typeof value === "object" &&
            Array.isArray((value as { entries?: unknown[] }).entries),
    );
}

function parseLineEntry(line: string, currentPath: string): WorkspaceEntry {
    const dirTagged = line.startsWith("[DIR] ");
    const fileTagged = line.startsWith("[FILE] ");
    const cleaned = line
        .replace(/^\[DIR\]\s*/, "")
        .replace(/^\[FILE\]\s*/, "")
        .replace(/\s+\(directory\)$/i, "")
        .trim();
    const type: WorkspaceEntry["type"] =
        dirTagged || cleaned.endsWith("/") ? "dir" : fileTagged ? "file" : "file";
    const name = cleaned.endsWith("/") ? cleaned.slice(0, -1) : cleaned;
    return { name, path: joinPath(currentPath, name), type };
}
