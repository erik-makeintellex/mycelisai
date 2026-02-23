/**
 * GET /docs-api
 *
 * Returns the curated documentation manifest (sections + entries).
 * Uses /docs-api prefix instead of /api/docs to avoid:
 *   1. The Go backend proxy rewrite in next.config.ts (/api/* → backend)
 *   2. Next.js private folder convention (_prefix = unreachable)
 */
import { NextResponse } from "next/server";
import { DOC_MANIFEST } from "@/lib/docsManifest";

export async function GET() {
    return NextResponse.json({ sections: DOC_MANIFEST });
}
