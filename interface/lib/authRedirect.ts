// Post-login redirect guard shared by every auth route and the login page.

const NEXT_PATH_BASE = "http://next-path.mycelis.invalid";
const UNSAFE_NEXT_CHARS = /[\\\u0000-\u0020\u007f]/;

/**
 * The single post-login redirect guard. Returns a same-origin
 * pathname+search+hash, or "" when the value could leave the origin. Rejects
 * backslashes and control/whitespace characters in raw and percent-decoded
 * forms, and anything that does not start with exactly one "/".
 */
function passesDecodedChecks(value: string): boolean {
    let decoded = value;
    for (let depth = 0; depth < 4; depth++) {
        if (UNSAFE_NEXT_CHARS.test(decoded)) return false;
        if (!decoded.startsWith("/") || decoded.startsWith("//")) return false;
        let next: string;
        try {
            next = decodeURIComponent(decoded);
        } catch {
            return false;
        }
        if (next === decoded) return true;
        decoded = next;
    }
    return false;
}

export function safeNextPath(value: string | null | undefined): string {
    if (typeof value !== "string" || !value || !passesDecodedChecks(value)) return "";
    try {
        const url = new URL(value, NEXT_PATH_BASE);
        if (url.origin !== NEXT_PATH_BASE) return "";
        const path = `${url.pathname}${url.search}${url.hash}`;
        // Dot segments can normalize into "//host" (e.g. "/.//evil.example"), so the
        // normalized result must pass the same raw and decoded checks.
        return passesDecodedChecks(path) ? path : "";
    } catch {
        return "";
    }
}
