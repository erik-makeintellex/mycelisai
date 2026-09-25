package server

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func newOrganizationRepoMock(t *testing.T) (*PostgresOrganizationRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewPostgresOrganizationRepository(db, ""), mock
}

func testOrganizationHome(id, name string) OrganizationHomePayload {
	return OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			ID:                id,
			Name:              name,
			Purpose:           "Ship durable organizations",
			StartMode:         OrganizationStartModeTemplate,
			TemplateID:        "engineering-starter",
			AIEngineProfileID: "starter_defaults",
			Status:            "ready",
		},
		Departments: []OrganizationDepartmentSummary{{ID: "platform", Name: "Platform"}},
	}
}

func organizationDoc(t *testing.T, home OrganizationHomePayload) []byte {
	t.Helper()
	raw, err := json.Marshal(home)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return raw
}

func TestPostgresOrganizationRepository_SaveUpserts(t *testing.T) {
	repo, mock := newOrganizationRepoMock(t)
	home := testOrganizationHome("org-1", "Northstar")

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO organizations")).
		WithArgs("org-1", "default", "Northstar", "Ship durable organizations", "engineering-starter", string(organizationDoc(t, home))).
		WillReturnResult(sqlmock.NewResult(0, 1))

	saved, err := repo.SaveOrganization(context.Background(), home)
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if saved.ID != "org-1" {
		t.Fatalf("unexpected saved home: %+v", saved)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresOrganizationRepository_SaveRejectsMissingIDAndForeignTenant(t *testing.T) {
	repo, mock := newOrganizationRepoMock(t)
	if _, err := repo.SaveOrganization(context.Background(), OrganizationHomePayload{}); !errors.Is(err, ErrOrganizationIDRequired) {
		t.Fatalf("expected ErrOrganizationIDRequired, got %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO organizations")).WillReturnResult(sqlmock.NewResult(0, 0))
	if _, err := repo.SaveOrganization(context.Background(), testOrganizationHome("org-x", "Other")); err == nil {
		t.Fatal("expected error when the upsert touches no row")
	}
}

func TestPostgresOrganizationRepository_Get(t *testing.T) {
	repo, mock := newOrganizationRepoMock(t)
	home := testOrganizationHome("org-1", "Northstar")

	mock.ExpectQuery(regexp.QuoteMeta(organizationGetSQL)).
		WithArgs("org-1", "default").
		WillReturnRows(sqlmock.NewRows([]string{"document"}).AddRow(organizationDoc(t, home)))

	got, ok, err := repo.GetOrganization(context.Background(), "org-1")
	if err != nil || !ok {
		t.Fatalf("get: ok=%v err=%v", ok, err)
	}
	if got.Name != "Northstar" || len(got.Departments) != 1 || got.Departments[0].ID != "platform" {
		t.Fatalf("document did not round-trip: %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresOrganizationRepository_GetNotFound(t *testing.T) {
	repo, mock := newOrganizationRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta(organizationGetSQL)).
		WithArgs("missing", "default").
		WillReturnRows(sqlmock.NewRows([]string{"document"}))

	_, ok, err := repo.GetOrganization(context.Background(), "missing")
	if err != nil || ok {
		t.Fatalf("expected clean not-found, ok=%v err=%v", ok, err)
	}
}

func TestPostgresOrganizationRepository_GetDriverError(t *testing.T) {
	repo, mock := newOrganizationRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta(organizationGetSQL)).WillReturnError(errors.New("connection refused"))

	if _, _, err := repo.GetOrganization(context.Background(), "org-1"); err == nil {
		t.Fatal("expected driver error to surface")
	}
}

func TestPostgresOrganizationRepository_List(t *testing.T) {
	repo, mock := newOrganizationRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta(organizationListSQL)).
		WithArgs("default").
		WillReturnRows(sqlmock.NewRows([]string{"document"}).
			AddRow(organizationDoc(t, testOrganizationHome("org-b", "Beta"))).
			AddRow(organizationDoc(t, testOrganizationHome("org-a", "Alpha"))))

	items, err := repo.ListOrganizations(context.Background())
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 2 || items[0].Name != "Alpha" || items[1].Name != "Beta" {
		t.Fatalf("unexpected list: %+v", items)
	}
}

func TestPostgresOrganizationRepository_UpdateIsTransactional(t *testing.T) {
	repo, mock := newOrganizationRepoMock(t)
	home := testOrganizationHome("org-1", "Northstar")

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(organizationLockSQL)).
		WithArgs("org-1", "default").
		WillReturnRows(sqlmock.NewRows([]string{"document"}).AddRow(organizationDoc(t, home)))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE organizations")).
		WithArgs("org-1", "default", "Northstar", "Ship durable organizations", "engineering-starter", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	updated, ok, err := repo.UpdateOrganization(context.Background(), "org-1", func(h OrganizationHomePayload) OrganizationHomePayload {
		h.AIEngineProfileID = "high_reasoning"
		h.ID = "tampered"
		return h
	})
	if err != nil || !ok {
		t.Fatalf("update: ok=%v err=%v", ok, err)
	}
	if updated.AIEngineProfileID != "high_reasoning" || updated.ID != "org-1" {
		t.Fatalf("unexpected update result: %+v", updated)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresOrganizationRepository_UpdateNotFoundRollsBack(t *testing.T) {
	repo, mock := newOrganizationRepoMock(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(organizationLockSQL)).
		WithArgs("missing", "default").
		WillReturnRows(sqlmock.NewRows([]string{"document"}))
	mock.ExpectRollback()

	called := false
	_, ok, err := repo.UpdateOrganization(context.Background(), "missing", func(h OrganizationHomePayload) OrganizationHomePayload {
		called = true
		return h
	})
	if err != nil || ok || called {
		t.Fatalf("expected not-found without applying update, ok=%v err=%v called=%v", ok, err, called)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresOrganizationRepository_UpdateWriteFailureRollsBack(t *testing.T) {
	repo, mock := newOrganizationRepoMock(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(organizationLockSQL)).
		WillReturnRows(sqlmock.NewRows([]string{"document"}).AddRow(organizationDoc(t, testOrganizationHome("org-1", "Northstar"))))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE organizations")).WillReturnError(errors.New("disk full"))
	mock.ExpectRollback()

	if _, _, err := repo.UpdateOrganization(context.Background(), "org-1", func(h OrganizationHomePayload) OrganizationHomePayload { return h }); err == nil {
		t.Fatal("expected write failure to surface")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
