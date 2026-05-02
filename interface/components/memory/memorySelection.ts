import type { Artifact } from "@/store/useCortexStore";

export interface SearchResult {
  id: string;
  content: string;
  score?: number;
  similarity?: number;
  source: string;
  created_at: string;
}

export function memoryResultScore(result: SearchResult): number {
  const score = result.score ?? result.similarity ?? 0;
  return Number.isFinite(score) ? score : 0;
}

export type MemorySelection =
  | { kind: "search"; result: SearchResult }
  | { kind: "artifact"; artifact: Artifact };
