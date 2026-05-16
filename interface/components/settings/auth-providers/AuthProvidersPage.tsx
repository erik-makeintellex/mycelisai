"use client";

import { useMemo, useState } from "react";
import {
    Building2,
    CheckCircle2,
    Clock3,
    ExternalLink,
    Github,
    Globe2,
    KeyRound,
    LockKeyhole,
    ShieldCheck,
    UsersRound,
} from "lucide-react";

type ProviderStatus = "available" | "planned";

interface AuthProviderConcept {
    id: string;
    name: string;
    category: string;
    status: ProviderStatus;
    icon: React.ComponentType<{ className?: string }>;
    secretRefs: string[];
    setupItems: string[];
    notes: string;
    posture: string;
}

const AUTH_PROVIDER_CONCEPTS: AuthProviderConcept[] = [
    {
        id: "local",
        name: "Local",
        category: "Built-in identity",
        status: "available",
        icon: LockKeyhole,
        secretRefs: [],
        setupItems: ["Username/password policy", "Recovery owner", "Admin bootstrap review"],
        notes: "Keeps self-hosted access available without an external identity provider.",
        posture: "Active release path",
    },
    {
        id: "oidc",
        name: "OIDC / OAuth",
        category: "Generic federation",
        status: "planned",
        icon: Globe2,
        secretRefs: ["MYCELIS_AUTH_OIDC_CLIENT_SECRET_REF"],
        setupItems: ["Issuer URL", "Client ID", "Redirect URI", "Scopes", "Group claim mapping"],
        notes: "Use a secret manager reference for client credentials; never paste client secrets into this UI.",
        posture: "Enterprise adapter path",
    },
    {
        id: "saml",
        name: "SAML",
        category: "Enterprise SSO",
        status: "planned",
        icon: ShieldCheck,
        secretRefs: ["MYCELIS_AUTH_SAML_CERT_REF", "MYCELIS_AUTH_SAML_PRIVATE_KEY_REF"],
        setupItems: ["Metadata URL", "Entity ID", "ACS URL", "NameID format", "Attribute mapping"],
        notes: "Certificate and signing key material should resolve through referenced secrets only.",
        posture: "Enterprise adapter path",
    },
    {
        id: "entra",
        name: "Entra ID",
        category: "Microsoft identity",
        status: "planned",
        icon: Building2,
        secretRefs: ["MYCELIS_AUTH_ENTRA_CLIENT_SECRET_REF"],
        setupItems: ["Tenant ID", "Application ID", "Redirect URI", "Directory group mapping"],
        notes: "Designed for Microsoft Entra federation with group-to-role mapping.",
        posture: "Enterprise adapter path",
    },
    {
        id: "google",
        name: "Google Workspace",
        category: "Workspace identity",
        status: "planned",
        icon: UsersRound,
        secretRefs: ["MYCELIS_AUTH_GOOGLE_CLIENT_SECRET_REF"],
        setupItems: ["Workspace domain", "Client ID", "Redirect URI", "Hosted domain policy"],
        notes: "Workspace domain restrictions should be explicit before activation.",
        posture: "Enterprise adapter path",
    },
    {
        id: "github",
        name: "GitHub",
        category: "Developer identity",
        status: "planned",
        icon: Github,
        secretRefs: ["MYCELIS_AUTH_GITHUB_CLIENT_SECRET_REF"],
        setupItems: ["Organization allowlist", "OAuth app ID", "Callback URL", "Team mapping"],
        notes: "Useful for engineering teams where GitHub org or team membership is the access source.",
        posture: "Enterprise adapter path",
    },
    {
        id: "scim",
        name: "SCIM",
        category: "Future provisioning",
        status: "planned",
        icon: Clock3,
        secretRefs: ["MYCELIS_AUTH_SCIM_BEARER_TOKEN_REF"],
        setupItems: ["Provisioning endpoint", "Bearer token reference", "User lifecycle policy", "Group sync rules"],
        notes: "Lifecycle provisioning target for automated user and group sync after SSO is wired.",
        posture: "Future provisioning path",
    },
];

const statusStyles: Record<ProviderStatus, string> = {
    available: "border-cortex-success/30 bg-cortex-success/10 text-cortex-success",
    planned: "border-cortex-border bg-cortex-bg text-cortex-text-muted",
};

