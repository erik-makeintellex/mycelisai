# Registry Service
> Navigation: [Project README](../../../README.md) | [Docs Home](../../../docs/README.md)

## TOC

- [Purpose](#purpose)
- [Primary Files](#primary-files)
- [Current Status](#current-status)

## Purpose

This package provides the connector registry and wiring service for Mycelis.

Current responsibilities:
- store reusable `connector_templates`
- validate connector configs against JSON schema during installation
- refuse installs honestly: a valid install returns `ErrConnectorDeploymentUnavailable` and writes no `active_connectors` row
- expose a simple wiring graph view of connector topic inputs and outputs

## Primary Files

Primary files:
- `service.go`: template listing/registration, connector installation, wiring lookup
- `types.go`: connector and blueprint data types

## Current Status

Current status:
- database-backed template and active-connector records exist
- schema validation is enforced during install
- no connector deployer exists yet, so `POST /api/v1/teams/{id}/connectors` is a 501 `connector_deployment_unavailable` blocker
