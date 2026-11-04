import type { Artifact } from "@/store/useCortexStore";

export interface SearchResult {
  id: string;
  content: string;
  score?: number;
  similarity?: number;
  adjusted_score?: number;
  source: string;
  created_at: string;
  memory_layer?: string;
  verification_status?: string;
  evidence_anchors?: Array<{
    source_kind?: string;
    ref_id?: string;
    label?: string;
    verified?: boolean;
  }>;
  contradiction_halt?: {
    halted: boolean;
    reason: string;
    conflicting_id?: string;
    conflicting_layer?: string;
  };
}

export interface ArbitrationSummary {
  total_candidates?: number;
  returned_count?: number;
  superseded_count?: number;
  contradicted_count?: number;
  verified_count?: number;
  candidate_count?: number;
  dominant_layer?: string;
  has_contradictions?: boolean;
}

const LAYER_LABELS: Record<string, string> = {
  GOVERNED_DOCTRINE: "Governed Doctrine",
  PROJECT_MEMORY: "Project",
  AGENT_MEMORY: "Team / Agent",
  SOMA_MEMORY: "Soma",
  REFLECTION_MEMORY: "Reflection",
};

const VERIFICATION_LABELS: Record<string, string> = {
  verified: "Verified",
  candidate: "Candidate",
  contradicted: "Contradicted",
  superseded: "Superseded",
};

export function formatMemoryLayerLabel(layer?: string): string {
  if (!layer) return "";
  return LAYER_LABELS[layer] ?? layer;
}

export function formatVerificationStatusLabel(status?: string): string {
  if (!status) return "";
  return VERIFICATION_LABELS[status.toLowerCase()] ?? status;
}

export function memoryResultScore(result: SearchResult): number {
  const score = result.score ?? result.similarity ?? 0;
  return Number.isFinite(score) ? score : 0;
}

export type MemorySelection =
  | { kind: "search"; result: SearchResult }
  | { kind: "artifact"; artifact: Artifact };
