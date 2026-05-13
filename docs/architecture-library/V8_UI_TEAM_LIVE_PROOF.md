# V8 UI Team Live Proof
> Navigation: [V8 UI Team Full Test Set](V8_UI_TEAM_FULL_TEST_SET.md) | [Remote User Testing](../REMOTE_USER_TESTING.md)

Status: ACTIVE live-backend proof detail.

## Stable Compose-Hosted Browser Proof

Bring up:

```bash
uv run inv compose.up --build --wait-timeout=240
uv run inv compose.status
uv run inv compose.health
```

Then run focused browser proof against the delivered UI address.

For the guarded Windows -> WSL release lane, use:

```bash
uv run inv wsl.validate --lane=release
uv run inv wsl.validate --lane=release --headed-browser
```

The first command is the headless deployment-mimic gate; the second opens visible Playwright windows for the same focused live-backend specs when browser-visible release evidence is required.

## Live Governed Browser Proof

Use when Soma, governance, Core proxy, retained outputs, groups, or runtime contracts changed:

```bash
uv run inv interface.e2e --headed --live-backend --server-mode=start --project=chromium --spec=e2e/specs/soma-governance-live.spec.ts
uv run inv interface.e2e --headed --live-backend --server-mode=start --project=chromium --spec=e2e/specs/groups-live-backend.spec.ts
uv run inv interface.e2e --headed --live-backend --server-mode=start --project=chromium --spec=e2e/specs/team-creation.spec.ts
```

When the Compose UI is already running and must remain the delivered surface under test, use `--server-mode=external` instead of replacing it with a managed `next start` server.

## Windows Self-Hosted Browser Proof

Use the Windows browser and delivered UI address:
- same-machine Compose/WSL: `http://localhost:3000`
- second machine: real host/IP/hostname

AI endpoint must be explicit and non-loopback from the runtime's perspective.

## Pass Conditions

The live proof passes only when:
- readiness checks pass first
- browser reaches the delivered UI address
- direct answer works
- governed proposal cancel and execute work
- retained outputs are visible after refresh
- failure/recovery behavior is user-readable

## Blocker Handling

Record blockers explicitly:
- missing media provider
- unreachable AI endpoint
- unhealthy storage
- unavailable Core or Interface
- raw backend error shown in UI
