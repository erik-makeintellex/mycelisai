package swarm

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mycelis/core/internal/governance"
)

func TestHandleCreateTeamRejectsCallerSuppliedWorkerProfileLineage(t *testing.T) {
	_, nc := startTestNATS(t)
	soma := NewSoma(nc, &governance.Guard{}, NewRegistryFromManifests(nil), nil, nil, nil, nil)

	for _, body := range []string{
		`{"id":"forged-ref","name":"Forged Ref","members":[{"id":"worker","profile_ref":"custom.builder"}]}`,
		`{"id":"forged-snapshot","name":"Forged Snapshot","members":[{"id":"worker","profile_snapshot":{"id":"custom.builder","digest":"sha256:forged"}}]}`,
	} {
		req := httptest.NewRequest(http.MethodPost, "/api/swarm/teams", strings.NewReader(body))
		recorder := httptest.NewRecorder()
		soma.HandleCreateTeam(recorder, req)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
		}
		if !strings.Contains(recorder.Body.String(), "must be resolved through Soma") {
			t.Fatalf("body = %q", recorder.Body.String())
		}
	}
}
