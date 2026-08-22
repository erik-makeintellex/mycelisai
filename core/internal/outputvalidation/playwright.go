package outputvalidation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultTimeout = 30 * time.Second
	maxTimeout     = 2 * time.Minute
	maxOutputBytes = 256 * 1024
)

// PlaywrightConfig identifies the installed Node runtime and retained validator script.
type PlaywrightConfig struct {
	NodeBinary string
	ScriptPath string
	WorkingDir string
	Timeout    time.Duration
}

type processResult struct {
	stdout   []byte
	stderr   []byte
	exitCode int
	err      error
}

type processRunner func(context.Context, string, []string, string, []byte) processResult

// PlaywrightValidator executes a replaceable Node adapter without invoking a shell.
type PlaywrightValidator struct {
	config PlaywrightConfig
	run    processRunner
}

// NewPlaywrightValidator creates a bounded validator. ScriptPath must be explicit.
func NewPlaywrightValidator(config PlaywrightConfig) (*PlaywrightValidator, error) {
	config.NodeBinary = strings.TrimSpace(config.NodeBinary)
	if config.NodeBinary == "" {
		config.NodeBinary = "node"
	}
	if strings.TrimSpace(config.ScriptPath) == "" {
		return nil, errors.New("output validation script path is required")
	}
	absScript, err := filepath.Abs(config.ScriptPath)
	if err != nil {
		return nil, fmt.Errorf("resolve output validation script: %w", err)
	}
	config.ScriptPath = absScript
	if config.WorkingDir != "" {
		config.WorkingDir, err = filepath.Abs(config.WorkingDir)
		if err != nil {
			return nil, fmt.Errorf("resolve output validation working directory: %w", err)
		}
	}
	config.Timeout = boundedTimeout(config.Timeout)
	return &PlaywrightValidator{config: config, run: runProcess}, nil
}

// Validate sends one JSON request over stdin and decodes one typed JSON report.
func (v *PlaywrightValidator) Validate(ctx context.Context, request Request) (Report, error) {
	normalized, err := normalizeRequest(request)
	if err != nil {
		return Report{}, err
	}
	payload, err := json.Marshal(normalized)
	if err != nil {
		return Report{}, fmt.Errorf("encode browser validation request: %w", err)
	}

	runCtx, cancel := context.WithTimeout(ctx, v.config.Timeout)
	defer cancel()
	startedAt := time.Now().UTC()
	result := v.run(runCtx, v.config.NodeBinary, []string{v.config.ScriptPath}, v.config.WorkingDir, payload)
	if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
		return localReport(normalized, startedAt, StatusUnavailable, "validator_timeout", "Browser validation exceeded its bounded runtime."), nil
	}
	if result.err != nil && len(bytes.TrimSpace(result.stdout)) == 0 {
		message := "Browser validation runtime is unavailable."
		if result.exitCode >= 0 && len(result.stderr) > 0 {
			message = boundedText(result.stderr)
		}
		return localReport(normalized, startedAt, StatusUnavailable, "validator_unavailable", message), nil
	}

	var report Report
	if err := json.Unmarshal(result.stdout, &report); err != nil {
		return localReport(normalized, startedAt, StatusError, "invalid_validator_report", "Browser validator returned an invalid report."), fmt.Errorf("decode browser validation report: %w", err)
	}
	if report.ContentDigest != normalized.ContentDigest || report.LaunchURL != normalized.LaunchURL {
		return Report{}, errors.New("browser validation report does not match request correlation")
	}
	if report.Status == "" {
		return Report{}, errors.New("browser validation report status is required")
	}
	return report, nil
}

func normalizeRequest(request Request) (Request, error) {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(request.LaunchURL))
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return Request{}, errors.New("output validation launch URL must be an absolute HTTP(S) URL")
	}
	request.LaunchURL = parsed.String()
	request.ContentDigest = strings.TrimSpace(request.ContentDigest)
	if request.ContentDigest == "" || len(request.ContentDigest) > 256 {
		return Request{}, errors.New("output validation content digest is required and must be at most 256 characters")
	}
	if strings.TrimSpace(request.EvidencePath) == "" {
		return Request{}, errors.New("retained evidence path is required")
	}
	absEvidence, err := filepath.Abs(request.EvidencePath)
	if err != nil {
		return Request{}, fmt.Errorf("resolve retained evidence path: %w", err)
	}
	request.EvidencePath = absEvidence
	if err := validatePlan(request.Plan); err != nil {
		return Request{}, err
	}
	return request, nil
}

func validatePlan(plan Plan) error {
	if err := plan.Validate(); err != nil {
		return fmt.Errorf("invalid browser validation plan: %w", err)
	}
	p := plan.Probe
	if p.Action.DurationMS < 0 || p.Action.DurationMS > 10_000 {
		return errors.New("browser validation probe duration_ms must be between 0 and 10000")
	}
	if p.Action.Kind != ActionClick && p.Action.Kind != ActionKeyPress &&
		p.Action.Kind != ActionKeyHold && p.Action.Kind != ActionFill &&
		p.Action.Kind != ActionPointer {
		return fmt.Errorf("browser adapter does not support action %q", p.Action.Kind)
	}
	return nil
}

func runProcess(ctx context.Context, executable string, args []string, dir string, input []byte) processResult {
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir = dir
	cmd.Stdin = bytes.NewReader(input)
	stdout := newLimitWriter(maxOutputBytes)
	stderr := newLimitWriter(maxOutputBytes)
	cmd.Stdout, cmd.Stderr = stdout, stderr
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		exitCode = -1
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
	}
	return processResult{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: exitCode, err: err}
}

func boundedTimeout(value time.Duration) time.Duration {
	if value <= 0 {
		return defaultTimeout
	}
	if value > maxTimeout {
		return maxTimeout
	}
	return value
}

func localReport(request Request, startedAt time.Time, status Status, code, message string) Report {
	return Report{
		Status: status, ContentDigest: request.ContentDigest, LaunchURL: request.LaunchURL,
		StartedAt: startedAt, FinishedAt: time.Now().UTC(),
		Diagnostics: []Diagnostic{{Code: code, Message: message, Severity: "error"}},
	}
}

func boundedText(value []byte) string {
	text := strings.TrimSpace(string(value))
	if len(text) > 2_000 {
		return text[:2_000] + "..."
	}
	return text
}

type limitWriter struct {
	buffer bytes.Buffer
	limit  int
}

func newLimitWriter(limit int) *limitWriter { return &limitWriter{limit: limit} }

func (w *limitWriter) Write(p []byte) (int, error) {
	original := len(p)
	remaining := w.limit - w.buffer.Len()
	if remaining > 0 {
		if len(p) > remaining {
			p = p[:remaining]
		}
		_, _ = w.buffer.Write(p)
	}
	return original, nil
}

func (w *limitWriter) Bytes() []byte { return w.buffer.Bytes() }
