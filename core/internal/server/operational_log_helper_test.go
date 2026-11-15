package server

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestBuildOperationalErrorContext_NormalizesAgentryAndHumanFields(t *testing.T) {
	ctx := BuildOperationalErrorContext(
		"runs_timeline",
		"Failed to load run lifecycle stages.",
		"database timeout while reading mission_events",
		"Verify PostgreSQL connectivity and retry GET /api/v1/runs/{id}/stages.",
		WithOperationalRunID("run-901"),
		WithOperationalTeamID("team-alpha"),
		WithOperationalWhyItMatters("Operator and Soma timeline stage strip cannot verify run progression."),
		WithOperationalSource(protocol.SourceKindWebAPI, "api.v1.runs.stages"),
	)

	if ctx.SchemaVersion != protocol.OperationalLogSchemaVersion {
		t.Fatalf("expected schema_version %q, got %q", protocol.OperationalLogSchemaVersion, ctx.SchemaVersion)
	}
	if ctx.ReviewScope != protocol.LogReviewScopeCentralReview {
		t.Fatalf("expected central_review scope, got %q", ctx.ReviewScope)
	}
	if !ctx.CentralizedReview {
		t.Fatal("expected CentralizedReview=true for central_review error context")
	}
	if ctx.Service != "mycelis-core" || ctx.Component != "runs_timeline" {
		t.Fatalf("unexpected service/component: %s / %s", ctx.Service, ctx.Component)
	}
	if ctx.RunID != "run-901" || ctx.TeamID != "team-alpha" {
		t.Fatalf("unexpected run_id/team_id: %s / %s", ctx.RunID, ctx.TeamID)
	}
	if ctx.SourceKind != protocol.SourceKindWebAPI || ctx.SourceChannel != "api.v1.runs.stages" {
		t.Fatalf("unexpected source metadata: %s / %s", ctx.SourceKind, ctx.SourceChannel)
	}
	if len(ctx.Audiences) == 0 || len(ctx.ReviewChannels) == 0 {
		t.Fatalf("expected normalized audiences and review channels, got %+v / %+v", ctx.Audiences, ctx.ReviewChannels)
	}
}

func TestLogOperationalError_EmitsStandardizedComponentLog(t *testing.T) {
	var buf bytes.Buffer
	origOut := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(origOut)

	ctx := LogOperationalError(
		"auth_middleware",
		"Deployment auth posture rejected request.",
		"enterprise posture requires MYCELIS_BREAK_GLASS_API_KEY",
		"Set MYCELIS_BREAK_GLASS_API_KEY via uv run inv auth.break-glass-key and restart services.",
	)

	output := buf.String()
	for _, expected := range []string{
		"[auth_middleware]",
		"status=error",
		"review_scope=central_review",
		`summary="Deployment auth posture rejected request."`,
		`suggested_action="Set MYCELIS_BREAK_GLASS_API_KEY via uv run inv auth.break-glass-key and restart services."`,
		"review_channels=",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected log output %q to contain %q", output, expected)
		}
	}
	if ctx.Status != "error" {
		t.Fatalf("expected status=error, got %q", ctx.Status)
	}
}
