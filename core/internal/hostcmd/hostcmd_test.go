package hostcmd

import (
	"context"
	"testing"
	"time"
)

func TestAllowedCommands_Defaults(t *testing.T) {
	t.Setenv("MYCELIS_LOCAL_COMMAND_ALLOWLIST", "")
	cmds := AllowedCommands()
	if len(cmds) == 0 {
		t.Fatal("expected default allowlist commands")
	}
}

func TestExecute_Disallowed(t *testing.T) {
	t.Setenv("MYCELIS_LOCAL_COMMAND_ALLOWLIST", "hostname")
	_, err := Execute(context.Background(), "whoami", nil, time.Second)
	if err == nil {
		t.Fatal("expected disallowed command error")
	}
}

func TestExecute_Success(t *testing.T) {
	t.Setenv("MYCELIS_LOCAL_COMMAND_ALLOWLIST", "hostname")
	out, err := Execute(context.Background(), "hostname", nil, 3*time.Second)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if out.Status != "success" {
		t.Fatalf("status = %q, want success", out.Status)
	}
	if out.Command != "hostname" {
		t.Fatalf("command = %q, want hostname", out.Command)
	}
}
