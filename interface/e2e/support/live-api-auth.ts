import fs from "node:fs";
import path from "node:path";

const repoRoot = path.resolve(__dirname, "../../..");

let cachedAPIKey: string | null | undefined;

function envAPIKey() {
  const direct = process.env.MYCELIS_API_KEY?.trim();
  if (direct) return direct;

  if (cachedAPIKey !== undefined) return cachedAPIKey;

  const envPath = path.join(repoRoot, ".env");
  try {
    const content = fs.readFileSync(envPath, "utf8");
    for (const line of content.split(/\r?\n/)) {
      const match = line.match(/^\s*MYCELIS_API_KEY=(.*)$/);
      if (!match) continue;
      cachedAPIKey = match[1].trim().replace(/^["']|["']$/g, "") || null;
      return cachedAPIKey;
    }
  } catch {
    cachedAPIKey = null;
  }

  cachedAPIKey = null;
  return cachedAPIKey;
}

export function liveAPIHeaders(): Record<string, string> | undefined {
  const apiKey = envAPIKey();
  if (!apiKey) return undefined;
  return {
    Authorization: apiKey.startsWith("Bearer ") ? apiKey : `Bearer ${apiKey}`,
  };
}

export function liveAPIURL(apiPath: string) {
  if (/^https?:\/\//i.test(apiPath)) return apiPath;
  const base = process.env.PLAYWRIGHT_API_BASE_URL ?? "http://127.0.0.1:8081";
  return `${base.replace(/\/$/, "")}/${apiPath.replace(/^\//, "")}`;
}
