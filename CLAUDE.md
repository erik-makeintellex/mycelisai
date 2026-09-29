# Mycelis: Claude Code orientation

This file orients the Claude Code harness. It adds no rules. [`AGENTS.md`](AGENTS.md) is the operating contract and wins on any conflict.

## Read first

1. [`AGENTS.md`](AGENTS.md): the contract, Operating Doctrine, model routing and delivery-target team rule.
2. The area `AGENTS.md` of every folder you touch: index in the "Area" table of the root `AGENTS.md`.
3. [`.state/V8_DEV_STATE.md`](.state/V8_DEV_STATE.md): live truth. Start with the Resume Guide, then "Delivery Targets And Teams".
4. [`docs/architecture-library/MYCELIS_CANONICAL_PRD.md`](docs/architecture-library/MYCELIS_CANONICAL_PRD.md): the only architecture document.
5. [`docs/TESTING.md`](docs/TESTING.md) and [`docs/architecture/OPERATIONS.md`](docs/architecture/OPERATIONS.md) before running proofs.

## Surface and tooling

- The WSL `dev` checkout is the single development and proof surface. Windows is only a host (Ollama, browser, editors through `\\wsl.localhost\...`).
- The task runner is `uv run inv <namespace>.<task>`. Python goes through `uv` only.
- The lead changes Git topology. The owner runs `git push`.

## Agents

- Role roster and tier routing: [`.claude/agents/README.md`](.claude/agents/README.md).
- Writers follow [`.claude/agents/WRITER_BRIEF.md`](.claude/agents/WRITER_BRIEF.md).
- Packets are transient; outcomes are recorded in the scoreboard.
- `.claude/settings.local.json` is local and stays ignored.
