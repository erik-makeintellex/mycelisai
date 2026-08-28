import type { SomaThreadEvent } from "@/store/useCortexStore";

export type PackageRecoveryPresentation = {
  detail: string;
  trust: string;
  actionLabel: string;
};

function eventText(event: SomaThreadEvent) {
  return [event.status, event.label, event.title, event.detail, event.target_reference]
    .filter((value): value is string => typeof value === "string")
    .join(" ")
    .toLowerCase();
}

export function packageRecoveryPresentation(event: SomaThreadEvent): PackageRecoveryPresentation | null {
  const text = eventText(event);
  const isPackageFailure = [
    "runtime_validation_",
    "result_contract_unsatisfied",
    "deliverable needs runtime repair",
    "output is not playable",
  ].some((token) => text.includes(token));
  if (!isPackageFailure) return null;

  let detail = "The retained package did not pass its approved interaction check.";
  if (text.includes("intercepts pointer") || (text.includes("click") && text.includes("timeout"))) {
    detail = "The retained page opened, but its approved primary control could not be used.";
  } else if (text.includes("did not change") || text.includes("unchanged") || text.includes("text_change")) {
    detail = "The retained page opened, but the approved interaction did not change the validated application surface.";
  } else if (text.includes("page error") || text.includes("javascript error")) {
    detail = "The retained page opened with an application error during its approved interaction check.";
  } else if (text.includes("local asset") || text.includes("missing file")) {
    detail = "The retained package could not load all files required by its approved workflow.";
  }

  return {
    detail,
    trust: "This retained candidate is unverified and is not ready to use.",
    actionLabel: "Ask Soma to have the same team repair it",
  };
}
