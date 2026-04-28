package identity

import "testing"

func TestAuthorizationResolverChecksPermissionsAndRoles(t *testing.T) {
	ctx := &UserContext{
		Roles: []Role{
			{Key: "account.owner"},
			{Key: "workspace.operator"},
		},
		Permissions: []string{
			"identity.users.read",
			"identity.users.write",
		},
	}

	authz := NewAuthorizationResolver(ctx)
	if !authz.HasPermission("identity.users.read") {
		t.Fatal("expected identity.users.read permission")
	}
	if authz.HasPermission("identity.users.delete") {
		t.Fatal("did not expect identity.users.delete permission")
	}
	if !authz.HasRole("account.owner") {
		t.Fatal("expected account.owner role")
	}
	if authz.HasRole("system.admin") {
		t.Fatal("did not expect system.admin role")
	}
}

func TestAuthorizationResolverChecksAnyAndAll(t *testing.T) {
	authz := NewAuthorizationResolver(&UserContext{
		Roles: []Role{
			{Key: "account.owner"},
			{Key: "workspace.operator"},
		},
		Permissions: []string{
			"identity.users.read",
			"identity.users.write",
		},
	})

	if !authz.HasAnyPermission("missing.permission", "identity.users.write") {
		t.Fatal("expected any permission match")
	}
	if !authz.HasAllPermissions("identity.users.read", "identity.users.write") {
		t.Fatal("expected all permissions to match")
	}
	if authz.HasAllPermissions("identity.users.read", "missing.permission") {
		t.Fatal("did not expect all permissions to match")
	}
	if !authz.HasAnyRole("missing.role", "workspace.operator") {
		t.Fatal("expected any role match")
	}
	if !authz.HasAllRoles("account.owner", "workspace.operator") {
		t.Fatal("expected all roles to match")
	}
	if authz.HasAllRoles("account.owner", "missing.role") {
		t.Fatal("did not expect all roles to match")
	}
}

func TestAuthorizationResolverNormalizesBlankAndWhitespaceKeys(t *testing.T) {
	authz := NewAuthorizationResolver(&UserContext{
		Roles: []Role{
			{Key: " account.owner "},
			{Key: ""},
		},
		Permissions: []string{
			" identity.users.read ",
			"",
			"identity.users.read",
		},
	})

	if !authz.HasPermission("identity.users.read") {
		t.Fatal("expected trimmed permission lookup to match")
	}
	if authz.HasPermission("") {
		t.Fatal("did not expect blank permission lookup to match")
	}
	if !authz.HasRole("account.owner") {
		t.Fatal("expected trimmed role lookup to match")
	}
	if authz.HasRole(" ") {
		t.Fatal("did not expect blank role lookup to match")
	}
}

func TestAuthorizationResolverHandlesNilContextAndEmptyRequirements(t *testing.T) {
	authz := NewAuthorizationResolver(nil)
	if authz.HasPermission("identity.users.read") {
		t.Fatal("did not expect permission on nil context")
	}
	if authz.HasRole("account.owner") {
		t.Fatal("did not expect role on nil context")
	}
	if authz.HasAnyPermission() {
		t.Fatal("did not expect any permission match with no requirements")
	}
	if authz.HasAllPermissions() {
		t.Fatal("did not expect all permission match with no requirements")
	}
	if authz.HasAnyRole() {
		t.Fatal("did not expect any role match with no requirements")
	}
	if authz.HasAllRoles() {
		t.Fatal("did not expect all role match with no requirements")
	}
}

func TestUserContextAuthorizationBuildsResolver(t *testing.T) {
	ctx := &UserContext{
		Roles:       []Role{{Key: "account.owner"}},
		Permissions: []string{"identity.users.read"},
	}

	if !ctx.Authorization().HasRole("account.owner") {
		t.Fatal("expected resolver from context to include role")
	}
	if !ctx.Authorization().HasPermission("identity.users.read") {
		t.Fatal("expected resolver from context to include permission")
	}
}
