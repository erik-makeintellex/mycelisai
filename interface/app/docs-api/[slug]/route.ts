/**
 * GET /docs-api/[slug]
 *
 * Returns the raw markdown content of a documentation file identified by slug.
 * The slug is validated against DOC_BY_SLUG before any filesystem access —
 * no arbitrary path traversal is possible.
 *
 * Uses /docs-api prefix instead of /api/docs to avoid the Go backend
 * proxy rewrite in next.config.ts which catches all /api/* paths.
 * params is a Promise in Next.js 15+ — must be awaited.
 *
 * Response:
 *   200  { slug, label, content }   — markdown text
 *   404  { error }                  — slug not in manifest
 *   500  { error }                  — file unreadable
 */
import { NextRequest, NextResponse } from "next/server";
import { readFile } from "fs/promises";
import path from "path";
import { DOC_BY_SLUG } from "@/lib/docsManifest";

export async function GET(
    _req: NextRequest,
    { params }: { params: Promise<{ slug: string }> }
) {
    const { slug } = await params;
    const entry = DOC_BY_SLUG.get(slug);

    if (!entry) {
        return NextResponse.json({ error: `Unknown doc slug: ${slug}` }, { status: 404 });
    }

    // process.cwd() is the `interface/` directory when running `next dev`.
    // The docs live one level up in the scratch/ project root.
    const filePath = path.join(process.cwd(), "..", entry.path);

    try {
        const content = await readFile(filePath, "utf-8");
        return NextResponse.json({ slug, label: entry.label, content });
    } catch (err) {
        const message = err instanceof Error ? err.message : String(err);
        return NextResponse.json(
            { error: `Failed to read ${entry.path}: ${message}` },
            { status: 500 }
        );
    }
}
