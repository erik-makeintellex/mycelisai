# V8 Config Migration Dehardening
> Navigation: [V8 Config and Bootstrap Model](V8_CONFIG_AND_BOOTSTRAP_MODEL.md)

Status: canonical migration/de-hardening detail.

## Purpose

V7 behavior relied on standing teams, YAML defaults, and startup assumptions that were useful during development but too rigid for V8 AI Organizations.

V8 replaces those assumptions with bundle-driven bootstrap and explicit instantiation.

## Observed V7 Assumptions

- one expected standing-team shape
- config files treated as live runtime truth
- provider routing implied by local defaults
- startup tolerated ambiguous or missing bootstrap inputs
- operator creation flows could drift from backend config truth

## V8 Replacement

- mount explicit bootstrap bundles
- select a bundle when more than one exists
- fail closed when required bootstrap truth is missing
- instantiate organizations before runtime use
- apply inheritance and precedence deterministically
- keep deployment/env overrides limited to endpoint/profile wiring

## Startup Rule

Normal startup should not silently invent a canonical organization when required bootstrap inputs are missing or ambiguous.

## Migration Checklist

When translating V7 assets:
- identify reusable template material
- remove raw secrets
- separate deployment endpoint assumptions
- define override boundaries
- add source metadata
- add tests for invalid/missing/ambiguous cases
- update `.state/V8_DEV_STATE.md`

## Proof

Proof should include:
- startup with one valid bundle
- startup with missing bundle fails closed
- multi-bundle startup requires explicit selection
- env provider override changes endpoint/profile wiring only
