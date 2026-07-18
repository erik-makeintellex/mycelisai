import type { OutcomeHealthState } from "@/lib/outcomeHealth";

export type OutcomeProjectSummary = {
  title: string;
  detail: string;
  health: OutcomeHealthState;
  ownerLabel?: string;
  leadLabel?: string;
  registryOwnerLabel?: string;
  teamCount: number;
  workCount: number;
  outputCount: number;
  recoveryCount: number;
  href: string;
  hrefLabel: string;
  targetReference?: string;
  outputHref?: string;
  outputLabel?: string;
};
