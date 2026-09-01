import type { WorkOutputContractData } from "@/store/useCortexStore";

export type DeliverableKind = "game" | "app" | "report" | "document" | "data" | "media" | "image" | "result";

export type DeliverablePresentation = {
  kind: DeliverableKind;
  actionLabel: "Play game" | "Open app" | "Read report" | "Review document" | "View data" | "Play media" | "View image" | "Open result";
};

export type DeliverablePresentationInput = {
  outputContract?: WorkOutputContractData;
  kind?: string | null;
  type?: string | null;
  contentType?: string | null;
  title?: string | null;
  entrypoint?: string | null;
  path?: string | null;
};

const PRESENTATIONS: Record<DeliverableKind, DeliverablePresentation> = {
  game: { kind: "game", actionLabel: "Play game" },
  app: { kind: "app", actionLabel: "Open app" },
  report: { kind: "report", actionLabel: "Read report" },
  document: { kind: "document", actionLabel: "Review document" },
  data: { kind: "data", actionLabel: "View data" },
  media: { kind: "media", actionLabel: "Play media" },
  image: { kind: "image", actionLabel: "View image" },
  result: { kind: "result", actionLabel: "Open result" },
};

export function deliverablePresentation(input: DeliverablePresentationInput): DeliverablePresentation {
  const contractSignals = [
    input.outputContract?.primary_deliverable,
    input.outputContract?.shape,
    input.outputContract?.launch_hint,
  ].map(normalizedSignal).filter(Boolean);
  const outputSignals = [
    input.kind,
    input.type,
    input.contentType,
  ].map(normalizedSignal).filter(Boolean);

  for (const signal of contractSignals) {
    const kind = explicitDeliverableKind(signal);
    if (kind) return PRESENTATIONS[kind];
  }
  for (const signal of outputSignals) {
    const kind = explicitDeliverableKind(signal, false);
    if (kind) return PRESENTATIONS[kind];
  }

  const titleKind = explicitDeliverableKind(normalizedSignal(input.title), false);
  if (titleKind) return PRESENTATIONS[titleKind];

  for (const signal of outputSignals) {
    const kind = explicitDeliverableKind(signal);
    if (kind) return PRESENTATIONS[kind];
  }

  if (contractSignals.length > 0 || outputSignals.length > 0) return PRESENTATIONS.result;
  return PRESENTATIONS[pathDeliverableKind(input.entrypoint ?? input.path)];
}

function normalizedSignal(value?: string | null) {
  return value?.trim().toLowerCase() ?? "";
}

function explicitDeliverableKind(signal: string, includeGenericPackage = true): DeliverableKind | null {
  if (!signal) return null;
  const words = ` ${signal.replace(/[^a-z0-9]+/g, " ")} `;
  const has = (...tokens: string[]) => tokens.some((token) => words.includes(` ${token} `));

  if (has("game")) return "game";
  if (has("image", "photo", "picture", "illustration")) return "image";
  if (has("audio", "video", "media", "podcast", "animation")) return "media";
  if (has("data", "dataset", "table", "csv", "spreadsheet", "xlsx", "parquet")) return "data";
  if (has("report")) return "report";
  if (has("document", "markdown", "pdf", "doc", "docx")) return "document";
  if (has("app", "site", "website", "webpage", "microsite", "html", "dashboard")
    || words.includes(" app package ")
    || (includeGenericPackage && words.includes(" project package "))) return "app";
  return null;
}

function pathDeliverableKind(value?: string | null): DeliverableKind {
  const path = value?.trim().toLowerCase().split(/[?#]/, 1)[0] ?? "";
  if (/\.(?:html?|xhtml)$/.test(path)) return "app";
  if (/\.(?:png|jpe?g|gif|webp|svg|avif|bmp|tiff?)$/.test(path)) return "image";
  if (/\.(?:mp3|wav|ogg|m4a|aac|flac|mp4|webm|mov|m4v|avi)$/.test(path)) return "media";
  if (/\.(?:csv|tsv|json|jsonl|xlsx?|parquet|sqlite|db)$/.test(path)) return "data";
  if (/\.(?:md|mdx|txt|pdf|docx?|rtf|odt)$/.test(path)) return "document";
  return "result";
}
