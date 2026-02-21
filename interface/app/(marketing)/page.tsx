import Link from "next/link";
import {
    ArrowRight,
    Terminal,
    Cpu,
    Shield,
    Zap,
    Layers,
    GitBranch,
    Activity,
    Lock,
    Users,
    Globe,
    Server,
    Brain,
    Network,
    Eye,
    Workflow,
} from "lucide-react";

export default function LandingPage() {
    return (
        <div className="min-h-screen bg-cortex-bg text-cortex-text-main font-sans selection:bg-cyan-500/30">

            {/* ── Navigation ── */}
            <nav className="fixed top-0 w-full z-50 border-b border-cortex-border bg-cortex-bg/80 backdrop-blur-md">
                <div className="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between">
                    <div className="flex items-center gap-2">
                        <div className="w-3 h-3 bg-cortex-primary rounded-full animate-pulse shadow-[0_0_10px_rgba(6,182,212,0.5)]" />
                        <span className="font-mono font-bold text-zinc-100 tracking-widest text-sm">MYCELIS</span>
                    </div>
                    <div className="hidden md:flex items-center gap-8 text-sm font-mono text-cortex-text-muted">
                        <Link href="#what" className="hover:text-cortex-primary transition-colors">WHAT</Link>
                        <Link href="#spectrum" className="hover:text-cortex-primary transition-colors">SPECTRUM</Link>
                        <Link href="#architecture" className="hover:text-cortex-primary transition-colors">ARCHITECTURE</Link>
                        <Link href="#governance" className="hover:text-cortex-primary transition-colors">GOVERNANCE</Link>
                    </div>
                    <div className="flex items-center gap-4">
                        <Link href="/dashboard" className="px-4 py-2 text-xs font-mono border border-cortex-border hover:border-cortex-primary/50 hover:text-cortex-primary hover:bg-cortex-primary/10 transition-all rounded text-cortex-text-muted font-medium">
                            LAUNCH CONSOLE
                        </Link>
                    </div>
                </div>
            </nav>

            {/* ── Hero Section ── */}
            <section className="relative pt-32 pb-20 md:pt-48 md:pb-32 overflow-hidden">
                {/* Background Grid */}
                <div className="absolute inset-0 bg-[linear-gradient(to_right,#80808012_1px,transparent_1px),linear-gradient(to_bottom,#80808012_1px,transparent_1px)] bg-[size:24px_24px]" />
                <div className="absolute top-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-cortex-primary/20 to-transparent" />

                <div className="max-w-7xl mx-auto px-6 relative z-10">
                    <div className="grid lg:grid-cols-2 gap-12 items-center">

                        {/* Text Content */}
                        <div className="space-y-8">
                            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-cortex-primary/20 bg-cyan-950/30 text-cortex-primary text-xs font-mono shadow-sm backdrop-blur-sm">
                                <span className="relative flex h-2 w-2">
                                    <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-cortex-primary opacity-75" />
                                    <span className="relative inline-flex rounded-full h-2 w-2 bg-cortex-primary" />
                                </span>
                                CORTEX V7.1 ONLINE — BRAIN PROVENANCE
                            </div>

                            <h1 className="text-5xl md:text-7xl font-bold tracking-tight text-zinc-100 leading-tight">
                                Your AI Agents. <br />
                                <span className="text-transparent bg-clip-text bg-gradient-to-r from-cyan-400 to-blue-500">
                                    Your Rules.
                                </span>
                            </h1>

                            <p className="text-xl text-cortex-text-muted max-w-lg leading-relaxed font-light">
                                Mycelis is a <strong className="text-cortex-text-main">Recursive Swarm Operating System</strong> that
                                lets you deploy, orchestrate, and govern AI agent teams — from a single localhost model
                                to a fleet of commercial LLMs — with <strong className="text-cortex-text-main">full brain provenance</strong>,
                                transparent routing, and zero vendor lock-in.
                            </p>

                            <div className="flex flex-col sm:flex-row gap-4 pt-4">
                                <Link href="/dashboard" className="group relative px-6 py-3 bg-cortex-primary text-cortex-bg font-semibold rounded hover:bg-cortex-primary/90 transition-all flex items-center justify-center gap-2 shadow-[0_0_20px_rgba(6,182,212,0.2)]">
                                    Deploy Cortex
                                    <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
                                </Link>
                                <Link href="https://github.com/erik-makeintellex/mycelisai" target="_blank" className="px-6 py-3 border border-cortex-border text-cortex-text-main rounded hover:bg-cortex-surface transition-all flex items-center justify-center gap-2 font-mono bg-cortex-surface/50 backdrop-blur-sm">
                                    <GitBranch className="w-4 h-4" />
                                    View Source
                                </Link>
                            </div>
                        </div>

                        {/* Visual: Agent Orchestration Terminal */}
                        <div className="relative">
                            <div className="absolute -inset-4 bg-gradient-to-r from-cyan-500/20 to-blue-600/20 rounded-2xl blur-xl opacity-50 animate-pulse" />
                            <div className="relative bg-cortex-bg rounded-xl border border-cortex-border p-1 font-mono text-xs md:text-sm shadow-2xl ring-1 ring-cortex-border">
                                <div className="bg-cortex-surface rounded-t-lg border-b border-cortex-border p-3 flex items-center gap-2">
                                    <div className="flex gap-1.5">
                                        <div className="w-2.5 h-2.5 rounded-full bg-red-500/50" />
                                        <div className="w-2.5 h-2.5 rounded-full bg-yellow-500/50" />
                                        <div className="w-2.5 h-2.5 rounded-full bg-green-500/50" />
                                    </div>
                                    <span className="text-cortex-text-muted ml-2 text-xs">soma@cortex:~/missions</span>
                                </div>

                                <div className="p-4 space-y-2 h-[300px] overflow-hidden text-cortex-text-muted bg-cortex-bg rounded-b-lg">
                                    <div className="flex gap-2">
                                        <span className="text-cortex-success font-bold">➜</span>
                                        <span className="text-cortex-text-main">myc blueprint &quot;Analyze quarterly sales data&quot;</span>
                                    </div>
                                    <div className="pl-4 border-l-2 border-cortex-border ml-1 space-y-1.5 py-1">
                                        <p className="text-cortex-primary">[SOMA] Routing to council-architect...</p>
                                        <p className="text-zinc-500">[BRAIN] Ollama (Local) · qwen2.5-coder:7b · data: local_only</p>
                                        <p className="text-cortex-text-muted">[AXON] Blueprint generated: 3 teams, 7 agents</p>
                                        <p className="text-cortex-text-muted">[MODE] ANSWER · Governance: PASSIVE · C:0.85</p>
                                        <p className="text-amber-400 bg-amber-500/10 inline-block px-1 rounded">[GOV] Trust 0.7 — file_write requires approval</p>
                                        <p className="text-cortex-text-main">[AGENT] data-analyst → MCP warehouse query</p>
                                        <p className="text-cortex-text-main">[AGENT] chart-renderer → Observable Plot</p>
                                        <p className="text-cortex-success font-bold">✔ Mission activated. 7 agents, all brains local.</p>
                                        <p className="animate-pulse w-2 h-4 bg-cortex-text-muted inline-block align-middle" />
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </section>

            {/* ── What Is Mycelis ── */}
            <section id="what" className="py-24 bg-cortex-surface/30 border-y border-cortex-border">
                <div className="max-w-7xl mx-auto px-6">
                    <div className="text-center mb-16">
                        <h2 className="text-3xl font-bold text-zinc-100 mb-4">What Is Mycelis?</h2>
                        <p className="text-cortex-text-muted max-w-3xl mx-auto text-lg leading-relaxed">
                            Mycelis is a <strong className="text-cortex-text-main">self-hosted agent operating system</strong> that
                            treats AI models as biological cells in a living organism. You define <em>what</em> needs to happen,
                            and the system decomposes your intent into teams of specialized agents that coordinate, reason,
                            use tools, and report back — all governed by policies you control.
                        </p>
                    </div>

                    <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6">
                        <ConceptCard
                            title="Intent → Blueprint → Teams"
                            description="Describe a goal in natural language. The Meta-Architect decomposes it into a mission blueprint with teams, roles, and data flows. You approve the plan before any agent moves."
                            icon={<Workflow className="w-5 h-5" />}
                        />
                        <ConceptCard
                            title="Agents With Real Tools"
                            description="Each agent runs a ReAct reasoning loop with access to 18+ built-in tools and unlimited MCP servers — file I/O, database queries, API calls, image generation, memory search, and more."
                            icon={<Terminal className="w-5 h-5" />}
                        />
                        <ConceptCard
                            title="Brain Provenance"
                            description="Every response shows which AI model powered it, where it ran (local or remote), and its data boundary. No silent provider switching — you always know which brain is thinking."
                            icon={<Brain className="w-5 h-5" />}
                        />
                        <ConceptCard
                            title="Human-in-the-Loop by Default"
                            description="The Governance Valve intercepts every actuation. A trust score (0.0–1.0) determines what auto-executes and what requires your explicit approval. Nothing happens without your consent."
                            icon={<Shield className="w-5 h-5" />}
                        />
                    </div>
                </div>
            </section>

            {/* ── The Spectrum ── */}
            <section id="spectrum" className="py-24 bg-cortex-bg">
                <div className="max-w-7xl mx-auto px-6">
                    <div className="text-center mb-16">
                        <h2 className="text-3xl font-bold text-zinc-100 mb-4">From Localhost to Enterprise</h2>
                        <p className="text-cortex-text-muted max-w-3xl mx-auto text-lg">
                            One platform, every scale. Mycelis runs the same architecture whether you&apos;re
                            experimenting on a laptop or deploying across a production cluster.
                        </p>
                    </div>

                    <div className="grid md:grid-cols-3 gap-8">
                        {/* Local / Personal */}
                        <div className="group p-8 bg-cortex-surface border border-cortex-border rounded-xl hover:border-cortex-primary/30 transition-all">
                            <div className="flex items-center gap-3 mb-6">
                                <div className="p-2.5 rounded-lg bg-cortex-primary/10 text-cortex-primary border border-cortex-primary/20">
                                    <Lock className="w-5 h-5" />
                                </div>
                                <div>
                                    <h3 className="text-lg font-bold text-zinc-100">Local &amp; Air-Gapped</h3>
                                    <span className="text-xs font-mono text-cortex-primary">PERSONAL</span>
                                </div>
                            </div>
                            <ul className="space-y-3 text-sm text-cortex-text-muted">
                                <li className="flex gap-2"><span className="text-cortex-primary mt-0.5">▸</span>Run entirely on localhost with Ollama, LM Studio, or vLLM</li>
                                <li className="flex gap-2"><span className="text-cortex-primary mt-0.5">▸</span>Zero cloud dependency — your data never leaves your machine</li>
                                <li className="flex gap-2"><span className="text-cortex-primary mt-0.5">▸</span>Personal automation: email triage, code review, research synthesis</li>
                                <li className="flex gap-2"><span className="text-cortex-primary mt-0.5">▸</span>Ideal for security-conscious developers and researchers</li>
                            </ul>
                            <div className="mt-6 pt-4 border-t border-cortex-border">
                                <span className="text-xs font-mono text-cortex-text-muted">Default model: qwen2.5-coder:7b via Ollama</span>
                            </div>
                        </div>

                        {/* Team / Hybrid */}
                        <div className="group p-8 bg-cortex-surface border border-cortex-border rounded-xl hover:border-cortex-primary/30 transition-all relative">
                            <div className="absolute -top-3 left-1/2 -translate-x-1/2 px-3 py-0.5 bg-cortex-primary text-cortex-bg text-[10px] font-mono font-bold rounded-full uppercase tracking-wider">
                                Most Common
                            </div>
                            <div className="flex items-center gap-3 mb-6">
                                <div className="p-2.5 rounded-lg bg-cortex-primary/10 text-cortex-primary border border-cortex-primary/20">
                                    <Users className="w-5 h-5" />
                                </div>
                                <div>
                                    <h3 className="text-lg font-bold text-zinc-100">Hybrid &amp; Team</h3>
                                    <span className="text-xs font-mono text-cortex-primary">PROFESSIONAL</span>
                                </div>
                            </div>
                            <ul className="space-y-3 text-sm text-cortex-text-muted">
                                <li className="flex gap-2"><span className="text-cortex-primary mt-0.5">▸</span>Mix local models with cloud APIs (OpenAI, Anthropic, Gemini)</li>
                                <li className="flex gap-2"><span className="text-cortex-primary mt-0.5">▸</span>Cognitive Matrix: route tasks by cost, capability, and data boundary</li>
                                <li className="flex gap-2"><span className="text-cortex-primary mt-0.5">▸</span>Standing teams: architect, coder, creative, sentry — always ready</li>
                                <li className="flex gap-2"><span className="text-cortex-primary mt-0.5">▸</span>MCP tool servers extend agents with any API or service</li>
                            </ul>
                            <div className="mt-6 pt-4 border-t border-cortex-border">
                                <span className="text-xs font-mono text-cortex-text-muted">6 LLM providers, brain provenance, policy-gated routing</span>
                            </div>
                        </div>

                        {/* Enterprise / Production */}
                        <div className="group p-8 bg-cortex-surface border border-cortex-border rounded-xl hover:border-cortex-primary/30 transition-all">
                            <div className="flex items-center gap-3 mb-6">
                                <div className="p-2.5 rounded-lg bg-cortex-primary/10 text-cortex-primary border border-cortex-primary/20">
                                    <Globe className="w-5 h-5" />
                                </div>
                                <div>
                                    <h3 className="text-lg font-bold text-zinc-100">Enterprise &amp; Production</h3>
                                    <span className="text-xs font-mono text-cortex-primary">COMMERCIAL</span>
                                </div>
                            </div>
                            <ul className="space-y-3 text-sm text-cortex-text-muted">
                                <li className="flex gap-2"><span className="text-cortex-primary mt-0.5">▸</span>Kubernetes-native with Helm charts and Kind dev clusters</li>
                                <li className="flex gap-2"><span className="text-cortex-primary mt-0.5">▸</span>NATS JetStream for durable, high-throughput agent messaging</li>
                                <li className="flex gap-2"><span className="text-cortex-primary mt-0.5">▸</span>Postgres + pgvector for relational state and semantic memory</li>
                                <li className="flex gap-2"><span className="text-cortex-primary mt-0.5">▸</span>Brain provenance audit trails, RBAC, and policy-gated routing</li>
                            </ul>
                            <div className="mt-6 pt-4 border-t border-cortex-border">
                                <span className="text-xs font-mono text-cortex-text-muted">50+ API endpoints, 30+ NATS topics, 21 DB migrations</span>
                            </div>
                        </div>
                    </div>
                </div>
            </section>

            {/* ── Architecture ── */}
            <section id="architecture" className="py-24 bg-cortex-surface/30 border-y border-cortex-border">
                <div className="max-w-7xl mx-auto px-6">
                    <div className="text-center mb-16">
                        <h2 className="text-3xl font-bold text-zinc-100 mb-4">The Cortex Architecture</h2>
                        <p className="text-cortex-text-muted max-w-2xl mx-auto text-lg">
                            Four layers of synthetic biology — from infrastructure to the visual cortex.
                        </p>
                    </div>

                    <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6">
                        <ArchLayer
                            layer="L0"
                            name="Sovereign Base"
                            icon={<Server className="w-5 h-5" />}
                            items={["Go 1.26 static binary", "Postgres 16 + pgvector", "Kubernetes (Kind)", "Cognitive Registry"]}
                        />
                        <ArchLayer
                            layer="L1"
                            name="Nervous System"
                            icon={<Activity className="w-5 h-5" />}
                            items={["NATS JetStream 2.12", "30+ pub/sub topics", "Council request-reply", "Sensor ingress hierarchy"]}
                        />
                        <ArchLayer
                            layer="L2"
                            name="Fractal Fabric"
                            icon={<Brain className="w-5 h-5" />}
                            items={["Soma → Axon → Teams", "18 built-in agent tools", "Brain provenance pipeline", "Overseer DAG engine"]}
                        />
                        <ArchLayer
                            layer="L3"
                            name="Conscious Face"
                            icon={<Eye className="w-5 h-5" />}
                            items={["Next.js 16 + React 19", "ReactFlow circuit board", "Mode Ribbon + Inspector", "Midnight Cortex theme"]}
                        />
                    </div>
                </div>
            </section>

            {/* ── Governance ── */}
            <section id="governance" className="py-24 bg-cortex-bg">
                <div className="max-w-7xl mx-auto px-6">
                    <div className="grid md:grid-cols-2 gap-16 items-center">
                        <div>
                            <h2 className="text-3xl font-bold text-zinc-100 mb-6">Zero-Trust Agent Governance</h2>
                            <p className="text-cortex-text-muted mb-8 leading-relaxed text-lg">
                                Every agent action passes through the Governance Valve. You set the trust threshold,
                                define policy rules by intent pattern, and approve or deny actuations in real-time.
                                The system defaults to <strong className="text-cortex-text-main">deny</strong> — agents must earn autonomy.
                            </p>

                            <div className="space-y-4">
                                <GovernanceItem
                                    title="Trust Economy"
                                    description="Each CTS envelope carries a trust score (0.0–1.0). Actions above your threshold auto-execute. Below it, they queue for human review."
                                />
                                <GovernanceItem
                                    title="Brain Governance"
                                    description="Every provider has a location (local/remote), data boundary (local_only/leaves_org), and usage policy. Remote providers require explicit enablement — no silent escalation."
                                />
                                <GovernanceItem
                                    title="Policy-as-Code"
                                    description="YAML-defined rules per intent pattern — file.write, api.call, deploy.* — with ALLOW, DENY, or REQUIRE_APPROVAL actions."
                                />
                                <GovernanceItem
                                    title="Approval Queue"
                                    description="Real-time governance modal intercepts dangerous operations. Approve or deny with full context: who, what, why, and risk level."
                                />
                            </div>
                        </div>

                        {/* Visual: Trust Spectrum */}
                        <div className="bg-cortex-surface border border-cortex-border rounded-xl p-8 shadow-2xl">
                            <div className="space-y-6 font-mono text-sm">
                                <div className="text-cortex-text-muted text-xs uppercase tracking-wider mb-4">Trust Spectrum</div>
                                <TrustLevel score="0.0" label="Full Lockdown" desc="Every action requires approval" color="bg-red-500" />
                                <TrustLevel score="0.3" label="Cautious" desc="Read ops auto, writes blocked" color="bg-amber-500" />
                                <TrustLevel score="0.7" label="Balanced" desc="Standard ops auto, deploys blocked" color="bg-cortex-primary" />
                                <TrustLevel score="1.0" label="Full Autonomy" desc="All actions auto-execute" color="bg-cortex-success" />
                            </div>
                        </div>
                    </div>
                </div>
            </section>

            {/* ── Use Cases ── */}
            <section className="py-24 bg-cortex-surface/30 border-y border-cortex-border">
                <div className="max-w-7xl mx-auto px-6">
                    <div className="text-center mb-16">
                        <h2 className="text-3xl font-bold text-zinc-100 mb-4">What People Build</h2>
                        <p className="text-cortex-text-muted max-w-2xl mx-auto text-lg">
                            From solo developers to multi-team organizations.
                        </p>
                    </div>

                    <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
                        <UseCaseCard title="Code Review Pipelines" desc="Architect analyzes PRs, Coder suggests fixes, Sentry checks for vulnerabilities — all coordinated through NATS." tag="TEAM" />
                        <UseCaseCard title="Research Synthesis" desc="Spawn a research team that queries databases, reads papers, and produces structured reports with citations." tag="PERSONAL" />
                        <UseCaseCard title="Data Analysis Workflows" desc="Sensor agents ingest live data, analysts process it, chart renderers produce Observable Plot visualizations." tag="HYBRID" />
                        <UseCaseCard title="Automated Monitoring" desc="Sensor agents poll APIs on schedule, publish to NATS, trigger alerts when thresholds are breached." tag="ENTERPRISE" />
                        <UseCaseCard title="Content Generation" desc="Creative agents draft content, architect agents structure it, human approves the final output." tag="TEAM" />
                        <UseCaseCard title="Custom Tool Chains" desc="Connect any API via MCP servers. Agents discover and invoke tools dynamically at runtime." tag="ANY" />
                    </div>
                </div>
            </section>

            {/* ── CTA ── */}
            <section className="py-24 bg-gradient-to-b from-cortex-surface to-cortex-bg relative overflow-hidden border-t border-cortex-border">
                <div className="max-w-4xl mx-auto px-6 text-center relative z-10">
                    <h2 className="text-4xl font-bold mb-6 text-zinc-100">Deploy Your Own Nervous System</h2>
                    <p className="text-cortex-text-muted mb-10 text-lg max-w-2xl mx-auto">
                        Self-hosted. Open source. Air-gapped or cloud-connected.
                        Every response shows which brain powered it. Every routing decision is transparent.
                    </p>
                    <div className="flex justify-center gap-4">
                        <Link href="/dashboard" className="px-8 py-4 bg-cortex-primary text-cortex-bg font-bold rounded hover:bg-cortex-primary/90 transition-all shadow-xl shadow-cyan-900/20 ring-1 ring-cortex-primary/50">
                            Launch Console
                        </Link>
                        <Link href="https://github.com/erik-makeintellex/mycelisai" target="_blank" className="px-8 py-4 border border-cortex-border text-cortex-text-main font-bold rounded hover:bg-cortex-surface transition-all">
                            Read the Docs
                        </Link>
                    </div>
                </div>
            </section>

            {/* ── Footer ── */}
            <footer className="py-12 bg-cortex-bg text-cortex-text-muted border-t border-cortex-border">
                <div className="max-w-7xl mx-auto px-6 flex flex-col md:flex-row justify-between items-center gap-4">
                    <div className="text-sm">
                        &copy; 2026 Mycelis AI. All systems nominal.
                    </div>
                    <div className="flex gap-6 text-sm">
                        <Link href="https://github.com/erik-makeintellex/mycelisai#readme" target="_blank" className="hover:text-cortex-text-main transition-colors">Documentation</Link>
                        <Link href="/dashboard" className="hover:text-cortex-text-main transition-colors">Dashboard</Link>
                        <Link href="/telemetry" className="hover:text-cortex-text-main transition-colors">Status</Link>
                    </div>
                </div>
            </footer>
        </div>
    );
}

