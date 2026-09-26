import { createHash, timingSafeEqual } from "node:crypto";
import fs from "node:fs";
import path from "node:path";
import { safeNextPath } from "./authRedirect";

export { safeNextPath };

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
    /** Empty when MYCELIS_WEB_SESSION_SECRET or the forward secret is missing, short or reused. */
    sessionSecret: string;
    /** MYCELIS_WEB_IDENTITY_FORWARD_SECRET only; never falls back to another secret. */
    forwardSecret: string;
    localUsername: string;
    localPassword: string;
    localPasswordSha256: string;
    googleClientId: string;
    googleClientSecret: string;
    googleRedirectUri: string;
    googleHostedDomain: string;
    allowedDomains: string[];
    adminEmails: string[];
    /** Operator-facing setup problems. They name variables and never contain values. */
    setupIssues: string[];
}

export interface GoogleWorkspacePolicy {
    hostedDomain: string;
    allowedDomains: string[];
    displayDomains: string[];
    domainLabel: string;
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

const MIN_SECRET_BYTES = 32;
const DEV_KEY_HINT = "uv run inv auth.dev-key";

// Secrets are read from exactly one variable each. There is no fallback chain:
// a missing, short or reused secret disables the provider that needs it.
export function getWebAuthConfig(): WebAuthConfig {
    const issues: string[] = [];
    // Trimmed like Core so both sides agree on length and distinctness.
    const apiKeys = [envValue("MYCELIS_API_KEY").trim(), envValue("MYCELIS_BREAK_GLASS_API_KEY").trim()].filter(Boolean);
    const rawSession = envValue("MYCELIS_WEB_SESSION_SECRET").trim();
    const rawForward = envValue("MYCELIS_WEB_IDENTITY_FORWARD_SECRET").trim();
    const sessionIssue = secretIssue("MYCELIS_WEB_SESSION_SECRET", rawSession, apiKeys);
    const forwardIssue = secretIssue("MYCELIS_WEB_IDENTITY_FORWARD_SECRET", rawForward, apiKeys)
        || (rawForward && rawForward === rawSession ? "MYCELIS_WEB_IDENTITY_FORWARD_SECRET must differ from MYCELIS_WEB_SESSION_SECRET." : "");
    if (sessionIssue) issues.push(sessionIssue);
    if (forwardIssue) issues.push(forwardIssue);
    const sessionReady = !sessionIssue && !forwardIssue;

    const reserved = [...apiKeys, rawSession, rawForward].filter(Boolean);
    let localPassword = envValue("MYCELIS_LOCAL_ADMIN_PASSWORD");
    let localPasswordSha256 = envValue("MYCELIS_LOCAL_ADMIN_PASSWORD_SHA256").trim().toLowerCase();
    if (localPassword && reserved.includes(localPassword)) {
        issues.push("MYCELIS_LOCAL_ADMIN_PASSWORD must not reuse MYCELIS_API_KEY, MYCELIS_BREAK_GLASS_API_KEY or a web secret.");
        localPassword = "";
    }
    if (localPasswordSha256 && (!/^[a-f0-9]{64}$/.test(localPasswordSha256) || reserved.some((value) => sha256HexSync(value) === localPasswordSha256))) {
        issues.push("MYCELIS_LOCAL_ADMIN_PASSWORD_SHA256 must be a 64-character hex SHA-256 of a password that is not an API key or web secret.");
        localPasswordSha256 = "";
    }
    if (!localPassword && !localPasswordSha256 && !issues.some((issue) => issue.startsWith("MYCELIS_LOCAL_ADMIN_PASSWORD"))) {
        issues.push("Set MYCELIS_LOCAL_ADMIN_PASSWORD_SHA256 (preferred) or MYCELIS_LOCAL_ADMIN_PASSWORD to enable local sign-in. MYCELIS_API_KEY is no longer accepted as a password.");
    }

    return {
        sessionSecret: sessionReady ? rawSession : "",
        forwardSecret: sessionReady ? rawForward : "",
        localUsername: envValue("MYCELIS_LOCAL_ADMIN_USERNAME") || "admin",
        localPassword,
        localPasswordSha256,
        googleClientId: envValue("MYCELIS_AUTH_GOOGLE_CLIENT_ID") || "",
        googleClientSecret: envValue("MYCELIS_AUTH_GOOGLE_CLIENT_SECRET") || "",
        googleRedirectUri: envValue("MYCELIS_AUTH_GOOGLE_REDIRECT_URI") || "",
        googleHostedDomain: envValue("MYCELIS_AUTH_GOOGLE_HOSTED_DOMAIN") || "",
        allowedDomains: splitList(envValue("MYCELIS_AUTH_ALLOWED_DOMAINS") || envValue("MYCELIS_AUTH_GOOGLE_HOSTED_DOMAIN") || ""),
        adminEmails: splitList(envValue("MYCELIS_AUTH_ADMIN_EMAILS") || ""),
        setupIssues: issues,
    };
}

function secretIssue(name: string, value: string, apiKeys: string[]): string {
    if (!value) return `Set ${name} (at least ${MIN_SECRET_BYTES} bytes; ${DEV_KEY_HINT} generates it).`;
    if (encoder.encode(value).length < MIN_SECRET_BYTES) return `${name} must be at least ${MIN_SECRET_BYTES} bytes (${DEV_KEY_HINT}).`;
    if (apiKeys.includes(value)) return `${name} must differ from MYCELIS_API_KEY and MYCELIS_BREAK_GLASS_API_KEY.`;
    return "";
}

export function localLoginConfigured(config: WebAuthConfig): boolean {
    return Boolean(config.sessionSecret && (config.localPassword || config.localPasswordSha256));
}

/** Constant-time local password check over SHA-256 digests. */
export function verifyLocalPassword(password: string, config: WebAuthConfig): boolean {
    const expectedHex = config.localPasswordSha256 || (config.localPassword ? sha256HexSync(config.localPassword) : "");
    if (!expectedHex) return false;
    const expected = Buffer.from(expectedHex, "hex");
    const actual = createHash("sha256").update(password, "utf8").digest();
    return expected.length === actual.length && timingSafeEqual(expected, actual);
}

/**
 * CSRF guard for state-changing auth POSTs. A browser request is accepted only
 * when its Origin matches the configured public origin or the server's own
 * origin, and Fetch Metadata does not mark it cross-site.
 */
export function requestOriginAllowed(headers: Headers, serverOrigin: string): boolean {
    if ((headers.get("sec-fetch-site") || "").toLowerCase() === "cross-site") return false;
    const origin = headers.get("origin");
    if (!origin) return true;
    const allowed = new Set([webAuthBaseOrigin(serverOrigin), safeOrigin(serverOrigin)].filter(Boolean));
    return allowed.has(origin.trim());
}

/**
 * The origin this server was addressed at (scheme plus Host header), never the
 * Origin header. NextURL rewrites loopback hosts to "localhost", so the Host
 * header is used to keep 127.0.0.1 and localhost distinct.
 */
export function serverOrigin(request: { headers: Headers; nextUrl: URL }): string {
    const host = (request.headers.get("host") || "").trim();
    return (host && safeOrigin(`${request.nextUrl.protocol}//${host}`)) || request.nextUrl.origin;
}

function safeOrigin(value: string): string {
    try {
        return new URL(value).origin;
    } catch {
        return "";
    }
}

function sha256HexSync(value: string): string {
    return createHash("sha256").update(value, "utf8").digest("hex");
}

export function googleConfigured(config = getWebAuthConfig()): boolean {
    return Boolean(config.googleClientId && config.googleClientSecret && config.googleRedirectUri);
}

export function googleWorkspacePolicy(config = getWebAuthConfig()): GoogleWorkspacePolicy {
    const allowedDomains = [...config.allowedDomains];
    const hostedDomain = config.googleHostedDomain.trim().toLowerCase();
    const displayDomains = allowedDomains.length ? allowedDomains : (hostedDomain ? [hostedDomain] : []);
    return {
        hostedDomain,
        allowedDomains,
        displayDomains,
        domainLabel: displayDomains.join(", "),
    };
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
    const parts = token.split(".");
    if (parts.length !== 2) return null;
    const [payload, signature] = parts;
    if (!BASE64URL.test(payload) || !BASE64URL.test(signature)) return null;
    try {
        // HMAC verify is constant time; there is no string comparison of signatures.
        if (!(await verifySignature(payload, signature, secret))) return null;
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
    for (const candidate of [envValue("MYCELIS_PUBLIC_ORIGIN"), fallbackOrigin]) {
        try {
            const url = new URL((candidate || "").trim());
            if (!["http:", "https:"].includes(url.protocol)) continue;
            if (["0.0.0.0", "[::]"].includes(url.hostname)) continue;
            if (url.username || url.password) continue;
            return url.origin;
        } catch {
            // Invalid or missing origins cannot become browser destinations.
        }
    }
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
        const next = safeNextPath(typeof parsed.next === "string" ? parsed.next : "") || "/dashboard";
        return { state, next };
    } catch {
        return { state: "", next: "/dashboard" };
    }
}

export async function sha256Hex(input: string): Promise<string> {
    const hash = await webCryptoSubtle().digest("SHA-256", encoder.encode(input));
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
    const subtle = webCryptoSubtle();
    const key = await subtle.importKey("raw", encoder.encode(secret), { name: "HMAC", hash: "SHA-256" }, false, ["sign"]);
    const signature = await subtle.sign("HMAC", key, encoder.encode(payload));
    return base64UrlEncodeBytes(new Uint8Array(signature));
}

const BASE64URL = /^[A-Za-z0-9_-]+$/;

async function verifySignature(payload: string, signature: string, secret: string): Promise<boolean> {
    const subtle = webCryptoSubtle();
    const key = await subtle.importKey("raw", encoder.encode(secret), { name: "HMAC", hash: "SHA-256" }, false, ["verify"]);
    const bytes = base64UrlDecodeBytes(signature);
    return subtle.verify("HMAC", key, bytes, encoder.encode(payload));
}

function webCryptoSubtle(): SubtleCrypto {
    if (globalThis.crypto?.subtle) return globalThis.crypto.subtle;
    throw new Error("Web Crypto is unavailable for Mycelis web auth");
}

function base64UrlEncodeString(value: string): string {
    return base64UrlEncodeBytes(encoder.encode(value));
}

function base64UrlEncodeBytes(bytes: Uint8Array): string {
    if (typeof btoa !== "function") return Buffer.from(bytes).toString("base64").replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/g, "");
    let binary = "";
    bytes.forEach((byte) => {
        binary += String.fromCharCode(byte);
    });
    return btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/g, "");
}

function base64UrlDecodeString(value: string): string {
    return decoder.decode(base64UrlDecodeBytes(value));
}

function base64UrlDecodeBytes(value: string): Uint8Array<ArrayBuffer> {
    const padded = value.replace(/-/g, "+").replace(/_/g, "/").padEnd(Math.ceil(value.length / 4) * 4, "=");
    if (typeof atob !== "function") return new Uint8Array(Buffer.from(padded, "base64"));
    const binary = atob(padded);
    return Uint8Array.from(binary, (char) => char.charCodeAt(0));
}
