package server

import (
	"errors"
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/internal/deploymentcontext"
	"github.com/mycelis/core/internal/memory"
)

func TestMemoryLifecycleDenial(t *testing.T) {
	private := &deploymentcontext.EntryRecord{KnowledgeClass: "customer_context", Visibility: "private", OwnerUserID: "user-ada", LoadedBy: "ada"}
	team := &deploymentcontext.EntryRecord{KnowledgeClass: "customer_context", Visibility: "team", TeamID: "team-a", OwnerUserID: "user-ada"}
	legacy := &deploymentcontext.EntryRecord{KnowledgeClass: "user_private_context", Visibility: "private", LoadedBy: "ada"}
	company := &deploymentcontext.EntryRecord{KnowledgeClass: "company_knowledge", Visibility: "team", OwnerUserID: "user-ada"}
	operating := &deploymentcontext.EntryRecord{KnowledgeClass: "soma_operating_context", Visibility: "global", OwnerUserID: "user-root"}
	globalCustomer := &deploymentcontext.EntryRecord{KnowledgeClass: "customer_context", Visibility: "global", OwnerUserID: "user-ada"}
	forgedLabel := memoryUser("user-mallory", "ada", "operator")

	cases := []struct {
		name   string
		who    *RequestIdentity
		record *deploymentcontext.EntryRecord
		want   string
	}{
		{"owner private", adaOwner, private, ""},
		{"owner team", adaOwner, team, ""},
		{"owner legacy row by label", adaOwner, legacy, ""},
		{"other user private", bobOther, private, codeMemoryEntryNotOwned},
		{"same username different user id", forgedLabel, private, codeMemoryEntryNotOwned},
		{"admin is not the owner of a private entry", adminWithWrite, private, codeMemoryEntryNotOwned},
		{"owner of org-wide class", adaOwner, company, codeAdminRequired},
		{"admin without memory:write", adminNoMemory, operating, codeAdminRequired},
		{"admin with memory:write", adminWithWrite, operating, ""},
		{"global customer entry is org-wide", adaOwner, globalCustomer, codeAdminRequired},
		{"admin with memory:write on global customer entry", adminWithWrite, globalCustomer, ""},
		{"no identity", nil, private, codeMemoryEntryNotOwned},
	}
	for _, tc := range cases {
		if got, _ := memoryLifecycleDenial(tc.who, tc.record); got != tc.want {
			t.Errorf("%s: got %q want %q", tc.name, got, tc.want)
		}
	}
}

func TestDeploymentContextLifecycle_AuditFailureChangesNothing(t *testing.T) {
	for _, action := range []string{"archive", "restore", "delete"} {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		state := `{"knowledge_store":"governed_context_store","knowledge_class":"customer_context","visibility":"private","owner_user_id":"user-ada","loaded_by":"ada","lifecycle_state":"active"}`
		if action == "restore" {
			state = `{"knowledge_store":"governed_context_store","knowledge_class":"customer_context","visibility":"private","owner_user_id":"user-ada","loaded_by":"ada","lifecycle_state":"archived"}`
		}
		mock.ExpectQuery("FROM artifacts").WillReturnRows(sqlmock.NewRows([]string{"id", "title", "metadata", "chunks"}).
			AddRow("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "Ada private note", []byte(state), 1))
		mock.ExpectExec("INSERT INTO log_entries").WillReturnError(errors.New("disk full"))

		s := &AdminServer{DB: db, Artifacts: artifacts.NewService(db, t.TempDir()), Mem: memory.NewServiceWithDB(db), Cognitive: newDeploymentContextBrain()}
		rr := lifecycleCall(t, lifecycleMux(s), adaOwner, action, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
		assertStatus(t, rr, http.StatusServiceUnavailable)
		if code := blockerCode(t, rr); code != codeServiceUnavailable {
			t.Fatalf("%s: code %q", action, code)
		}
		// sqlmock fails on any statement after the failed audit insert.
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("%s: %v", action, err)
		}
		db.Close()
	}
}

func TestDeploymentContextLifecycle_StoreOfflineIs503(t *testing.T) {
	s := &AdminServer{}
	rr := lifecycleCall(t, lifecycleMux(s), adaOwner, "archive", "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	assertStatus(t, rr, http.StatusServiceUnavailable)
}
