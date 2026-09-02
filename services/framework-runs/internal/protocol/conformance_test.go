package protocol

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSharedV1GoldensParseStrictly(t *testing.T) {
	root := os.Getenv("MYCELIS_CONTRACT_ROOT")
	if root == "" {
		root = filepath.Join("..", "..", "..", "..", "contracts", "framework-runs", "v1")
	}
	cases := map[string]any{
		"create_request.json": &CreateRequest{}, "run_snapshot.json": &Run{},
		"event.json": &Event{}, "stop_request.json": &StopRequest{},
		"approval_request.json": &ApprovalDecisionRequest{},
		"control_receipt.json":  &ControlReceipt{}, "error.json": &ErrorEnvelope{},
	}
	for name, target := range cases {
		raw, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatalf("shared contract %s is required: %v", name, err)
		}
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(target); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if err := validateGolden(target); err != nil {
			t.Fatalf("%s violates the production contract: %v", name, err)
		}
	}
}

func validateGolden(target any) error {
	switch value := target.(type) {
	case *CreateRequest:
		NormalizeCreate(value)
		return ValidateCreate(*value)
	case *Event:
		return ValidateEvent(*value)
	case *StopRequest:
		return ValidateStop(*value)
	case *ApprovalDecisionRequest:
		return ValidateApprovalDecision(*value)
	case *Run:
		if value.Version < 1 || value.CreatedAt.IsZero() || value.UpdatedAt.IsZero() {
			return os.ErrInvalid
		}
		if err := validateCorrelation(value.Correlation, value.RunID); err != nil {
			return err
		}
		if value.Result == nil || value.Result.Metadata["completion_authority"] != "candidate" ||
			value.Result.Metadata["requires_core_validation"] != true || value.Result.Metadata["verified"] != false {
			return os.ErrInvalid
		}
		candidate := *value.Result
		candidate.Metadata = map[string]any{}
		candidate.Outputs = append([]Output(nil), value.Result.Outputs...)
		for index := range candidate.Outputs {
			candidate.Outputs[index].Metadata = map[string]any{}
		}
		return ForceAndValidateCandidate(&candidate, value.RunID)
	case *ControlReceipt:
		if value.CommandID == "" || value.RunID == "" || value.Version < 1 || value.State != "applied" {
			return os.ErrInvalid
		}
	case *ErrorEnvelope:
		if value.Error.Code == "" || value.Error.Message == "" {
			return os.ErrInvalid
		}
	}
	return nil
}

func TestCreateAndControlValidationFailClosed(t *testing.T) {
	request := validCreate()
	NormalizeCreate(&request)
	if err := ValidateCreate(request); err != nil {
		t.Fatal(err)
	}
	bad := request
	bad.RunID = " changed"
	if ValidateCreate(bad) == nil {
		t.Fatal("noncanonical run id accepted")
	}
	bad = request
	bad.Metadata = map[string]any{"correlation_id": "shadow"}
	if ValidateCreate(bad) == nil {
		t.Fatal("duplicated correlation metadata accepted")
	}
	bad = request
	bad.Input = map[string]any{"token": "raw-secret"}
	if ValidateCreate(bad) == nil {
		t.Fatal("raw secret accepted")
	}
	if ValidateStop(StopRequest{CommandID: "stop-1", ExpectedVersion: 0, ActorID: "core"}) == nil {
		t.Fatal("zero expected_version accepted")
	}
	bad = request
	bad.OrgID = string(make([]byte, 257))
	if ValidateCreate(bad) == nil {
		t.Fatal("oversized org_id accepted")
	}
	bad = request
	bad.RequiredProtocols = make([]string, 33)
	if ValidateCreate(bad) == nil {
		t.Fatal("oversized required_protocols accepted")
	}
	event := Event{EventID: "event-1", Sequence: 1, Version: 1, RunID: request.RunID,
		Correlation: request.Correlation, Kind: EventAccepted, Status: StatusAccepted,
		Timestamp: time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)}
	event.Correlation.WorkItemID = ""
	if ValidateEvent(event) == nil {
		t.Fatal("event with incomplete correlation accepted")
	}
}

func TestCandidateManifestIsForcedAndScoped(t *testing.T) {
	result := Result{Outputs: []Output{{
		ID: "output-1", Kind: "document", URI: "candidate://run-1/output-1",
		ContentType: "application/json", SizeBytes: 2,
		SHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}}, Metadata: map[string]any{"verified": true},
		FinishedAt: time.Date(2026, 9, 2, 12, 0, 1, 0, time.UTC)}
	if err := ForceAndValidateCandidate(&result, "run-1"); err != nil {
		t.Fatal(err)
	}
	if result.Metadata["completion_authority"] != "candidate" || result.Metadata["verified"] != false {
		t.Fatalf("candidate authority was not forced: %#v", result.Metadata)
	}
	result.Outputs[0].URI = "https://example.invalid/output"
	if ForceAndValidateCandidate(&result, "run-1") == nil {
		t.Fatal("open URL accepted as candidate URI")
	}
	result.Outputs[0].URI = ""
	if ForceAndValidateCandidate(&result, "run-1") == nil {
		t.Fatal("URI-less candidate bypassed immutable manifest")
	}
}

func TestOutcomeRejectsReservedAuthorityAndUnsafeEvidence(t *testing.T) {
	outcome := ExecutorOutcome{Status: StatusRunning, Metadata: map[string]any{"verified": true}}
	if ValidateOutcome(&outcome, "run-1") == nil {
		t.Fatal("executor authority override was accepted")
	}
	outcome = ExecutorOutcome{Status: StatusFailed, Error: &Error{Code: "", Message: "failed"}}
	if ValidateOutcome(&outcome, "run-1") == nil {
		t.Fatal("incomplete executor error was accepted")
	}
	outcome = ExecutorOutcome{Status: StatusRunning, Usage: &Usage{DurationMS: -1}}
	if ValidateOutcome(&outcome, "run-1") == nil {
		t.Fatal("negative usage was accepted")
	}
}

func validCreate() CreateRequest {
	return CreateRequest{
		RunID: "run-1", Intent: "Produce a candidate", Correlation: Correlation{
			RunID: "run-1", IntentProofID: "proof-1", ExecutionContractID: "contract-1",
			WorkItemID: "work-1", IdempotencyKey: "idem:1", SourceKind: "web_api",
			SourceChannel: "api.intent", PayloadKind: "command", GraphRevision: "graph-1",
		},
	}
}