// ── Sub-Components ────────────────────────────────────────────

function ConceptCard({ title, description, icon }: { title: string; description: string; icon: React.ReactNode }) {
    return (
        <div className="p-6 bg-cortex-surface border border-cortex-border rounded-xl hover:border-cortex-primary/30 transition-all">
            <div className="mb-4 p-2.5 rounded-lg bg-cortex-primary/10 text-cortex-primary inline-block border border-cortex-primary/20">
                {icon}
            </div>
            <h3 className="text-lg font-bold text-zinc-100 mb-2">{title}</h3>
            <p className="text-sm text-cortex-text-muted leading-relaxed">{description}</p>
        </div>
    );
}

function ArchLayer({ layer, name, icon, items }: { layer: string; name: string; icon: React.ReactNode; items: string[] }) {
    return (
        <div className="p-6 bg-cortex-bg border border-cortex-border rounded-xl hover:border-cortex-primary/20 transition-all group">
            <div className="flex items-center gap-3 mb-4">
                <span className="text-xs font-mono font-bold text-cortex-primary bg-cortex-primary/10 px-2 py-0.5 rounded border border-cortex-primary/20">{layer}</span>
                <div className="text-cortex-text-muted group-hover:text-cortex-primary transition-colors">{icon}</div>
            </div>
            <h3 className="text-base font-bold text-zinc-100 mb-3">{name}</h3>
            <ul className="space-y-1.5">
                {items.map((item) => (
                    <li key={item} className="text-xs text-cortex-text-muted flex gap-2">
                        <span className="text-cortex-border mt-0.5">•</span>{item}
                    </li>
                ))}
            </ul>
        </div>
    );
}

