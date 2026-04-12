# V8 MVP Media, Team Output, And Template Registry
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md)

> Status: ACTIVE
> Last Updated: 2026-04-11
> Purpose: Define the MVP path for user-output-first Soma work, media/team-managed output proof, Ollama model-role routing, and DB-registered conversation templates.
> Supporting Docs: `V8_UI_TEAM_FULL_TEST_SET.md`, `V8_UNIVERSAL_SOMA_AND_CONTEXT_MODEL_PRD.md`, `V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md`, `docs/WORKFLOWS.md`, `docs/COGNITIVE_ARCHITECTURE.md`

## TOC

- [User-Output Posture](#user-output-posture)
- [Current Local Model Inventory](#current-local-model-inventory)
- [MVP Model Roles](#mvp-model-roles)
- [Media Generation Boundary](#media-generation-boundary)
- [Media Provider Contract](#media-provider-contract)
- [Team-Managed Output Demonstration](#team-managed-output-demonstration)
- [Conversation Template Registry](#conversation-template-registry)
- [Workflow Proof Set](#workflow-proof-set)
- [Implementation Slices](#implementation-slices)

## User-Output Posture

Soma's default posture is user-output-first.

Soma should not judge ordinary creative, business, or operator-desired outputs by taste, ideology, or whether Soma would personally choose the same output. Soma should help the user get the requested output into a more useful, attributed, executable, reviewable, and expandable form.

Governance remains real, but it is scoped to operational boundaries rather than content preference judgment:

- capability fit: whether the requested output can be produced by the available local or configured engines
- source and attribution: whether material needs citation, provenance, license notes, or user-provided context separation
- security and privacy: whether the workflow would expose private data, credentials, tenant data, or internal files
- side effects: whether the workflow writes, deletes, deploys, spends money, calls external services, or changes durable state
- auditability: whether the workflow needs a retained trace because it affected company knowledge, customer context, artifacts, or scheduled work
- cost and resource use: whether a model, media engine, browser run, or external service use is materially expensive or long-running

When no concrete boundary is crossed, Soma should stay helpful and proceed with clarification, planning, specialist routing, or output generation.

## Current Local Model Inventory

The current hosted Ollama inventory observed for this release lane includes:

- embeddings: `nomic-embed-text:latest`, `bge-m3:latest`
- general and reasoning: `qwen3:8b`, `qwen3:14b`, `llama3.1:8b`, `llama3:8b`, `deepseek-r1:8b`, `qwen2.5:7b-instruct-q4_K_M`, `minimax-m2.7:cloud`
- code and web work: `qwen2.5-coder:7b-instruct`, `qwen2.5-coder:14b`, `deepseek-coder-v2:16b`, `deepseek-coder:6.7b-instruct`, `kirito1/qwen3-coder:latest`, `web-architect-qwen:latest`
- local custom orchestration candidates: `orchestrator:latest`, `mother-brain:latest`, `symbios-orchestrator:latest`, `wonderluma-qwen:latest`
- vision analysis candidates: `llava:7b`, `wonderluma-vision:latest`, `prism:latest`

Current repo config sets the enabled `ollama` provider to `qwen2.5-coder:7b-instruct` at `http://127.0.0.1:11434/v1`. Several execution profiles still name `local-ollama-dev`, but that provider is disabled in the checked-in local config, so runtime availability handling must continue to fall back to the enabled local provider instead of surfacing a generic chat failure.

The official Ollama library currently exposes a wider multimodal model surface, including current vision/tool/reasoning families such as Gemma 4 and Qwen vision-language families. Treat those as candidate additions, not as already-proven local runtime dependencies, until they are pulled and validated on the target host.

References:

- [Ollama model library](https://ollama.com/models)
- [Ollama vision model filter](https://ollama.com/search?c=vision)
- [Gemma 4 in the Ollama library](https://ollama.com/library/gemma4)

## MVP Model Roles

The MVP should route by output role before chasing a single universal model:

- Soma orchestration and operator-facing planning: use the configured local default first, then validate `qwen3:14b`, `qwen3:8b`, `llama3.1:8b`, `orchestrator:latest`, and `mother-brain:latest` as candidate routing models.
- Council architecture and planning: prefer `qwen3:14b` or `llama3.1:8b` where latency allows; use `qwen3:8b` when speed matters.
- Code and website artifact generation: prefer `qwen2.5-coder:14b`, `deepseek-coder-v2:16b`, or `web-architect-qwen:latest`; keep `qwen2.5-coder:7b-instruct` as the fast fallback.
- Vision review, OCR, and image critique: prefer `llava:7b` as the current built-in catalog baseline, then validate `wonderluma-vision:latest`, `prism:latest`, or a newly pulled current Ollama vision-language model.
- Embeddings and RAG: keep `nomic-embed-text:latest` as the proven current default; validate `bge-m3:latest` where multilingual or richer retrieval quality matters.
- Media planning and prompt writing: use the creative/council models to create the prompt, storyboard, campaign context, and acceptance rubric; do not confuse this with the media engine that actually produces pixels or audio.

Current organization output-model routing already has a useful baseline: `general_text -> qwen3:8b`, `research_reasoning -> llama3.1:8b`, `code_generation -> qwen2.5-coder:7b`, and `vision_analysis -> llava:7b`. The live model-review lane now also exposes stronger installed or pull-candidate options such as `qwen3:14b`, `qwen2.5-coder:14b`, `deepseek-coder-v2:16b`, and `gemma3:12b` when they fit the requested behavior.

When an output type is not pinned, Soma should select from installed self-hosted models using explicit criteria:

- prefer an installed model that declares fit for the detected output type before suggesting a pull or remote provider
- prefer larger local reasoning/code models when latency and memory budget are acceptable
- keep the media boundary explicit: Ollama text/vision models can plan prompts, write website/code artifacts, and review images, but image or voice generation requires the configured media engine
- ask the owner/admin before reviewing model behavior for a requested output or changing durable routing policy

## Media Generation Boundary

Ollama is not the current pixel/audio generation engine in this repo. The current image tool calls the configured media endpoint through `media.endpoint` and `media.model_id`, then stores a returned image artifact for preview/save/download.

Current repo contract:

- image generation: `generate_image` calls `media.endpoint + /images/generations` and stores an `image` artifact
- image save: `save_cached_image` writes the cached image to the governed workspace path and records the saved file path
- website generation: code/web models can generate HTML/CSS/TypeScript artifacts that return as document/code/file outputs
- voice and audio: `docs/WORKFLOWS.md` treats voice input and TTS output as a future rendering concern; no current hosted Ollama voice-generation path is proven in this repo
- ComfyUI or comparable external media graph: treat as an external workflow contract until a native connector is implemented and tested

This gives Soma a clean answer when a user asks for media:

- if a media engine is configured, build the creative team plan, call the media engine, and return the artifact visibly
- if only Ollama is available, create the prompt pack, storyboard, image-review rubric, website assets, or implementation files that the available models can actually produce
- if voice/audio is requested and no local audio engine is configured, return a blocker that names the missing engine and offers a local setup path rather than pretending the work completed

## Media Provider Contract

Media providers must support both local/self-hosted and hosted execution. Pinokio, ComfyUI, Stable Diffusion WebUI, and the repo's Diffusers server are local provider launch or endpoint options; Replicate, FLUX, DALL-E, ElevenLabs, and comparable API-backed services are hosted provider options.

All provider paths must normalize into the same Soma-visible artifact contract:

- provider identity: `provider_id`, provider kind (`local_process`, `local_http`, `mcp`, or `hosted_api`), endpoint origin, model/workflow id, and whether the provider was owner-configured or chosen by Soma for the output type
- locality and exposure: `local_only`, `self_hosted_remote`, or `hosted_external`, plus whether user/customer/project context leaves the self-hosted boundary
- output block: local/self-hosted providers that write files directly must target the configured Core `/data` mount. In Compose that is `MYCELIS_OUTPUT_BLOCK_MODE=local_hosted` plus `MYCELIS_OUTPUT_HOST_PATH`; in Kubernetes it is the chart-managed `outputBlock.mode=cluster_generated` PVC unless the operator explicitly chooses a hostPath exception.
- credential posture: required environment variables, secret presence without secret echoing, and whether the provider requires owner approval before first use
- cost and runtime posture: estimated cost class, long-running status, queue/run id where available, cancellation capability, and retry posture
- output normalization: generated image/audio/video/document artifacts return previewable content when safe, saved paths or download URLs when binary, and lineage metadata that ties the artifact back to the team/run/provider
- audit metadata: provider, model/workflow, request class, target team, source context class, and final artifact id must be visible in managed exchange/audit without exposing raw secrets or unnecessary prompt internals by default

Hosted providers are not a fallback loophole. They are normal first-class media providers when the owner/admin configures them or approves their use for a specific output class. Soma should prefer local/self-hosted providers when policy says local-first, but it must also support hosted media providers when the user wants capability breadth, speed, or higher-quality output and the relevant credential/cost/external-exposure posture is acceptable.

MVP provider routing order:

- if the admin pins an output type to a provider, use that provider after checking health and required credentials
- if no provider is pinned, prefer an online local/self-hosted provider that can produce the requested media type
- if no local provider is online, present configured hosted providers as executable options with clear exposure/cost notes
- if no executable provider exists, preserve value by returning the prompt pack, storyboard, team plan, and precise setup action instead of pretending media was generated

Pinokio-style app launchers should be treated as local process supervisors and model/app installers, not as the media provider protocol itself. Mycelis should connect to the service Pinokio launches, such as a ComfyUI or Stable Diffusion WebUI HTTP endpoint, or register an MCP connector that calls that endpoint through the curated MCP library.

## Team-Managed Output Demonstration

The MVP demonstration must show why a team-managed workflow is different from asking one agent to do everything.

Direct Soma output is enough when the user wants a short answer, draft, or simple plan. Team-managed output is required when the user asks for a durable deliverable package, multiple specialist steps, media plus review, attribution, retained outputs, or repeatable workflows.

For investor/user testing, the clearest proof is a marketing or product-presentation request:

- user asks Soma for a launch/presentation output package
- Soma identifies target outputs such as messaging brief, hero-image prompt, image artifact, landing-page draft, and review notes
- Soma creates or proposes a bounded temporary workflow group with a named team lead
- the team lead coordinates the work and the output remains attached to the group
- binary media is shown as an artifact/download path; text and code are displayed inline when readable
- the temporary group can be archived while retained outputs remain reviewable and downloadable

## Conversation Template Registry

The MVP registry should store reusable conversation templates in the database so users, Soma, and Council can use proven asks without hardcoding every workflow in source.

This must be a new surface, not an overload of existing template systems:

- keep `/api/v1/templates` for CE-1 orchestration templates and organization starter bundle views
- keep `connector_templates` for connector registry definitions
- keep inception recipes focused on RAG prompt recipes
- keep conversation templates focused on reusable Soma/Council/team asks, output contracts, and optional temporary-group drafts

Minimum template fields:

- `id`
- `tenant_id`
- `name`
- `description`
- `scope`: `soma`, `council`, `team`, or `temporary_group`
- `created_by`
- `creator_kind`: `user`, `soma`, `council`, or `system`
- `status`: `active`, `archived`, or `draft`
- `template_body`
- `variables` as JSON
- `output_contract` as JSON
- `recommended_team_shape` as JSON
- `model_routing_hint` as JSON
- `governance_tags` as JSON
- `created_at`
- `updated_at`
- `last_used_at`

Required API shape:

- `GET /api/v1/conversation-templates`
- `POST /api/v1/conversation-templates`
- `GET /api/v1/conversation-templates/{id}`
- `PATCH /api/v1/conversation-templates/{id}`
- `POST /api/v1/conversation-templates/{id}/instantiate`

Instantiation should return a bounded ask package that Soma can use to start a direct answer, consult Council, create a temporary group, or recommend a retained team.

The MVP instantiation path should be non-executing by default: render the template with supplied variables, update `last_used_at`, and return the ask package plus any workflow-group draft for the caller to approve or launch through the existing group endpoint.

Implementation checkpoint:

- `IN_REVIEW` the backend MVP slice now has migration `038_conversation_templates`, protocol/store types, admin API routes, internal tools, standing-team manifests, and focused tests for create and non-executing instantiation.
- `NEXT` expose template discovery/launch in the Soma/team UI.

## Workflow Proof Set

The release proof should include both automated and visible-browser validation:

- model inventory proof: UI/API can show configured output model routing and at least two popular local self-hostable candidates for each supported output family
- direct-vs-team proof: one prompt returns a direct inline Soma answer, while a deliverable-package prompt produces a team-managed output contract. `interface/e2e/specs/v8-ui-testing-agentry.spec.ts` now covers this as a mocked browser proof.
- media proof: image request either returns a real image artifact from the configured local/self-hosted or hosted media provider, or a clear missing-engine / missing-credential / approval-needed blocker with next setup action. `interface/e2e/specs/v8-ui-testing-agentry.spec.ts` now covers preview/save/download rendering with generated media artifact payloads; live configured-provider proof remains separate.
- website proof: generated website assets are returned as readable files/artifacts, not as invisible agent chatter
- voice proof: if no local audio engine exists, the UI shows an honest missing-engine blocker; once an engine exists, it must return an audio artifact or playback/download reference
- MCP proof: a team workflow uses an MCP-backed capability and the operator can see which MCP service/tool was used in the connected-tools/activity surface. `interface/e2e/specs/mcp-connected-tools.spec.ts` now covers the Connected Tools browser side for persisted activity, expanded server tools, and curated install; live team-run-to-MCP activity correlation remains separate.
- temporary group proof: a Soma-created temporary group can produce multiple outputs, be archived, and retain outputs after closure
- template proof: a registered conversation template can instantiate a direct Soma ask and a temporary-group ask

## Implementation Slices

Recommended order:

1. `COMPLETE` document and expose the user-output posture, model-role map, and media boundary in product docs and in-app docs.
2. `IN_REVIEW` add DB-backed conversation templates with backend CRUD and focused tests.
3. `IN_REVIEW` add an instantiation endpoint that turns a template into a bounded Soma/council/team ask package.
4. `COMPLETE` add a browser-visible demo workflow for direct answer vs team-managed marketing/media package in the mocked Soma workspace.
5. `IN_REVIEW` add media smoke proof that either returns an artifact or a precise missing-engine blocker; mocked browser proof now covers artifact preview/save/download, and live configured-engine proof remains `NEXT`.
6. `IN_REVIEW` add MCP usage proof tied to connected-tools activity; browser proof now covers the Connected Tools side, and live team-run-to-MCP correlation remains `NEXT`.
7. `NEXT` validate candidate Ollama model additions on target hardware before adding them to default catalog recommendations.
8. `NEXT` implement a provider registry for media execution that supports local Diffusers, Pinokio-launched ComfyUI / Stable Diffusion endpoints, MCP-backed media servers, and hosted providers with explicit credential, cost, locality, and audit metadata.
