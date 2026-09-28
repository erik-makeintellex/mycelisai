package swarm

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

func TestAgentTriggerRequestReplyReturnsDegradedTruthInsteadOfModelProse(t *testing.T) {
	server, nc := startTestNATS(t)
	defer server.Shutdown()
	defer nc.Close()

	provider := &resultContractProvider{responses: []string{"The requested package is complete."}}
	agent := resultContractTestAgent(provider, nil)
	agent.nc = nc
	// F16b: execution posture needs a claim the proof store confirms.
	f16bExpectLookup(f16bAttachProofStore(t, agent), "delivery-team").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	subject := "test.team.agent.result-contract"
	if _, err := nc.Subscribe(subject, agent.handleTrigger); err != nil {
		t.Fatal(err)
	}
	if err := nc.Flush(); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(protocol.TeamAsk{
		Goal: "Create retained work.",
		Context: map[string]any{
			"run_id": f16bRun, "contract_id": f16bContract, "intent_proof_id": f16bProof, "work_item_id": f16bWorkItem,
			"result_contract": map[string]any{"kind": "project_package", "entrypoint_required": true},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	reply, err := nc.Request(subject, raw, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	text := string(reply.Data)
	for _, want := range []string{"Work unavailable", "result_contract_unsatisfied", "Recovery:"} {
		if !strings.Contains(text, want) {
			t.Fatalf("reply = %q, missing %q", text, want)
		}
	}
	if strings.Contains(text, "The requested package is complete") {
		t.Fatalf("degraded reply leaked model completion prose: %q", text)
	}
}
