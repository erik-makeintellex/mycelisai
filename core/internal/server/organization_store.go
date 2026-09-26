package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// organizationTenantID matches current Core usage: every organization query
// is scoped to the default tenant.
const organizationTenantID = "default"

var (
	ErrOrganizationNotFound         = errors.New("organization not found")
	ErrOrganizationConflict         = errors.New("organization id already exists")
	ErrOrganizationStoreUnavailable = errors.New("organization storage unavailable")
)

// OrganizationRepository is the persistence seam behind OrganizationStore.
// Production uses the PostgreSQL repository; tests inject an explicit memory
// repository. There is no silent in-memory fallback.
type OrganizationRepository interface {
	List(ctx context.Context) ([]OrganizationSummary, error)
	Insert(ctx context.Context, home OrganizationHomePayload) error
	Get(ctx context.Context, id string) (OrganizationHomePayload, error)
	// Update runs update under the row lock; an update error aborts with no write.
	Update(ctx context.Context, id string, update func(OrganizationHomePayload) (OrganizationHomePayload, error)) (OrganizationHomePayload, error)
	Delete(ctx context.Context, id string) (bool, error)
}

type OrganizationStore struct {
	repo OrganizationRepository
}

// NewOrganizationStore returns the PostgreSQL-backed store. A nil db yields a
// store whose every call fails with ErrOrganizationStoreUnavailable.
func NewOrganizationStore(db *sql.DB) *OrganizationStore {
	if db == nil {
		return &OrganizationStore{}
	}
	return &OrganizationStore{repo: &postgresOrganizationRepository{db: db}}
}

// NewOrganizationStoreWithRepository wires an explicit repository (tests).
func NewOrganizationStoreWithRepository(repo OrganizationRepository) *OrganizationStore {
	return &OrganizationStore{repo: repo}
}

func (s *OrganizationStore) repository() (OrganizationRepository, error) {
	if s == nil || s.repo == nil {
		return nil, ErrOrganizationStoreUnavailable
	}
	return s.repo, nil
}

func (s *OrganizationStore) List(ctx context.Context) ([]OrganizationSummary, error) {
	repo, err := s.repository()
	if err != nil {
		return nil, err
	}
	return repo.List(ctx)
}

// Save creates a new organization. It never upserts: an existing id is
// ErrOrganizationConflict.
func (s *OrganizationStore) Save(ctx context.Context, home OrganizationHomePayload) (OrganizationHomePayload, error) {
	repo, err := s.repository()
	if err != nil {
		return OrganizationHomePayload{}, err
	}
	if err := repo.Insert(ctx, home); err != nil {
		return OrganizationHomePayload{}, err
	}
	return home, nil
}

func (s *OrganizationStore) Get(ctx context.Context, id string) (OrganizationHomePayload, error) {
	repo, err := s.repository()
	if err != nil {
		return OrganizationHomePayload{}, err
	}
	return repo.Get(ctx, id)
}

// Update applies update atomically under a row lock. id and the QA fixture
// scope are preserved regardless of what update returns.
func (s *OrganizationStore) Update(ctx context.Context, id string, update func(OrganizationHomePayload) OrganizationHomePayload) (OrganizationHomePayload, error) {
	return s.UpdateChecked(ctx, id, func(home OrganizationHomePayload) (OrganizationHomePayload, error) {
		return update(home), nil
	})
}

// UpdateChecked is Update whose callback may abort (returned error, no write).
func (s *OrganizationStore) UpdateChecked(ctx context.Context, id string, update func(OrganizationHomePayload) (OrganizationHomePayload, error)) (OrganizationHomePayload, error) {
	repo, err := s.repository()
	if err != nil {
		return OrganizationHomePayload{}, err
	}
	return repo.Update(ctx, id, update)
}

func (s *OrganizationStore) Delete(ctx context.Context, id string) (bool, error) {
	repo, err := s.repository()
	if err != nil {
		return false, err
	}
	return repo.Delete(ctx, id)
}

// QAFixtureScope returns the authoritative fixture scope column for id.
func (s *OrganizationStore) QAFixtureScope(ctx context.Context, id string) (string, bool, error) {
	home, err := s.Get(ctx, id)
	if errors.Is(err, ErrOrganizationNotFound) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return home.QAFixtureScopeID, home.QAFixtureScopeID != "", nil
}

func preserveOrganizationIdentity(current, updated OrganizationHomePayload) OrganizationHomePayload {
	updated.ID = current.ID
	updated.QAFixtureScopeID = current.QAFixtureScopeID
	return updated
}

func wrapOrganizationStorageError(op string, err error) error {
	if err == nil || errors.Is(err, ErrOrganizationNotFound) || errors.Is(err, ErrOrganizationConflict) {
		return err
	}
	return fmt.Errorf("%w: %s: %v", ErrOrganizationStoreUnavailable, op, err)
}

// respondOrganizationStoreError maps store errors to the normalized envelope.
// Storage failures are 503, never 404 or a fake success.
func respondOrganizationStoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrOrganizationNotFound):
		respondAPIError(w, "organization not found", http.StatusNotFound)
	case errors.Is(err, errOrganizationDepartmentNotFound), errors.Is(err, errOrganizationAgentTypeNotFound):
		respondAPIError(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, ErrOrganizationConflict):
		respondAPIError(w, "organization already exists", http.StatusConflict)
	default:
		log.Printf("[organizations] storage error: %v", err)
		respondAPIError(w, "Organization storage unavailable. Check Core database connectivity and retry.", http.StatusServiceUnavailable)
	}
}

// loadOrganizationForRequest reads one organization or writes 404/503.
func (s *AdminServer) loadOrganizationForRequest(w http.ResponseWriter, r *http.Request, id string) (OrganizationHomePayload, bool) {
	home, err := s.organizationStore().Get(r.Context(), id)
	if err != nil {
		respondOrganizationStoreError(w, err)
		return OrganizationHomePayload{}, false
	}
	return home, true
}

// lookupOrganizationBestEffort is for context enrichment (chat, runtime
// summaries) where a missing organization is not an API error. Storage
// failures are logged, not reported as success.
func (s *AdminServer) lookupOrganizationBestEffort(ctx context.Context, id, caller string) (OrganizationHomePayload, bool) {
	home, err := s.organizationStore().Get(ctx, id)
	if err != nil {
		if !errors.Is(err, ErrOrganizationNotFound) {
			log.Printf("[%s] organization %s lookup failed: %v", caller, id, err)
		}
		return OrganizationHomePayload{}, false
	}
	return home, true
}

func (s *AdminServer) templateBundlesPath() string {
	if strings.TrimSpace(s.TemplateBundlesPath) != "" {
		return s.TemplateBundlesPath
	}
	return "config/templates"
}

func (s *AdminServer) organizationStore() *OrganizationStore {
	if s.Organizations != nil {
		return s.Organizations
	}
	db := s.getDB()
	if db == nil {
		// Not cached: storage wired later must take effect. Every call is 503.
		return NewOrganizationStore(nil)
	}
	s.Organizations = NewOrganizationStore(db)
	return s.Organizations
}
