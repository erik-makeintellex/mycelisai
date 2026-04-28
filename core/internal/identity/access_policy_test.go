package identity

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetUserContextRequiresActiveAccountAndUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT a.id, a.tenant_id").
		WithArgs(accountID, userID).
		WillReturnError(sql.ErrNoRows)

	if _, err := NewStore(db).GetUserContext(context.Background(), accountID, userID); err == nil {
		t.Fatalf("expected inactive or missing account/user to be rejected")
	}
	assertExpectations(t, mock)
}

func TestGetUserContextFiltersExpiredAndCrossAccountMemberships(t *testing.T) {
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
				now, now, userID, accountID, "erik", "", "Erik", UserStatusActive, "", now, now, nil))
	mock.ExpectQuery("SELECT account_id, user_id, profile").
		WithArgs(accountID, userID).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("expires_at IS NULL OR m.expires_at > NOW").
		WithArgs(accountID, userID).
		WillReturnRows(sqlmock.NewRows(roleGroupColumns()))
	mock.ExpectQuery("SELECT DISTINCT rp.permission_key").
		WithArgs(accountID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"permission_key"}))

	ctx, err := NewStore(db).GetUserContext(context.Background(), accountID, userID)
	if err != nil {
		t.Fatalf("GetUserContext: %v", err)
	}
	if len(ctx.Roles) != 0 || len(ctx.Permissions) != 0 {
		t.Fatalf("expected filtered memberships, got %+v", ctx)
	}
	assertExpectations(t, mock)
}

func TestAssignRoleRejectsRoleFromAnotherAccount(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("WITH valid_role").
		WithArgs(sqlmock.AnyArg(), accountID, userID, "", roleID, UserStatusActive, nil).
		WillReturnError(sql.ErrNoRows)

	if _, err := NewStore(db).AssignRole(context.Background(), OrgMembership{
		AccountID: accountID,
		UserID:    userID,
		RoleID:    roleID,
	}); err == nil {
		t.Fatalf("expected cross-account role assignment to be rejected")
	}
	assertExpectations(t, mock)
}
