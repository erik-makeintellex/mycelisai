# Writer brief (all slice writers)

The root `AGENTS.md` wins on any conflict, including its Operating Doctrine. Your packet is the path the lead gives you (transient, not tracked); your role file is `.claude/agents/<role>.md`; your worktree is the one the lead names (WSL: `~/Projects/mycelisai-worktrees/<slice-dir>`).

Rules
1. Read your role file, your packet and `AGENTS.md` first, plus the area `AGENTS.md` of every folder you will touch (for example `core/AGENTS.md`, `interface/AGENTS.md`), where one exists. Stay inside your packet's owned files. Another writer may work in the same Go package in a different worktree: prefix every new test helper/type with your packet's helper prefix so merges do not collide.
2. Edit through the `\\wsl.localhost\mother-brain\...` path and keep LF endings. Run commands only inside WSL, from a script file: write the script to the scratch location the lead names, then run it from PowerShell with `wsl -d mother-brain -- bash -l <script>` (`-l` puts uv on PATH). Never use Git Bash for WSL paths, never bare `python3`, and never inline `$VAR`s in a `wsl -- bash -c '...'` string (Windows expands them). Script preamble: `export PATH="$HOME/.local/bin:$HOME/go/bin:/usr/local/go/bin:$PATH"; export MYCELIS_PROJECT_CACHE_ROOT="$HOME/Projects/mycelisai/workspace/tool-cache"; cd ~/Projects/mycelisai-worktrees/<slice-dir>`.
3. Git is read-only for you (status, diff, log). The lead commits and merges; the owner runs `git push`.
4. Never start, stop or recreate containers. The owner's `mycelis-home-*` stack and the Windows Ollama root are shared: do not call them, restart them or run compose/lifecycle tasks. The only allowed container is one disposable `docker run --rm` Postgres for DB tests, if your packet says so.
5. Never edit `.github/workflows`. Never print or commit secrets. Files stay within the 385-line policy (legacy caps are exact no-regression).
6. Test-first: turn the packet's evidence into a failing test, watch it fail, then fix. Give authority changes positive, negative and adversarial cases. Run targeted `go test -race -count=1 ./internal/<pkg>/ -run '<pattern>'` and `go vet ./internal/<pkg>/...`. Do not run full `uv run inv core.test`, first boots or image builds: the lead runs full gates once at merge (CPU is shared).
7. Verify a check can fail before trusting its pass (root `AGENTS.md`, Operating Doctrine, "Verify that a check can fail"). Use `set -o pipefail` in gate scripts so a `| tail` cannot hide a failure, and confirm a log source or container exists before reading zero matches as "none".
8. Docs: update the docs your packet names in the same slice. `docs/TESTING.md` and `docs/architecture/OPERATIONS.md` are at the line cap: net-zero edits there.
9. No coordination bus. Report only through your final handback.
10. Do not use a local model for your own work unless the packet says so; do the work yourself.
11. Never search from `/`, `/c`, `/mnt/c` or a drive root. Scope searches to your worktree and wrap them in `timeout 60`.
12. No placeholder or fake functionality: no templated content standing in for real output, no stubs returning success, no validations that always pass, no UI or API state claiming completion without evidence. If the real behavior cannot happen, return an honest normalized blocker (`respondBlocker`) or stop and report the blocker to the lead.
13. Never spawn sub-agents or background agents.
14. Python only through uv: `uv run python ...`, `uv run pytest ...`, `uv run inv ...`, `uvx <tool>`; installs via `uv add`/`uv pip`. Never bare `python3`/`python`/`pip`.
15. Final report to the lead, at most 300 words before the COMMIT BODY: files changed, test names and counts (fail-before/pass-after), docs changed and docs reviewed-but-unchanged, decisions made and open risks.
16. End with a COMMIT BODY block: subject line (`type(scope): ... (SLICE)`), What changed; How it works (3-8 lines: mechanism, invariants, authority); Proof (tests with counts); Follow-ups.
