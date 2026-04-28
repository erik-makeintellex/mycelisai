package identity

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

const (
	accountID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	userID    = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	roleID    = "cccccccc-cccc-cccc-cccc-cccccccccccc"
	groupID   = "dddddddd-dddd-dddd-dddd-dddddddddddd"
)

func TestCreateAccountDefaultsAndScans(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	mock.ExpectQuery("INSERT INTO accounts").
		WithArgs(sqlmock.AnyArg(), DefaultTenantID, "acme", "Acme", AccountStatusActive, "{}").
		WillReturnRows(sqlmock.NewRows(accountColumns()).
			AddRow(accountID, DefaultTenantID, "acme", "Acme", AccountStatusActive, `{"tier":"dev"}`, now, now))

	account, err := NewStore(db).CreateAccount(context.Background(), Account{Slug: "acme", Name: "Acme"})
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	if account.TenantID != DefaultTenantID || account.Status != AccountStatusActive {
		t.Fatalf("unexpected account defaults: %+v", account)
	}
	if account.Settings["tier"] != "dev" {
		t.Fatalf("expected settings.tier=dev, got %v", account.Settings["tier"])
	}
	assertExpectations(t, mock)
}

func TestCreateUserDefaultsDisplayName(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	mock.ExpectQuery("INSERT INTO users").
		WithArgs(sqlmock.AnyArg(), accountID, "erik", "erik@example.test", "erik", UserStatusActive, "").
		WillReturnRows(sqlmock.NewRows(userColumns()).
			AddRow(userID, accountID, "erik", "erik@example.test", "erik",
				UserStatusActive, "", now, now, nil))

	user, err := NewStore(db).CreateUser(context.Background(), User{
		AccountID: accountID,
		Username:  "erik",
		Email:     "erik@example.test",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if user.DisplayName != "erik" || user.Status != UserStatusActive {
		t.Fatalf("unexpected user defaults: %+v", user)
	}
	assertExpectations(t, mock)
}

func TestCreateSessionStoresMetadata(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	expires := now.Add(time.Hour)
	mock.ExpectQuery("INSERT INTO sessions").
		WithArgs(sqlmock.AnyArg(), accountID, userID, "", "hash-1",
			SessionStatusActive, `{"ip":"127.0.0.1"}`, expires).
		WillReturnRows(sqlmock.NewRows(sessionColumns()).
			AddRow("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee", accountID, userID, "",
				"hash-1", SessionStatusActive, `{"ip":"127.0.0.1"}`,
				now, now, expires, nil, nil, "", ""))

	session, err := NewStore(db).CreateSession(context.Background(), Session{
		AccountID: accountID,
		UserID:    userID,
		TokenHash: "hash-1",
		Metadata:  map[string]any{"ip": "127.0.0.1"},
		ExpiresAt: expires,
	})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if session.Metadata["ip"] != "127.0.0.1" {
		t.Fatalf("expected metadata.ip, got %+v", session.Metadata)
	}
	assertExpectations(t, mock)
}

func TestAssignRoleUpsertsMembership(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	mock.ExpectQuery("INSERT INTO org_memberships").
		WithArgs(sqlmock.AnyArg(), accountID, userID, groupID, roleID, UserStatusActive, nil).
		WillReturnRows(sqlmock.NewRows(membershipColumns()).
			AddRow("ffffffff-ffff-ffff-ffff-ffffffffffff", accountID, userID,
				groupID, roleID, UserStatusActive, now, now, nil))

	membership, err := NewStore(db).AssignRole(context.Background(), OrgMembership{
		AccountID: accountID,
		UserID:    userID,
		GroupID:   groupID,
		RoleID:    roleID,
	})
	if err != nil {
		t.Fatalf("AssignRole: %v", err)
	}
	if membership.Status != UserStatusActive {
		t.Fatalf("expected active membership, got %+v", membership)
	}
	assertExpectations(t, mock)
}

func TestGetUserContextAggregatesProfileRolesGroupsPermissions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	mock.ExpectQuery("SELECT a.id, a.tenant_id").
		WithArgs(accountID, userID).
		WillReturnRows(sqlmock.NewRows(accountUserColumns()).
			AddRow(accountID, DefaultTenantID, "default", "Default", AccountStatusActive, "{}",
				now, now, userID, accountID, "erik", "erik@example.test", "Erik",
				UserStatusActive, "oidc|erik", now, now, nil))
	mock.ExpectQuery("SELECT account_id, user_id, profile").
		WithArgs(accountID, userID).
		WillReturnRows(sqlmock.NewRows(profileColumns()).
			AddRow(accountID, userID, `{"name":"Erik"}`, `{"workspace":"ops"}`, now, now))
	mock.ExpectQuery("SELECT r.id, COALESCE\\(r.account_id::text").
		WithArgs(accountID, userID).
		WillReturnRows(sqlmock.NewRows(roleGroupColumns()).
			AddRow(roleID, accountID, "owner", "Owner", "Owns account", "account", now, now,
				groupID, accountID, "admins", "Admins", "Administrators", now, now))
	mock.ExpectQuery("SELECT DISTINCT rp.permission_key").
		WithArgs(accountID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"permission_key"}).
			AddRow("identity.users.read").AddRow("identity.users.write"))

	ctx, err := NewStore(db).GetUserContext(context.Background(), accountID, userID)
	if err != nil {
		t.Fatalf("GetUserContext: %v", err)
	}
	if ctx.User.Username != "erik" || ctx.Account.Slug != "default" {
		t.Fatalf("unexpected context subject: %+v", ctx)
	}
	if len(ctx.Roles) != 1 || ctx.Roles[0].Key != "owner" {
		t.Fatalf("unexpected roles: %+v", ctx.Roles)
	}
	if len(ctx.Groups) != 1 || ctx.Groups[0].Key != "admins" {
		t.Fatalf("unexpected groups: %+v", ctx.Groups)
	}
	if len(ctx.Permissions) != 2 || ctx.Profile.Context["workspace"] != "ops" {
		t.Fatalf("unexpected permissions/profile: %+v", ctx)
	}
	assertExpectations(t, mock)
}

