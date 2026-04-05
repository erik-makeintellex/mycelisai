# Mycelis CLI
> Navigation: [Project README](../README.md) | [Docs Home](../docs/README.md)

## TOC

- [Purpose](#purpose)
- [Current Scope](#current-scope)
- [Development](#development)

## Purpose

This directory holds the repo-local Python CLI surface for lightweight command-line interaction with Mycelis.

## Current Scope

- `main.py` contains the current CLI entrypoint
- `tests/` holds CLI-specific validation
- the CLI is a secondary operator/developer surface, not the primary runtime or product shell

## Development

Use the repo task runner for validation and broader automation:

```bash
uv run inv -l
uv run inv ci.test
```
