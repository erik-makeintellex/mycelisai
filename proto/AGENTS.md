# Proto

> Navigation: [Repo AGENTS.md](../AGENTS.md) | [SDK AGENTS.md](../sdk/AGENTS.md)

## Owns / does not own
- `proto/envelope.proto` (the `scip.SignalEnvelope` message) and `proto/swarm/v1/swarm.proto` -> `mycelis-core-execution` (lead decision 2026-09-28: the wire shape is an execution contract, and its Go consumers are the swarm/router/governance/bootstrap/state runtime this role already owns).
- Generated Go output lands in `core/pkg/scip/envelope.pb.go` (consumed by `core/internal/scip`, envelope validation) and `core/pkg/pb/swarm/swarm.pb.go` (consumed by `core/internal/{swarm,router,governance,bootstrap,state}` plus `core/cmd/{server,probe,smoke}`; `core/internal/server` uses it only in `_test.go` files — verified on dev `38e74c3f`).
- Generated Python output (`sdk/python/src/relay/proto/swarm/v1/swarm_pb2.py`, `sdk/python/src/scip/proto/envelope_pb2.py`) is regenerated alongside the Go stubs; see [`sdk/AGENTS.md`](../sdk/AGENTS.md) (owned by `mycelis-platform-ops`, lead decision 2026-09-28) — coordinate with core-execution whenever the regenerated output changes.

## Contracts
- `.proto` files are the single source of truth for the wire schema; hand-editing a generated `_pb2.py` or `.pb.go` file instead of regenerating from `.proto` is not acceptable.
- `envelope.proto`'s `SignalEnvelope` fields: `trace_id`, `timestamp`, `sender_id`, `target_id`, `intent`, `signature` (HMAC-SHA256), `meta` (map), `data_type` (the `DataType` enum: `TEXT_UTF8`/`TENSOR_FLOAT`/`IMAGE_BINARY`/`STATE_DIFF`), `payload` (bytes). Any change to this shape is a wire-compatibility decision, not a local refactor.

## Gates
- `uv run inv quality.max-lines` covers `proto/**` `.proto` files.
- Regenerate with `uv run inv proto.generate` (Go and Python stubs, `ops/proto_relay.py`); never hand-run `protoc` or hand-edit generated output.

## Gotchas
- Regenerated `.pb.go` / `_pb2.py` / `_pb2_grpc.py` files are excluded from the line-cap gate (`GENERATED_SUFFIXES` in `ops/quality.py`) — do not hand-trim them to satisfy a line count.
