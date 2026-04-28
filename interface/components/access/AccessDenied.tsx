import Link from "next/link";
import { ArrowLeft, LockKeyhole, ShieldAlert } from "lucide-react";

interface AccessDeniedProps {
    title?: string;
    message?: string;
    requiredAccess?: string;
    returnHref?: string;
    returnLabel?: string;
    supportHref?: string;
    supportLabel?: string;
}

export default function AccessDenied({
    title = "Access denied",
    message = "Your current role cannot open this area. Ask an owner or administrator to update your access before continuing.",
    requiredAccess = "Authorized workspace role",
    returnHref = "/dashboard",
    returnLabel = "Back to dashboard",
    supportHref,
    supportLabel = "Request access",
}: AccessDeniedProps) {
    return (
        <main className="min-h-full bg-cortex-bg px-4 py-10 text-cortex-text-main sm:px-6 lg:px-8">
            <section className="mx-auto flex min-h-[calc(100vh-7rem)] max-w-3xl items-center justify-center">
                <div className="w-full rounded-lg border border-cortex-border bg-cortex-surface px-6 py-7 shadow-sm sm:px-8 sm:py-8">
                    <div className="flex flex-col gap-6 sm:flex-row sm:items-start">
                        <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-lg border border-cortex-danger/25 bg-cortex-danger/10 text-cortex-danger">
                            <ShieldAlert className="h-6 w-6" aria-hidden="true" />
                        </div>

                        <div className="min-w-0 flex-1">
                            <p className="text-xs font-semibold uppercase tracking-[0.16em] text-cortex-danger">
                                Permission required
                            </p>
                            <h1 className="mt-3 text-3xl font-semibold tracking-tight text-cortex-text-main">
                                {title}
                            </h1>
                            <p className="mt-4 max-w-2xl text-sm leading-6 text-cortex-text-muted">
                                {message}
                            </p>

                            <div className="mt-6 rounded-lg border border-cortex-border bg-cortex-bg px-4 py-3">
                                <div className="flex items-center gap-3">
                                    <LockKeyhole className="h-4 w-4 shrink-0 text-cortex-primary" aria-hidden="true" />
                                    <div>
                                        <p className="text-xs font-medium uppercase tracking-[0.14em] text-cortex-text-muted">
                                            Required access
                                        </p>
                                        <p className="mt-1 text-sm font-semibold text-cortex-text-main">
                                            {requiredAccess}
                                        </p>
                                    </div>
                                </div>
                            </div>

                            <div className="mt-7 flex flex-col gap-3 sm:flex-row">
                                <Link
                                    href={returnHref}
                                    className="inline-flex items-center justify-center gap-2 rounded-lg bg-cortex-primary px-4 py-2.5 text-sm font-semibold text-cortex-bg transition-colors hover:bg-cortex-primary/90"
                                >
                                    <ArrowLeft className="h-4 w-4" aria-hidden="true" />
                                    {returnLabel}
                                </Link>
                                {supportHref ? (
                                    <Link
                                        href={supportHref}
                                        className="inline-flex items-center justify-center rounded-lg border border-cortex-border bg-cortex-bg px-4 py-2.5 text-sm font-semibold text-cortex-text-main transition-colors hover:border-cortex-primary/40 hover:text-cortex-primary"
                                    >
                                        {supportLabel}
                                    </Link>
                                ) : null}
                            </div>
                        </div>
                    </div>
                </div>
            </section>
        </main>
    );
}
