# V8 Config Bootstrap Resolution
> Navigation: [V8 Config and Bootstrap Model](V8_CONFIG_AND_BOOTSTRAP_MODEL.md)

Status: canonical bootstrap resolution detail.

## Definition

Bootstrap resolution is the staged transformation from configuration, templates, database state, operator input, and deployment wiring into a runtime-ready AI Organization.

## Stages

1. Load source inputs.
2. Validate source integrity and secret references.
3. Select or create the target Inception.
4. Resolve Soma Kernel defaults.
5. Resolve Central Council/advisor composition.
6. Resolve team/department defaults.
7. Resolve agent/specialist overrides.
8. Resolve provider policy and profile routing.
9. Resolve memory, identity, and continuity defaults.
10. Produce effective runtime state.

## Source Precedence

Conceptual source order:
1. templates
2. static configuration
3. database/runtime state
4. operator input
5. derived context

Deployment/env overrides wire endpoints and profile defaults. They do not silently replace organization truth.

## Policy Before Override

Each candidate value must pass:
- schema validation
- higher-scope governance
- provider/capability constraints
- secret-reference rules
- compatibility checks

Only then may a more specific source or scope refine a value.

## Output

Resolution outputs:
- effective organization shape
- effective Soma/council/team/agent defaults
- effective provider policy
- effective continuity and memory posture
- rejected/blocked candidate details where relevant

## Testing

Test:
- deterministic conflict resolution
- invalid override blocking
- missing bundle fail-closed behavior
- multi-bundle explicit selection
- env override endpoint/profile wiring without organization-truth replacement
