# Registry Service

This package provides the connector registry and wiring service for Mycelis.

Current responsibilities:
- store reusable `connector_templates`
- validate connector configs against JSON schema during installation
- create `active_connectors` bound to teams
- expose a simple wiring graph view of connector topic inputs and outputs

Primary files:
- `service.go`: template listing/registration, connector installation, wiring lookup
- `types.go`: connector and blueprint data types

Current status:
- database-backed template and active-connector records exist
- schema validation is enforced during install
- runtime deployment of installed connectors is still a later implementation step
