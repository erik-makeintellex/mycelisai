const PENDING_SOMA_PROMPT_KEY = "mycelis:pending-soma-prompt";

export function requestSomaPromptHandoff(prompt: string) {
  if (typeof window === "undefined") return;
  const normalized = prompt.trim();
  if (!normalized) return;
  window.sessionStorage.setItem(PENDING_SOMA_PROMPT_KEY, normalized);
  if (window.location.pathname !== "/dashboard") {
    window.location.assign("/dashboard");
  }
}

export function takePendingSomaPrompt(): string | null {
  if (typeof window === "undefined") return null;
  const prompt = window.sessionStorage.getItem(PENDING_SOMA_PROMPT_KEY)?.trim() ?? "";
  if (!prompt) return null;
  window.sessionStorage.removeItem(PENDING_SOMA_PROMPT_KEY);
  return prompt;
}

export function newWorkerProfilePrompt() {
  return "Help me create a reusable Worker Profile. Ask only the questions needed to define its purpose, access, context, and quality checks. Then preview the profile and ask before saving or activating it.";
}

export function customizeWorkerProfilePrompt(name: string) {
  return `Help me create a custom Worker Profile based on \"${name}\". Show the important changes, preview the profile, and ask before saving or activating it.`;
}

export function saveWorkerProfileDraftPrompt(content: string) {
  return [
    "Save this Worker Profile as a governed reusable teammate template.",
    "Preview it first if needed, then ask me before saving the immutable revision.",
    "",
    "```yaml",
    content.trim(),
    "```",
  ].join("\n");
}

export function activateWorkerProfileDraftPrompt(content: string) {
  return [
    "Preview, save, and activate this Worker Profile for future team creation.",
    "Ask me before saving or activating. Do not start a team or agent from this request.",
    "",
    "```yaml",
    content.trim(),
    "```",
  ].join("\n");
}
