import { existsSync } from "fs";
import path from "path";

function candidateDocRoots(startDirs?: string[]): string[] {
    const seen = new Set<string>();
    const roots: string[] = [];
    const explicitRoot = process.env.MYCELIS_PROJECT_ROOT?.trim();
    const moduleDir = typeof __dirname === "string" ? __dirname : process.cwd();
    for (const rawStart of startDirs ?? [explicitRoot, process.cwd(), moduleDir]) {
        if (!rawStart) {
            continue;
        }
        let current = path.resolve(rawStart);
        while (!seen.has(current)) {
            seen.add(current);
            roots.push(current);
            const parent = path.dirname(current);
            if (parent === current) {
                break;
            }
            current = parent;
        }
    }
    return roots;
}

export function resolveDocFilePath(relativePath: string, startDirs?: string[]): string {
    const roots = candidateDocRoots(startDirs);
    for (const root of roots) {
        const candidate = path.join(root, relativePath);
        if (existsSync(candidate)) {
            return candidate;
        }
    }
    throw new Error(`Unable to resolve doc path ${relativePath} from runtime roots: ${roots.join(", ")}`);
}
