# V8 UI API Screen Contracts
> Navigation: [V8 UI/API Contract](V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md)

Status: canonical screen/API detail.

## Envelope Rules

UI routes expect normalized payloads with:
- stable IDs
- human labels
- role/type metadata
- loading, empty, error, and blocker states
- no raw backend/provider exception text

## Screen Families

| Surface | Required contract |
| --- | --- |
| `/dashboard` | re-entry, status overview, create/open actions |
| `/organizations/[id]` | Soma workspace, organization context, memory/activity/settings |
| `/groups` | temporary/standing collaboration and retained outputs |
| `/teams` | team roster and specialization review |
| `/teams/create` | guided compact team creation |
| `/resources` | resources, deployment context, Connected Tools |
| `/memory` | continuity and memory review |
| `/system` | health, degraded state, recovery guidance |
| `/settings` | profile, access, providers, tools, advanced controls |
| `/runs` and detail routes | timeline, status, chain, artifacts, retry/error states |

## Terminal States

Accepted terminal states:
- `answer`
- `proposal`
- `execution_result`
- `blocker`
- `empty`
- `error`

Long-running work may show `loading`, but it must not become an unbounded terminal state.

## Backend/API Change Rule

Every backend/API behavior change must name affected screens, terminal states, failure/recovery behavior, and evidence commands before review.
