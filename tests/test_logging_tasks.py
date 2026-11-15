from __future__ import annotations

from pathlib import Path

from ops import logging as logging_tasks


def test_read_event_types_extracts_declared_constants(tmp_path: Path):
    events_file = tmp_path / "events.go"
    events_file.write_text(
        """
package protocol
const (
    EventMissionStarted EventType = "mission.started"
    EventToolInvoked EventType = "tool.invoked"
)
""".strip()
        + "\n",
        encoding="utf-8",
    )

    event_types = logging_tasks._read_event_types(events_file)

    assert event_types == {"mission.started", "tool.invoked"}


def test_collect_schema_violations_flags_unknown_event_literals(tmp_path: Path):
    source = tmp_path / "sample.go"
    source.write_text(
        """
package sample
func f() {
    _ = map[string]any{
        "event": EventPayload{EventType: "mission.started"},
        "bad": EventPayload{EventType: "mission.unknown"},
    }
}
""".strip()
        + "\n",
        encoding="utf-8",
    )

    violations = logging_tasks._collect_schema_violations(
        [source], {"mission.started", "tool.invoked"}
    )

    assert any("mission.unknown" in item for item in violations)
    assert not any("mission.started" in item for item in violations)


def test_check_doc_event_coverage_reports_missing_entries(tmp_path: Path):
    doc = tmp_path / "logging.md"
    doc.write_text("mission.started\n", encoding="utf-8")

    missing = logging_tasks._check_doc_event_coverage(
        {"mission.started", "tool.invoked"}, doc
    )

    assert missing == ["tool.invoked"]


def test_check_doc_contract_terms_reports_missing_terms(tmp_path: Path):
    doc = tmp_path / "logging.md"
    doc.write_text("OperationalLogContext\nmemory.stream\n", encoding="utf-8")

    missing = logging_tasks._check_doc_contract_terms(doc)

    assert missing == [
        "central_review",
        "mcp",
        "mission_events",
        "review_channels",
        "schema_version",
        "service_local",
        "source_channel",
        "suggested_action",
        "team_lead",
        "team_local",
    ]


def test_collect_topic_literal_violations_respects_allowed_file(tmp_path: Path):
    src = tmp_path / "runtime.go"
    src.write_text('subject := "swarm.global.broadcast"\n', encoding="utf-8")

    violations = logging_tasks._collect_topic_literal_violations([src], allowed_files=set())
    assert violations

    allowed = {src.as_posix()}
    violations_allowed = logging_tasks._collect_topic_literal_violations([src], allowed_files=allowed)
    assert violations_allowed == []


def test_build_and_format_operational_log_context_supports_agentry_and_human_review():
    ctx = logging_tasks.build_operational_log_context(
        component="compose.up",
        summary="PostgreSQL readiness probe timed out.",
        detail="container mycelis-postgres did not become healthy within 60s",
        suggested_action="Run uv run inv compose.logs and inspect database startup.",
        why_it_matters="Core runtime migrations cannot execute without PostgreSQL.",
        run_id="run-123",
        team_id="team-ops",
    )

    assert ctx["schema_version"] == "v1"
    assert ctx["review_scope"] == "central_review"
    assert ctx["centralized_review"] is True
    assert "soma" in ctx["audiences"]
    assert "memory.stream" in ctx["review_channels"]
    assert ctx["run_id"] == "run-123"

    formatted = logging_tasks.format_operator_diagnostic(ctx)
    assert '[compose.up] status=ERROR review_scope=central_review summary="PostgreSQL readiness probe timed out."' in formatted
    assert "suggested_action: Run uv run inv compose.logs and inspect database startup." in formatted
    assert "review_channels: memory.stream,mission_events" in formatted