func TestGetUserContextUsesEmptyProfileWhenMissing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	mock.ExpectQuery("SELECT a.id, a.tenant_id").
		WithArgs(accountID, userID).
		WillReturnRows(sqlmock.NewRows(accountUserColumns()).
			AddRow(accountID, DefaultTenantID, "default", "Default", AccountStatusActive, "{}",
				now, now, userID, accountID, "erik", "", "Erik",
				UserStatusActive, "", now, now, nil))
	mock.ExpectQuery("SELECT account_id, user_id, profile").
		WithArgs(accountID, userID).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT r.id, COALESCE\\(r.account_id::text").
		WithArgs(accountID, userID).
		WillReturnRows(sqlmock.NewRows(roleGroupColumns()))
	mock.ExpectQuery("SELECT DISTINCT rp.permission_key").
		WithArgs(accountID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"permission_key"}))

	ctx, err := NewStore(db).GetUserContext(context.Background(), accountID, userID)
	if err != nil {
		t.Fatalf("GetUserContext: %v", err)
	}
	if ctx.Profile.Profile == nil || len(ctx.Roles) != 0 || len(ctx.Permissions) != 0 {
		t.Fatalf("expected empty profile, roles, permissions; got %+v", ctx)
	}
	assertExpectations(t, mock)
}

func TestRecordAuditEventDefaultsSource(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	mock.ExpectQuery("INSERT INTO identity_audit_events").
		WithArgs(sqlmock.AnyArg(), accountID, userID, "", "identity.user.created",
			"user", userID, "system", "identity.store", `{"status":"active"}`).
		WillReturnRows(sqlmock.NewRows(auditColumns()).
			AddRow("99999999-9999-9999-9999-999999999999", accountID, userID, "",
				"identity.user.created", "user", userID, "system",
				"identity.store", `{"status":"active"}`, now))

	event, err := NewStore(db).RecordAuditEvent(context.Background(), AuditEvent{
		AccountID:   accountID,
		ActorUserID: userID,
		EventType:   "identity.user.created",
		TargetKind:  "user",
		TargetID:    userID,
		Payload:     map[string]any{"status": "active"},
	})
	if err != nil {
		t.Fatalf("RecordAuditEvent: %v", err)
	}
	if event.SourceKind != "system" || event.Payload["status"] != "active" {
		t.Fatalf("unexpected audit event: %+v", event)
	}
	assertExpectations(t, mock)
}

func accountColumns() []string { return []string{"id", "tenant_id", "slug", "name", "status", "settings", "created_at", "updated_at"} }

func userColumns() []string {
	return []string{
		"id", "account_id", "username", "email", "display_name", "status",
		"external_subject", "created_at", "updated_at", "last_login_at",
	}
}

func sessionColumns() []string {
	return []string{
		"id", "account_id", "user_id", "provider_id", "token_hash", "status",
		"metadata", "created_at", "updated_at", "expires_at", "last_seen_at",
		"revoked_at", "revoked_by", "revoke_reason",
	}
}

func membershipColumns() []string {
	return []string{
		"id", "account_id", "user_id", "group_id", "role_id", "status",
		"created_at", "updated_at", "expires_at",
	}
}

func accountUserColumns() []string {
	return append(accountColumns(), userColumns()...)
}

func profileColumns() []string {
	return []string{"account_id", "user_id", "profile", "context", "created_at", "updated_at"}
}

func roleGroupColumns() []string {
	return []string{
		"role_id", "role_account_id", "role_key", "role_name", "role_description",
		"role_scope", "role_created_at", "role_updated_at", "group_id", "group_account_id",
		"group_key", "group_name", "group_description", "group_created_at", "group_updated_at",
	}
}

func auditColumns() []string {
	return []string{"id", "account_id", "actor_user_id", "actor_session_id", "event_type",
		"target_kind", "target_id", "source_kind", "source_channel", "payload", "created_at"}
}

func assertExpectations(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet DB expectations: %v", err)
	}
}
