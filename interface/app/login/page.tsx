import Link from "next/link";
import { redirect } from "next/navigation";
import { cookies } from "next/headers";
import { ArrowRight, Building2, KeyRound, ShieldCheck, Sparkles } from "lucide-react";
import type { ReactNode } from "react";
import { ThemeSync } from "@/components/shell/ThemeSync";
import { WEB_SESSION_COOKIE, getWebAuthConfig, googleConfigured, verifySessionToken } from "@/lib/webAuth";

interface LoginPageProps {
    searchParams?: Promise<{ error?: string; next?: string }>;
}

export const dynamic = "force-dynamic";

export default async function LoginPage({ searchParams }: LoginPageProps) {
    const params = await searchParams;
    const config = getWebAuthConfig();
    const cookieStore = await cookies();
    const session = await verifySessionToken(cookieStore.get(WEB_SESSION_COOKIE)?.value, config.sessionSecret);
    if (session) redirect(safeNext(params?.next) || "/dashboard");

    const next = safeNext(params?.next) || "/dashboard";
    const localReady = Boolean(config.sessionSecret && config.localPassword);
    const googleReady = Boolean(config.sessionSecret && googleConfigured(config));
    const allowedDomains = config.allowedDomains;
    const errorText = errorMessage(params?.error, allowedDomains);

    return (
        <main className="min-h-screen bg-cortex-bg text-cortex-text-main">
            <ThemeSync />
            <div className="mx-auto grid min-h-screen max-w-6xl gap-8 px-6 py-8 lg:grid-cols-[0.88fr_1fr] lg:items-center">
                <section className="space-y-6">
                    <span className="inline-flex items-center gap-2 rounded-full border border-cortex-border bg-cortex-surface px-4 py-2 text-xs font-mono uppercase tracking-[0.22em] text-cortex-primary">
                        <Sparkles className="h-4 w-4" />
                        Soma
                    </span>
                    <div>
                        <h1 className="text-4xl font-semibold leading-tight sm:text-5xl">Sign in to operate Mycelis.</h1>
                        <p className="mt-4 max-w-xl text-lg leading-8 text-cortex-text-muted">
                            Every edition starts behind an identity boundary. Free nodes use a local owner login; enterprise deployments can add Google Workspace SSO without changing the Soma workflow.
                        </p>
                    </div>
                    <div className="grid gap-3 sm:grid-cols-2">
                        <TrustTile icon={<ShieldCheck className="h-5 w-5" />} title="Admin" body="Configures identity, people, providers, tools, deployment trust, and recovery." />
                        <TrustTile icon={<Building2 className="h-5 w-5" />} title="Standard user" body="Works with Soma, teams, outputs, runs, and proof inside assigned organizations." />
                    </div>
                </section>

                <section className="rounded-2xl border border-cortex-border bg-cortex-surface p-6 shadow-[0_18px_50px_rgba(0,0,0,0.18)]">
                    <div className="mb-5">
                        <p className="text-[11px] font-mono uppercase tracking-[0.22em] text-cortex-primary">Access required</p>
                        <h2 className="mt-2 text-2xl font-semibold">Continue to Soma</h2>
                    </div>
                    {errorText ? (
                        <p role="alert" className="mb-4 rounded-xl border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-200">
                            {errorText}
                        </p>
                    ) : null}

                    <form action={`/auth/local?next=${encodeURIComponent(next)}`} method="post" className="space-y-4">
                        <label className="block">
                            <span className="text-sm font-medium text-cortex-text-muted">Local admin username</span>
                            <input name="username" autoComplete="username" defaultValue={config.localUsername} className="mt-2 w-full rounded-xl border border-cortex-border bg-cortex-bg px-4 py-3 text-cortex-text-main outline-none focus:border-cortex-primary" />
                        </label>
                        <label className="block">
                            <span className="text-sm font-medium text-cortex-text-muted">Password or local API key</span>
                            <input name="password" type="password" autoComplete="current-password" className="mt-2 w-full rounded-xl border border-cortex-border bg-cortex-bg px-4 py-3 text-cortex-text-main outline-none focus:border-cortex-primary" />
                        </label>
                        <button disabled={!localReady} type="submit" className="inline-flex w-full items-center justify-center gap-2 rounded-xl bg-cortex-primary px-4 py-3 font-semibold text-cortex-bg hover:bg-cortex-primary/90 disabled:cursor-not-allowed disabled:opacity-50">
                            Sign in as local admin
                            <ArrowRight className="h-4 w-4" />
                        </button>
                    </form>

                    <div className="my-6 flex items-center gap-3 text-xs uppercase tracking-[0.18em] text-cortex-text-muted">
                        <span className="h-px flex-1 bg-cortex-border" />
                        Enterprise SSO
                        <span className="h-px flex-1 bg-cortex-border" />
                    </div>
                    {googleReady ? (
                        <div className="space-y-3">
                            <Link href={`/auth/google/start?next=${encodeURIComponent(next)}`} className="inline-flex w-full items-center justify-center gap-2 rounded-xl border border-cortex-border bg-cortex-bg px-4 py-3 font-semibold hover:border-cortex-primary/40">
                                <KeyRound className="h-4 w-4" />
                                Sign in with Google Workspace
                            </Link>
                            {allowedDomains.length ? (
                                <p className="rounded-xl border border-cortex-border bg-cortex-bg px-4 py-3 text-xs leading-5 text-cortex-text-muted">
                                    Use a Google account from: <span className="font-mono text-cortex-text-main">{allowedDomains.join(", ")}</span>.
                                    Personal Gmail accounts are rejected for this deployment.
                                </p>
                            ) : null}
                        </div>
                    ) : (
                        <p className="rounded-xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm leading-6 text-cortex-text-muted">
                            Google Workspace login appears after `MYCELIS_AUTH_GOOGLE_CLIENT_ID`, `MYCELIS_AUTH_GOOGLE_CLIENT_SECRET`, and `MYCELIS_AUTH_GOOGLE_REDIRECT_URI` are configured.
                        </p>
                    )}
                </section>
            </div>
        </main>
    );
}

function TrustTile({ icon, title, body }: { icon: ReactNode; title: string; body: string }) {
    return (
        <div className="rounded-xl border border-cortex-border bg-cortex-surface p-4">
            <div className="text-cortex-primary">{icon}</div>
            <h3 className="mt-3 font-semibold">{title}</h3>
            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{body}</p>
        </div>
    );
}

function safeNext(value?: string): string {
    return value && value.startsWith("/") && !value.startsWith("//") ? value : "";
}

function errorMessage(code: string | undefined, allowedDomains: string[]): string {
    if (code === "invalid") return "The local username or password was not accepted.";
    if (code === "google") return "Google Workspace could not complete sign-in for this deployment.";
    if (code === "google_state") return "Google sign-in state expired or did not match. Start Google sign-in again from this browser tab.";
    if (code === "google_token") return "Google rejected the OAuth token exchange. Verify the client ID, client secret, and authorized redirect URI for this deployment.";
    if (code === "google_tokeninfo") return "Google returned a token, but Mycelis could not verify the Google identity token.";
    if (code === "google_identity") return "Google sign-in returned an identity token without a verified email for this deployment.";
    if (code === "google_exception") return "Google sign-in could not complete because the callback request failed. Check Interface logs for the sanitized auth phase.";
    if (code === "domain") {
        const domains = allowedDomains.length ? ` Use ${allowedDomains.join(", ")}.` : "";
        return `That Google account is outside the allowed Workspace domain.${domains}`;
    }
    if (code === "config") return "Authentication is not fully configured for this deployment.";
    return "";
}
