# V8 Mycelis Search Capability Delivery Plan
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md)

> Status: ACTIVE
> Last Updated: 2026-04-26
> Module Boundary: capability/MCP
> Purpose: Put the delivery teams together for a Mycelis-owned search capability that does not require Brave tokens as the only web-search path.

## Outcome

Mycelis should expose one governed search capability to Soma, Council, and selected teams:

`Soma -> Mycelis Search API -> local_sources | searxng | local_api | brave | disabled`

The user-facing promise is simple:
- Soma can search user-shared sources and governed deployment context without external search tokens.
- Soma can search the public web through a self-hosted SearXNG option or a narrow self-hosted HTTP search API when the operator enables that service.
- Brave or another hosted search provider remains optional, not mandatory.
- If no search provider is enabled, Soma returns a concrete capability blocker instead of saying web requests are impossible.

## Delivery Teams

| Team | Status | Owner Scope | First Deliverable |
| --- | --- | --- | --- |
| Architecture Lead | ACTIVE | Own provider contract, trust boundaries, API shape, and promotion rules. | `SearchProvider` contract covering `local_sources`, `searxng`, `local_api`, `brave`, and `disabled`. |
| Runtime Development | IN_REVIEW | Implement Go backend search service, provider config, internal tool exposure, and audit metadata. | `/api/v1/search`, `/api/v1/search/status`, and `web_search` internal tool with provider status in results/blockers. |
| Data/Memory Development | ACTIVE | Connect `local_sources` to governed context/vector stores without mixing it into ordinary chat memory. | Local-source search over user-private, customer, company, Soma-operating, reflection, artifact, and team memory lanes with scope filters. |
| Interface Development | IN_REVIEW | Make capability state visible without turning Resources into a tool console. | Connected Tools/Search status card plus Soma blocker copy for missing provider/credential. |
| Ops/Runtime Delivery | ACTIVE | Add self-hosted SearXNG or local HTTP search endpoint configuration as optional Compose/Kubernetes runtime wiring. | `MYCELIS_SEARCH_PROVIDER`, `MYCELIS_SEARXNG_ENDPOINT`, `MYCELIS_SEARCH_LOCAL_API_ENDPOINT`, and service-health checks. |
| Validation | ACTIVE | Prove behavior across disabled, local-only, SearXNG, and hosted-provider modes. | Contract, unit, API, and browser workflow tests listed below. |

## Architecture Contract

The backend owns a small search abstraction:

```text
SearchRequest
- query
- source_scope: local_sources | web | all
- max_results
- recency/time_range
- allowed_domains
- blocked_domains
- tenant/user/org context

SearchResult
- title
- url or local_source_id
- snippet
- source_kind: local_source | searxng | local_api | brave | hosted_search
- trust_class
- sensitivity
- retrieved_at
- provider_metadata
```

Provider rules:
- `local_sources` uses Mycelis-governed stores and never requires external tokens.
- `searxng` calls an operator-owned SearXNG endpoint and requires JSON output to be enabled there.
- `local_api` calls an operator-owned HTTP JSON search endpoint and never requires hosted search tokens from Mycelis.
- `brave` uses the existing curated MCP/provider path and requires `BRAVE_API_KEY`.
- `disabled` returns a structured blocker naming the missing provider configuration.

Soma rules:
- Direct Soma should call `web_search` for search intent when available.
- `fetch` remains the explicit URL retrieval path.
- `brave-search` remains an MCP path, but it is not the only search path.
- Missing provider state must be reported as a capability/configuration blocker with a Connected Tools or System next action.

## Local Developer Contract

Required env/config:

```text
MYCELIS_SEARCH_PROVIDER=disabled|local_sources|searxng|local_api|brave
MYCELIS_SEARXNG_ENDPOINT=http://searxng:8080
MYCELIS_SEARCH_LOCAL_API_ENDPOINT=http://search.local/api/search
MYCELIS_SEARCH_MAX_RESULTS=8
```

Local development should support:
- `disabled`: no public web search, but clear Soma blocker.
- `local_sources`: no external service required.
- `searxng`: optional Compose profile for self-hosted web metasearch.
- `local_api`: optional operator-owned HTTP search endpoint.
- `brave`: optional hosted-provider path requiring `BRAVE_API_KEY`.

## Testing Contract

Runtime tests:
- provider config resolves `disabled`, `local_sources`, `searxng`, `local_api`, and `brave` deterministically.
- disabled provider returns a structured blocker.
- local-source provider respects knowledge class, sensitivity, visibility, and target-goal filters.
- SearXNG provider calls `/search?q=...&format=json`, normalizes results, and handles disabled JSON/403 as a readable blocker.
- local API provider calls the configured endpoint, normalizes JSON results, and handles missing endpoint/unreachable service as a readable blocker.
- Brave path remains governed and credential-gated.

API tests:
- `POST /api/v1/search` returns normalized results with `source_kind`, trust, sensitivity, and provider metadata.
- `GET /api/v1/search/status` exposes whether search is configured, degraded, or disabled, plus the direct Soma `web_search` binding and token requirement posture.

Soma/tool tests:
- asking "can you search the web?" returns capability status, not a blanket no.
- asking for web research calls `web_search` when configured.
- asking to fetch an explicit URL uses `fetch`/retrieval path when installed.
- missing provider or credential produces a blocker naming the missing setting.

Browser workflow tests:
- Resources/Connected Tools shows the active search provider posture.
- Central Soma can report search capability status.
- A configured local-source search returns governed source results.
- A disabled web-search setup shows a clear recovery action.

## Acceptance Gate

The slice is acceptable when:
- local-source search works without Brave or any hosted token
- SearXNG can be enabled through self-hosted runtime config
- local API search can be enabled through self-hosted runtime config
- Brave remains optional
- Soma, Council, and teams use one governed search capability contract
- tests prove disabled, local-source, and at least one web-provider path
- docs and in-app docs expose the delivery contract
