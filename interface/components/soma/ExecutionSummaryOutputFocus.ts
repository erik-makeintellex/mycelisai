import type { ExecutionSummaryData, ExecutionSummaryItem } from "@/store/useCortexStore";
import { asItems, compactText, itemText } from "./ExecutionSummaryCardModel";

export type OutputFocus = { label: string; detail: string };

const OUTPUT_LABELS: Record<string, OutputFocus> = {
    table: { label: "Table / report", detail: "Use clear columns, rows, and source/assumption notes where relevant." },
    app_package: { label: "App / package", detail: "Return an openable package with launch, validation, folder, and follow-up support." },
    code_app: { label: "Code / script", detail: "Return reviewable files with usage and validation notes." },
    media: { label: "Media", detail: "Save the artifact, show a preview when possible, and state provider/proof boundaries." },
    document: { label: "Document", detail: "Use the requested structure and keep assumptions, sources, and next steps readable." },
    data: { label: "Dataset", detail: "Keep schema, source boundary, and validation/recovery notes visible." },
    package: { label: "Project package", detail: "Include direct launch/open controls plus folder and proof access." },
};

export function outputFocus(summary: ExecutionSummaryData): OutputFocus | null {
    const outputs = asItems(summary.outputs);
    const outputText = outputs.map(outputSignal).filter(Boolean).join(" ").toLowerCase();
    const requestText = [
        typeof summary.intent === "string" ? summary.intent : summary.intent?.original,
        typeof summary.intent === "string" ? "" : summary.intent?.resolved,
        typeof summary.understanding === "string" ? summary.understanding : summary.understanding?.summary,
        summary.execution?.summary,
        summary.execution_summary,
        summary.work_intent?.objective,
    ].map((value) => compactText(value)).filter(Boolean).join(" ").toLowerCase();
    const text = `${outputText} ${requestText}`;

    if (/\b(application_package|project_package|browser app|web app|mobile app|executable|launch)\b/.test(text)) return outputType("app_package");
    if (/\b(table|spreadsheet|csv|matrix|row|columns?|dataset|data extract)\b/.test(text)) return outputType("table");
    if (/\b(source code|code|script|repository)\b/.test(text)) return outputType("code_app");
    if (/\b(image|audio|video|media|music|sprite|sound)\b/.test(text)) return outputType("media");
    if (/\b(json|schema|records?|dataset|data file)\b/.test(text)) return outputType("data");
    if (/\b(markdown|document|report|brief|readme|plan|proposal|copy)\b/.test(text)) return outputType("document");
    return outputs.some((item) => typeof item !== "string" && (item.entrypoint || item.folder)) ? outputType("app_package") : null;
}

function outputSignal(item: string | ExecutionSummaryItem) {
    if (typeof item === "string") return item;
    return [
        itemText(item),
        item.kind,
        item.type,
        item.content_type,
        item.output_class,
        item.path,
        item.entrypoint,
        item.folder,
    ].map((value) => compactText(value)).filter(Boolean).join(" ");
}

function outputType(key: keyof typeof OUTPUT_LABELS) {
    return OUTPUT_LABELS[key];
}
