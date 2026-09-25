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

## Starter And Delivery Defaults Blocks (MYC-RC-004A)

Template bundles may carry two optional blocks. Bundles without them, such as `core/config/templates/v8-migration-standing-team-bridge.yaml`, load unchanged.

```yaml
starter:                        # labels only; never shapes runtime
  label: Client delivery studio # required when the block exists, max 80 chars
  persona: studio               # studio | saas_product_team | operations | regulated_enterprise
  order: 10                     # chip sort order, >= 0
  sample_name: Brightline Studio
  sample_purpose: Draft client work and hold it for approval.
  soma_intro: "Your studio is ready. Nothing is sent until you approve."
delivery_defaults:              # shapes organization delivery posture
  work_mode: propose_only       # read_only | propose_only | execute_with_approval | execute_bounded | resume_continuity
  approval_policy_ref: delivery-approve-external  # must be a group name in core/config/policy.yaml
  approval_posture: approve_external_sends        # see postures below
  audit: standard               # standard (default) | full
  outputs: { retain: true, proof_required: true }
  max_team_members: 3           # 1..5; no team in the bundle may exceed it
```

Graduated approval postures:

| Posture | Company type | Meaning | Constraint |
| --- | --- | --- | --- |
| `approve_external_sends` | agency | nothing leaves the organization unapproved | not with `execute_bounded` |
| `approve_changes` | SaaS | deploy and write actions need approval | none |
| `approve_high_risk` | operations | bounded execution; high-risk actions need approval | none |
| `approve_all_mutations` | regulated | every mutation needs approval, full audit | requires `audit: full`; not with `execute_bounded` or `resume_continuity` |

Other rules: `outputs.proof_required: true` needs `outputs.retain: true`. The work-mode set lives in `core/pkg/protocol/work_modes.go`.

Loading and failure policy:

- Boot reads only the top level of `core/config/templates/` (no recursion), so starters never change which bundle Core boots with. A host run without `MYCELIS_BOOTSTRAP_TEMPLATE_ID` still boots the single top-level bundle. A top-level bundle with a structurally invalid block fails boot like any other schema error.
- Starters live in `core/config/templates/starters/`. They must declare both blocks. The organization-starters catalog reads them separately and never fails as a whole because of one starter: a starter that does not parse, fails validation, reuses a top-level or other starter id, or names an `approval_policy_ref` that is not a `policy.yaml` group is listed with `available: false` and an `unavailable_reason`, and it cannot be used to create an organization. Policy references are resolved against the live governance guard, or `config/policy.yaml` when no guard is wired, and are never checked at boot.
- `core/config/templates/starters/testdata/` holds a test-only fixture that the non-recursive loader never lists.

`GET /api/v1/templates?view=organization-starters` keeps every existing summary field and adds `source` (`bundle` or `starter`), `available`, `unavailable_reason`, `starter`, `delivery_defaults`, and `delivery_summary` (`work_mode_label`, `approval_posture`, `output_proof`, `team_shape`). Labelled starters sort first by `starter.order`.

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

Clone and modify (RC-004G): an operator clones a file bundle or starter into an operator template (`operator_templates`, migration 041) through the governed clone endpoint, then edits it through governed apply. The clone is a separate template with its own id and `source_template_id`; cloning or applying never rewrites the source file or any organization already created from it. See [Config clone and apply](V8_CONFIG_AND_BOOTSTRAP_MODEL.md#config-clone-and-apply-rc-004g).

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
