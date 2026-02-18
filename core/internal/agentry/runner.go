package agentry

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

const maxRetries = 3

// Runner executes an agent's work loop with built-in proof-of-work verification.
// It implements a 3-step cycle: Draft -> Verify -> Evaluate, retrying on failure.
type Runner struct {
	Brain *cognitive.Router
}

// NewRunner creates a Runner wired to the cognitive Router.
func NewRunner(brain *cognitive.Router) *Runner {
	return &Runner{Brain: brain}
}

// Run executes the agent defined by manifest against the given input.
// It returns a ProofEnvelope containing the artifact and verification evidence.
func (r *Runner) Run(ctx context.Context, manifest protocol.AgentManifest, input string) (*protocol.ProofEnvelope, error) {
	var lastDraft string
	var failureLogs []string

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Step 1: Draft
		prompt := manifest.SystemPrompt
		if prompt == "" {
			prompt = fmt.Sprintf("You are a %s.", manifest.Role)
		}

		fullPrompt := fmt.Sprintf("%s\n\nInput: %s", prompt, input)
		if len(failureLogs) > 0 {
			fullPrompt += fmt.Sprintf("\n\nPrevious attempt failed. Failure context:\n%s\nPlease correct your output.", strings.Join(failureLogs, "\n"))
		}

		req := cognitive.InferRequest{
			Profile: manifest.Model,
			Prompt:  fullPrompt,
		}
		if req.Profile == "" {
			req.Profile = "chat"
		}

		resp, err := r.Brain.InferWithContract(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("draft failed (attempt %d): %w", attempt, err)
		}
		lastDraft = resp.Text

		// Step 2: Verify
		if manifest.Verification == nil {
			// No verification required - pass through
			return &protocol.ProofEnvelope{
				Artifact: lastDraft,
				Proof: protocol.Proof{
					Method:      "none",
					Logs:        "no verification configured",
					RubricScore: "n/a",
					Pass:        true,
				},
			}, nil
		}

		pass, verifyLogs, score, err := r.verify(ctx, manifest, lastDraft)
		if err != nil {
			return nil, fmt.Errorf("verification error (attempt %d): %w", attempt, err)
		}

		// Step 3: Evaluate
		if pass {
			log.Printf("Runner: Verification PASSED (attempt %d/%d)", attempt, maxRetries)
			return &protocol.ProofEnvelope{
				Artifact: lastDraft,
				Proof: protocol.Proof{
					Method:      manifest.Verification.Strategy,
					Logs:        verifyLogs,
					RubricScore: score,
					Pass:        true,
				},
			}, nil
		}

		log.Printf("Runner: Verification FAILED (attempt %d/%d): %s", attempt, maxRetries, verifyLogs)
		failureLogs = append(failureLogs, fmt.Sprintf("Attempt %d: %s", attempt, verifyLogs))
	}

	// All retries exhausted
	return &protocol.ProofEnvelope{
		Artifact: lastDraft,
		Proof: protocol.Proof{
			Method:      manifest.Verification.Strategy,
			Logs:        strings.Join(failureLogs, "\n"),
			RubricScore: "0/" + fmt.Sprintf("%d", len(manifest.Verification.Rubric)),
			Pass:        false,
		},
	}, fmt.Errorf("verification failed after %d attempts", maxRetries)
}

func (r *Runner) verify(ctx context.Context, manifest protocol.AgentManifest, draft string) (pass bool, logs string, score string, err error) {
	switch manifest.Verification.Strategy {
	case protocol.VerifyEmpirical:
		return r.verifyEmpirical(ctx, manifest.Verification.ValidationCommand, draft)
	case protocol.VerifySemantic:
		return r.verifySemantic(ctx, manifest.Verification.Rubric, draft)
	default:
		return true, "unknown strategy, defaulting to pass", "n/a", nil
	}
}

func (r *Runner) verifyEmpirical(ctx context.Context, command string, draft string) (bool, string, string, error) {
	if command == "" {
		return false, "no validation command configured", "0/1", fmt.Errorf("empirical verification requires a validation_command")
	}

	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = strings.NewReader(draft)

	err := cmd.Run()
	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\nSTDERR: " + stderr.String()
	}

	if err != nil {
		return false, output, "0/1", nil // Command failed = verification failed (not an error)
	}

	return true, output, "1/1", nil
}

func (r *Runner) verifySemantic(ctx context.Context, rubric []string, draft string) (bool, string, string, error) {
	if len(rubric) == 0 {
		return true, "empty rubric, auto-pass", "0/0", nil
	}

	rubricText := strings.Join(rubric, "\n- ")
	prompt := fmt.Sprintf(`You are a strict quality assessor. Grade the following draft against the rubric.

RUBRIC:
- %s

DRAFT:
%s

For each rubric item, respond PASS or FAIL with a one-line reason.
End with a final line: "VERDICT: PASS" or "VERDICT: FAIL"`, rubricText, draft)

	req := cognitive.InferRequest{
		Profile: "sentry",
		Prompt:  prompt,
	}

	resp, err := r.Brain.InferWithContract(ctx, req)
	if err != nil {
		return false, fmt.Sprintf("semantic verification inference failed: %v", err), "0/" + fmt.Sprintf("%d", len(rubric)), err
	}

	logs := resp.Text
	pass := strings.Contains(strings.ToUpper(logs), "VERDICT: PASS")

	// Count passes
	passCount := strings.Count(strings.ToUpper(logs), "PASS")
	// Subtract the verdict line itself if it passed
	if pass && passCount > 0 {
		passCount--
	}
	score := fmt.Sprintf("%d/%d", passCount, len(rubric))

	return pass, logs, score, nil
}
