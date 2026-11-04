export type EntryType = "file" | "dir";

export interface WorkspaceEntry {
    name: string;
    path: string;
    type: EntryType;
}

export function normalizePath(path: string): string {
    const raw = (path || ".").replaceAll("\\", "/").trim();
    const parts = raw.split("/").filter(Boolean);
    const stack: string[] = [];
    for (const p of parts) {
        if (p === ".") continue;
        if (p === "..") {
            stack.pop();
            continue;
        }
        stack.push(p);
    }
    return stack.length === 0 ? "." : stack.join("/");
}

export function joinPath(base: string, child: string): string {
    return normalizePath(base === "." ? child : `${base}/${child}`);
}

export function parseListOutput(raw: string, currentPath: string): WorkspaceEntry[] {
    const text = raw.trim();
    if (!text) return [];

    try {
        const parsed = JSON.parse(text) as unknown;
        if (Array.isArray(parsed)) {
            return parsed
                .filter((v) => typeof v === "string")
                .map((name) => ({
                    name: name as string,
                    path: joinPath(currentPath, name as string),
                    type: (name as string).endsWith("/") ? "dir" : "file",
                }));
        }
        if (parsed && typeof parsed === "object" && Array.isArray((parsed as { entries?: unknown[] }).entries)) {
            return ((parsed as { entries: unknown[] }).entries ?? [])
                .filter((v) => v && typeof v === "object")
                .map((v) => {
                    const entry = v as Record<string, unknown>;
                    const name = String(entry.name ?? "");
                    const isDir = entry.type === "directory" || entry.type === "dir" || Boolean(entry.isDirectory);
                    return {
                        name,
                        path: joinPath(currentPath, name),
                        type: isDir ? "dir" : "file",
                    } satisfies WorkspaceEntry;
                })
                .filter((e) => e.name.length > 0);
        }
    } catch {
        // Non-JSON output, parse line format.
    }

    return text
        .split(/\r?\n/)
        .map((line) => line.trim())
        .filter(Boolean)
        .map((line) => {
            const dirTagged = line.startsWith("[DIR] ");
            const fileTagged = line.startsWith("[FILE] ");
            const cleaned = line
                .replace(/^\[DIR\]\s*/, "")
                .replace(/^\[FILE\]\s*/, "")
                .replace(/\s+\(directory\)$/i, "")
                .trim();
            const type: EntryType = dirTagged || cleaned.endsWith("/") ? "dir" : fileTagged ? "file" : "file";
            const name = cleaned.endsWith("/") ? cleaned.slice(0, -1) : cleaned;
            return { name, path: joinPath(currentPath, name), type } satisfies WorkspaceEntry;
        })
        .filter((entry) => entry.name.length > 0);
}
