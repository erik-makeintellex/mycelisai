"use client";

import {
    Building2,
    CheckCircle2,
    Clock3,
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
    },
    {
        id: "scim",
        name: "SCIM",
        category: "Future provisioning",
        status: "planned",
        icon: Clock3,
        secretRefs: ["MYCELIS_AUTH_SCIM_BEARER_TOKEN_REF"],
        setupItems: ["Provisioning endpoint", "Bearer token reference", "User lifecycle policy", "Group sync rules"],
        notes: "Future scaffold for automated user and group provisioning after SSO is wired.",
    },
];

const statusStyles: Record<ProviderStatus, string> = {
    available: "border-cortex-success/30 bg-cortex-success/10 text-cortex-success",
    planned: "border-cortex-border bg-cortex-bg text-cortex-text-muted",
};

function StatusBadge({ status }: { status: ProviderStatus }) {
    const Icon = status === "available" ? CheckCircle2 : Clock3;
    const label = status === "available" ? "Available scaffold" : "Future integration";

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

function ProviderConceptCard({ concept }: { concept: AuthProviderConcept }) {
    const Icon = concept.icon;

    return (
        <article className="rounded-lg border border-cortex-border bg-cortex-surface/60 p-4">
            <div className="flex flex-wrap items-start justify-between gap-3">
                <div className="flex min-w-0 items-start gap-3">
                    <div className="rounded border border-cortex-border bg-cortex-bg p-2 text-cortex-primary">
                        <Icon className="h-4 w-4" aria-hidden="true" />
                    </div>
                    <div className="min-w-0">
                        <h3 className="text-sm font-semibold text-cortex-text-main">{concept.name}</h3>
                        <p className="mt-1 text-xs text-cortex-text-muted">{concept.category}</p>
                    </div>
                </div>
                <StatusBadge status={concept.status} />
            </div>

            <p className="mt-3 text-xs leading-5 text-cortex-text-muted">{concept.notes}</p>

            <div className="mt-4 grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(16rem,0.85fr)]">
                <div>
                    <h4 className="text-[10px] font-semibold uppercase text-cortex-text-muted">Setup concepts</h4>
                    <ul className="mt-2 flex flex-wrap gap-1.5">
                        {concept.setupItems.map((item) => (
                            <li
                                key={item}
                                className="rounded border border-cortex-border bg-cortex-bg px-2 py-1 text-[11px] text-cortex-text-muted"
                            >
                                {item}
                            </li>
                        ))}
                    </ul>
                </div>
                <div>
                    <h4 className="text-[10px] font-semibold uppercase text-cortex-text-muted">Secret references only</h4>
                    <div className="mt-2">
                        <SecretReferenceList refs={concept.secretRefs} />
                    </div>
                </div>
            </div>
        </article>
    );
}

export default function AuthProvidersPage() {
    return (
        <section className="space-y-5" aria-labelledby="auth-providers-heading">
            <div className="flex flex-col gap-3 border-b border-cortex-border pb-4 md:flex-row md:items-end md:justify-between">
                <div>
                    <p className="text-[10px] font-semibold uppercase text-cortex-text-muted">Settings scaffold</p>
                    <h2 id="auth-providers-heading" className="mt-1 text-xl font-semibold text-cortex-text-main">
                        Auth Providers
                    </h2>
                    <p className="mt-2 max-w-3xl text-sm leading-6 text-cortex-text-muted">
                        Read-only planning surface for enterprise identity setup. Provider credentials are represented as
                        secret manager references, not inline values.
                    </p>
                </div>
                <div className="rounded border border-cortex-primary/30 bg-cortex-primary/10 px-3 py-2 text-xs text-cortex-primary">
                    No provider changes are submitted from this scaffold.
                </div>
            </div>

            <div className="rounded-lg border border-amber-400/30 bg-amber-400/5 p-4 text-xs leading-5 text-amber-300">
                Store secrets in the deployment secret manager and expose only reference names here. Client secrets,
                private keys, certificates, and bearer tokens should never be entered as raw values.
            </div>

            <div className="grid gap-3">
                {AUTH_PROVIDER_CONCEPTS.map((concept) => (
                    <ProviderConceptCard key={concept.id} concept={concept} />
                ))}
            </div>
        </section>
    );
}
