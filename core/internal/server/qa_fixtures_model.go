package server

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	qaFixtureDefaultTTL  = 24 * time.Hour
	qaFixtureMaximumTTL  = 7 * 24 * time.Hour
	qaFixtureScopeHeader = "X-Mycelis-QA-Fixture-Scope"
)

var (
	errQAFixtureScopeMismatch        = errors.New("fixture scope owner or execution does not match")
	errQAFixtureScopeClosed          = errors.New("fixture scope is already purged")
	errQAFixtureResourceUnowned      = errors.New("fixture resource is not owned by this execution")
	errQAFixtureTeamPreexisting      = errors.New("fixture team already has runtime, durable, or workspace state")
	errQAFixtureTeamIdentityMismatch = errors.New("created fixture team identity does not match the approved team")
	qaFixtureIdentityPattern         = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:@/-]{0,127}$`)
	qaFixtureKinds                   = map[string]bool{
		"organization":   true,
		"group":          true,
		"team":           true,
		"run":            true,
		"outcome":        true,
		"artifact":       true,
		"workspace_path": true,
	}
)

type qaFixtureScope struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	OwnerRef     string    `json:"owner_ref"`
	ExecutionRef string    `json:"execution_ref"`
	Status       string    `json:"status"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type qaFixtureResource struct {
	Kind string `json:"kind"`
	Ref  string `json:"ref"`
}

type qaFixtureScopeRequest struct {
	OwnerRef     string `json:"owner_ref"`
	ExecutionRef string `json:"execution_ref"`
	TTLSeconds   int    `json:"ttl_seconds,omitempty"`
}

type qaFixtureResourcesRequest struct {
	OwnerRef     string              `json:"owner_ref"`
	ExecutionRef string              `json:"execution_ref"`
	Resources    []qaFixtureResource `json:"resources"`
}

type qaFixturePurgeRequest struct {
	OwnerRef     string `json:"owner_ref"`
	ExecutionRef string `json:"execution_ref"`
	Confirm      bool   `json:"confirm"`
}

type qaFixturePurgeResult struct {
	Scope                qaFixtureScope   `json:"scope"`
	Confirmed            bool             `json:"confirmed"`
	RegisteredResources  int              `json:"registered_resources"`
	ResourceCounts       map[string]int   `json:"resource_counts"`
	DeletedRows          map[string]int64 `json:"deleted_rows,omitempty"`
	RemovedPaths         []string         `json:"removed_paths,omitempty"`
	StoppedTeams         []string         `json:"stopped_teams,omitempty"`
	RemovedOrganizations []string         `json:"removed_organizations,omitempty"`
	Warnings             []string         `json:"warnings,omitempty"`
	NATSUntouched        bool             `json:"nats_untouched"`
}

func qaFixtureManagementEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("MYCELIS_QA_FIXTURE_MANAGEMENT"))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func normalizeQAFixtureIdentity(label, value string) (string, error) {
	value = strings.TrimSpace(value)
	if !qaFixtureIdentityPattern.MatchString(value) {
		return "", fmt.Errorf("%s must use 1-128 letters, numbers, or ._:@/- characters", label)
	}
	return value, nil
}

func normalizeQAFixtureResource(resource qaFixtureResource) (qaFixtureResource, error) {
	resource.Kind = strings.ToLower(strings.TrimSpace(resource.Kind))
	resource.Ref = strings.TrimSpace(resource.Ref)
	if !qaFixtureKinds[resource.Kind] {
		return qaFixtureResource{}, fmt.Errorf("unsupported fixture resource kind %q", resource.Kind)
	}
	if resource.Ref == "" || len(resource.Ref) > 512 {
		return qaFixtureResource{}, errors.New("fixture resource ref must contain 1-512 characters")
	}
	if resource.Kind == "workspace_path" {
		requested := strings.Trim(strings.ReplaceAll(resource.Ref, "\\", "/"), "/")
		if !strings.HasPrefix(requested, groupWorkspaceRoot+"/") {
			return qaFixtureResource{}, errors.New("fixture workspace path must explicitly start with groups/")
		}
		normalized, err := normalizeRequestedGroupWorkspaceFolder(resource.Ref)
		if err != nil || normalized == groupWorkspaceRoot {
			return qaFixtureResource{}, errors.New("fixture workspace path must target one governed groups/... folder")
		}
		resource.Ref = normalized
	}
	if resource.Kind == "group" || resource.Kind == "run" || resource.Kind == "artifact" {
		if _, err := uuid.Parse(resource.Ref); err != nil {
			return qaFixtureResource{}, fmt.Errorf("fixture %s ref must be a UUID", resource.Kind)
		}
	}
	return resource, nil
}

func qaFixtureExpiry(ttlSeconds int) (time.Time, error) {
	if ttlSeconds < 0 {
		return time.Time{}, errors.New("fixture scope TTL cannot be negative")
	}
	ttl := qaFixtureDefaultTTL
	if ttlSeconds > 0 {
		ttl = time.Duration(ttlSeconds) * time.Second
	}
	if ttl > qaFixtureMaximumTTL {
		return time.Time{}, errors.New("fixture scope TTL cannot exceed 7 days")
	}
	return time.Now().UTC().Add(ttl), nil
}
