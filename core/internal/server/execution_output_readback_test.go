package server

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

const readbackTestRequest = "Write a short launch note for the Mycelis beta release"

func useTestWorkspace(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", root)
	return root
}

func writeWorkspaceTestFile(t *testing.T, root, rel, content string) {
	t.Helper()
	target := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func diskChecksum(t *testing.T, root, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func TestReadbackWorkspaceOutputStatuses(t *testing.T) {
	root := useTestWorkspace(t)
	real := "# Beta launch\n\nMycelis beta opens Monday with governed teams, retained outputs, and audit trails."
	writeWorkspaceTestFile(t, root, "notes/real.md", real)
	writeWorkspaceTestFile(t, root, "notes/empty.md", "  \n\t")
	writeWorkspaceTestFile(t, root, "notes/echo.html", "<html><body><h1>Request</h1><p>"+readbackTestRequest+"</p></body></html>")
	echoes := []string{readbackTestRequest}

	cases := []struct {
		name, path, expected, want string
	}{
		{"real file", "notes/real.md", "", outputReadbackVerified},
		{"real file with matching checksum", "notes/real.md", "sha256:" + sha256Hex(real), outputReadbackVerified},
		{"changed bytes", "notes/real.md", sha256Hex(real + " edited"), outputReadbackMismatch},
		{"missing", "notes/missing.md", "", outputReadbackMissing},
		{"empty", "notes/empty.md", "", outputReadbackEmpty},
		{"echo only", "notes/echo.html", "", outputReadbackEchoOnly},
		{"escape", "../outside.md", "", outputReadbackOutOfBounds},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := readbackWorkspaceOutput(tc.path, tc.expected, echoes)
			if got.Status != tc.want {
				t.Fatalf("status = %q (%s), want %q", got.Status, got.Detail, tc.want)
			}
			if got.verified() != (tc.want == outputReadbackVerified) {
				t.Fatalf("verified() = %v for %q", got.verified(), got.Status)
			}
		})
	}
	if got := readbackWorkspaceOutput("notes/real.md", "", echoes); got.Checksum != diskChecksum(t, root, "notes/real.md") || got.Bytes != int64(len(real)) {
		t.Fatalf("checksum/bytes must describe disk bytes: %+v", got)
	}
}

func TestOutputOnlyEchoesRequestKeepsSubstantiveOutput(t *testing.T) {
	substantive := "<h1>" + readbackTestRequest + "</h1><p>Mycelis beta opens Monday. Teams keep retained outputs with audit trails and recovery paths.</p>"
	if outputOnlyEchoesRequest([]byte(substantive), []string{readbackTestRequest}) {
		t.Fatal("output with real substance beyond the request must not count as an echo")
	}
	if outputOnlyEchoesRequest([]byte("short"), []string{"tiny"}) {
		t.Fatal("request texts shorter than the echo floor must be ignored")
	}
}

func TestConfirmActionProofReadsBackDisk(t *testing.T) {
	root := useTestWorkspace(t)
	approved := "<!doctype html><title>Launch</title><main>Beta opens Monday with governed teams and audit trails.</main>"
	writeWorkspaceTestFile(t, root, "generated/launch.html", approved)
	scope := &protocol.ScopeValidation{WorkIntent: &protocol.WorkIntent{Objective: readbackTestRequest}}
	result := plannedToolExecutionResult{Name: "write_file", Arguments: map[string]any{"path": "workspace/generated/launch.html", "content": approved}}

	summary := buildConfirmActionExecutionSummary("proof-1", "contract-1", "artifact-1", "run-1", "audit-1", scope, []plannedToolExecutionResult{result})
	if summary.Proof.Verified == nil || !*summary.Proof.Verified || summary.AuditRecovery.RecoveryState != "verified" {
		t.Fatalf("real file must verify: proof=%+v recovery=%+v", summary.Proof, summary.AuditRecovery)
	}
	proof := summary.Outputs[0].Proof
	if proof.ReadbackStatus != outputReadbackVerified || proof.Checksum != diskChecksum(t, root, "generated/launch.html") {
		t.Fatalf("proof must carry the disk checksum: %+v", proof)
	}

	// The disk now differs from the approved bytes: the checksum follows the disk and the proof fails.
	writeWorkspaceTestFile(t, root, "generated/launch.html", approved+"<!-- tampered -->")
	summary = buildConfirmActionExecutionSummary("proof-1", "contract-1", "artifact-1", "run-1", "audit-1", scope, []plannedToolExecutionResult{result})
	proof = summary.Outputs[0].Proof
	if *summary.Proof.Verified || proof.ReadbackStatus != outputReadbackMismatch || proof.Checksum != diskChecksum(t, root, "generated/launch.html") {
		t.Fatalf("changed bytes must not verify and checksum must reflect disk: verified=%v proof=%+v", *summary.Proof.Verified, proof)
	}
	if summary.AuditRecovery.Degradation == nil || summary.AuditRecovery.Degradation.Code != "output_readback_mismatch" {
		t.Fatalf("degradation = %+v", summary.AuditRecovery.Degradation)
	}
}

func TestConfirmActionResponseIsUnverifiedWhenOutputMissingEmptyOrEcho(t *testing.T) {
	root := useTestWorkspace(t)
	writeWorkspaceTestFile(t, root, "generated/empty.md", "")
	writeWorkspaceTestFile(t, root, "generated/echo.md", "# Request\n\n"+readbackTestRequest+"\n")
	scope := &protocol.ScopeValidation{WorkIntent: &protocol.WorkIntent{Objective: readbackTestRequest}}
	for path, want := range map[string]string{
		"generated/missing.md": "output_readback_missing",
		"generated/empty.md":   "output_readback_empty",
		"generated/echo.md":    "output_readback_echo_only",
	} {
		result := plannedToolExecutionResult{Name: "write_file", Arguments: map[string]any{"path": path}}
		data := confirmActionResponseData("proof-1", "contract-1", "artifact-1", "run-1", "audit-1", scope, []plannedToolExecutionResult{result}, nil, nil)
		summary := data["execution_summary"].(*protocol.ExecutionSummary)
		if data["verified"] != false || data["execution_state"] != "unverified" || *summary.Proof.Verified {
			t.Fatalf("%s: verified=%v state=%v", path, data["verified"], data["execution_state"])
		}
		if summary.AuditRecovery.RecoveryState != "unverified" || summary.AuditRecovery.Degradation == nil || summary.AuditRecovery.Degradation.Code != want {
			t.Fatalf("%s: recovery=%+v", path, summary.AuditRecovery)
		}
	}
}

func TestConfirmActionWithoutWorkspaceOutputKeepsRunAuditProof(t *testing.T) {
	useTestWorkspace(t)
	result := plannedToolExecutionResult{Name: "create_team", Arguments: map[string]any{"team_id": "ops-team", "name": "Ops"}}
	summary := buildConfirmActionExecutionSummary("proof-1", "contract-1", "artifact-1", "run-1", "audit-1", nil, []plannedToolExecutionResult{result})
	if summary.Proof.Verified == nil || !*summary.Proof.Verified {
		t.Fatalf("team creation proof rests on run+audit: %+v", summary.Proof)
	}
	if proof := summary.Outputs[0].Proof; proof.ReadbackStatus != outputReadbackNotApplicable || proof.Checksum != "" {
		t.Fatalf("non-file output must not claim readback: %+v", proof)
	}
}
