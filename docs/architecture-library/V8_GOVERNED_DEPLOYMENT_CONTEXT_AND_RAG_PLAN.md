# V8 Governed Deployment Context and RAG Plan

> Status: ACTIVE
> Last Updated: 2026-04-04
> Owners: Soma runtime, memory/RAG, governance, Resources UI, release verification

## Purpose

Turn the new governed deployment-context intake into a complete operating model for Mycelis.

This plan covers how Soma, Council, and teams should use pgvector-backed deployment knowledge without blurring it into ordinary remembered facts.

## Canonical Memory Boundary

Mycelis now has three distinct long-lived reasoning substrates plus temporary continuity:

1. `Soma memory`
   - durable remembered facts, decisions, summaries, recipes, and continuity recall
   - tools: `remember`, `recall`, `search_memory`, conversation summaries
   - purpose: what Soma has learned or deliberately promoted as memory

2. `customer_context`
   - operator- or customer-provided documents, briefs, constraints, handoff notes, policies, and curated research
   - purpose: deployment-shaping input context that should influence reasoning without becoming ordinary Soma memory

3. `company_knowledge`
   - approved company-authored guidance, playbooks, reference procedures, and organization-authored deployment knowledge
   - purpose: durable company reference material that Soma and teams may reuse only after explicit approval

4. `temporary continuity`
   - restart-safe planning checkpoints and in-flight work state
   - purpose: current work continuity without silent promotion into durable memory

## Product Rules

1. Soma memory must never silently absorb customer-provided deployment documents.
2. Customer-provided context must default into `customer_context`, not `company_knowledge`.
3. Company-authored knowledge must enter `company_knowledge` only through an explicit approved path.
4. Web research is allowed only as governed input, with explicit trust and sensitivity posture.
5. MCP/fetch/web access must stay securable by policy, source kind, trust class, and context-security settings.
6. Council and teams must follow the same context-boundary rules as Soma.

## Phase Order

### Phase 1: Ingest Truth

Goal:
prove all governed deployment knowledge enters the correct store with correct metadata.

Required:
- `customer_context` as default ingest class
- explicit `company_knowledge` selection for approved company-authored content
- artifact + vector metadata parity
- provenance fields on every ingest:
  - `knowledge_class`
  - `source_kind`
  - `source_label`
  - `visibility`
  - `sensitivity_class`
  - `trust_class`
  - `loaded_by`

Current state:
- complete for manual operator intake via Resources and API
- complete for the internal `load_deployment_context` tool

### Phase 2: Runtime Recall Truth

Goal:
prove Soma, Council, and teams all understand which store to use and how to interpret results.

Required:
- runtime prompts explain the difference between Soma memory, customer context, and company knowledge
- governed context recall must search only `customer_context` and `company_knowledge`
- recall output must remain grouped by store
- default reasoning should prefer:
  1. relevant Soma memory
  2. relevant customer context
  3. relevant company knowledge
  4. only then additional external research when needed and permitted

Next work:
- verify council-specialist prompts inherit the same memory-boundary section
- add focused tests for team and council prompt composition if any paths bypass the shared prompt builder

### Phase 3: Promotion and Conversion Workflow

Goal:
define how material moves between stores without ambiguity.

Required:
- no direct silent promotion from `customer_context` to `company_knowledge`
- explicit approval flow for promoting approved synthesized content into `company_knowledge`
- clear distinction between:
  - source document
  - Soma-authored synthesis
  - approved company reference

Current state:
- first minimal governed promotion path is implemented
- promotion runs through the stored confirm-action plan, not a side-channel bypass
- promotion creates a new `company_knowledge` artifact/vector record and preserves lineage back to the source `customer_context` artifact
- operator-facing dedicated promotion UI is still future work

Implementation target:
- add a promotion workflow such as `promote_context_to_company_knowledge`
- require proposal + approval for promotion
- preserve lineage:
  - source artifact IDs
  - source knowledge class
  - approving user
  - run/proposal IDs

### Phase 4: Governed Web and MCP Intake

Goal:
make web research and MCP-powered input useful without making it implicitly trusted.

Required:
- `source_kind=web_research` stays explicit
- context-security settings control whether external research may be loaded automatically, optionally, or only with approval
- MCP/fetch/web access remains policy-bound and inspectable
- operator can see whether a piece of context came from:
  - user document
  - user note
  - workspace file
  - web research
  - future MCP/import surfaces

Implementation target:
- add a direct path for Soma to research, summarize, and propose loading results into `customer_context`
- add policy/config switches for:
  - external research allowed
  - trusted domains / trust posture
  - approval required for external context loading
  - visibility defaults for external context

### Phase 5: Retrieval Quality and Query Strategy

Goal:
make the stores accurate and useful at runtime rather than just present.

Required:
- better chunking guidance for large docs
- optional tags and structured retrieval hints
- ranking strategy that can distinguish:
  - deployment architecture
  - security constraints
  - MCP/tooling constraints
  - operating procedures
  - approved company guidance

Implementation target:
- review chunk sizing and overlap against real docs
- add tests for recall relevance across mixed stores
- consider lightweight reranking or store-aware ranking rules if needed

### Phase 6: UI and Operator Explainability

Goal:
make the store model obvious to operators.

Required:
- Resources clearly teaches:
  - Soma memory vs customer context vs company knowledge
  - when to choose each store
  - why approved company knowledge is higher-trust and higher-risk
- future inspect surfaces show recent governed context and promotion lineage
- avoid exposing raw vector internals by default

Implementation target:
- add short store-specific guidance and examples in Resources
- add inspect-only lineage details for promoted company knowledge
- keep advanced details behind inspect surfaces

### Phase 7: Release Proof

Goal:
prove the deployment-context model is real from committed state.

Required backend proof:
- operator/API ingest creates governed artifact + vectors
- `customer_context` and `company_knowledge` are stored under separate vector types
- governance risk differs by store and source kind
- runtime recall surfaces grouped store results

Required UI proof:
- Resources intake supports both stores
- success states and list states show the selected knowledge class
- operator guidance explains the boundary clearly

Required live proof:
- Soma can ingest customer-provided material through the governed path
- approved company-authored content can be loaded only through the higher-governance path
- external/web research respects context-security posture

## Acceptance Gates

The plan is complete when:

1. all governed deployment knowledge flows through the two-store model
2. all runtime prompt paths teach the same boundary
3. promotion into `company_knowledge` is explicit and approval-backed
4. web/MCP research intake is policy-bound and inspectable
5. operator docs and UI use the same vocabulary as runtime and tests
6. committed-state validation passes for backend, UI, and docs

## Immediate Next Steps

1. Verify that no Council or delegated team runtime path bypasses the shared memory-boundary prompt section.
2. Design the explicit promotion workflow from `customer_context` to `company_knowledge`.
3. Add context-security configuration for governed external/web ingestion posture.
4. Add live/browser proof for governed external research intake.
5. Fold this plan into the broader memory continuity and investor-demo lanes so deployment context becomes a visible product strength, not just a backend capability.
