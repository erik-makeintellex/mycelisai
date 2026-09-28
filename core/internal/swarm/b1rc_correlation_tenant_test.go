package swarm

import (
	"testing"

	"github.com/mycelis/core/internal/cognitive"
)

// B1R-C: runtime agent inference is charged to its tenant ("default" today),
// so token-budget counters and ledger rows are tenant-keyed.
func TestB1rcBuildInferRequestCorrelationCarriesTenant(t *testing.T) {
	agent := resultContractTestAgent(&boundedInferenceProvider{response: "done"}, &resultContractToolExecutor{})
	agent.runID = "run-123"
	req, _, _ := agent.buildInferRequest("complete the work", nil)
	want := cognitive.InferenceCorrelation{RunID: "run-123", TeamID: "delivery-team", AgentID: "worker", TenantID: cognitive.DefaultBudgetTenant}
	if req.Correlation != want {
		t.Fatalf("correlation = %#v, want %#v", req.Correlation, want)
	}
}
