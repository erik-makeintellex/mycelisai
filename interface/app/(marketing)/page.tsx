import Link from "next/link";
import { ArrowRight, Terminal, Cpu, Shield, Zap, Layers, GitBranch, Activity } from "lucide-react";

export default function LandingPage() {
    return (
        <div className="min-h-screen bg-zinc-950 text-zinc-300 font-sans selection:bg-cyan-500/30">

            {/* ── Navigation ── */}
            <nav className="fixed top-0 w-full z-50 border-b border-white/5 bg-zinc-950/80 backdrop-blur-md">
                <div className="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between">
                    <div className="flex items-center gap-2">
                        <div className="w-3 h-3 bg-cyan-500 rounded-full animate-pulse shadow-[0_0_10px_rgba(6,182,212,0.5)]" />
                        <span className="font-mono font-bold text-zinc-100 tracking-widest text-sm">MYCELIS</span>
                    </div>
                    <div className="hidden md:flex items-center gap-8 text-sm font-mono text-zinc-400">
                        <Link href="#features" className="hover:text-cyan-400 transition-colors">ARCHITECTURE</Link>
                        <Link href="#workflow" className="hover:text-cyan-400 transition-colors">WORKFLOW</Link>
                        <Link href="https://github.com/erik-makeintellex/mycelisai" target="_blank" className="hover:text-cyan-400 transition-colors">GITHUB</Link>
                    </div>
                    <div className="flex items-center gap-4">
                        <Link href="/dashboard" className="px-4 py-2 text-xs font-mono border border-white/10 hover:border-cyan-500/50 hover:text-cyan-400 hover:bg-cyan-500/10 transition-all rounded text-zinc-400 font-medium">
                            LAUNCH CONSOLE
                        </Link>
                    </div>
                </div>
            </nav>

            {/* ── Hero Section ── */}
            <section className="relative pt-32 pb-20 md:pt-48 md:pb-32 overflow-hidden">
                {/* Background Grid */}
                <div className="absolute inset-0 bg-[linear-gradient(to_right,#80808012_1px,transparent_1px),linear-gradient(to_bottom,#80808012_1px,transparent_1px)] bg-[size:24px_24px]"></div>
                <div className="absolute top-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-cyan-500/20 to-transparent" />

                <div className="max-w-7xl mx-auto px-6 relative z-10">
                    <div className="grid lg:grid-cols-2 gap-12 items-center">

                        {/* Text Content */}
                        <div className="space-y-8">
                            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-cyan-500/20 bg-cyan-950/30 text-cyan-400 text-xs font-mono shadow-sm backdrop-blur-sm">
                                <span className="relative flex h-2 w-2">
                                    <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-cyan-400 opacity-75"></span>
                                    <span className="relative inline-flex rounded-full h-2 w-2 bg-cyan-500"></span>
                                </span>
                                CORTEX V6.2 ONLINE
                            </div>

                            <h1 className="text-5xl md:text-7xl font-bold tracking-tight text-white leading-tight">
                                Orchestrate <br />
                                <span className="text-transparent bg-clip-text bg-gradient-to-r from-cyan-400 to-blue-500">
                                    Intelligence.
                                </span>
                            </h1>

                            <p className="text-xl text-zinc-400 max-w-lg leading-relaxed font-light">
                                The central nervous system for autonomous agentic workforces.
                                Secure, scalable, and self-organizing cognitive architecture.
                            </p>

                            <div className="flex flex-col sm:flex-row gap-4 pt-4">
                                <Link href="/dashboard" className="group relative px-6 py-3 bg-zinc-100 text-zinc-950 font-semibold rounded hover:bg-white transition-all flex items-center justify-center gap-2 shadow-[0_0_20px_rgba(255,255,255,0.1)]">
                                    Deploy Cortex
                                    <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
                                </Link>
                                <Link href="https://github.com/erik-makeintellex/mycelisai" target="_blank" className="px-6 py-3 border border-white/10 text-zinc-300 rounded hover:bg-white/5 transition-all flex items-center justify-center gap-2 font-mono bg-white/5 backdrop-blur-sm">
                                    <GitBranch className="w-4 h-4" />
                                    View Source
                                </Link>
                            </div>
                        </div>

                        {/* Visual: Simulated Terminal/Graph */}
                        <div className="relative">
                            <div className="absolute -inset-4 bg-gradient-to-r from-cyan-500/20 to-blue-600/20 rounded-2xl blur-xl opacity-50 animate-pulse"></div>
                            <div className="relative bg-[#0F0F12] rounded-xl border border-white/10 p-1 font-mono text-xs md:text-sm shadow-2xl ring-1 ring-white/5">
                                <div className="bg-white/5 rounded-t-lg border-b border-white/5 p-3 flex items-center gap-2">
                                    <div className="flex gap-1.5">
                                        <div className="w-2.5 h-2.5 rounded-full bg-red-500/50" />
                                        <div className="w-2.5 h-2.5 rounded-full bg-yellow-500/50" />
                                        <div className="w-2.5 h-2.5 rounded-full bg-green-500/50" />
                                    </div>
                                    <span className="text-zinc-500 ml-2 text-xs">archivist@cortex:~/missions</span>
                                </div>

                                <div className="p-4 space-y-2 h-[300px] overflow-hidden text-zinc-400 bg-[#0a0a0c] rounded-b-lg">
                                    <div className="flex gap-2">
                                        <span className="text-emerald-500 font-bold">➜</span>
                                        <span className="text-zinc-300">./spawn_agent --role="architect" --mode="planning"</span>
                                    </div>
                                    <div className="pl-4 border-l-2 border-white/5 ml-1 space-y-1.5 py-1">
                                        <p className="text-cyan-400">[INFO] Initializing Cognitive Matrix...</p>
                                        <p className="text-zinc-500">[DEBUG] Loaded 4 active models (OpenAI, Anthropic, Ollama)</p>
                                        <p className="text-zinc-500">[INFO] NATS Connection established (ws://localhost:4222)</p>
                                        <p className="text-amber-400 bg-amber-500/10 inline-block px-1 rounded">[WARN] Governance valve active. Approval required.</p>
                                        <p className="text-zinc-300">[INFO] Agent 'Arch-01' online. Listening on swarm.events.architect</p>
                                        <p className="text-blue-400">[THOUGHT] Analyzing project structure...</p>
                                        <p className="text-blue-400">[THOUGHT] Generating implementation plan...</p>
                                        <p className="text-emerald-400 font-bold">✔ Plan 'refactor_core' created (artifacts/plan.md)</p>
                                        <p className="animate-pulse w-2 h-4 bg-zinc-500 inline-block align-middle"></p>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </section>

            {/* ── Features Grid ── */}
            <section id="features" className="py-24 bg-zinc-900/30 border-y border-white/5">
                <div className="max-w-7xl mx-auto px-6">
                    <div className="text-center mb-16">
                        <h2 className="text-3xl font-bold text-white mb-4">The Cortex Architecture</h2>
                        <p className="text-zinc-400 max-w-2xl mx-auto text-lg">
                            Built on first principles of distributed systems engineering and cognitive science.
                        </p>
                    </div>

                    <div className="grid md:grid-cols-3 gap-8">
                        <FeatureCard
                            icon={<Cpu className="w-6 h-6 text-cyan-400" />}
                            title="Cognitive Matrix"
                            description="Plug-and-play LLM routing. Switch between OpenAI, Ollama, and Anthropic in real-time without code changes."
                        />
                        <FeatureCard
                            icon={<Zap className="w-6 h-6 text-amber-400" />}
                            title="Mycelial Network"
                            description="NATS-based nervous system. Micro-services signal propagation with < 10ms latency across the swarm."
                        />
                        <FeatureCard
                            icon={<Shield className="w-6 h-6 text-red-400" />}
                            title="Governance Valve"
                            description="Human-in-the-loop rejection. Agents propose, you ratify. Nothing happens without explicit trust."
                        />
                        <FeatureCard
                            icon={<Layers className="w-6 h-6 text-purple-400" />}
                            title="Memory Stream"
                            description="Vector-embedded semantic recall. Agents remember context, decisions, and outcomes across sessions."
                        />
                        <FeatureCard
                            icon={<Activity className="w-6 h-6 text-emerald-400" />}
                            title="Deep Telemetry"
                            description="Real-time observability into agent thought processes, tool usage, and resource consumption."
                        />
                        <FeatureCard
                            icon={<Terminal className="w-6 h-6 text-zinc-400" />}
                            title="CLI First"
                            description="Full control via the 'inv' ops toolkit. Manage infrastructure, deployments, and testing from the terminal."
                        />
                    </div>
                </div>
            </section>

            {/* ── Workflow / CI/CD ── */}
            <section id="workflow" className="py-24 bg-zinc-950">
                <div className="max-w-7xl mx-auto px-6">
                    <div className="grid md:grid-cols-2 gap-16 items-center">
                        <div>
                            <h2 className="text-3xl font-bold text-white mb-6">Built for Production</h2>
                            <p className="text-zinc-400 mb-8 leading-relaxed text-lg">
                                Mycelis employs a rigorous CI/CD pipeline ensuring stability from local development to production deployment.
                            </p>

                            <div className="space-y-6">
                                <WorkflowItem
                                    title="Development Build"
                                    status="Active"
                                    color="bg-blue-500"
                                    description="Triggers on every push to 'develop'. Runs unit tests, builds debug binaries, and pushes ':dev' Docker images."
                                    link="https://github.com/erik-makeintellex/mycelisai/actions/workflows/core-ci.yaml"
                                />
                                <WorkflowItem
                                    title="Production Release"
                                    status="Stable"
                                    color="bg-green-500"
                                    description="Triggers on 'v*' tags. Builds optimized artifacts, generates changelogs, and pushes ':latest' images to GHCR."
                                    link="https://github.com/erik-makeintellex/mycelisai/actions/workflows/release.yaml"
                                />
                            </div>
                        </div>

                        {/* Visual representation of pipeline */}
                        <div className="bg-[#0a0a0c] border border-white/10 rounded-xl p-8 relative overflow-hidden shadow-2xl">
                            <div className="absolute top-0 right-0 p-6 opacity-10">
                                <GitBranch className="w-40 h-40 text-white" />
                            </div>
                            <div className="space-y-6 relative z-10">
                                <div className="flex items-center gap-4 text-sm font-mono text-zinc-500">
                                    <span className="w-16 font-bold text-zinc-600">commit</span>
                                    <span className="text-cyan-400 bg-cyan-950/30 px-2 py-0.5 rounded font-bold border border-cyan-500/20">7f2a9c1</span>
                                    <span className="text-zinc-400">feat: implement neural bridge</span>
                                </div>
                                <div className="h-8 w-px bg-white/5 ml-3"></div>
                                <div className="flex items-center gap-4 text-sm font-mono text-zinc-500">
                                    <span className="w-16 font-bold text-zinc-600">build</span>
                                    <span className="text-amber-400 bg-amber-950/30 px-2 py-0.5 rounded font-bold flex items-center gap-1 border border-amber-500/20">
                                        <span className="animate-spin text-amber-500">⟳</span> running...
                                    </span>
                                    <span className="text-zinc-400">core-ci // test-and-lint</span>
                                </div>
                                <div className="h-8 w-px bg-white/5 ml-3"></div>
                                <div className="flex items-center gap-4 text-sm font-mono text-zinc-500">
                                    <span className="w-16 font-bold text-zinc-600">deploy</span>
                                    <span className="text-zinc-500 bg-white/5 px-2 py-0.5 rounded border border-white/5">pending</span>
                                    <span className="text-zinc-400">release // build-and-push</span>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </section>

            {/* ── CTA ── */}
            <section className="py-24 bg-gradient-to-b from-zinc-900 to-black relative overflow-hidden border-t border-white/5">
                <div className="absolute inset-0 bg-[url('/grid.svg')] opacity-5"></div>
                <div className="max-w-4xl mx-auto px-6 text-center relative z-10">
                    <h2 className="text-4xl font-bold mb-6 text-white">Ready to upgrade your cognition?</h2>
                    <p className="text-zinc-400 mb-10 text-lg max-w-2xl mx-auto">
                        Join the hive mind. Deploy your own Mycelis node today and start building the future of autonomous work.
                    </p>
                    <div className="flex justify-center gap-4">
                        <Link href="/dashboard" className="px-8 py-4 bg-cyan-600 text-white font-bold rounded hover:bg-cyan-500 transition-all shadow-xl shadow-cyan-900/20 ring-1 ring-cyan-500/50">
                            Launch Interface
                        </Link>
                    </div>
                </div>
            </section>

            {/* ── Footer ── */}
            <footer className="py-12 bg-black text-zinc-600 border-t border-white/5">
                <div className="max-w-7xl mx-auto px-6 flex flex-col md:flex-row justify-between items-center gap-4">
                    <div className="text-sm">
                        © 2026 Mycelis AI. All systems nominal.
                    </div>
                    <div className="flex gap-6 text-sm">
                        <Link href="https://github.com/erik-makeintellex/mycelisai#readme" target="_blank" className="hover:text-zinc-300 transition-colors">Documentation</Link>
                        <Link href="#" className="hover:text-zinc-300 transition-colors">API Reference</Link>
                        <Link href="#" className="hover:text-zinc-300 transition-colors">Status</Link>
                    </div>
                </div>
            </footer>
        </div>
    );
}

function FeatureCard({ icon, title, description }: { icon: React.ReactNode, title: string, description: string }) {
    return (
        <div className="p-8 bg-[#0F0F12] border border-white/5 rounded-xl hover:border-cyan-500/30 hover:shadow-[0_0_30px_rgba(6,182,212,0.05)] transition-all group">
            <div className="mb-5 p-3 bg-zinc-900/50 rounded-lg inline-block border border-white/5 group-hover:bg-cyan-950/30 group-hover:border-cyan-500/20 transition-colors">
                {icon}
            </div>
            <h3 className="text-xl font-bold text-white mb-3 group-hover:text-cyan-400 transition-colors">{title}</h3>
            <p className="text-zinc-500 leading-relaxed font-light">
                {description}
            </p>
        </div>
    );
}

function WorkflowItem({ title, status, color, description, link }: { title: string, status: string, color: string, description: string, link: string }) {
    return (
        <div className="flex gap-4 p-5 rounded-lg bg-[#0F0F12] border border-white/5 hover:border-cyan-500/20 hover:shadow-lg transition-all">
            <div className={`w-1.5 h-full ${color} rounded-full flex-shrink-0 opacity-80`} />
            <div>
                <div className="flex items-center gap-3 mb-1">
                    <h3 className="font-bold text-zinc-200">{title}</h3>
                    <span className="text-[10px] uppercase tracking-wider px-2 py-0.5 rounded-full bg-white/5 text-zinc-500 border border-white/5 font-semibold">
                        {status}
                    </span>
                </div>
                <p className="text-sm text-zinc-500 mb-3">{description}</p>
                <Link href={link} className="text-xs text-cyan-500 hover:text-cyan-400 flex items-center gap-1 font-medium">
                    View Workflow <ArrowRight className="w-3 h-3" />
                </Link>
            </div>
        </div>
    );
}