function StatusBadge({ status }: { status: ProviderStatus }) {
    const Icon = status === "available" ? CheckCircle2 : Clock3;
    const label = status === "available" ? "Available" : "Future integration";

    return (
        <span className={`inline-flex items-center gap-1 rounded border px-2 py-0.5 text-[10px] uppercase ${statusStyles[status]}`}>
            <Icon className="h-3 w-3" aria-hidden="true" />
            {label}
        </span>
    );
}

function SecretReferenceList({ refs }: { refs: string[] }) {
    if (refs.length === 0) {
        return (
            <p className="rounded border border-cortex-border bg-cortex-bg px-3 py-2 text-xs text-cortex-text-muted">
                No external secret reference required.
            </p>
        );
    }

    return (
        <ul className="space-y-1.5" aria-label="Secret references">
            {refs.map((ref) => (
                <li
                    key={ref}
                    className="flex min-w-0 items-center gap-2 rounded border border-cortex-border bg-cortex-bg px-3 py-2 font-mono text-[11px] text-cortex-text-main"
                >
                    <KeyRound className="h-3.5 w-3.5 flex-shrink-0 text-cortex-primary" aria-hidden="true" />
                    <span className="break-all">{ref}</span>
                </li>
            ))}
        </ul>
    );
}

export default function AuthProvidersPage() {
    const [selectedProviderId, setSelectedProviderId] = useState(AUTH_PROVIDER_CONCEPTS[0].id);
    const selectedProvider = useMemo(
        () => AUTH_PROVIDER_CONCEPTS.find((concept) => concept.id === selectedProviderId) ?? AUTH_PROVIDER_CONCEPTS[0],
        [selectedProviderId],
    );
    const SelectedIcon = selectedProvider.icon;
    const availableCount = AUTH_PROVIDER_CONCEPTS.filter((concept) => concept.status === "available").length;
    const plannedCount = AUTH_PROVIDER_CONCEPTS.length - availableCount;

    return (
        <section className="space-y-5" aria-labelledby="auth-providers-heading">
            <div className="flex flex-col gap-3 border-b border-cortex-border pb-4 md:flex-row md:items-end md:justify-between">
                <div>
                    <p className="text-[10px] font-semibold uppercase text-cortex-text-muted">Identity configuration</p>
                    <h2 id="auth-providers-heading" className="mt-1 text-xl font-semibold text-cortex-text-main">
                        Auth Providers
                    </h2>
                    <p className="mt-2 max-w-3xl text-sm leading-6 text-cortex-text-muted">
                        Read-only configuration contract for local auth, enterprise SSO, and lifecycle provisioning.
                        Provider credentials are represented as secret manager references, not inline values.
                    </p>
                </div>
                <div className="grid grid-cols-2 gap-2 text-xs md:min-w-[14rem]">
                    <div className="rounded border border-cortex-success/30 bg-cortex-success/10 px-3 py-2 text-cortex-success">
                        <span className="block text-[10px] uppercase">Available</span>
                        <span className="text-base font-semibold">{availableCount}</span>
                    </div>
                    <div className="rounded border border-cortex-border bg-cortex-bg px-3 py-2 text-cortex-text-muted">
                        <span className="block text-[10px] uppercase">Planned</span>
                        <span className="text-base font-semibold">{plannedCount}</span>
                    </div>
                </div>
            </div>

            <div className="grid gap-4 lg:grid-cols-[18rem_minmax(0,1fr)]">
                <nav className="rounded-lg border border-cortex-border bg-cortex-surface/70 p-2" aria-label="Auth provider menu">
                    <div className="px-2 pb-2 pt-1 text-[10px] font-semibold uppercase text-cortex-text-muted">
                        Provider menu
                    </div>
                    <div className="space-y-1">
                        {AUTH_PROVIDER_CONCEPTS.map((concept) => {
                            const Icon = concept.icon;
                            const selected = concept.id === selectedProvider.id;

                            return (
                                <button
                                    key={concept.id}
                                    type="button"
                                    aria-current={selected ? "page" : undefined}
                                    onClick={() => setSelectedProviderId(concept.id)}
                                    className={`flex w-full items-center gap-3 rounded border px-3 py-2 text-left transition ${
                                        selected
                                            ? "border-cortex-primary/50 bg-cortex-primary/10 text-cortex-text-main"
                                            : "border-transparent text-cortex-text-muted hover:border-cortex-border hover:bg-cortex-bg"
                                    }`}
                                >
                                    <Icon className="h-4 w-4 flex-shrink-0 text-cortex-primary" aria-hidden="true" />
                                    <span className="min-w-0 flex-1">
                                        <span className="block truncate text-sm font-medium">{concept.name}</span>
                                        <span className="block truncate text-[11px]">{concept.category}</span>
                                    </span>
                                    <span
                                        className={`h-2 w-2 flex-shrink-0 rounded-full ${
                                            concept.status === "available" ? "bg-cortex-success" : "bg-cortex-text-muted"
                                        }`}
                                        aria-hidden="true"
                                    />
                                </button>
                            );
                        })}
                    </div>
                </nav>

                <article className="rounded-lg border border-cortex-border bg-cortex-surface/70 p-4" aria-live="polite">
                    <div className="flex flex-wrap items-start justify-between gap-3">
                        <div className="flex min-w-0 items-start gap-3">
                            <div className="rounded border border-cortex-border bg-cortex-bg p-2 text-cortex-primary">
                                <SelectedIcon className="h-5 w-5" aria-hidden="true" />
                            </div>
                            <div className="min-w-0">
                                <p className="text-[10px] font-semibold uppercase text-cortex-text-muted">
                                    {selectedProvider.category}
                                </p>
                                <h3 className="mt-1 text-lg font-semibold text-cortex-text-main">{selectedProvider.name}</h3>
                            </div>
                        </div>
                        <StatusBadge status={selectedProvider.status} />
                    </div>

                    <div className="mt-4 grid gap-3 md:grid-cols-2">
                        <div className="rounded border border-cortex-border bg-cortex-bg p-3">
                            <h4 className="text-[10px] font-semibold uppercase text-cortex-text-muted">Current posture</h4>
                            <p className="mt-2 text-sm text-cortex-text-main">{selectedProvider.posture}</p>
                        </div>
                        <div className="rounded border border-cortex-border bg-cortex-bg p-3">
                            <h4 className="text-[10px] font-semibold uppercase text-cortex-text-muted">Write boundary</h4>
                            <p className="mt-2 text-sm text-cortex-text-main">Deployment-owned</p>
                        </div>
                    </div>

                    <p className="mt-4 text-sm leading-6 text-cortex-text-muted">{selectedProvider.notes}</p>

                    <div className="mt-5 grid gap-4 xl:grid-cols-[minmax(0,1fr)_minmax(18rem,0.9fr)]">
                        <div>
                            <h4 className="text-[10px] font-semibold uppercase text-cortex-text-muted">Setup checklist</h4>
                            <ul className="mt-2 grid gap-2 sm:grid-cols-2">
                                {selectedProvider.setupItems.map((item) => (
                                    <li
                                        key={item}
                                        className="rounded border border-cortex-border bg-cortex-bg px-3 py-2 text-xs text-cortex-text-muted"
                                    >
                                        {item}
                                    </li>
                                ))}
                            </ul>
                        </div>
                        <div>
                            <h4 className="text-[10px] font-semibold uppercase text-cortex-text-muted">Secret references only</h4>
                            <div className="mt-2">
                                <SecretReferenceList refs={selectedProvider.secretRefs} />
                            </div>
                        </div>
                    </div>

                    <div className="mt-5 rounded border border-amber-400/30 bg-amber-400/5 p-3 text-xs leading-5 text-amber-300">
                        Store secrets in the deployment secret manager and expose only reference names here. Client
                        secrets, private keys, certificates, and bearer tokens should never be entered as raw values.
                    </div>

                    <a
                        href="/docs?doc=auth-modes"
                        className="mt-4 inline-flex items-center gap-2 text-xs font-medium text-cortex-primary underline"
                    >
                        Open authentication modes
                        <ExternalLink className="h-3.5 w-3.5" aria-hidden="true" />
                    </a>
                </article>
            </div>
        </section>
    );
}
