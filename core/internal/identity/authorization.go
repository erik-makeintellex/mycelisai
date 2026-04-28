package identity

import "strings"

// AuthorizationResolver answers role and permission checks for a loaded UserContext.
type AuthorizationResolver struct {
	permissions map[string]struct{}
	roles       map[string]struct{}
}

func NewAuthorizationResolver(userContext *UserContext) AuthorizationResolver {
	resolver := AuthorizationResolver{
		permissions: map[string]struct{}{},
		roles:       map[string]struct{}{},
	}
	if userContext == nil {
		return resolver
	}
	for _, permission := range userContext.Permissions {
		resolver.addPermission(permission)
	}
	for _, role := range userContext.Roles {
		resolver.addRole(role.Key)
	}
	return resolver
}

func (r AuthorizationResolver) HasPermission(permission string) bool {
	_, ok := r.permissions[strings.TrimSpace(permission)]
	return ok
}

func (r AuthorizationResolver) HasAnyPermission(permissions ...string) bool {
	for _, permission := range permissions {
		if r.HasPermission(permission) {
			return true
		}
	}
	return false
}

func (r AuthorizationResolver) HasAllPermissions(permissions ...string) bool {
	if len(permissions) == 0 {
		return false
	}
	for _, permission := range permissions {
		if !r.HasPermission(permission) {
			return false
		}
	}
	return true
}

func (r AuthorizationResolver) HasRole(role string) bool {
	_, ok := r.roles[strings.TrimSpace(role)]
	return ok
}

func (r AuthorizationResolver) HasAnyRole(roles ...string) bool {
	for _, role := range roles {
		if r.HasRole(role) {
			return true
		}
	}
	return false
}

func (r AuthorizationResolver) HasAllRoles(roles ...string) bool {
	if len(roles) == 0 {
		return false
	}
	for _, role := range roles {
		if !r.HasRole(role) {
			return false
		}
	}
	return true
}

func (c *UserContext) Authorization() AuthorizationResolver {
	return NewAuthorizationResolver(c)
}

func (r AuthorizationResolver) addPermission(permission string) {
	permission = strings.TrimSpace(permission)
	if permission != "" {
		r.permissions[permission] = struct{}{}
	}
}

func (r AuthorizationResolver) addRole(role string) {
	role = strings.TrimSpace(role)
	if role != "" {
		r.roles[role] = struct{}{}
	}
}