function GovernanceItem({ title, description }: { title: string; description: string }) {
    return (
        <div className="flex gap-4 p-4 rounded-lg bg-cortex-surface border border-cortex-border">
            <div className="w-1.5 h-auto bg-cortex-primary rounded-full flex-shrink-0 opacity-80" />
            <div>
                <h3 className="font-bold text-zinc-200 mb-1">{title}</h3>
                <p className="text-sm text-cortex-text-muted">{description}</p>
            </div>
        </div>
    );
}

function TrustLevel({ score, label, desc, color }: { score: string; label: string; desc: string; color: string }) {
    return (
        <div className="flex items-center gap-4">
            <span className="text-cortex-text-muted w-8 text-right">{score}</span>
            <div className={`w-3 h-3 rounded-full ${color}`} />
            <div>
                <span className="text-cortex-text-main font-bold">{label}</span>
                <span className="text-cortex-text-muted ml-2 text-xs">— {desc}</span>
            </div>
        </div>
    );
}

function UseCaseCard({ title, desc, tag }: { title: string; desc: string; tag: string }) {
    return (
        <div className="p-6 bg-cortex-bg border border-cortex-border rounded-xl hover:border-cortex-primary/20 transition-all">
            <div className="flex items-center justify-between mb-3">
                <h3 className="font-bold text-zinc-100">{title}</h3>
                <span className="text-[9px] font-mono font-bold text-cortex-primary bg-cortex-primary/10 px-2 py-0.5 rounded-full border border-cortex-primary/20 uppercase tracking-wider">{tag}</span>
            </div>
            <p className="text-sm text-cortex-text-muted leading-relaxed">{desc}</p>
        </div>
    );
}
