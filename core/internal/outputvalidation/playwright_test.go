package outputvalidation

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"
)

func validRequest(t *testing.T) Request {
	t.Helper()
	return Request{
		LaunchURL:     "http://127.0.0.1:3000/output",
		ContentDigest: "sha256:abc123",
		EvidencePath:  t.TempDir(),
		Plan: Plan{
			Kind: KindInteractiveBrowser, Required: true,
			Checks: []Check{CheckLoad, CheckNoPageErrors, CheckNoFailedLocalAsset},
			Probe: &Probe{
				Action:  ProbeAction{Kind: ActionClick, Target: "#run"},
				Observe: ProbeObservation{Kind: ObserveTextChange, Target: "#status"},
			},
		},
	}
}

func TestPlaywrightValidatorPassesTypedRequestWithoutShell(t *testing.T) {
	validator, err := NewPlaywrightValidator(PlaywrightConfig{NodeBinary: "configured-node", ScriptPath: "validator.mjs"})
	if err != nil {
		t.Fatal(err)
	}
	request := validRequest(t)
	validator.run = func(_ context.Context, executable string, args []string, dir string, input []byte) processResult {
		if executable != "configured-node" || len(args) != 1 || args[0] != validator.config.ScriptPath || dir != "" {
			t.Fatalf("unexpected direct process invocation: %q %#v %q", executable, args, dir)
		}
		var received Request
		if err := json.Unmarshal(input, &received); err != nil {
			t.Fatal(err)
		}
		if received.ContentDigest != request.ContentDigest || !filepath.IsAbs(received.EvidencePath) {
			t.Fatalf("request was not preserved and normalized: %#v", received)
		}
		report := Report{
			Status: StatusPassed, ContentDigest: received.ContentDigest, LaunchURL: received.LaunchURL,
			StartedAt: time.Now().UTC(), FinishedAt: time.Now().UTC(),
		}
		encoded, _ := json.Marshal(report)
		return processResult{stdout: encoded}
	}

	report, err := validator.Validate(context.Background(), request)
	if err != nil || report.Status != StatusPassed {
		t.Fatalf("Validate() = (%#v, %v), want passed", report, err)
	}
}

func TestPlaywrightValidatorPreservesCriterionMappings(t *testing.T) {
	validator, err := NewPlaywrightValidator(PlaywrightConfig{ScriptPath: "validator.mjs"})
	if err != nil {
		t.Fatal(err)
	}
	request := validRequest(t)
	request.AcceptanceCriteria = []string{"Primary interaction changes the application state"}
	request.CriterionMappings = []CriterionMapping{{Criterion: request.AcceptanceCriteria[0], Source: CriterionSourceProbe}}
	validator.run = func(_ context.Context, _ string, _ []string, _ string, input []byte) processResult {
		var received Request
		if err := json.Unmarshal(input, &received); err != nil {
			t.Fatal(err)
		}
		if len(received.CriterionMappings) != 1 || received.CriterionMappings[0].Criterion != request.AcceptanceCriteria[0] {
			t.Fatalf("criterion mapping lost: %#v", received)
		}
		report := Report{Status: StatusPassed, ContentDigest: received.ContentDigest, LaunchURL: received.LaunchURL,
			StartedAt: time.Now().UTC(), FinishedAt: time.Now().UTC(), CriterionEvidence: []CriterionEvidence{{
				Criterion: received.AcceptanceCriteria[0], Passed: true, EvidenceRefs: []string{"proof/report.json"},
			}}}
		encoded, _ := json.Marshal(report)
		return processResult{stdout: encoded}
	}
	report, err := validator.Validate(context.Background(), request)
	if err != nil || len(report.CriterionEvidence) != 1 {
		t.Fatalf("Validate() = (%#v, %v)", report, err)
	}
}

func TestPlaywrightValidatorRejectsIncompleteCriterionMapping(t *testing.T) {
	validator, err := NewPlaywrightValidator(PlaywrightConfig{ScriptPath: "validator.mjs"})
	if err != nil {
		t.Fatal(err)
	}
	request := validRequest(t)
	request.AcceptanceCriteria = []string{"Win state is testable"}
	if _, err := validator.Validate(context.Background(), request); err == nil {
		t.Fatal("expected unmapped criterion rejection")
	}
}

func TestPlaywrightValidatorUnavailableWhenRunnerMissing(t *testing.T) {
	validator, err := NewPlaywrightValidator(PlaywrightConfig{
		NodeBinary: "definitely-not-a-mycelis-node-runtime",
		ScriptPath: "validator.mjs",
		Timeout:    time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	report, err := validator.Validate(context.Background(), validRequest(t))
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusUnavailable || report.Diagnostics[0].Code != "validator_unavailable" {
		t.Fatalf("unexpected unavailable report: %#v", report)
	}
}

func TestPlaywrightValidatorTimeoutIsTyped(t *testing.T) {
	validator, err := NewPlaywrightValidator(PlaywrightConfig{ScriptPath: "validator.mjs", Timeout: 20 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	validator.run = func(ctx context.Context, _ string, _ []string, _ string, _ []byte) processResult {
		<-ctx.Done()
		return processResult{err: ctx.Err(), exitCode: -1}
	}
	report, err := validator.Validate(context.Background(), validRequest(t))
	if err != nil || report.Status != StatusUnavailable || report.Diagnostics[0].Code != "validator_timeout" {
		t.Fatalf("Validate() = (%#v, %v), want typed timeout", report, err)
	}
}

func TestPlaywrightValidatorRejectsUncorrelatedReport(t *testing.T) {
	validator, err := NewPlaywrightValidator(PlaywrightConfig{ScriptPath: "validator.mjs"})
	if err != nil {
		t.Fatal(err)
	}
	validator.run = func(context.Context, string, []string, string, []byte) processResult {
		return processResult{stdout: []byte(`{"status":"passed","content_digest":"wrong","launch_url":"http://127.0.0.1:3000/output","started_at":"2026-01-01T00:00:00Z","finished_at":"2026-01-01T00:00:01Z"}`)}
	}
	_, err = validator.Validate(context.Background(), validRequest(t))
	if err == nil || err.Error() != "browser validation report does not match request correlation" {
		t.Fatalf("expected correlation error, got %v", err)
	}
}

func TestPlaywrightValidatorRejectsInvalidPlans(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Request)
	}{
		{"non-http URL", func(r *Request) { r.LaunchURL = "file:///tmp/output.html" }},
		{"missing digest", func(r *Request) { r.ContentDigest = "" }},
		{"missing evidence path", func(r *Request) { r.EvidencePath = "" }},
		{"unknown check", func(r *Request) { r.Plan.Checks = []Check{"execute_everything"} }},
		{"click without selector", func(r *Request) { r.Plan.Probe.Action.Target = "" }},
		{"observation without target", func(r *Request) { r.Plan.Probe.Observe.Target = "" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, err := NewPlaywrightValidator(PlaywrightConfig{ScriptPath: "validator.mjs"})
			if err != nil {
				t.Fatal(err)
			}
			request := validRequest(t)
			tt.mutate(&request)
			if _, err := validator.Validate(context.Background(), request); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
