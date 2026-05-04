"use client";
import Link from "next/link";
import { ArrowRight, Blocks, Bot, BrainCircuit, Building2, Layers3, Loader2, Sparkles } from "lucide-react";
import { rememberLastOrganization } from "@/lib/lastOrganization";
import { ActionableState, ActionButton, EmptyState, HiddenLater, LoadingState, Metric, StartModeCard, WhyStartHere } from "@/components/organizations/createOrganizationEntryUi";
import { useCreateOrganizationEntryState } from "@/components/organizations/useCreateOrganizationEntryState";
import { CreateOrganizationRecentList } from "@/components/organizations/createOrganizationRecentList";
export default function CreateOrganizationEntry() {
    const state = useCreateOrganizationEntryState();
    const {
        templates, lastOrganization, templatesState, organizationsState, templatesError, organizationsError, selectedMode, selectedTemplateId, selectedTemplate, name, purpose, submitError, isSubmitting, openingOrganizationId,
        showDiagnosticOrganizations, diagnosticOrganizations, visibleDiagnosticOrganizations, hiddenDiagnosticCount, testingSetupCountLabel, visibleOrganizations, isTemplateMode, canSubmit, creationReadinessMessage,
        createSectionRef, nameInputRef, setSelectedTemplateId, setSelectedMode, setName, setPurpose, setShowDiagnosticOrganizations, focusCreateFlow, rememberOrganization, openOrganization, handleCreate,
        reloadTemplates, reloadOrganizations, showRecentOrganizations } = state;
    return (
        <div className="space-y-8">
                <section className="rounded-3xl border border-cortex-border bg-cortex-surface px-6 py-8 shadow-[0_18px_40px_rgba(148,163,184,0.16)]">
                    <div className="flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between">
                        <div className="max-w-3xl space-y-4">
                            <div className="inline-flex items-center gap-2 rounded-full border border-cortex-primary/25 bg-cortex-primary/10 px-3 py-1 text-[11px] font-mono uppercase tracking-[0.24em] text-cortex-primary">
                                <Sparkles className="h-3.5 w-3.5" />
                                AI Organization Setup
                            </div>
                            <div className="space-y-3">
                                <h1 className="text-4xl font-semibold tracking-tight text-cortex-text-main">
                                    Create AI Organization
                                </h1>
                                <p className="max-w-2xl text-base leading-7 text-cortex-text-muted">
                                    Start Mycelis by creating an AI Organization, not a blank assistant session. Choose a starting point, define the purpose, and open the new organization home.
                                </p>
                            </div>
                            <div className="flex flex-wrap gap-3">
                                <ActionButton onClick={() => focusCreateFlow("template")} kind="primary">
                                    Explore Templates
                                </ActionButton>
                                <ActionButton onClick={() => focusCreateFlow("empty")} kind="secondary">
                                    Start Empty
                                </ActionButton>
                            </div>
                        </div>
                            <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted lg:max-w-sm">
                                <p className="font-medium text-cortex-text-main">Hidden until Advanced mode</p>
                                <p className="mt-1 leading-6">
                                    AI Engine Settings and Memory &amp; Continuity stay out of the default flow until the operator intentionally opens advanced controls.
                                </p>
                            </div>
                    </div>
                </section>
                {lastOrganization && (
                    <section className="rounded-3xl border border-cortex-primary/25 bg-cortex-surface px-6 py-5">
                        <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
                            <div>
                                <p className="text-[11px] font-mono uppercase tracking-[0.2em] text-cortex-primary">Return to Organization</p>
                                <p className="mt-2 text-lg font-semibold text-cortex-text-main">{lastOrganization.name}</p>
                                <p className="mt-1 text-sm leading-6 text-cortex-text-muted">
                                    Re-enter the current AI Organization in one click, even if it does not appear in the recent list below yet.
                                </p>
                            </div>
                            <Link
                                href={`/organizations/${lastOrganization.id}`}
                                onClick={() => rememberLastOrganization(lastOrganization)}
                                className="inline-flex items-center gap-2 rounded-xl border border-cortex-primary/35 bg-cortex-primary px-4 py-2.5 text-sm font-semibold text-cortex-bg transition-colors hover:bg-cortex-primary/90"
                            >
                                Return to Organization
                                <ArrowRight className="h-4 w-4" />
                            </Link>
                        </div>
                    </section>
                )}
                <section id="create-ai-organization" ref={createSectionRef} className="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
                    <div className="space-y-6 rounded-3xl border border-cortex-border bg-cortex-surface p-6">
                        <div>
                            <h2 className="text-xl font-semibold text-cortex-text-main">Choose how to start</h2>
                            <p className="mt-1 text-sm text-cortex-text-muted">
                                Shape an AI Organization first so Mycelis feels like an operating environment, not a generic chat.
                            </p>
                        </div>
                        <div className="grid gap-3 md:grid-cols-2">
                            <StartModeCard
                                active={selectedMode === "template"}
                                title="Start from template"
                                description="Use a starter that already defines a Team Lead, initial Departments, and Specialist structure."
                                icon={<Layers3 className="h-5 w-5" />}
                                onClick={() => focusCreateFlow("template")}
                            />
                            <StartModeCard
                                active={selectedMode === "empty"}
                                title="Start empty"
                                description="Begin with a clean AI Organization and shape the structure after the organization exists."
                                icon={<Building2 className="h-5 w-5" />}
                                onClick={() => focusCreateFlow("empty")}
                            />
                        </div>
                        {templatesState === "loading" && (
                            <LoadingState label="Loading starter templates..." />
                        )}
                        {templatesState === "error" && templatesError && (
                            <ActionableState
                                title="Starter templates are unavailable"
                                message={templatesError}
                                guidance="You can retry, start empty now, or reopen a recent AI Organization if one is available above."
                                actions={
                                    <>
                                        <ActionButton onClick={reloadTemplates} kind="secondary" showRetryIcon>
                                            Retry starters
                                        </ActionButton>
                                        <ActionButton onClick={() => setSelectedMode("empty")} kind="ghost">
                                            Start empty instead
                                        </ActionButton>
                                    </>
                                }
                            />
                        )}
                        {templatesState === "ready" && isTemplateMode && templates.length === 0 && (
                            <EmptyState
                                title="No starter templates available"
                                detail="You can still create an AI Organization with Start empty while starter templates are unavailable."
                                actions={
                                    <ActionButton onClick={() => setSelectedMode("empty")} kind="secondary">
                                        Start empty instead
                                    </ActionButton>
                                }
                            />
                        )}
                        {templatesState === "ready" && isTemplateMode && templates.length > 0 && (
                            <div className="space-y-3">
                                <div>
                                    <h3 className="text-sm font-semibold uppercase tracking-[0.18em] text-cortex-text-muted">Starter templates</h3>
                                    <p className="mt-1 text-sm text-cortex-text-muted">Choose an AI Organization starter using operator-facing terms only.</p>
                                </div>
                                <div className="grid gap-3">
                                    {templates.map((template) => (
                                        <button
                                            key={template.id}
                                            onClick={() => setSelectedTemplateId(template.id)}
                                            className={`rounded-2xl border p-4 text-left transition-colors ${
                                                selectedTemplateId === template.id
                                                    ? "border-cortex-primary/40 bg-cortex-primary/10"
                                                    : "border-cortex-border bg-cortex-bg hover:border-cortex-primary/20"
                                            }`}
                                        >
                                            <div className="flex items-start justify-between gap-3">
                                                <div>
                                                    <p className="text-base font-semibold text-cortex-text-main">{template.name}</p>
                                                    <p className="mt-1 text-sm leading-6 text-cortex-text-muted">{template.description}</p>
                                                </div>
                                                <span className="rounded-full border border-cortex-border px-2 py-0.5 text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">
                                                    {template.organization_type}
                                                </span>
                                            </div>
                                            <div className="mt-4 grid gap-2 sm:grid-cols-2 xl:grid-cols-3">
                                                <Metric label="Team Lead" value={template.team_lead_label} />
                                                <Metric label="Advisors" value={String(template.advisor_count)} />
                                                <Metric label="Departments" value={String(template.department_count)} />
                                                <Metric label="Specialists" value={String(template.specialist_count)} />
                                                <Metric label="AI Engine Settings" value={template.ai_engine_settings_summary} />
                                                <Metric label="Memory & Continuity" value={template.memory_personality_summary} />
                                            </div>
                                        </button>
                                    ))}
                                </div>
                            </div>
                        )}
                    </div>
                    <div className="space-y-4 rounded-3xl border border-cortex-border bg-cortex-surface p-6">
                        <div>
                            <h2 className="text-xl font-semibold text-cortex-text-main">Define the AI Organization</h2>
                            <p className="mt-1 text-sm text-cortex-text-muted">
                                Focus on the organization name, purpose, and starting point. Advanced details stay tucked away until you choose to open them.
                            </p>
                        </div>
                        <label className="block space-y-2">
                            <span className="text-sm font-medium text-cortex-text-main">AI Organization name</span>
                            <input
                                ref={nameInputRef}
                                value={name}
                                onChange={(event) => setName(event.target.value)}
                                placeholder="Northstar Labs"
                                className="w-full rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-main outline-none transition-colors placeholder:text-cortex-text-muted focus:border-cortex-primary/40"
                            />
                        </label>
                        <label className="block space-y-2">
                            <span className="text-sm font-medium text-cortex-text-main">Purpose</span>
                            <textarea
                                value={purpose}
                                onChange={(event) => setPurpose(event.target.value)}
                                placeholder="Ship a focused AI engineering organization for product delivery."
                                className="min-h-[120px] w-full rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-main outline-none transition-colors placeholder:text-cortex-text-muted focus:border-cortex-primary/40"
                            />
                        </label>
                        <div className="space-y-2">
                            <p className="text-sm font-medium text-cortex-text-main">Suggested starting points</p>
                            <div className="flex flex-wrap gap-2">
                                <button
                                    type="button"
                                    onClick={() => {
                                        setName("Northstar Labs");
                                        setPurpose("Launch a focused AI product organization that can plan priorities, review work, and execute governed changes.");
                                    }}
                                    className="rounded-full border border-cortex-border bg-cortex-bg px-3 py-1.5 text-xs font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/25 hover:text-cortex-primary"
                                >
                                    Product delivery organization
                                </button>
                                <button
                                    type="button"
                                    onClick={() => {
                                        setName("Skylight Works");
                                        setPurpose("Create an AI operations organization that can review workflows, coordinate specialists, and keep execution visible.");
                                    }}
                                    className="rounded-full border border-cortex-border bg-cortex-bg px-3 py-1.5 text-xs font-medium text-cortex-text-main transition-colors hover:border-cortex-primary/25 hover:text-cortex-primary"
                                >
                                    Operations organization
                                </button>
                            </div>
                        </div>
                        <div className="rounded-2xl border border-cortex-border bg-cortex-bg p-4 text-sm text-cortex-text-muted">
                            <p className="font-medium text-cortex-text-main">Selected start</p>
                            <p className="mt-1 leading-6">
                                {selectedMode === "template" && selectedTemplate
                                    ? `${selectedTemplate.name} will shape the first Team Lead, Departments, and Specialists for this AI Organization.`
                                    : selectedMode === "empty"
                                      ? "Start empty creates the AI Organization first so you can shape Advisors, Departments, and Specialists afterward."
                                      : "Choose Start from template or Start empty to continue."}
                            </p>
                        </div>
                        {submitError && (
                            <ActionableState
                                title="Unable to create AI Organization"
                                message={submitError}
                                guidance="Check the organization name, purpose, and chosen starting point, then try again."
                                actions={
                                    <>
                                        <ActionButton onClick={handleCreate} kind="secondary" disabled={!canSubmit || isSubmitting} showRetryIcon>
                                            Try again
                                        </ActionButton>
                                        {isTemplateMode && (
                                            <ActionButton onClick={() => setSelectedMode("empty")} kind="ghost">
                                                Start empty instead
                                            </ActionButton>
                                        )}
                                    </>
                                }
                            />
                        )}
                        <button
                            onClick={handleCreate}
                            disabled={!canSubmit || isSubmitting}
                            className="inline-flex w-full items-center justify-center gap-2 rounded-2xl border border-cortex-primary/35 bg-cortex-primary/10 px-4 py-3 text-sm font-semibold text-cortex-primary transition-colors hover:bg-cortex-primary/15 disabled:cursor-not-allowed disabled:opacity-50"
                        >
                            {isSubmitting ? <Loader2 className="h-4 w-4 animate-spin" /> : <Blocks className="h-4 w-4" />}
                            Create AI Organization
                        </button>
                        {creationReadinessMessage && !submitError && (
                            <p className="text-sm text-cortex-text-muted">{creationReadinessMessage}</p>
                        )}
                        {isSubmitting && (
                            <p className="text-sm text-cortex-text-muted">
                                Opening your Soma workspace as soon as the organization is ready.
                            </p>
                        )}
                        <div className="rounded-2xl border border-cortex-border bg-cortex-bg p-4 text-sm text-cortex-text-muted">
                            <p className="font-medium text-cortex-text-main">What stays hidden for now</p>
                            <div className="mt-3 grid gap-2 sm:grid-cols-2">
                                <HiddenLater label="AI Engine Settings" icon={<Bot className="h-4 w-4" />} />
                                <HiddenLater label="Memory & Continuity" icon={<BrainCircuit className="h-4 w-4" />} />
                            </div>
                            <p className="mt-3 text-xs leading-6 text-cortex-text-muted">
                                Advanced setup opens later so you can focus on the organization name, purpose, and starting point first.
                            </p>
                        </div>
                    </div>
                </section>
                <CreateOrganizationRecentList
                    showRecentOrganizations={showRecentOrganizations}
                    diagnosticOrganizations={diagnosticOrganizations}
                    showDiagnosticOrganizations={showDiagnosticOrganizations}
                    visibleDiagnosticOrganizations={visibleDiagnosticOrganizations}
                    hiddenDiagnosticCount={hiddenDiagnosticCount}
                    testingSetupCountLabel={testingSetupCountLabel}
                    organizationsState={organizationsState}
                    organizationsError={organizationsError}
                    visibleOrganizations={visibleOrganizations}
                    openingOrganizationId={openingOrganizationId}
                    isSubmitting={isSubmitting}
                    onToggleDiagnostics={() => setShowDiagnosticOrganizations((value) => !value)}
                    onReloadOrganizations={reloadOrganizations}
                    onOpenOrganization={openOrganization}
                />
                <WhyStartHere />
        </div>
    );
}
