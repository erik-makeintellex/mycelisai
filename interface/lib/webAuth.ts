import fs from "node:fs";
import path from "node:path";

export type WebUserRole = "admin" | "standard";

export type WebAuthProvider = "local" | "google";

export interface WebSession {
    sub: string;
    email: string;
    name: string;
    role: WebUserRole;
    provider: WebAuthProvider;
    hd?: string;
    iat: number;
    exp: number;
}

export interface WebAuthConfig {
    sessionSecret: string;
    localUsername: string;
    localPassword: string;
    localPasswordSha256: string;
    googleClientId: string;
    googleClientSecret: string;
    googleRedirectUri: string;
    googleHostedDomain: string;
    allowedDomains: string[];
    adminEmails: string[];
}

export interface ForwardedWebIdentity {
    sub: string;
    email: string;
    name: string;
    role: WebUserRole;
    provider: WebAuthProvider;
    hd?: string;
    iat: number;
}

export const WEB_SESSION_COOKIE = "mycelis_web_session";
const encoder = new TextEncoder();
const decoder = new TextDecoder();
let rootEnvCache: Record<string, string> | null = null;

export function getWebAuthConfig(): WebAuthConfig {
    return {
        sessionSecret: envValue("MYCELIS_WEB_SESSION_SECRET") || envValue("MYCELIS_API_KEY") || "",
        localUsername: envValue("MYCELIS_LOCAL_ADMIN_USERNAME") || "admin",
        localPassword: envValue("MYCELIS_LOCAL_ADMIN_PASSWORD") || envValue("MYCELIS_API_KEY") || "",
        localPasswordSha256: envValue("MYCELIS_LOCAL_ADMIN_PASSWORD_SHA256") || "",
        googleClientId: envValue("MYCELIS_AUTH_GOOGLE_CLIENT_ID") || "",
        googleClientSecret: envValue("MYCELIS_AUTH_GOOGLE_CLIENT_SECRET") || "",
        googleRedirectUri: envValue("MYCELIS_AUTH_GOOGLE_REDIRECT_URI") || "",
        googleHostedDomain: envValue("MYCELIS_AUTH_GOOGLE_HOSTED_DOMAIN") || "",
        allowedDomains: splitList(envValue("MYCELIS_AUTH_ALLOWED_DOMAINS") || envValue("MYCELIS_AUTH_GOOGLE_HOSTED_DOMAIN") || ""),
        adminEmails: splitList(envValue("MYCELIS_AUTH_ADMIN_EMAILS") || ""),
    };
}

export function googleConfigured(config = getWebAuthConfig()): boolean {
    return Boolean(config.googleClientId && config.googleClientSecret && config.googleRedirectUri);
}

export function roleForEmail(email: string, config = getWebAuthConfig()): WebUserRole {
    return config.adminEmails.includes(email.trim().toLowerCase()) ? "admin" : "standard";
}

export async function createSessionToken(session: WebSession, secret: string): Promise<string> {
    const payload = base64UrlEncodeString(JSON.stringify(session));
    const signature = await sign(payload, secret);
    return `${payload}.${signature}`;
}

export async function createForwardedWebIdentityHeaders(session: WebSession, secret: string): Promise<Record<string, string>> {
    if (!secret) return {};
    const payload: ForwardedWebIdentity = {
        sub: session.sub,
        email: session.email,
        name: session.name,
        role: session.role,
        provider: session.provider,
        hd: session.hd,
        iat: Math.floor(Date.now() / 1000),
    };
    const encodedPayload = base64UrlEncodeString(JSON.stringify(payload));
    return {
        "x-mycelis-web-identity": encodedPayload,
        "x-mycelis-web-identity-signature": await sign(encodedPayload, secret),
    };
}

export async function verifySessionToken(token: string | undefined, secret: string): Promise<WebSession | null> {
    if (!token || !secret) return null;
    const [payload, signature] = token.split(".");
    if (!payload || !signature) return null;
    if ((await sign(payload, secret)) !== signature) return null;
    try {
        const session = JSON.parse(base64UrlDecodeString(payload)) as WebSession;
        if (!session.exp || session.exp < Math.floor(Date.now() / 1000)) return null;
        if (session.role !== "admin" && session.role !== "standard") return null;
        return session;
    } catch {
        return null;
    }
}

export function sessionCookieOptions(maxAgeSeconds = 60 * 60 * 8) {
    return {
        httpOnly: true,
        sameSite: "lax" as const,
        secure: secureCookieEnabled(),
        path: "/",
        maxAge: maxAgeSeconds,
    };
}

