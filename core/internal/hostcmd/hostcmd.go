package hostcmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

var (
	// ErrCommandNotAllowed is returned when a command is outside the configured allowlist.
	ErrCommandNotAllowed = errors.New("command is not in local allowlist")
	// ErrInvalidArguments is returned when command arguments are invalid.
	ErrInvalidArguments = errors.New("invalid local command arguments")
)

const (
	defaultTimeout = 5 * time.Second
	minTimeout     = 250 * time.Millisecond
	maxTimeout     = 30 * time.Second
	maxArgCount    = 8
	maxArgLen      = 256
	maxOutputBytes = 16 * 1024
)

var defaultAllowedCommands = []string{"hostname", "whoami"}

// Result is the normalized local command execution result.
type Result struct {
	Command    string   `json:"command"`
	Args       []string `json:"args"`
	Status     string   `json:"status"` // success | error | timeout
	ExitCode   int      `json:"exit_code"`
	Stdout     string   `json:"stdout,omitempty"`
	Stderr     string   `json:"stderr,omitempty"`
	DurationMS int64    `json:"duration_ms"`
}

// AllowedCommands returns the currently effective sorted command allowlist.
func AllowedCommands() []string {
	raw := strings.TrimSpace(os.Getenv("MYCELIS_LOCAL_COMMAND_ALLOWLIST"))
	if raw == "" {
		out := make([]string, len(defaultAllowedCommands))
		copy(out, defaultAllowedCommands)
		sort.Strings(out)
		return out
	}

	seen := map[string]struct{}{}
	out := make([]string, 0)
	for _, part := range strings.Split(raw, ",") {
		cmd := strings.ToLower(strings.TrimSpace(part))
		if cmd == "" {
			continue
		}
		if _, ok := seen[cmd]; ok {
			continue
		}
		seen[cmd] = struct{}{}
		out = append(out, cmd)
	}
	sort.Strings(out)
	return out
}

func isAllowed(command string) bool {
	cmd := strings.ToLower(strings.TrimSpace(command))
	if cmd == "" {
		return false
	}
	for _, allowed := range AllowedCommands() {
		if cmd == allowed {
			return true
		}
	}
	return false
}

func clampTimeout(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return defaultTimeout
	}
	if timeout < minTimeout {
		return minTimeout
	}
	if timeout > maxTimeout {
		return maxTimeout
	}
	return timeout
}

func validateArgs(args []string) error {
	if len(args) > maxArgCount {
		return fmt.Errorf("%w: max %d args", ErrInvalidArguments, maxArgCount)
	}
	for _, arg := range args {
		if len(arg) > maxArgLen {
			return fmt.Errorf("%w: arg too long", ErrInvalidArguments)
		}
	}
	return nil
}

func truncateOutput(s string) string {
	if len(s) <= maxOutputBytes {
		return s
	}
	return s[:maxOutputBytes] + "\n... [truncated]"
}

// Execute runs one allowlisted command without shell interpolation.
func Execute(ctx context.Context, command string, args []string, timeout time.Duration) (Result, error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return Result{}, fmt.Errorf("%w: command is required", ErrInvalidArguments)
	}
	if !isAllowed(command) {
		return Result{}, fmt.Errorf("%w: %s", ErrCommandNotAllowed, command)
	}
	if err := validateArgs(args); err != nil {
		return Result{}, err
	}

	timeout = clampTimeout(timeout)
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, command, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	res := Result{
		Command:    command,
		Args:       args,
		ExitCode:   0,
		Stdout:     truncateOutput(stdout.String()),
		Stderr:     truncateOutput(stderr.String()),
		DurationMS: time.Since(start).Milliseconds(),
	}

	if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
		res.Status = "timeout"
		res.ExitCode = -1
		return res, nil
	}

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			res.Status = "error"
			res.ExitCode = exitErr.ExitCode()
			return res, nil
		}
		return Result{}, fmt.Errorf("failed to execute %q: %w", command, err)
	}

	res.Status = "success"
	return res, nil
}
