# SDK (`sdk/python`)

> Navigation: [Repo AGENTS.md](../AGENTS.md) | [Proto AGENTS.md](../proto/AGENTS.md)

## Owns / does not own
- `sdk/python/src/relay/**` (the relay client and persistence helpers) and `sdk/python/tests/test_relay_client.py` -> `mycelis-platform-ops` (lead decision 2026-09-28: it is Python, run only through `ops` tasks, and has no Core authority surface). Coordinate with `mycelis-core-execution` whenever the regenerated `.proto` output changes (see `proto/AGENTS.md`).
- It is exercised directly by `uv run pytest -q sdk/python/tests/test_relay_client.py` (the `relay.test` and `relay.demo` tasks were removed in TASKS pass 2); `ops/config.py`, `ops/quality.py` and `ops/misc*.py` only reference its path, they don't consume its API.
- Generated protobuf modules under `sdk/python/src/relay/proto/**` and `sdk/python/src/scip/proto/**` are regenerated from `proto/**` (see [`proto/AGENTS.md`](../proto/AGENTS.md)); never hand-edit a `_pb2.py` file.

## Contracts
- The relay client's wire shape follows `proto/swarm/v1/swarm.proto` and `proto/envelope.proto`; a client-side field change without a matching `.proto` change breaks wire compatibility.
- Python only through `uv`: any local run/test of this SDK uses `uv run pytest`/`uv run python`, never bare `python3`/`pip`.

## Gates
- `uv run pytest -q sdk/python/tests/test_relay_client.py` (it has no dedicated task-runner wrapper).
- `uv run inv quality.max-lines` covers `sdk/python/src/**` (generated `_pb2.py` files are excluded, see `proto/AGENTS.md` gotchas).

## Gotchas
- This is a small, single-test-file package; do not assume broader relay/session-management functionality exists beyond `client.py` and `persistence.py` without reading them first.