export function secureCookieEnabled(): boolean {
    if (envValue("MYCELIS_WEB_COOKIE_SECURE")) return envValue("MYCELIS_WEB_COOKIE_SECURE") === "true";
    return Boolean(envValue("MYCELIS_PUBLIC_ORIGIN").startsWith("https://"));
}

export function webAuthRedirectURL(path: string, fallbackOrigin?: string | null): URL {
    return new URL(path, webAuthBaseOrigin(fallbackOrigin));
}

export function webAuthBaseOrigin(fallbackOrigin?: string | null): string {
    const fallback = (fallbackOrigin || "").trim();
    if (fallback && !fallback.includes("[::]")) return fallback.replace(/\/+$/, "");
    const configured = envValue("MYCELIS_PUBLIC_ORIGIN").trim().replace(/\/+$/, "");
    if (configured) return configured;
    return "http://127.0.0.1:3000";
}

export function encodeOAuthStateCookie(state: string, nextPath: string): string {
    return encodeURIComponent(JSON.stringify({ state, next: nextPath }));
}

export function decodeOAuthStateCookie(value: string | undefined): { state: string; next: string } {
    if (!value) return { state: "", next: "/dashboard" };
    try {
        const parsed = JSON.parse(decodeURIComponent(value)) as { state?: unknown; next?: unknown };
        const state = typeof parsed.state === "string" ? parsed.state : "";
        const next = typeof parsed.next === "string" && parsed.next.startsWith("/") && !parsed.next.startsWith("//")
            ? parsed.next
            : "/dashboard";
        return { state, next };
    } catch {
        return { state: "", next: "/dashboard" };
    }
}

export async function sha256Hex(input: string): Promise<string> {
    const hash = await crypto.subtle.digest("SHA-256", encoder.encode(input));
    return Array.from(new Uint8Array(hash)).map((b) => b.toString(16).padStart(2, "0")).join("");
}

export function splitList(value: string): string[] {
    return value.split(",").map((item) => item.trim().toLowerCase()).filter(Boolean);
}

function envValue(key: string): string {
    return process.env[key] ?? rootEnv()[key] ?? "";
}

function rootEnv(): Record<string, string> {
    if (rootEnvCache) return rootEnvCache;
    rootEnvCache = {};
    for (const candidate of rootEnvCandidates()) {
        if (!fs.existsSync(candidate)) continue;
        rootEnvCache = parseEnvFile(candidate);
        break;
    }
    return rootEnvCache;
}

function rootEnvCandidates(): string[] {
    return [
        path.resolve(process.cwd(), ".env"),
        path.resolve(process.cwd(), "..", ".env"),
    ];
}

function parseEnvFile(envPath: string): Record<string, string> {
    const values: Record<string, string> = {};
    for (const line of fs.readFileSync(envPath, "utf8").split(/\r?\n/)) {
        const trimmed = line.trim();
        if (!trimmed || trimmed.startsWith("#")) continue;
        const equals = trimmed.indexOf("=");
        if (equals <= 0) continue;
        const key = trimmed.slice(0, equals).trim();
        let value = trimmed.slice(equals + 1).trim();
        if ((value.startsWith("\"") && value.endsWith("\"")) || (value.startsWith("'") && value.endsWith("'"))) {
            value = value.slice(1, -1);
        }
        values[key] = value;
    }
    return values;
}

async function sign(payload: string, secret: string): Promise<string> {
    const key = await crypto.subtle.importKey("raw", encoder.encode(secret), { name: "HMAC", hash: "SHA-256" }, false, ["sign"]);
    const signature = await crypto.subtle.sign("HMAC", key, encoder.encode(payload));
    return base64UrlEncodeBytes(new Uint8Array(signature));
}

function base64UrlEncodeString(value: string): string {
    return base64UrlEncodeBytes(encoder.encode(value));
}

function base64UrlEncodeBytes(bytes: Uint8Array): string {
    let binary = "";
    bytes.forEach((byte) => {
        binary += String.fromCharCode(byte);
    });
    return btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/g, "");
}

function base64UrlDecodeString(value: string): string {
    const padded = value.replace(/-/g, "+").replace(/_/g, "/").padEnd(Math.ceil(value.length / 4) * 4, "=");
    const binary = atob(padded);
    const bytes = Uint8Array.from(binary, (char) => char.charCodeAt(0));
    return decoder.decode(bytes);
}
