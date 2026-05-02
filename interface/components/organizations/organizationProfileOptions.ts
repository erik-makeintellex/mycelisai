import type {
    OrganizationAIEngineProfileId,
    OrganizationOutputModelRoutingPayload,
    OrganizationOutputTypeId,
    ResponseContractProfileId,
} from "@/lib/organizations";

export const AI_ENGINE_OPTIONS: Array<{
    id: OrganizationAIEngineProfileId;
    label: string;
    description: string;
    goodFor: string;
}> = [
    {
        id: "starter_defaults",
        label: "Starter Defaults",
        description: "Keeps the guided starter profile that already came with this AI Organization.",
        goodFor: "Best for keeping the original setup intact while Soma settles the first workflow.",
    },
    {
        id: "balanced",
        label: "Balanced",
        description: "Steady planning depth and response quality for everyday Soma work.",
        goodFor: "Best for most organizations that want dependable guidance across planning and execution.",
    },
    {
        id: "high_reasoning",
        label: "High Reasoning",
        description: "Adds more careful thinking when planning and tradeoffs need extra attention.",
        goodFor: "Best for complex decisions, deeper review, and more deliberate next-step planning.",
    },
    {
        id: "fast_lightweight",
        label: "Fast & Lightweight",
        description: "Keeps responses quick and planning lighter for rapid iteration.",
        goodFor: "Best for quick reviews, check-ins, and lighter day-to-day coordination.",
    },
    {
        id: "deep_planning",
        label: "Deep Planning",
        description: "Leans into longer multi-step planning and more deliberate organization shaping.",
        goodFor: "Best for designing larger workstreams and sequencing bigger efforts.",
    },
];

export const RESPONSE_CONTRACT_OPTIONS: Array<{
    id: ResponseContractProfileId;
    label: string;
    toneStyle: string;
    structure: string;
    verbosity: string;
    bestFor: string;
}> = [
    {
        id: "clear_balanced",
        label: "Clear & Balanced",
        toneStyle: "Straightforward and steady without sounding cold.",
        structure: "Uses clear sections and practical takeaways when helpful.",
        verbosity: "Balanced detail with enough context to act confidently.",
        bestFor: "Best for everyday Soma guidance, reviews, and general coordination.",
    },
    {
        id: "structured_analytical",
        label: "Structured & Analytical",
        toneStyle: "Measured, methodical, and reasoning-forward.",
        structure: "Organizes answers into clear steps, comparisons, or frameworks.",
        verbosity: "Moderate-to-detailed when structure improves decision-making.",
        bestFor: "Best for planning, tradeoffs, diagnosis, and deeper review work.",
    },
    {
        id: "concise_direct",
        label: "Concise & Direct",
        toneStyle: "Focused, efficient, and low-friction.",
        structure: "Keeps responses short and action-led unless more detail is needed.",
        verbosity: "Intentionally brief with only the highest-signal details.",
        bestFor: "Best for quick decisions, status checks, and fast-moving execution work.",
    },
    {
        id: "warm_supportive",
        label: "Warm & Supportive",
        toneStyle: "Encouraging, collaborative, and reassuring.",
        structure: "Still organized, but written to feel more human and supportive.",
        verbosity: "Balanced detail with a little more guidance and framing.",
        bestFor: "Best for onboarding, operator guidance, and people-facing support work.",
    },
];

export const OUTPUT_TYPE_OPTIONS: Array<{ id: OrganizationOutputTypeId; label: string; description: string }> = [
    { id: "general_text", label: "General text", description: "Default chat, writing, and broad team communication outputs." },
    { id: "research_reasoning", label: "Research & reasoning", description: "Planning, review, synthesis, and deeper reasoning-heavy outputs." },
    { id: "code_generation", label: "Code generation", description: "Implementation, code repair, and execution-heavy output lanes." },
    { id: "vision_analysis", label: "Vision analysis", description: "Image understanding, OCR, and visual review support." },
];

export function responseContractDetailItems(summary: string) {
    const option = RESPONSE_CONTRACT_OPTIONS.find((item) => item.label === summary.trim()) ?? RESPONSE_CONTRACT_OPTIONS[0];
    return [
        {
            name: "Current response style",
            purpose: `Current profile: ${option.label}.`,
            supportCue: option.bestFor,
        },
        {
            name: "Tone and style",
            purpose: option.toneStyle,
            supportCue: "Guides how supportive, direct, or analytical Soma should sound by default.",
        },
        {
            name: "Structure and detail",
            purpose: `${option.structure} ${option.verbosity}`,
            supportCue: "Guides how organized and how detailed responses should feel across the AI Organization.",
        },
    ];
}

export function outputModelOptions(routing: OrganizationOutputModelRoutingPayload | null) {
    const options = routing?.available_models ?? [];
    return options.length > 0 ? options : [];
}
