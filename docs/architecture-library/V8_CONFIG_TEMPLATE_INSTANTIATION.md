# V8 Config Template Instantiation
> Navigation: [V8 Config and Bootstrap Model](V8_CONFIG_AND_BOOTSTRAP_MODEL.md)

Status: canonical template and instantiation detail.

## Definition

Templates are reusable blueprints. Instantiated organizations are live runtime objects. They must not be collapsed into one concept.

## Template May Contain

- display name and purpose defaults
- Soma profile defaults
- council/advisor roles
- team/department patterns
- provider policy defaults
- memory/continuity defaults
- allowed override hints
- capability posture

## Template Must Not Contain

- live organization identity
- raw secrets
- run-specific execution state
- unreviewed private data
- persisted conversation state
- deployment-only endpoint assumptions

## Instantiation Paths

Supported paths:
- create from template
- create empty
- create from config/API
- clone and modify a template later

Instantiation binds:
- stable organization identity
- selected template source metadata
- initial runtime state
- effective policy
- operator-provided purpose/settings

## UI Rule

Beginner UI may hide inheritance and source details, but it must not imply that editing a live organization is the same thing as editing a reusable template.

## Testing

Prove:
- template creation does not create runtime state by itself
- organization creation produces stable identity
- edits to instantiated organization do not mutate reusable templates unless explicitly requested
- invalid template inputs produce blockers
