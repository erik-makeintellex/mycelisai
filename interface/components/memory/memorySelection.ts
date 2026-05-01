import type { Artifact } from "@/store/useCortexStore";

export interface SearchResult {
  id: string;
  content: string;
  similarity: number;
  source: string;
  created_at: string;
}

export type MemorySelection =
  | { kind: "search"; result: SearchResult }
  | { kind: "artifact"; artifact: Artifact };
