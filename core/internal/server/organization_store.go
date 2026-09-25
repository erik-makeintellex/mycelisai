package server

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"sort"
	"strings"
	"sync"
)

// ErrOrganizationIDRequired is returned when a store write has no organization id.
var ErrOrganizationIDRequired = errors.New("organization id is required")

// OrganizationUpdateFunc mutates one organization home inside an atomic store update.
type OrganizationUpdateFunc func(OrganizationHomePayload) OrganizationHomePayload

// OrganizationRepository is the persistence contract for AI Organization homes.
// Implementations: Postgres (migration 040) and the in-memory fallback used
// when Core runs without a database.
type OrganizationRepository interface {
	ListOrganizations(ctx context.Context) ([]OrganizationSummary, error)
	SaveOrganization(ctx context.Context, home OrganizationHomePayload) (OrganizationHomePayload, error)
	GetOrganization(ctx context.Context, id string) (OrganizationHomePayload, bool, error)
	// UpdateOrganization applies update atomically; ok is false when the id is unknown.
	UpdateOrganization(ctx context.Context, id string, update OrganizationUpdateFunc) (OrganizationHomePayload, bool, error)
	Backend() string
}

// OrganizationStore is the handler-facing facade over an OrganizationRepository.
// The *Context methods surface storage errors; the short methods keep the
// original call shape for read-mostly callers and log storage errors.
type OrganizationStore struct {
	repo OrganizationRepository
}

// NewOrganizationStore returns an in-memory store (no durability across restarts).
func NewOrganizationStore() *OrganizationStore {
	return NewOrganizationStoreWithRepository(newMemoryOrganizationRepository())
}

// NewOrganizationStoreForDB returns a Postgres-backed store, or the in-memory
// fallback when db is nil.
func NewOrganizationStoreForDB(db *sql.DB) *OrganizationStore {
	if db == nil {
		log.Println("[organizations] database unavailable; using in-memory organization store (not durable)")
		return NewOrganizationStore()
	}
	return NewOrganizationStoreWithRepository(NewPostgresOrganizationRepository(db, ""))
}

func NewOrganizationStoreWithRepository(repo OrganizationRepository) *OrganizationStore {
	if repo == nil {
		repo = newMemoryOrganizationRepository()
	}
	return &OrganizationStore{repo: repo}
}

func (s *OrganizationStore) Backend() string { return s.repo.Backend() }

func (s *OrganizationStore) ListContext(ctx context.Context) ([]OrganizationSummary, error) {
	return s.repo.ListOrganizations(ctx)
}

func (s *OrganizationStore) SaveContext(ctx context.Context, home OrganizationHomePayload) (OrganizationHomePayload, error) {
	return s.repo.SaveOrganization(ctx, home)
}

func (s *OrganizationStore) GetContext(ctx context.Context, id string) (OrganizationHomePayload, bool, error) {
	return s.repo.GetOrganization(ctx, id)
}

func (s *OrganizationStore) UpdateContext(ctx context.Context, id string, update OrganizationUpdateFunc) (OrganizationHomePayload, bool, error) {
	return s.repo.UpdateOrganization(ctx, id, update)
}

func (s *OrganizationStore) List() []OrganizationSummary {
	items, err := s.ListContext(context.Background())
	if err != nil {
		log.Printf("[organizations] list failed backend=%s error=%v", s.Backend(), err)
		return []OrganizationSummary{}
	}
	return items
}

func (s *OrganizationStore) Save(home OrganizationHomePayload) OrganizationHomePayload {
	saved, err := s.SaveContext(context.Background(), home)
	if err != nil {
		log.Printf("[organizations] save failed backend=%s id=%s error=%v", s.Backend(), home.ID, err)
		return home
	}
	return saved
}

func (s *OrganizationStore) Get(id string) (OrganizationHomePayload, bool) {
	home, ok, err := s.GetContext(context.Background(), id)
	if err != nil {
		log.Printf("[organizations] get failed backend=%s id=%s error=%v", s.Backend(), id, err)
		return OrganizationHomePayload{}, false
	}
	return home, ok
}

func (s *OrganizationStore) Update(id string, update func(OrganizationHomePayload) OrganizationHomePayload) (OrganizationHomePayload, bool) {
	home, ok, err := s.UpdateContext(context.Background(), id, update)
	if err != nil {
		log.Printf("[organizations] update failed backend=%s id=%s error=%v", s.Backend(), id, err)
		return OrganizationHomePayload{}, false
	}
	return home, ok
}

// memoryOrganizationRepository is the no-database fallback.
type memoryOrganizationRepository struct {
	mu    sync.RWMutex
	items map[string]OrganizationHomePayload
}

func newMemoryOrganizationRepository() *memoryOrganizationRepository {
	return &memoryOrganizationRepository{items: make(map[string]OrganizationHomePayload)}
}

func (m *memoryOrganizationRepository) Backend() string { return "memory" }

func (m *memoryOrganizationRepository) ListOrganizations(context.Context) ([]OrganizationSummary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	summaries := make([]OrganizationSummary, 0, len(m.items))
	for _, item := range m.items {
		summaries = append(summaries, item.OrganizationSummary)
	}
	sortOrganizationSummaries(summaries)
	return summaries, nil
}

func (m *memoryOrganizationRepository) SaveOrganization(_ context.Context, home OrganizationHomePayload) (OrganizationHomePayload, error) {
	if strings.TrimSpace(home.ID) == "" {
		return OrganizationHomePayload{}, ErrOrganizationIDRequired
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[home.ID] = home
	return home, nil
}

func (m *memoryOrganizationRepository) GetOrganization(_ context.Context, id string) (OrganizationHomePayload, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	item, ok := m.items[id]
	return item, ok, nil
}

func (m *memoryOrganizationRepository) UpdateOrganization(_ context.Context, id string, update OrganizationUpdateFunc) (OrganizationHomePayload, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	item, ok := m.items[id]
	if !ok {
		return OrganizationHomePayload{}, false, nil
	}
	item = update(item)
	item.ID = id
	m.items[id] = item
	return item, true, nil
}

func sortOrganizationSummaries(summaries []OrganizationSummary) {
	sort.SliceStable(summaries, func(i, j int) bool {
		if summaries[i].Name != summaries[j].Name {
			return summaries[i].Name < summaries[j].Name
		}
		return summaries[i].ID < summaries[j].ID
	})
}

func (s *AdminServer) templateBundlesPath() string {
	if strings.TrimSpace(s.TemplateBundlesPath) != "" {
		return s.TemplateBundlesPath
	}
	return "config/templates"
}

func (s *AdminServer) organizationStore() *OrganizationStore {
	if s.Organizations == nil {
		s.Organizations = NewOrganizationStore()
	}
	return s.Organizations
}
