import Link from "next/link";
import type { CatalogueAgent } from "@/store/useCortexStore";
import TeamsIntroPanel from "./TeamsIntroPanel";
import { TeamMemberTemplatesPanel } from "./TeamMemberTemplatesPanel";

export function TeamsSetupPanels({
  highlightedTemplates,
  templateCoverage,
  isFetchingCatalogue,
  onNewTemplate,
  onEditTemplate,
}: {
  highlightedTemplates: CatalogueAgent[];
  templateCoverage: Array<[string, CatalogueAgent[]]>;
  isFetchingCatalogue: boolean;
  onNewTemplate: () => void;
  onEditTemplate: (agent: CatalogueAgent) => void;
}) {
  return (
    <div className="grid gap-4 xl:grid-cols-[1.15fr_0.85fr]">
      <TeamsIntroPanel />
      <TeamMemberTemplatesPanel
        highlightedTemplates={highlightedTemplates}
        templateCoverage={templateCoverage}
        isFetchingCatalogue={isFetchingCatalogue}
        onNewTemplate={onNewTemplate}
        onEditTemplate={onEditTemplate}
      />
    </div>
  );
}

export function TeamsOutputCollaboration() {
  return (
    <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-4">
      <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div>
          <p className="text-sm font-semibold text-cortex-text-main">
            Outputs and active collaboration
          </p>
          <p className="mt-1 text-sm leading-6 text-cortex-text-muted">
            Keep active team work here. Use Groups when a team has retained
            outputs, archived collaboration records, or multi-team coordination
            to review.
          </p>
        </div>
        <Link
          href="/groups"
          className="inline-flex items-center justify-center rounded-2xl border border-cortex-primary/30 px-4 py-2 text-sm font-semibold text-cortex-primary hover:bg-cortex-primary/10"
        >
          Review group outputs
        </Link>
      </div>
    </div>
  );
}

export function TeamsWorkReviewIntro() {
  return (
    <div className="rounded-2xl border border-cortex-primary/30 bg-cortex-primary/10 px-4 py-4">
      <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div>
          <p className="text-sm font-semibold text-cortex-text-main">
            Decide what happens to this work
          </p>
          <p className="mt-1 text-sm leading-6 text-cortex-text-muted">
            Start with the review item. It shows why attention is needed, what
            remains trusted, and the safest next move. Clear stale test data from review;
            recover only when you want Soma to retry from retained context.
          </p>
        </div>
        <Link
          href="/teams"
          className="inline-flex items-center justify-center rounded-2xl border border-cortex-border px-4 py-2 text-sm font-semibold text-cortex-text-main hover:border-cortex-primary/40"
        >
          Open all teams
        </Link>
      </div>
    </div>
  );
}

export function TeamsReviewContextNote() {
  return (
    <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-4">
      <p className="text-sm font-semibold text-cortex-text-main">
        Team context
      </p>
      <p className="mt-1 text-sm leading-6 text-cortex-text-muted">
        Use these cards after the review item tells you which team, run, or
        output needs attention.
      </p>
    </div>
  );
}
