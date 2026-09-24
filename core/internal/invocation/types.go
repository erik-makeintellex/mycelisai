package invocation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	CapabilityID = "counting.increment"
	Permission   = "counting.increment"
	boundaryKey  = "invocation_boundary"
	maxBudget    = 100
)

var (
	ErrDenied      = errors.New("invocation denied")
	ErrConflict    = errors.New("invocation conflict")
	ErrInvalid     = errors.New("invalid invocation request")
	counterPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)
)

type Proposal struct {
	GroupID string `json:"group_id"`
	Counter string `json:"counter"`
	Budget  int    `json:"budget"`
}

type Input struct {
	Counter string `json:"counter"`
}

type Proposed struct {
	ProofID      string    `json:"proof_id"`
	ContractID   string    `json:"contract_id"`
	ConfirmToken string    `json:"confirm_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type Grant struct {
	ID           string    `json:"id"`
	Digest       string    `json:"digest"`
	ProofID      string    `json:"proof_id"`
	ContractID   string    `json:"contract_id"`
	CapabilityID string    `json:"capability_id"`
	UserID       string    `json:"user_id"`
	AccountID    string    `json:"account_id"`
	GroupID      string    `json:"group_id"`
	Budget       int       `json:"budget"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type Invocation struct {
	ID             string          `json:"id"`
	GrantID        string          `json:"grant_id"`
	GrantDigest    string          `json:"grant_digest"`
	UserID         string          `json:"user_id"`
	AccountID      string          `json:"account_id"`
	GroupID        string          `json:"group_id"`
	CapabilityID   string          `json:"capability_id"`
	Input          Input           `json:"input"`
	IdempotencyKey string          `json:"idempotency_key"`
	State          string          `json:"state"`
	Generation     int64           `json:"generation"`
	Attempt        int             `json:"attempt"`
	Result         json.RawMessage `json:"result,omitempty"`
	Reconciliation json.RawMessage `json:"reconciliation,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

type Evidence struct {
	Source        string `json:"source"`
	Summary       string `json:"summary"`
	InvocationID  string `json:"invocation_id"`
	ObservedCount *int   `json:"observed_count,omitempty"`
}

type binding struct {
	Endpoint string `json:"endpoint"`
	Adapter  string `json:"adapter"`
	Version  string `json:"version"`
	Method   string `json:"method"`
	Schema   string `json:"schema"`
	UnitCost int    `json:"unit_cost"`
}

type authority struct {
	AccountID         string    `json:"account_id"`
	AccountState      string    `json:"account_status"`
	AccountVersion    string    `json:"account_version"`
	UserID            string    `json:"user_id"`
	UserState         string    `json:"user_status"`
	UserVersion       string    `json:"user_version"`
	GroupID           string    `json:"group_id"`
	GroupVersion      string    `json:"group_version"`
	MembershipID      string    `json:"membership_id"`
	MemberState       string    `json:"membership_status"`
	MemberExpiry      time.Time `json:"membership_expires_at"`
	MemberVersion     string    `json:"membership_version"`
	RoleID            string    `json:"role_id"`
	RoleScope         string    `json:"role_scope"`
	RoleVersion       string    `json:"role_version"`
	Permission        string    `json:"permission"`
	PermissionVersion string    `json:"permission_version"`
}

type boundary struct {
	Version         string    `json:"version"`
	UserID          string    `json:"user_id"`
	AccountID       string    `json:"account_id"`
	GroupID         string    `json:"group_id"`
	MembershipID    string    `json:"membership_id"`
	Permission      string    `json:"permission"`
	CapabilityID    string    `json:"capability_id"`
	Authority       authority `json:"authority"`
	AuthorityDigest string    `json:"authority_digest"`
	Binding         binding   `json:"binding"`
	BindingDigest   string    `json:"binding_digest"`
	Input           Input     `json:"input"`
	InputDigest     string    `json:"input_digest"`
	Budget          int       `json:"budget"`
	ExpiresAt       time.Time `json:"expires_at"`
}

func normalizedInput(input Input) (Input, error) {
	input.Counter = strings.TrimSpace(input.Counter)
	if !counterPattern.MatchString(input.Counter) {
		return Input{}, fmt.Errorf("%w: counter must contain 1-64 letters, digits, underscores or hyphens", ErrInvalid)
	}
	return input, nil
}

func validBinding(b binding) error {
	if b.Adapter != "counting-http" || b.Version != "1" || b.Method != "POST" || b.Schema != "counting.v1" || b.UnitCost != 1 {
		return fmt.Errorf("%w: unsupported counting binding", ErrDenied)
	}
	u, err := url.Parse(b.Endpoint)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil || u.Fragment != "" || u.RawQuery != "" || u.Opaque != "" {
		return fmt.Errorf("%w: invalid counting endpoint", ErrDenied)
	}
	return nil
}

func digest(value any) string {
	raw, _ := json.Marshal(value)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func jsonBytes(value any) []byte {
	raw, _ := json.Marshal(value)
	return raw
}

func parsedUUID(value string) (string, error) {
	id, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil {
		return "", fmt.Errorf("%w: UUID required", ErrInvalid)
	}
	return id.String(), nil
}
